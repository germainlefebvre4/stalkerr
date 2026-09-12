package sonarr

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/circuitbreaker"
	"github.com/glefebvre/stalkeer/internal/retry"
)

func TestNew(t *testing.T) {
	cfg := Config{
		BaseURL: "http://localhost:8989",
		APIKey:  "test-key",
	}

	client := New(cfg)

	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.http.BaseURL != cfg.BaseURL {
		t.Errorf("expected baseURL %s, got %s", cfg.BaseURL, client.http.BaseURL)
	}
	if client.http.APIKey != cfg.APIKey {
		t.Errorf("expected apiKey %s, got %s", cfg.APIKey, client.http.APIKey)
	}
}

func TestGetMissingSeries(t *testing.T) {
	series := []Series{
		{ID: 1, Title: "Test Series 1", TvdbID: 101, Monitored: true, TotalEpisodeCount: 10, EpisodeFileCount: 5},
		{ID: 2, Title: "Test Series 2", TvdbID: 102, Monitored: true, TotalEpisodeCount: 10, EpisodeFileCount: 10},
		{ID: 3, Title: "Test Series 3", TvdbID: 103, Monitored: false, TotalEpisodeCount: 10, EpisodeFileCount: 5},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/series" {
			t.Errorf("expected path /api/v3/series, got %s", r.URL.Path)
		}
		if r.Header.Get("X-Api-Key") != "test-key" {
			t.Errorf("expected X-Api-Key header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(series)
	}))
	defer server.Close()

	client := New(Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Timeout: 5 * time.Second,
		RetryConfig: retry.Config{
			MaxAttempts: 1,
		},
	})

	ctx := context.Background()
	missing, err := client.GetMissingSeries(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only return monitored series with missing episodes
	if len(missing) != 1 {
		t.Errorf("expected 1 missing series, got %d", len(missing))
	}
	if missing[0].ID != 1 {
		t.Errorf("expected series ID 1, got %d", missing[0].ID)
	}
}

func TestGetAllSeries(t *testing.T) {
	series := []Series{
		{ID: 1, Title: "Monitored", TvdbID: 101, Monitored: true, TotalEpisodeCount: 10, EpisodeFileCount: 10},
		{ID: 3, Title: "Unmonitored", TvdbID: 103, Monitored: false, TotalEpisodeCount: 10, EpisodeFileCount: 0},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/series" {
			t.Errorf("expected path /api/v3/series, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(series)
	}))
	defer server.Close()

	client := New(Config{
		BaseURL:     server.URL,
		APIKey:      "test-key",
		Timeout:     5 * time.Second,
		RetryConfig: retry.Config{MaxAttempts: 1},
	})

	result, err := client.GetAllSeries(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected both monitored and unmonitored series to be returned, got %d", len(result))
	}
	byID := map[int]Series{}
	for _, s := range result {
		byID[s.ID] = s
	}
	if !byID[1].Monitored {
		t.Errorf("expected series 1 to be monitored")
	}
	if byID[3].Monitored {
		t.Errorf("expected series 3 to be unmonitored")
	}
}

func TestGetAllMonitoredSeries(t *testing.T) {
	series := []Series{
		{ID: 1, Title: "Fully Downloaded, Monitored", TvdbID: 101, Monitored: true, TotalEpisodeCount: 10, EpisodeFileCount: 10},
		{ID: 2, Title: "Missing Episodes, Monitored", TvdbID: 102, Monitored: true, TotalEpisodeCount: 10, EpisodeFileCount: 5},
		{ID: 3, Title: "Unmonitored", TvdbID: 103, Monitored: false, TotalEpisodeCount: 10, EpisodeFileCount: 0},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/series" {
			t.Errorf("expected path /api/v3/series, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(series)
	}))
	defer server.Close()

	client := New(Config{
		BaseURL:     server.URL,
		APIKey:      "test-key",
		Timeout:     5 * time.Second,
		RetryConfig: retry.Config{MaxAttempts: 1},
	})

	result, err := client.GetAllMonitoredSeries(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 monitored series (with and without missing episodes), got %d", len(result))
	}
	ids := map[int]bool{result[0].ID: true, result[1].ID: true}
	if !ids[1] || !ids[2] {
		t.Errorf("expected series 1 and 2 (both monitored) to be included, got %+v", result)
	}
}

func TestGetAllMonitoredSeriesDoesNotAffectGetMissingSeries(t *testing.T) {
	callCount := 0
	series := []Series{
		{ID: 1, Title: "Fully Downloaded, Monitored", TvdbID: 101, Monitored: true, TotalEpisodeCount: 10, EpisodeFileCount: 10},
		{ID: 2, Title: "Missing Episodes, Monitored", TvdbID: 102, Monitored: true, TotalEpisodeCount: 10, EpisodeFileCount: 5},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(series)
	}))
	defer server.Close()

	client := New(Config{
		BaseURL:     server.URL,
		APIKey:      "test-key",
		Timeout:     5 * time.Second,
		RetryConfig: retry.Config{MaxAttempts: 1},
	})

	all, err := client.GetAllMonitoredSeries(context.Background())
	if err != nil {
		t.Fatalf("unexpected error from GetAllMonitoredSeries: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 monitored series, got %d", len(all))
	}

	missing, err := client.GetMissingSeries(context.Background())
	if err != nil {
		t.Fatalf("unexpected error from GetMissingSeries: %v", err)
	}
	if len(missing) != 1 || missing[0].ID != 2 {
		t.Fatalf("expected GetMissingSeries to still only return series with missing episodes, got %+v", missing)
	}
}

func TestGetMissingEpisodes(t *testing.T) {
	episodes := []Episode{
		{ID: 1, SeriesID: 1, SeasonNumber: 1, EpisodeNumber: 1, HasFile: false, Monitored: true},
		{ID: 2, SeriesID: 1, SeasonNumber: 1, EpisodeNumber: 2, HasFile: false, Monitored: true},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/wanted/missing" {
			t.Errorf("expected path /api/v3/wanted/missing, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		response := struct {
			TotalRecords int       `json:"totalRecords"`
			Records      []Episode `json:"records"`
		}{
			TotalRecords: len(episodes),
			Records:      episodes,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := New(Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Timeout: 5 * time.Second,
		RetryConfig: retry.Config{
			MaxAttempts: 1,
		},
	})

	ctx := context.Background()
	missing, err := client.GetMissingEpisodes(ctx, FetchOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(missing) != 2 {
		t.Errorf("expected 2 missing episodes, got %d", len(missing))
	}
}

func TestGetEpisodeDetails(t *testing.T) {
	episode := Episode{
		ID:            1,
		SeriesID:      1,
		Title:         "Test Episode",
		SeasonNumber:  1,
		EpisodeNumber: 1,
		HasFile:       false,
		Monitored:     true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/episode/1" {
			t.Errorf("expected path /api/v3/episode/1, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(episode)
	}))
	defer server.Close()

	client := New(Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Timeout: 5 * time.Second,
		RetryConfig: retry.Config{
			MaxAttempts: 1,
		},
	})

	ctx := context.Background()
	result, err := client.GetEpisodeDetails(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != episode.ID {
		t.Errorf("expected ID %d, got %d", episode.ID, result.ID)
	}
	if result.Title != episode.Title {
		t.Errorf("expected title %s, got %s", episode.Title, result.Title)
	}
}

func TestUpdateEpisode(t *testing.T) {
	episode := &Episode{
		ID:            1,
		SeriesID:      1,
		Title:         "Test Episode",
		SeasonNumber:  1,
		EpisodeNumber: 1,
		HasFile:       true,
		Monitored:     true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v3/episode/1" {
			t.Errorf("expected path /api/v3/episode/1, got %s", r.URL.Path)
		}

		var received Episode
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if received.ID != episode.ID {
			t.Errorf("expected ID %d, got %d", episode.ID, received.ID)
		}

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(received)
	}))
	defer server.Close()

	client := New(Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Timeout: 5 * time.Second,
		RetryConfig: retry.Config{
			MaxAttempts: 1,
		},
	})

	ctx := context.Background()
	err := client.UpdateEpisode(ctx, episode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetMissingEpisodesMultiPage(t *testing.T) {
	// Three episodes total, pageSize=2 → should make 2 requests.
	allEpisodes := []Episode{
		{ID: 1, SeriesID: 1, SeasonNumber: 1, EpisodeNumber: 1},
		{ID: 2, SeriesID: 1, SeasonNumber: 1, EpisodeNumber: 2},
		{ID: 3, SeriesID: 2, SeasonNumber: 1, EpisodeNumber: 1},
	}
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/wanted/missing" {
			t.Errorf("expected path /api/v3/wanted/missing, got %s", r.URL.Path)
		}
		page := r.URL.Query().Get("page")
		pageSize := 2

		requestCount++
		var records []Episode
		switch page {
		case "1":
			records = allEpisodes[:pageSize]
		default:
			records = allEpisodes[pageSize:]
		}

		w.Header().Set("Content-Type", "application/json")
		response := struct {
			TotalRecords int       `json:"totalRecords"`
			Records      []Episode `json:"records"`
		}{
			TotalRecords: len(allEpisodes),
			Records:      records,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := New(Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Timeout: 5 * time.Second,
		RetryConfig: retry.Config{
			MaxAttempts: 1,
		},
	})

	ctx := context.Background()
	result, err := client.GetMissingEpisodes(ctx, FetchOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != len(allEpisodes) {
		t.Errorf("expected %d episodes, got %d", len(allEpisodes), len(result))
	}
	if requestCount != 2 {
		t.Errorf("expected 2 page requests, got %d", requestCount)
	}
}

func TestGetMissingEpisodesEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := struct {
			TotalRecords int       `json:"totalRecords"`
			Records      []Episode `json:"records"`
		}{TotalRecords: 0, Records: nil}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := New(Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Timeout: 5 * time.Second,
		RetryConfig: retry.Config{
			MaxAttempts: 1,
		},
	})

	ctx := context.Background()
	result, err := client.GetMissingEpisodes(ctx, FetchOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d episodes", len(result))
	}
}

func TestGetMissingEpisodesWithLimit(t *testing.T) {
	allEpisodes := []Episode{
		{ID: 1, SeriesID: 1, SeasonNumber: 1, EpisodeNumber: 1},
		{ID: 2, SeriesID: 1, SeasonNumber: 1, EpisodeNumber: 2},
		{ID: 3, SeriesID: 2, SeasonNumber: 1, EpisodeNumber: 1},
		{ID: 4, SeriesID: 2, SeasonNumber: 1, EpisodeNumber: 2},
		{ID: 5, SeriesID: 3, SeasonNumber: 1, EpisodeNumber: 1},
	}
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		const pageSize = 2
		requestCount++
		var records []Episode
		switch page {
		case "1":
			records = allEpisodes[:pageSize]
		case "2":
			records = allEpisodes[pageSize : pageSize*2]
		default:
			records = allEpisodes[pageSize*2:]
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			TotalRecords int       `json:"totalRecords"`
			Records      []Episode `json:"records"`
		}{TotalRecords: len(allEpisodes), Records: records})
	}))
	defer server.Close()

	client := New(Config{
		BaseURL:     server.URL,
		APIKey:      "test-key",
		Timeout:     5 * time.Second,
		RetryConfig: retry.Config{MaxAttempts: 1},
	})

	t.Run("limit within first page", func(t *testing.T) {
		requestCount = 0
		result, err := client.GetMissingEpisodes(context.Background(), FetchOptions{Limit: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("expected 1 episode, got %d", len(result))
		}
		if requestCount != 1 {
			t.Errorf("expected 1 request, got %d", requestCount)
		}
	})

	t.Run("limit on page boundary", func(t *testing.T) {
		requestCount = 0
		result, err := client.GetMissingEpisodes(context.Background(), FetchOptions{Limit: 2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 episodes, got %d", len(result))
		}
		if requestCount != 1 {
			t.Errorf("expected 1 request, got %d", requestCount)
		}
	})

	t.Run("limit spanning two pages", func(t *testing.T) {
		requestCount = 0
		result, err := client.GetMissingEpisodes(context.Background(), FetchOptions{Limit: 3})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 3 {
			t.Errorf("expected 3 episodes, got %d", len(result))
		}
		if requestCount != 2 {
			t.Errorf("expected 2 requests, got %d", requestCount)
		}
	})

	t.Run("limit larger than total", func(t *testing.T) {
		requestCount = 0
		result, err := client.GetMissingEpisodes(context.Background(), FetchOptions{Limit: 100})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != len(allEpisodes) {
			t.Errorf("expected %d episodes, got %d", len(allEpisodes), len(result))
		}
	})
}

func TestGetSeriesByTVDBID(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		series := Series{ID: 42, Title: "Breaking Bad", TvdbID: 81189}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v3/series" {
				t.Errorf("expected path /api/v3/series, got %s", r.URL.Path)
			}
			if r.URL.Query().Get("tvdbId") != "81189" {
				t.Errorf("expected tvdbId query param 81189, got %s", r.URL.Query().Get("tvdbId"))
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]Series{series})
		}))
		defer server.Close()

		client := New(Config{
			BaseURL:     server.URL,
			APIKey:      "test-key",
			Timeout:     5 * time.Second,
			RetryConfig: retry.Config{MaxAttempts: 1},
		})

		result, err := client.GetSeriesByTVDBID(context.Background(), 81189)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil || result.ID != series.ID {
			t.Fatalf("expected series ID %d, got %+v", series.ID, result)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]Series{})
		}))
		defer server.Close()

		client := New(Config{
			BaseURL:     server.URL,
			APIKey:      "test-key",
			Timeout:     5 * time.Second,
			RetryConfig: retry.Config{MaxAttempts: 1},
		})

		result, err := client.GetSeriesByTVDBID(context.Background(), 999)
		if err != nil {
			t.Fatalf("expected no error for not-found, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil series for not-found, got %+v", result)
		}
	})

	t.Run("transport error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("boom"))
		}))
		defer server.Close()

		client := New(Config{
			BaseURL:     server.URL,
			APIKey:      "test-key",
			Timeout:     5 * time.Second,
			RetryConfig: retry.Config{MaxAttempts: 1},
		})

		result, err := client.GetSeriesByTVDBID(context.Background(), 999)
		if err == nil {
			t.Fatal("expected error for transport failure")
		}
		if result != nil {
			t.Errorf("expected nil series on error, got %+v", result)
		}
	})
}

func TestGetEpisodesBySeriesID(t *testing.T) {
	episodes := []Episode{
		{ID: 1, SeriesID: 42, SeasonNumber: 1, EpisodeNumber: 1, Title: "Pilot"},
		{ID: 2, SeriesID: 42, SeasonNumber: 1, EpisodeNumber: 2, Title: "Cat's in the Bag..."},
		{ID: 3, SeriesID: 42, SeasonNumber: 2, EpisodeNumber: 1, Title: "Seven Thirty-Seven"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/episode" {
			t.Errorf("expected path /api/v3/episode, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("seriesId") != "42" {
			t.Errorf("expected seriesId query param 42, got %s", r.URL.Query().Get("seriesId"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(episodes)
	}))
	defer server.Close()

	client := New(Config{
		BaseURL:     server.URL,
		APIKey:      "test-key",
		Timeout:     5 * time.Second,
		RetryConfig: retry.Config{MaxAttempts: 1},
	})

	result, err := client.GetEpisodesBySeriesID(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != len(episodes) {
		t.Fatalf("expected %d episodes, got %d", len(episodes), len(result))
	}
}

func TestFindEpisodeByTVDBID(t *testing.T) {
	series := Series{ID: 42, Title: "Breaking Bad", TvdbID: 81189}
	episodes := []Episode{
		{ID: 1, SeriesID: 42, SeasonNumber: 1, EpisodeNumber: 1, Title: "Pilot"},
		{ID: 2, SeriesID: 42, SeasonNumber: 1, EpisodeNumber: 2, Title: "Cat's in the Bag..."},
	}

	newTestServer := func(seriesResp []Series, episodesResp []Episode) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/api/v3/series":
				json.NewEncoder(w).Encode(seriesResp)
			case "/api/v3/episode":
				json.NewEncoder(w).Encode(episodesResp)
			default:
				t.Errorf("unexpected path %s", r.URL.Path)
			}
		}))
	}

	t.Run("matching episode found", func(t *testing.T) {
		server := newTestServer([]Series{series}, episodes)
		defer server.Close()

		client := New(Config{
			BaseURL:     server.URL,
			APIKey:      "test-key",
			Timeout:     5 * time.Second,
			RetryConfig: retry.Config{MaxAttempts: 1},
		})

		gotSeries, gotEpisode, err := client.FindEpisodeByTVDBID(context.Background(), 81189, 1, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotSeries == nil || gotSeries.ID != series.ID {
			t.Fatalf("expected series %+v, got %+v", series, gotSeries)
		}
		if gotEpisode == nil || gotEpisode.ID != 2 {
			t.Fatalf("expected episode ID 2, got %+v", gotEpisode)
		}
	})

	t.Run("missing episode number", func(t *testing.T) {
		server := newTestServer([]Series{series}, episodes)
		defer server.Close()

		client := New(Config{
			BaseURL:     server.URL,
			APIKey:      "test-key",
			Timeout:     5 * time.Second,
			RetryConfig: retry.Config{MaxAttempts: 1},
		})

		gotSeries, gotEpisode, err := client.FindEpisodeByTVDBID(context.Background(), 81189, 5, 9)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotSeries != nil || gotEpisode != nil {
			t.Fatalf("expected not found, got series=%+v episode=%+v", gotSeries, gotEpisode)
		}
	})

	t.Run("missing series", func(t *testing.T) {
		server := newTestServer([]Series{}, episodes)
		defer server.Close()

		client := New(Config{
			BaseURL:     server.URL,
			APIKey:      "test-key",
			Timeout:     5 * time.Second,
			RetryConfig: retry.Config{MaxAttempts: 1},
		})

		gotSeries, gotEpisode, err := client.FindEpisodeByTVDBID(context.Background(), 999, 1, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotSeries != nil || gotEpisode != nil {
			t.Fatalf("expected not found, got series=%+v episode=%+v", gotSeries, gotEpisode)
		}
	})
}

func TestSystemStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v3/system/status" {
				t.Errorf("expected path /api/v3/system/status, got %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"version":"4.0.0"}`))
		}))
		defer server.Close()

		client := New(Config{BaseURL: server.URL, APIKey: "test-key", Timeout: 5 * time.Second})

		if err := client.SystemStatus(context.Background()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(200 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := New(Config{BaseURL: server.URL, APIKey: "test-key", Timeout: 5 * time.Second})

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()

		err := client.SystemStatus(ctx)
		if err == nil {
			t.Fatal("expected timeout error")
		}
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("expected a context.DeadlineExceeded error, got %v", err)
		}
	})

	t.Run("connection refused", func(t *testing.T) {
		client := New(Config{BaseURL: "http://127.0.0.1:1", APIKey: "test-key", Timeout: 5 * time.Second})

		if err := client.SystemStatus(context.Background()); err == nil {
			t.Fatal("expected connection error")
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message":"Unauthorized"}`))
		}))
		defer server.Close()

		client := New(Config{BaseURL: server.URL, APIKey: "bad-key", Timeout: 5 * time.Second})

		err := client.SystemStatus(context.Background())
		if err == nil {
			t.Fatal("expected unauthorized error")
		}
		var statusErr *StatusError
		if !errors.As(err, &statusErr) {
			t.Fatalf("expected a *StatusError, got %T: %v", err, err)
		}
		if statusErr.StatusCode() != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", statusErr.StatusCode())
		}
	})

	t.Run("open breaker short-circuits without an HTTP call", func(t *testing.T) {
		called := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		breaker := circuitbreaker.New(circuitbreaker.Config{
			MaxFailures:         1,
			Timeout:             time.Minute,
			MaxHalfOpenRequests: 1,
		})
		breaker.Execute(func() error { return errors.New("boom") })
		if breaker.State() != circuitbreaker.StateOpen {
			t.Fatalf("expected breaker to be open, got %s", breaker.State())
		}

		client := New(Config{BaseURL: server.URL, APIKey: "test-key", Timeout: 5 * time.Second, Breaker: breaker})

		err := client.SystemStatus(context.Background())
		if !errors.Is(err, circuitbreaker.ErrOpenState) {
			t.Fatalf("expected ErrOpenState, got %v", err)
		}
		if called {
			t.Error("expected no HTTP call to reach the test server while the breaker is open")
		}
	})
}
