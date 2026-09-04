package scheduler

import "github.com/glefebvre/stalkeer/internal/models"

// Tier classifies a Stream as new/missing content (Tier1) or already-downloaded
// content eligible for a forced re-download/upgrade attempt (Tier2).
type Tier int

const (
	// Tier1 is content that has never been successfully downloaded.
	Tier1 Tier = 1
	// Tier2 is content that already has a successful download but still has an
	// eligible (not-yet-downloaded) candidate, e.g. for a quality upgrade.
	Tier2 Tier = 2
)

// Item is a single downloadable unit within a Stream: one movie, or one TV episode.
type Item struct {
	// DisplayName is a human-readable label used for logging/dry-run output.
	DisplayName string
	// BaseDestDir is the untagged movie/season folder root the download should be
	// written under (without extension). The final per-attempt destination path is
	// computed lazily in cmd/download.go's downloadItem, once the candidate actually
	// being attempted is known, by appending that candidate's resolution/language/
	// French-variant quality tags to this root.
	BaseDestDir string
	// Candidates are the quality-ordered ProcessedLine rows eligible for download,
	// as returned by matcher.FindMovieDownloadCandidates/FindTVShowDownloadCandidates.
	Candidates []models.ProcessedLine
	// ResumeInfo is set when this item has an existing, incomplete DownloadInfo
	// record that should be resumed rather than started fresh.
	ResumeInfo *models.DownloadInfo
	// Episode is the episode number, used to order items within a season stream.
	// Always 0 for movie items.
	Episode int
}

// Stream is the atomic unit of scheduling: one movie, or one TV series' season.
// A Stream is claimed by exactly one worker at a time and drained fully (all of
// its Items, in order) before the worker returns to the pool.
type Stream struct {
	Tier Tier
	// SourceKey identifies the owning work unit (a movie or a TV series) across
	// all of its Streams (e.g. every season of the same series shares SourceKey).
	SourceKey string
	// SeriesID is the originating Sonarr series ID for a TV series-season stream.
	// It is 0 for movie streams and for Tier2 (upgrade) streams, neither of which
	// participate in the ascending-season-order queue.
	SeriesID int
	// Season is the season number for a series-season stream; 0 for movies.
	Season int
	Items  []Item
}
