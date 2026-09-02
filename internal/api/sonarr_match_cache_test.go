package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/glefebvre/stalkeer/internal/external/sonarr"
	"github.com/glefebvre/stalkeer/internal/models"
)

func TestSonarrMatchStatusCache_MissThenPopulate(t *testing.T) {
	db := setupTestDB(t)

	tvdbID := 5001
	season, episode := 1, 1
	if err := db.Create(&models.TVShow{TMDBID: 9999, TVDBID: &tvdbID, TMDBTitle: "Some Show", Season: &season, Episode: &episode}).Error; err != nil {
		t.Fatalf("failed to seed local episode: %v", err)
	}

	var episodeCalls int64
	sonarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&episodeCalls, 1)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]sonarr.Episode{
			{SeasonNumber: 1, EpisodeNumber: 1, Monitored: true},
		})
	}))
	defer sonarrServer.Close()

	client := sonarr.New(sonarr.Config{BaseURL: sonarrServer.URL, APIKey: "test"})
	series := sonarr.Series{ID: 1, Title: "Some Show", TvdbID: tvdbID, Monitored: true}

	cache := newSonarrMatchStatusCache()

	matched, err := cache.matchedStatus(context.Background(), client, db, series)
	if err != nil {
		t.Fatalf("matchedStatus returned error: %v", err)
	}
	if !matched {
		t.Errorf("expected series to be matched, got unmatched")
	}
	if calls := atomic.LoadInt64(&episodeCalls); calls != 1 {
		t.Fatalf("expected 1 episode fetch on cache miss, got %d", calls)
	}

	matched, err = cache.matchedStatus(context.Background(), client, db, series)
	if err != nil {
		t.Fatalf("matchedStatus returned error on cache hit: %v", err)
	}
	if !matched {
		t.Errorf("expected cached series to still report matched")
	}
	if calls := atomic.LoadInt64(&episodeCalls); calls != 1 {
		t.Errorf("expected no additional episode fetch on cache hit, got %d total calls", calls)
	}
}

func TestSonarrMatchStatusCache_ClearForcesRecompute(t *testing.T) {
	db := setupTestDB(t)

	tvdbID := 5002
	var episodeCalls int64
	sonarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&episodeCalls, 1)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]sonarr.Episode{})
	}))
	defer sonarrServer.Close()

	client := sonarr.New(sonarr.Config{BaseURL: sonarrServer.URL, APIKey: "test"})
	series := sonarr.Series{ID: 2, Title: "No Local Match", TvdbID: tvdbID, Monitored: true}

	cache := newSonarrMatchStatusCache()

	if _, err := cache.matchedStatus(context.Background(), client, db, series); err != nil {
		t.Fatalf("matchedStatus returned error: %v", err)
	}
	if calls := atomic.LoadInt64(&episodeCalls); calls != 1 {
		t.Fatalf("expected 1 episode fetch, got %d", calls)
	}

	cache.clear()

	if _, err := cache.matchedStatus(context.Background(), client, db, series); err != nil {
		t.Fatalf("matchedStatus returned error after clear: %v", err)
	}
	if calls := atomic.LoadInt64(&episodeCalls); calls != 2 {
		t.Errorf("expected a recompute (2 total episode fetches) after clear, got %d", calls)
	}
}
