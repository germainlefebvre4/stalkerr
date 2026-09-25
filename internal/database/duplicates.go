package database

import (
	"time"

	"github.com/glefebvre/stalkeer/internal/models"
	"gorm.io/gorm"
)

// DuplicateOccurrence is one ProcessedLine/DownloadInfo pair belonging to a
// DuplicateGroup.
type DuplicateOccurrence struct {
	ProcessedLineID uint
	DownloadInfoID  uint
	DownloadPath    string
	CompletedAt     *time.Time
}

// DuplicateGroup is a movie or TV episode with two or more automated
// completed occurrences at different download paths (see the
// duplicate-download-detection spec's Duplicate Group Definition). Exactly
// one of MovieID/TVShowID is set, matching which identity key the group was
// built from.
type DuplicateGroup struct {
	ContentType models.ContentType
	MovieID     *uint
	TVShowID    *uint
	Occurrences []DuplicateOccurrence
}

// duplicateCandidateRow is one completed, non-forced ProcessedLine/DownloadInfo
// pair, keyed by whichever identity column (movie_id or tv_show_id) the query
// selected on.
type duplicateCandidateRow struct {
	ProcessedLineID uint
	IdentityID      uint
	DownloadInfoID  uint
	DownloadPath    string
	CompletedAt     *time.Time
}

// findDuplicateCandidateRows fetches every ProcessedLine (identified by
// identityColumn, "movie_id" or "tv_show_id") whose state is "downloaded" and
// whose associated DownloadInfo is "completed" and not forced-origin — the
// pool of occurrences the Duplicate Group Definition compares. identityColumn
// and downloadPath are both already indexed (idx on ProcessedLine.MovieID/
// TVShowID; download_info is joined on its primary key), so this is a
// straightforward indexed lookup, not a full scan.
func findDuplicateCandidateRows(db *gorm.DB, identityColumn string) ([]duplicateCandidateRow, error) {
	var rows []duplicateCandidateRow
	err := db.Table("processed_lines AS pl").
		Select("pl.id AS processed_line_id, pl."+identityColumn+" AS identity_id, di.id AS download_info_id, di.download_path AS download_path, di.completed_at AS completed_at").
		Joins("JOIN download_info di ON di.id = pl.download_info_id").
		Where("pl."+identityColumn+" IS NOT NULL").
		Where("pl.state = ?", string(models.StateDownloaded)).
		Where("di.status = ?", string(models.DownloadStatusCompleted)).
		Where("di.is_forced = ?", false).
		Where("di.download_path IS NOT NULL AND di.download_path != ''").
		Scan(&rows).Error
	return rows, err
}

// groupDuplicateCandidates groups rows by IdentityID and keeps only the
// groups that qualify as a duplicate: 2+ occurrences with at least 2 distinct
// DownloadPath values (a same-path group is a tier-2 upgrade, not a
// duplicate).
func groupDuplicateCandidates(rows []duplicateCandidateRow, contentType models.ContentType) []DuplicateGroup {
	grouped := make(map[uint][]duplicateCandidateRow)
	for _, r := range rows {
		grouped[r.IdentityID] = append(grouped[r.IdentityID], r)
	}

	var groups []DuplicateGroup
	for identityID, occs := range grouped {
		if len(occs) < 2 {
			continue
		}
		paths := make(map[string]struct{}, len(occs))
		for _, o := range occs {
			paths[o.DownloadPath] = struct{}{}
		}
		if len(paths) < 2 {
			continue
		}

		id := identityID
		group := DuplicateGroup{ContentType: contentType}
		if contentType == models.ContentTypeMovies {
			group.MovieID = &id
		} else {
			group.TVShowID = &id
		}
		for _, o := range occs {
			group.Occurrences = append(group.Occurrences, DuplicateOccurrence{
				ProcessedLineID: o.ProcessedLineID,
				DownloadInfoID:  o.DownloadInfoID,
				DownloadPath:    o.DownloadPath,
				CompletedAt:     o.CompletedAt,
			})
		}
		groups = append(groups, group)
	}
	return groups
}

// FindDuplicateDownloadGroups implements the Duplicate Group Definition
// (capability duplicate-download-detection): it groups completed, non-forced
// ProcessedLine occurrences by movie and, separately, by TV episode (each
// TVShow row already represents one specific episode), and returns every
// group with 2+ occurrences whose stored download_path values differ. It is a
// pure local-database query — no Radarr/Sonarr call is made — and performs no
// mutation.
func FindDuplicateDownloadGroups(db *gorm.DB) ([]DuplicateGroup, error) {
	movieRows, err := findDuplicateCandidateRows(db, "movie_id")
	if err != nil {
		return nil, err
	}
	tvRows, err := findDuplicateCandidateRows(db, "tv_show_id")
	if err != nil {
		return nil, err
	}

	groups := groupDuplicateCandidates(movieRows, models.ContentTypeMovies)
	groups = append(groups, groupDuplicateCandidates(tvRows, models.ContentTypeTVShows)...)
	return groups, nil
}
