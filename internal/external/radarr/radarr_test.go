package radarr

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/retry"
)

func TestNew(t *testing.T) {
	cfg := Config{
		BaseURL: "http://localhost:7878",
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

func TestGetMissingMovies(t *testing.T) {
	movies := []Movie{
		{ID: 1, Title: "Test Movie 1", Year: 2020, TMDBID: 101, Monitored: true, HasFile: false},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/wanted/missing" {
			t.Errorf("expected path /api/v3/wanted/missing, got %s", r.URL.Path)
		}
		if r.Header.Get("X-Api-Key") != "test-key" {
			t.Errorf("expected X-Api-Key header")
		}

		w.Header().Set("Content-Type", "application/json")
		response := struct {
			TotalRecords int     `json:"totalRecords"`
			Records      []Movie `json:"records"`
		}{
			TotalRecords: len(movies),
			Records:      movies,
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
	missing, err := client.GetMissingMovies(ctx, FetchOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(missing) != 1 {
		t.Errorf("expected 1 missing movie, got %d", len(missing))
	}
	if missing[0].ID != 1 {
		t.Errorf("expected movie ID 1, got %d", missing[0].ID)
	}
}

func TestGetMovieDetails(t *testing.T) {
	movie := Movie{
		ID:        1,
		Title:     "Test Movie",
		Year:      2020,
		TMDBID:    101,
		Monitored: true,
		HasFile:   false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/movie/1" {
			t.Errorf("expected path /api/v3/movie/1, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(movie)
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
	result, err := client.GetMovieDetails(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != movie.ID {
		t.Errorf("expected ID %d, got %d", movie.ID, result.ID)
	}
	if result.Title != movie.Title {
		t.Errorf("expected title %s, got %s", movie.Title, result.Title)
	}
}

func TestUpdateMovie(t *testing.T) {
	movie := &Movie{
		ID:        1,
		Title:     "Test Movie",
		Year:      2020,
		TMDBID:    101,
		Monitored: true,
		HasFile:   true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v3/movie/1" {
			t.Errorf("expected path /api/v3/movie/1, got %s", r.URL.Path)
		}

		var received Movie
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if received.ID != movie.ID {
			t.Errorf("expected ID %d, got %d", movie.ID, received.ID)
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
	err := client.UpdateMovie(ctx, movie)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMovieTvdbIDDeserialization(t *testing.T) {
	t.Run("with tvdbId field", func(t *testing.T) {
		payload := `{"id":1,"title":"Test Movie","year":2020,"tvdbId":12345,"tmdbId":99}`
		var m Movie
		if err := json.Unmarshal([]byte(payload), &m); err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}
		if m.TvdbID != 12345 {
			t.Errorf("expected TvdbID 12345, got %d", m.TvdbID)
		}
	})

	t.Run("without tvdbId field", func(t *testing.T) {
		payload := `{"id":2,"title":"No TVDB Movie","year":2021,"tmdbId":42}`
		var m Movie
		if err := json.Unmarshal([]byte(payload), &m); err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}
		if m.TvdbID != 0 {
			t.Errorf("expected TvdbID 0 when field absent, got %d", m.TvdbID)
		}
	})
}

func TestGetMovieByTMDBID(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		movie := Movie{ID: 5, Title: "Dune", Year: 2021, TMDBID: 438631, HasFile: true}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v3/movie" {
				t.Errorf("expected path /api/v3/movie, got %s", r.URL.Path)
			}
			if r.URL.Query().Get("tmdbId") != "438631" {
				t.Errorf("expected tmdbId query param 438631, got %s", r.URL.Query().Get("tmdbId"))
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]Movie{movie})
		}))
		defer server.Close()

		client := New(Config{
			BaseURL:     server.URL,
			APIKey:      "test-key",
			Timeout:     5 * time.Second,
			RetryConfig: retry.Config{MaxAttempts: 1},
		})

		result, err := client.GetMovieByTMDBID(context.Background(), 438631)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil movie")
		}
		if result.ID != movie.ID {
			t.Errorf("expected ID %d, got %d", movie.ID, result.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]Movie{})
		}))
		defer server.Close()

		client := New(Config{
			BaseURL:     server.URL,
			APIKey:      "test-key",
			Timeout:     5 * time.Second,
			RetryConfig: retry.Config{MaxAttempts: 1},
		})

		result, err := client.GetMovieByTMDBID(context.Background(), 999)
		if err != nil {
			t.Fatalf("expected no error for not-found, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil movie for not-found, got %+v", result)
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

		result, err := client.GetMovieByTMDBID(context.Background(), 999)
		if err == nil {
			t.Fatal("expected error for transport failure")
		}
		if result != nil {
			t.Errorf("expected nil movie on error, got %+v", result)
		}
	})
}

func TestGetAllMovies(t *testing.T) {
	movies := []Movie{
		{ID: 1, Title: "Monitored With File", Year: 2020, TMDBID: 101, Monitored: true, HasFile: true},
		{ID: 2, Title: "Monitored Missing File", Year: 2021, TMDBID: 102, Monitored: true, HasFile: false},
		{ID: 3, Title: "Unmonitored", Year: 2022, TMDBID: 103, Monitored: false, HasFile: false},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/movie" {
			t.Errorf("expected path /api/v3/movie, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(movies)
	}))
	defer server.Close()

	client := New(Config{
		BaseURL:     server.URL,
		APIKey:      "test-key",
		Timeout:     5 * time.Second,
		RetryConfig: retry.Config{MaxAttempts: 1},
	})

	result, err := client.GetAllMovies(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != len(movies) {
		t.Fatalf("expected %d movies (regardless of monitored/hasFile), got %d", len(movies), len(result))
	}
}

func TestGetAllMoviesDoesNotAffectGetMissingMovies(t *testing.T) {
	missingCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v3/movie":
			json.NewEncoder(w).Encode([]Movie{{ID: 9, Title: "All Movie", Year: 2019}})
		case "/api/v3/wanted/missing":
			missingCalls++
			json.NewEncoder(w).Encode(struct {
				TotalRecords int     `json:"totalRecords"`
				Records      []Movie `json:"records"`
			}{TotalRecords: 1, Records: []Movie{{ID: 1, Title: "Missing Movie", Year: 2020}}})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := New(Config{
		BaseURL:     server.URL,
		APIKey:      "test-key",
		Timeout:     5 * time.Second,
		RetryConfig: retry.Config{MaxAttempts: 1},
	})

	all, err := client.GetAllMovies(context.Background())
	if err != nil {
		t.Fatalf("unexpected error from GetAllMovies: %v", err)
	}
	if len(all) != 1 || all[0].ID != 9 {
		t.Fatalf("unexpected GetAllMovies result: %+v", all)
	}

	missing, err := client.GetMissingMovies(context.Background(), FetchOptions{})
	if err != nil {
		t.Fatalf("unexpected error from GetMissingMovies: %v", err)
	}
	if len(missing) != 1 || missing[0].ID != 1 {
		t.Fatalf("unexpected GetMissingMovies result: %+v", missing)
	}
	if missingCalls != 1 {
		t.Errorf("expected 1 call to wanted/missing, got %d", missingCalls)
	}
}

func TestClientRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.Header().Set("Content-Type", "application/json")
		response := struct {
			TotalRecords int     `json:"totalRecords"`
			Records      []Movie `json:"records"`
		}{TotalRecords: 0, Records: nil}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := New(Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Timeout: 5 * time.Second,
		RetryConfig: retry.Config{
			MaxAttempts:       3,
			InitialBackoff:    10 * time.Millisecond,
			MaxBackoff:        100 * time.Millisecond,
			BackoffMultiplier: 2.0,
			JitterFraction:    0.1,
		},
	})

	ctx := context.Background()
	result, err := client.GetMissingMovies(ctx, FetchOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d movies", len(result))
	}
	if attempts != 1 {
		t.Errorf("expected 1 request for empty result set, got %d", attempts)
	}
}

func TestGetMissingMoviesMultiPage(t *testing.T) {
	allMovies := []Movie{
		{ID: 1, Title: "Movie A", Year: 2020, TMDBID: 101},
		{ID: 2, Title: "Movie B", Year: 2021, TMDBID: 102},
		{ID: 3, Title: "Movie C", Year: 2022, TMDBID: 103},
	}
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/wanted/missing" {
			t.Errorf("expected path /api/v3/wanted/missing, got %s", r.URL.Path)
		}
		page := r.URL.Query().Get("page")
		pageSize := 2

		requestCount++
		var records []Movie
		switch page {
		case "1":
			records = allMovies[:pageSize]
		default:
			records = allMovies[pageSize:]
		}

		w.Header().Set("Content-Type", "application/json")
		response := struct {
			TotalRecords int     `json:"totalRecords"`
			Records      []Movie `json:"records"`
		}{
			TotalRecords: len(allMovies),
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
	result, err := client.GetMissingMovies(ctx, FetchOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != len(allMovies) {
		t.Errorf("expected %d movies, got %d", len(allMovies), len(result))
	}
	if requestCount != 2 {
		t.Errorf("expected 2 page requests, got %d", requestCount)
	}
}

func TestGetMissingMoviesEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := struct {
			TotalRecords int     `json:"totalRecords"`
			Records      []Movie `json:"records"`
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
	result, err := client.GetMissingMovies(ctx, FetchOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d movies", len(result))
	}
}

func TestSystemStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v3/system/status" {
				t.Errorf("expected path /api/v3/system/status, got %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"version":"5.0.0"}`))
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
}

func TestGetMissingMoviesWithLimit(t *testing.T) {
	allMovies := []Movie{
		{ID: 1, Title: "Movie A", Year: 2020, TMDBID: 101},
		{ID: 2, Title: "Movie B", Year: 2021, TMDBID: 102},
		{ID: 3, Title: "Movie C", Year: 2022, TMDBID: 103},
		{ID: 4, Title: "Movie D", Year: 2023, TMDBID: 104},
		{ID: 5, Title: "Movie E", Year: 2024, TMDBID: 105},
	}
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		const pageSize = 2
		requestCount++
		var records []Movie
		switch page {
		case "1":
			records = allMovies[:pageSize]
		case "2":
			records = allMovies[pageSize : pageSize*2]
		default:
			records = allMovies[pageSize*2:]
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			TotalRecords int     `json:"totalRecords"`
			Records      []Movie `json:"records"`
		}{TotalRecords: len(allMovies), Records: records})
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
		result, err := client.GetMissingMovies(context.Background(), FetchOptions{Limit: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("expected 1 movie, got %d", len(result))
		}
		if requestCount != 1 {
			t.Errorf("expected 1 request, got %d", requestCount)
		}
	})

	t.Run("limit on page boundary", func(t *testing.T) {
		requestCount = 0
		result, err := client.GetMissingMovies(context.Background(), FetchOptions{Limit: 2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 movies, got %d", len(result))
		}
		if requestCount != 1 {
			t.Errorf("expected 1 request, got %d", requestCount)
		}
	})

	t.Run("limit spanning two pages", func(t *testing.T) {
		requestCount = 0
		result, err := client.GetMissingMovies(context.Background(), FetchOptions{Limit: 3})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 3 {
			t.Errorf("expected 3 movies, got %d", len(result))
		}
		if requestCount != 2 {
			t.Errorf("expected 2 requests, got %d", requestCount)
		}
	})

	t.Run("limit larger than total", func(t *testing.T) {
		requestCount = 0
		result, err := client.GetMissingMovies(context.Background(), FetchOptions{Limit: 100})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != len(allMovies) {
			t.Errorf("expected %d movies, got %d", len(allMovies), len(result))
		}
	})
}
