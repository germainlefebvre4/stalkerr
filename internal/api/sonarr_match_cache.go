package api

import (
	"context"
	"sync"

	"github.com/glefebvre/stalkeer/internal/external/sonarr"
	"github.com/glefebvre/stalkeer/internal/matcher"
	"gorm.io/gorm"
)

// sonarrMatchStatusCache is an in-memory, per-process cache of each Sonarr-
// monitored series' matched/unmatched status (at least one monitored episode
// has a playlist match). It exists solely to back the match-status filter on
// the Sonarr listing endpoint without a full per-series episode fan-out to
// Sonarr on every filtered request. It is populated lazily on a cache miss and
// invalidated only by the Séries section's manual refresh action - never by a
// background timer. See openspec/changes/radarr-sonarr-match-filters-and-counts.
type sonarrMatchStatusCache struct {
	mu       sync.RWMutex
	statuses map[int]bool // sonarr series ID -> matched
}

// sonarrMatchCache is the single process-wide instance used by the Sonarr
// listing endpoint.
var sonarrMatchCache = newSonarrMatchStatusCache()

func newSonarrMatchStatusCache() *sonarrMatchStatusCache {
	return &sonarrMatchStatusCache{statuses: make(map[int]bool)}
}

// clear invalidates every cached entry, so the next filtered request
// recomputes matched status from scratch.
func (c *sonarrMatchStatusCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.statuses = make(map[int]bool)
}

// matchedStatus returns whether the given series has at least one matched
// monitored episode, using the cached value when present. On a miss, it
// fetches the series' episodes from Sonarr, computes the status, and stores it
// before returning.
func (c *sonarrMatchStatusCache) matchedStatus(ctx context.Context, client *sonarr.Client, db *gorm.DB, series sonarr.Series) (bool, error) {
	c.mu.RLock()
	matched, ok := c.statuses[series.ID]
	c.mu.RUnlock()
	if ok {
		return matched, nil
	}

	episodes, err := client.GetEpisodesBySeriesID(ctx, series.ID)
	if err != nil {
		return false, err
	}

	var monitoredEpisodes []matcher.SeasonEpisode
	for _, ep := range episodes {
		if ep.Monitored {
			monitoredEpisodes = append(monitoredEpisodes, matcher.SeasonEpisode{
				Season:  ep.SeasonNumber,
				Episode: ep.EpisodeNumber,
			})
		}
	}

	matchedCount, _, err := matcher.MatchSeriesEpisodesAggregate(db, series.TvdbID, monitoredEpisodes)
	if err != nil {
		return false, err
	}

	matched = matchedCount > 0
	c.mu.Lock()
	c.statuses[series.ID] = matched
	c.mu.Unlock()

	return matched, nil
}
