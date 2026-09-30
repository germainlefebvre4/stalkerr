package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/external/radarr"
	"github.com/glefebvre/stalkeer/internal/external/sonarr"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/glefebvre/stalkeer/internal/retry"
	"github.com/glefebvre/stalkeer/internal/settings"
)

const (
	// Per-occurrence outcomes reported by the keep-duplicate endpoint.
	duplicateOutcomeKept                 = "kept"
	duplicateOutcomeDeleted              = "deleted"
	duplicateOutcomeCancelledFileMissing = "cancelled_file_missing"

	// Radarr/Sonarr "Refresh & Scan" outcomes.
	refreshRequested         = "requested"
	refreshFailed            = "failed"
	refreshNotConfigured     = "not_configured"
	refreshNotFoundInLibrary = "not_found_in_library"

	duplicateKeepErrorMessage = "Cancelled by duplicate-download keep-one action: duplicate file removed"
)

// DuplicateCandidate describes one distinct file of a duplicate group.
type DuplicateCandidate struct {
	DownloadPath       string     `json:"download_path"`
	Exists             bool       `json:"exists"`
	SizeBytes          *int64     `json:"size_bytes,omitempty"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
	DownloadInfoIDs    []uint     `json:"download_info_ids"`
	MatchesLibraryName *bool      `json:"matches_library_name,omitempty"`
	Suggested          bool       `json:"suggested"`
}

// DuplicateCandidatesResponse is returned by the duplicate-candidates endpoint.
type DuplicateCandidatesResponse struct {
	Candidates []DuplicateCandidate `json:"candidates"`
}

// KeepDuplicateResponse is returned by the keep-duplicate endpoint.
type KeepDuplicateResponse struct {
	Occurrences []DuplicateOccurrenceOutcome `json:"occurrences"`
	Refresh     string                       `json:"refresh"`
}

// duplicatePathGroup is one distinct (cleaned) download_path of a group with
// the occurrences that share it.
type duplicatePathGroup struct {
	Path        string
	Occurrences []database.DuplicateOccurrence
	CompletedAt *time.Time
}

// groupOccurrencesByPath groups occurrences by cleaned download_path,
// preserving first-seen order.
func groupOccurrencesByPath(occs []database.DuplicateOccurrence) []duplicatePathGroup {
	index := map[string]int{}
	var out []duplicatePathGroup
	for _, occ := range occs {
		p := filepath.Clean(occ.DownloadPath)
		i, ok := index[p]
		if !ok {
			i = len(out)
			index[p] = i
			out = append(out, duplicatePathGroup{Path: p})
		}
		out[i].Occurrences = append(out[i].Occurrences, occ)
		if occ.CompletedAt != nil && (out[i].CompletedAt == nil || occ.CompletedAt.After(*out[i].CompletedAt)) {
			t := *occ.CompletedAt
			out[i].CompletedAt = &t
		}
	}
	return out
}

// libraryEntry is the live Radarr/Sonarr entry of a group's movie/series.
type libraryEntry struct {
	ID    int
	Title string
	Year  int
	Movie bool
}

// libraryLookupState classifies the result of lookupLibraryEntry.
type libraryLookupState int

const (
	libraryFound libraryLookupState = iota
	libraryNotConfigured
	libraryNotFound
	libraryLookupFailed
)

// lookupLibraryEntry resolves the group's movie/series in Radarr/Sonarr the
// way resyncDownloadPath does (single attempt, server circuit breakers).
func (s *Server) lookupLibraryEntry(ctx context.Context, group *database.DuplicateGroup) (*libraryEntry, libraryLookupState) {
	db := database.Get()
	cfg := settings.Effective()

	switch {
	case group.MovieID != nil:
		if cfg.Radarr.URL == "" || cfg.Radarr.APIKey == "" {
			return nil, libraryNotConfigured
		}
		var movie models.Movie
		if err := db.First(&movie, *group.MovieID).Error; err != nil {
			return nil, libraryLookupFailed
		}
		client := radarr.New(radarr.Config{
			BaseURL:     cfg.Radarr.URL,
			APIKey:      cfg.Radarr.APIKey,
			Timeout:     existenceCheckTimeout,
			RetryConfig: retry.Config{MaxAttempts: 1},
			Breaker:     s.radarrBreaker,
		})
		m, err := client.GetMovieByTMDBID(ctx, movie.TMDBID)
		if err != nil {
			return nil, libraryLookupFailed
		}
		if m == nil {
			return nil, libraryNotFound
		}
		return &libraryEntry{ID: m.ID, Title: m.Title, Year: m.Year, Movie: true}, libraryFound

	case group.TVShowID != nil:
		if cfg.Sonarr.URL == "" || cfg.Sonarr.APIKey == "" {
			return nil, libraryNotConfigured
		}
		var show models.TVShow
		if err := db.First(&show, *group.TVShowID).Error; err != nil || show.TVDBID == nil {
			return nil, libraryLookupFailed
		}
		client := sonarr.New(sonarr.Config{
			BaseURL:     cfg.Sonarr.URL,
			APIKey:      cfg.Sonarr.APIKey,
			Timeout:     existenceCheckTimeout,
			RetryConfig: retry.Config{MaxAttempts: 1},
			Breaker:     s.sonarrBreaker,
		})
		sr, err := client.GetSeriesByTVDBID(ctx, *show.TVDBID)
		if err != nil {
			return nil, libraryLookupFailed
		}
		if sr == nil {
			return nil, libraryNotFound
		}
		return &libraryEntry{ID: sr.ID, Title: sr.Title, Year: sr.Year}, libraryFound
	}
	return nil, libraryLookupFailed
}

// normalizeName case-folds s and collapses every non letter/digit run into a
// single space.
func normalizeName(s string) string {
	var b strings.Builder
	space := false
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			space = false
		} else if !space && b.Len() > 0 {
			b.WriteByte(' ')
			space = true
		}
	}
	return strings.TrimSpace(b.String())
}

// matchesLibraryName applies design D2: a movie's base name must start with
// "Title (Year)"; an episode's base name must contain the series title.
func matchesLibraryName(path string, entry *libraryEntry) bool {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	name := normalizeName(base)
	title := normalizeName(entry.Title)
	if title == "" {
		return false
	}
	if entry.Movie {
		return strings.HasPrefix(name, normalizeName(fmt.Sprintf("%s (%d)", entry.Title, entry.Year)))
	}
	return strings.Contains(name, title)
}

// duplicateGroupForRequest parses :id and resolves its duplicate group,
// answering the error itself and returning nil when it cannot.
func duplicateGroupForRequest(c *gin.Context) (*database.DuplicateGroup, uint) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", "invalid download id")
		return nil, 0
	}
	groups, err := database.FindDuplicateDownloadGroups(database.Get())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "database_error", "failed to evaluate duplicate download groups")
		return nil, 0
	}
	group := findGroupContaining(groups, uint(id))
	if group == nil {
		respondError(c, http.StatusNotFound, "not_a_duplicate", fmt.Sprintf("download %d does not belong to any duplicate group", id))
		return nil, 0
	}
	return group, uint(id)
}

// getDuplicateCandidates describes the distinct files of the duplicate group
// containing :id. Read-only.
func (s *Server) getDuplicateCandidates(c *gin.Context) {
	group, _ := duplicateGroupForRequest(c)
	if group == nil {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), existenceCheckTimeout)
	defer cancel()
	entry, state := s.lookupLibraryEntry(ctx, group)

	paths := groupOccurrencesByPath(group.Occurrences)
	candidates := make([]DuplicateCandidate, 0, len(paths))
	for _, pg := range paths {
		cand := DuplicateCandidate{DownloadPath: pg.Path, CompletedAt: pg.CompletedAt, DownloadInfoIDs: []uint{}}
		if info, err := os.Stat(pg.Path); err == nil {
			cand.Exists = true
			size := info.Size()
			cand.SizeBytes = &size
		}
		for _, occ := range pg.Occurrences {
			cand.DownloadInfoIDs = append(cand.DownloadInfoIDs, occ.DownloadInfoID)
		}
		if state == libraryFound {
			m := matchesLibraryName(pg.Path, entry)
			cand.MatchesLibraryName = &m
		}
		candidates = append(candidates, cand)
	}

	markSuggestedCandidate(candidates)
	c.JSON(http.StatusOK, DuplicateCandidatesResponse{Candidates: candidates})
}

// markSuggestedCandidate flags at most one existing candidate: the canonical
// name match, otherwise the latest completion, otherwise none.
func markSuggestedCandidate(cands []DuplicateCandidate) {
	best := -1
	for i, cand := range cands {
		if !cand.Exists {
			continue
		}
		if best == -1 {
			best = i
			continue
		}
		b := cands[best]
		bMatch := b.MatchesLibraryName != nil && *b.MatchesLibraryName
		iMatch := cand.MatchesLibraryName != nil && *cand.MatchesLibraryName
		switch {
		case iMatch && !bMatch:
			best = i
		case iMatch == bMatch && laterThan(cand.CompletedAt, b.CompletedAt):
			best = i
		}
	}
	if best >= 0 {
		cands[best].Suggested = true
	}
}

func laterThan(a, b *time.Time) bool {
	if a == nil {
		return false
	}
	return b == nil || a.After(*b)
}

// keepDuplicateDownload applies the Keep One Duplicate Action (design D3) to
// the group containing :id, whose file is the one kept.
func (s *Server) keepDuplicateDownload(c *gin.Context) {
	group, id := duplicateGroupForRequest(c)
	if group == nil {
		return
	}
	db := database.Get()

	paths := groupOccurrencesByPath(group.Occurrences)
	var keptPath string
	for _, pg := range paths {
		for _, occ := range pg.Occurrences {
			if occ.DownloadInfoID == id {
				keptPath = pg.Path
			}
		}
	}

	// Step 2: the kept file must exist.
	if _, err := os.Stat(keptPath); err != nil {
		respondError(c, http.StatusConflict, "kept_file_missing", "the file to keep does not exist on disk; nothing was changed")
		return
	}

	// Step 3: preflight every other path before touching anything.
	toDelete := map[string]bool{}
	for _, pg := range paths {
		if pg.Path == keptPath {
			continue
		}
		info, err := os.Lstat(pg.Path)
		switch {
		case err == nil && info.Mode().IsRegular():
			toDelete[pg.Path] = true
		case err == nil:
			respondError(c, http.StatusConflict, "not_a_regular_file", fmt.Sprintf("%s is not a regular file; nothing was changed", pg.Path))
			return
		case errors.Is(err, os.ErrNotExist):
			// already missing
		default:
			respondError(c, http.StatusConflict, "file_check_failed", fmt.Sprintf("cannot inspect %s; nothing was changed", pg.Path))
			return
		}
	}

	// Step 4: per path, delete then cancel.
	outcomes := map[uint]string{}
	changed := false
	var failure error
	for _, pg := range paths {
		if pg.Path == keptPath {
			for _, occ := range pg.Occurrences {
				outcomes[occ.DownloadInfoID] = duplicateOutcomeKept
			}
			continue
		}
		outcome := duplicateOutcomeCancelledFileMissing
		if toDelete[pg.Path] {
			if err := os.Remove(pg.Path); err == nil {
				outcome = duplicateOutcomeDeleted
				changed = true
			} else if !errors.Is(err, os.ErrNotExist) {
				failure = fmt.Errorf("failed to delete %s: %w", pg.Path, err)
				break
			}
		}
		for _, occ := range pg.Occurrences {
			if err := cancelDuplicateOccurrenceWithMessage(db, occ.DownloadInfoID, duplicateKeepErrorMessage); err != nil {
				failure = fmt.Errorf("failed to cancel occurrence %d: %w", occ.DownloadInfoID, err)
				break
			}
			outcomes[occ.DownloadInfoID] = outcome
			changed = true
		}
		if failure != nil {
			break
		}
	}

	// Step 5: refresh once, even after a partial failure that changed something.
	refresh := refreshNotConfigured
	if changed {
		refresh = s.refreshLibrary(c.Request.Context(), group)
	}

	if failure != nil {
		respondError(c, http.StatusInternalServerError, "keep_duplicate_failed", failure.Error())
		return
	}

	resp := KeepDuplicateResponse{Occurrences: make([]DuplicateOccurrenceOutcome, 0, len(group.Occurrences)), Refresh: refresh}
	for _, occ := range group.Occurrences {
		resp.Occurrences = append(resp.Occurrences, DuplicateOccurrenceOutcome{
			DownloadInfoID: occ.DownloadInfoID,
			DownloadPath:   occ.DownloadPath,
			Outcome:        outcomes[occ.DownloadInfoID],
		})
	}
	sort.SliceStable(resp.Occurrences, func(i, j int) bool { return resp.Occurrences[i].DownloadInfoID < resp.Occurrences[j].DownloadInfoID })
	c.JSON(http.StatusOK, resp)
}

// refreshLibrary requests Radarr/Sonarr "Refresh & Scan" once for the group's
// movie/series and maps the result to a refresh outcome.
func (s *Server) refreshLibrary(parent context.Context, group *database.DuplicateGroup) string {
	ctx, cancel := context.WithTimeout(parent, existenceCheckTimeout)
	defer cancel()

	entry, state := s.lookupLibraryEntry(ctx, group)
	switch state {
	case libraryNotConfigured:
		return refreshNotConfigured
	case libraryNotFound:
		return refreshNotFoundInLibrary
	case libraryLookupFailed:
		return refreshFailed
	}

	cfg := settings.Effective()
	var err error
	if entry.Movie {
		err = radarr.New(radarr.Config{
			BaseURL:     cfg.Radarr.URL,
			APIKey:      cfg.Radarr.APIKey,
			Timeout:     existenceCheckTimeout,
			RetryConfig: retry.Config{MaxAttempts: 1},
			Breaker:     s.radarrBreaker,
		}).RefreshMovie(ctx, entry.ID)
	} else {
		err = sonarr.New(sonarr.Config{
			BaseURL:     cfg.Sonarr.URL,
			APIKey:      cfg.Sonarr.APIKey,
			Timeout:     existenceCheckTimeout,
			RetryConfig: retry.Config{MaxAttempts: 1},
			Breaker:     s.sonarrBreaker,
		}).RefreshSeries(ctx, entry.ID)
	}
	if err != nil {
		return refreshFailed
	}
	return refreshRequested
}
