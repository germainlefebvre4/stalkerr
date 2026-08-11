package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/external/tmdb"
	"github.com/glefebvre/stalkeer/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test SQLite DB: %v", err)
	}

	err = db.AutoMigrate(
		&models.Movie{},
		&models.TVShow{},
		&models.ProcessedLine{},
		&models.ProcessingLog{},
		&models.DownloadInfo{},
		&models.ManualMapping{},
		&models.FilterConfig{},
	)
	if err != nil {
		t.Fatalf("failed to migrate models: %v", err)
	}

	database.SetDB(db)
	return db
}

func TestListProcessingLogs(t *testing.T) {
	db := setupTestDB(t)

	// Seed logs
	log1 := models.ProcessingLog{
		Action:    "m3u-download",
		Status:    "success",
		StartedAt: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	log2 := models.ProcessingLog{
		Action:    "process",
		Status:    "in_progress",
		StartedAt: time.Now(),
		CreatedAt: time.Now().Add(time.Second),
		UpdatedAt: time.Now().Add(time.Second),
	}
	db.Create(&log1)
	db.Create(&log2)

	server := NewServer()

	// 1. Test un-filtered logs
	req, _ := http.NewRequest("GET", "/api/v1/processing-logs", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Total != 2 {
		t.Errorf("Expected total count 2, got %d", resp.Total)
	}

	// 2. Test status-filtered logs
	reqFiltered, _ := http.NewRequest("GET", "/api/v1/processing-logs?status=in_progress", nil)
	wFiltered := httptest.NewRecorder()
	server.router.ServeHTTP(wFiltered, reqFiltered)

	if wFiltered.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", wFiltered.Code)
	}

	var respFiltered PaginatedResponse
	if err := json.Unmarshal(wFiltered.Body.Bytes(), &respFiltered); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if respFiltered.Total != 1 {
		t.Errorf("Expected total count 1, got %d", respFiltered.Total)
	}
}

func TestListDownloads(t *testing.T) {
	db := setupTestDB(t)

	// Seed downloads
	dl1 := models.DownloadInfo{
		URL:       "http://example.com/movie.mkv",
		Status:    "completed",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	dl2 := models.DownloadInfo{
		URL:       "http://example.com/show.mkv",
		Status:    "downloading",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now().Add(time.Second),
	}
	db.Create(&dl1)
	db.Create(&dl2)

	server := NewServer()

	// 1. Test list downloads un-filtered
	req, _ := http.NewRequest("GET", "/api/v1/downloads", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Total != 2 {
		t.Errorf("Expected total count 2, got %d", resp.Total)
	}

	// 2. Test list downloads filtered
	reqFiltered, _ := http.NewRequest("GET", "/api/v1/downloads?status=downloading", nil)
	wFiltered := httptest.NewRecorder()
	server.router.ServeHTTP(wFiltered, reqFiltered)

	if wFiltered.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", wFiltered.Code)
	}

	var respFiltered PaginatedResponse
	if err := json.Unmarshal(wFiltered.Body.Bytes(), &respFiltered); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if respFiltered.Total != 1 {
		t.Errorf("Expected total count 1, got %d", respFiltered.Total)
	}
}

func TestListDownloadsEnriched(t *testing.T) {
	db := setupTestDB(t)

	// Seed Movies and TV Shows
	movie := models.Movie{
		TMDBID:     12345,
		TMDBTitle:  "The Matrix",
		TMDBYear:   1999,
		TMDBGenres: func(s string) *string { return &s }("Action, Sci-Fi"),
	}
	db.Create(&movie)

	tvShow := models.TVShow{
		TMDBID:     67890,
		TMDBTitle:  "Breaking Bad",
		TMDBYear:   2008,
		TMDBGenres: func(s string) *string { return &s }("Drama"),
		Season:     func(i int) *int { return &i }(5),
		Episode:    func(i int) *int { return &i }(14),
	}
	db.Create(&tvShow)

	// Seed Downloads with Processed Lines
	path1 := "/media/movies/The.Matrix.1999/matrix.mkv"
	dl1 := models.DownloadInfo{
		URL:          "http://example.com/matrix",
		Status:       "completed",
		DownloadPath: &path1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	db.Create(&dl1)

	pl1 := models.ProcessedLine{
		LineContent:    "matrix content",
		LineHash:       "hash1",
		TvgName:        "The Matrix",
		ContentType:    "movies",
		MovieID:        &movie.ID,
		DownloadInfoID: &dl1.ID,
		State:          "downloaded",
		ProcessedAt:    time.Now(),
	}
	db.Create(&pl1)

	path2 := "/media/tvshows/Breaking_Bad_S05E14.mp4"
	dl2 := models.DownloadInfo{
		URL:          "http://example.com/breakingbad",
		Status:       "completed",
		DownloadPath: &path2,
		CreatedAt:    time.Now().Add(time.Second),
		UpdatedAt:    time.Now().Add(time.Second),
	}
	db.Create(&dl2)

	pl2 := models.ProcessedLine{
		LineContent:    "breaking bad content",
		LineHash:       "hash2",
		TvgName:        "Breaking Bad",
		ContentType:    "tvshows",
		TVShowID:       &tvShow.ID,
		DownloadInfoID: &dl2.ID,
		State:          "downloaded",
		ProcessedAt:    time.Now(),
	}
	db.Create(&pl2)

	// Flat path with missing year
	path3 := "avatar.mkv"
	dl3 := models.DownloadInfo{
		URL:          "http://example.com/avatar",
		Status:       "completed",
		DownloadPath: &path3,
		CreatedAt:    time.Now().Add(time.Second * 2),
		UpdatedAt:    time.Now().Add(time.Second * 2),
	}
	db.Create(&dl3)

	pl3 := models.ProcessedLine{
		LineContent:    "avatar content",
		LineHash:       "hash3",
		TvgName:        "Avatar",
		ContentType:    "movies",
		DownloadInfoID: &dl3.ID,
		State:          "downloaded",
		ProcessedAt:    time.Now(),
	}
	db.Create(&pl3)

	server := NewServer()

	// 1. Query un-filtered enriched
	req, _ := http.NewRequest("GET", "/api/v1/downloads", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var resp struct {
		Data  []DownloadEnrichedResponse `json:"data"`
		Total int64                      `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.Total != 3 {
		t.Errorf("Expected 3, got %d", resp.Total)
	}

	// Verify the enriched content for Matrix
	foundMatrix := false
	for _, item := range resp.Data {
		if item.ID == dl1.ID {
			foundMatrix = true
			if item.Content == nil {
				t.Error("Matrix content should not be nil")
			} else {
				if item.Content.Title != "The Matrix" {
					t.Errorf("Expected Title 'The Matrix', got %q", item.Content.Title)
				}
				if item.Content.Type != "movies" {
					t.Errorf("Expected Type 'movies', got %q", item.Content.Type)
				}
			}
			if item.FileInfo == nil {
				t.Error("Matrix file_info should not be nil")
			} else {
				if item.FileInfo.FolderName != "The.Matrix.1999" {
					t.Errorf("Expected folder name 'The.Matrix.1999', got %q", item.FileInfo.FolderName)
				}
				if !item.FileInfo.HasYearInPath {
					t.Error("Expected HasYearInPath to be true")
				}
			}
		}
	}
	if !foundMatrix {
		t.Error("Did not find Matrix download in enriched response")
	}

	// 2. Query with problem=missing_year (should return dl3 and dl2 because they have no year in path)
	reqMissingYear, _ := http.NewRequest("GET", "/api/v1/downloads?problem=missing_year", nil)
	wMissingYear := httptest.NewRecorder()
	server.router.ServeHTTP(wMissingYear, reqMissingYear)

	var respMissingYear struct {
		Data  []DownloadEnrichedResponse `json:"data"`
		Total int64                      `json:"total"`
	}
	if err := json.Unmarshal(wMissingYear.Body.Bytes(), &respMissingYear); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(respMissingYear.Data) != 2 {
		t.Errorf("Expected 2 downloads with missing year, got %d", len(respMissingYear.Data))
	}

	// 3. Query with type=movies (should return dl1, dl3)
	reqType, _ := http.NewRequest("GET", "/api/v1/downloads?type=movies", nil)
	wType := httptest.NewRecorder()
	server.router.ServeHTTP(wType, reqType)

	var respType struct {
		Data  []DownloadEnrichedResponse `json:"data"`
		Total int64                      `json:"total"`
	}
	if err := json.Unmarshal(wType.Body.Bytes(), &respType); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if respType.Total != 2 {
		t.Errorf("Expected 2 movies, got %d", respType.Total)
	}
}

func TestGetConfigPaths(t *testing.T) {
	_ = setupTestDB(t)

	// Ensure config is set
	config.SetConfig(&config.Config{
		Downloads: config.DownloadsConfig{
			MoviesPath:  "/test/movies/path",
			TVShowsPath: "/test/tvshows/path",
		},
	})

	server := NewServer()

	req, _ := http.NewRequest("GET", "/api/v1/config/paths", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["movies_path"] != "/test/movies/path" {
		t.Errorf("Expected movies_path '/test/movies/path', got '%s'", resp["movies_path"])
	}
	if resp["tvshows_path"] != "/test/tvshows/path" {
		t.Errorf("Expected tvshows_path '/test/tvshows/path', got '%s'", resp["tvshows_path"])
	}
}

func TestMoveMovieFolder(t *testing.T) {
	db := setupTestDB(t)

	tempDir, err := os.MkdirTemp("", "stalkeer-test-move-movie")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	srcMovieParentDir := filepath.Join(tempDir, "movies")
	movieFolderName := "Interstellar (2014)"
	srcMovieFolder := filepath.Join(srcMovieParentDir, movieFolderName)
	dstMovieParentDir := filepath.Join(tempDir, "children-movies")

	// Create physical source file to move
	err = os.MkdirAll(srcMovieFolder, 0755)
	if err != nil {
		t.Fatalf("failed to create test movie folder: %v", err)
	}
	srcFilePath := filepath.Join(srcMovieFolder, "Interstellar (2014).mkv")
	err = os.WriteFile(srcFilePath, []byte("movie_content"), 0644)
	if err != nil {
		t.Fatalf("failed to write test movie file: %v", err)
	}

	// Create Database models
	movie := models.Movie{
		TMDBID:    157336,
		TMDBTitle: "Interstellar",
		TMDBYear:  2014,
	}
	db.Create(&movie)

	dl := models.DownloadInfo{
		URL:          "http://example.com/interstellar.mkv",
		Status:       "completed",
		DownloadPath: &srcFilePath,
	}
	db.Create(&dl)

	line := models.ProcessedLine{
		LineContent:    "#EXTINF:-1,Interstellar (2014)",
		LineHash:       "interstellar_hash",
		TvgName:        "Interstellar",
		GroupTitle:     "Movies",
		ProcessedAt:    time.Now(),
		ContentType:    models.ContentTypeMovies,
		State:          models.StateDownloaded,
		MovieID:        &movie.ID,
		DownloadInfoID: &dl.ID,
	}
	db.Create(&line)

	server := NewServer()

	// Execute move call
	reqBody, _ := json.Marshal(MoveDownloadRequest{
		DestinationParentDir: dstMovieParentDir,
	})
	req, _ := http.NewRequest("POST", "/api/v1/movies/1/move", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify disk state
	expectedNewPath := filepath.Join(dstMovieParentDir, movieFolderName, "Interstellar (2014).mkv")
	if _, err := os.Stat(expectedNewPath); os.IsNotExist(err) {
		t.Errorf("Expected file to be moved to %s, but it does not exist", expectedNewPath)
	}
	if _, err := os.Stat(srcFilePath); err == nil {
		t.Errorf("Expected source file at %s to be deleted, but it still exists", srcFilePath)
	}

	// Verify Database state
	var updatedDl models.DownloadInfo
	if err := db.First(&updatedDl, dl.ID).Error; err != nil {
		t.Fatalf("failed to retrieve updated download info: %v", err)
	}
	if updatedDl.DownloadPath == nil || *updatedDl.DownloadPath != expectedNewPath {
		t.Errorf("Expected database download path to be '%s', got '%s'", expectedNewPath, *updatedDl.DownloadPath)
	}
}

func TestMoveTVShowFolder(t *testing.T) {
	db := setupTestDB(t)

	tempDir, err := os.MkdirTemp("", "stalkeer-test-move-tv")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	srcTVParentDir := filepath.Join(tempDir, "tvshows")
	seriesFolderName := "Breaking Bad (2008)"
	srcSeriesFolder := filepath.Join(srcTVParentDir, seriesFolderName)
	srcSeasonFolder := filepath.Join(srcSeriesFolder, "Season 01")
	dstTVParentDir := filepath.Join(tempDir, "children-tv")

	// Create physical source folders and file
	err = os.MkdirAll(srcSeasonFolder, 0755)
	if err != nil {
		t.Fatalf("failed to create test series directories: %v", err)
	}
	srcFilePath := filepath.Join(srcSeasonFolder, "Breaking Bad (2008) - S01E01.mkv")
	err = os.WriteFile(srcFilePath, []byte("episode_content"), 0644)
	if err != nil {
		t.Fatalf("failed to write test episode file: %v", err)
	}

	// Create Database models
	tvShow := models.TVShow{
		TMDBID:    1396,
		TMDBTitle: "Breaking Bad",
		TMDBYear:  2008,
	}
	db.Create(&tvShow)

	dl := models.DownloadInfo{
		URL:          "http://example.com/breakingbad.mkv",
		Status:       "completed",
		DownloadPath: &srcFilePath,
	}
	db.Create(&dl)

	line := models.ProcessedLine{
		LineContent:    "#EXTINF:-1,Breaking Bad (2008)",
		LineHash:       "breakingbad_hash",
		TvgName:        "Breaking Bad",
		GroupTitle:     "Series",
		ProcessedAt:    time.Now(),
		ContentType:    models.ContentTypeTVShows,
		State:          models.StateDownloaded,
		TVShowID:       &tvShow.ID,
		DownloadInfoID: &dl.ID,
	}
	db.Create(&line)

	server := NewServer()

	// Execute move call
	reqBody, _ := json.Marshal(MoveDownloadRequest{
		DestinationParentDir: dstTVParentDir,
	})
	req, _ := http.NewRequest("POST", "/api/v1/tvshows/1/move", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify disk state
	expectedNewPath := filepath.Join(dstTVParentDir, seriesFolderName, "Season 01", "Breaking Bad (2008) - S01E01.mkv")
	if _, err := os.Stat(expectedNewPath); os.IsNotExist(err) {
		t.Errorf("Expected TV show file to be moved to %s, but it does not exist", expectedNewPath)
	}
	if _, err := os.Stat(srcFilePath); err == nil {
		t.Errorf("Expected source TV show file at %s to be deleted, but it still exists", srcFilePath)
	}

	// Verify Database state
	var updatedDl models.DownloadInfo
	if err := db.First(&updatedDl, dl.ID).Error; err != nil {
		t.Fatalf("failed to retrieve updated download info: %v", err)
	}
	if updatedDl.DownloadPath == nil || *updatedDl.DownloadPath != expectedNewPath {
		t.Errorf("Expected database download path to be '%s', got '%s'", expectedNewPath, *updatedDl.DownloadPath)
	}
}

func TestSearchTMDBProxy_And_OverrideItem(t *testing.T) {
	db := setupTestDB(t)

	// Mock TMDB Server
	mockTMDB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasPrefix(r.URL.Path, "/search/movie") {
			w.Write([]byte(`{"results":[{"id":101,"title":"Mock Inception","original_title":"Inception","release_date":"2010-07-16","overview":"A dream within a dream.","poster_path":"/poster1.jpg"}]}`))
		} else if strings.HasPrefix(r.URL.Path, "/search/tv") {
			w.Write([]byte(`{"results":[{"id":202,"name":"Mock Malcolm","original_name":"Malcolm","first_air_date":"2000-01-09","overview":"Malcolm in the middle.","poster_path":"/poster2.jpg"}]}`))
		} else if strings.HasPrefix(r.URL.Path, "/movie/101") {
			if strings.HasSuffix(r.URL.Path, "/external_ids") {
				w.Write([]byte(`{"imdb_id":"tt1375666","tvdb_id":12345}`))
			} else {
				w.Write([]byte(`{"id":101,"title":"Mock Inception","release_date":"2010-07-16","genres":[{"id":28,"name":"Action"}],"runtime":148,"poster_path":"/inception-poster.jpg","overview":"A dream within a dream."}`))
			}
		} else if strings.HasPrefix(r.URL.Path, "/tv/202") {
			if strings.HasSuffix(r.URL.Path, "/external_ids") {
				w.Write([]byte(`{"imdb_id":"tt0248654","tvdb_id":67890}`))
			} else {
				w.Write([]byte(`{"id":202,"name":"Mock Malcolm","first_air_date":"2000-01-09","genres":[{"id":35,"name":"Comedy"}],"poster_path":"/malcolm-poster.jpg","overview":"Malcolm in the middle."}`))
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockTMDB.Close()

	// Override TMDB baseURL
	tmdb.SetBaseURL(mockTMDB.URL)
	defer tmdb.SetBaseURL("https://api.themoviedb.org/3")

	// Set configuration so TMDB is enabled
	config.SetConfig(&config.Config{
		TMDB: config.TMDBConfig{
			Enabled:           true,
			APIKey:            "mock-key",
			Language:          "en-US",
			RequestsPerSecond: 0,
		},
	})

	server := NewServer()

	// 1. Test Search Proxy (Movie)
	req1, _ := http.NewRequest("GET", "/api/v1/tmdb/search?query=Inception&type=movie", nil)
	w1 := httptest.NewRecorder()
	server.router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w1.Code, w1.Body.String())
	}
	var searchResults []TMDBSearchResult
	if err := json.Unmarshal(w1.Body.Bytes(), &searchResults); err != nil {
		t.Fatalf("failed to unmarshal search results: %v", err)
	}
	if len(searchResults) != 1 || searchResults[0].ID != 101 || searchResults[0].Title != "Mock Inception" {
		t.Errorf("unexpected movie search results: %+v", searchResults)
	}

	// 2. Test Search Proxy (TV Show)
	req2, _ := http.NewRequest("GET", "/api/v1/tmdb/search?query=Malcolm&type=tvshow", nil)
	w2 := httptest.NewRecorder()
	server.router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w2.Code)
	}
	var showResults []TMDBSearchResult
	if err := json.Unmarshal(w2.Body.Bytes(), &showResults); err != nil {
		t.Fatalf("failed to unmarshal show results: %v", err)
	}
	if len(showResults) != 1 || showResults[0].ID != 202 || showResults[0].Title != "Mock Malcolm" {
		t.Errorf("unexpected tvshow search results: %+v", showResults)
	}

	// Seed item to override
	item := models.ProcessedLine{
		LineContent: "Test content",
		LineHash:    "hash123",
		TvgName:     "FR: Inception (2010)",
		GroupTitle:  "FR: FILMS ACTION",
		ProcessedAt: time.Now(),
		ContentType: models.ContentTypeUncategorized,
		State:       models.StateProcessed,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	db.Create(&item)

	// 3. Test Override (Movie)
	overrideBody := OverrideItemRequest{
		TMDBID: 101,
		Type:   "movie",
	}
	bodyBytes, _ := json.Marshal(overrideBody)
	req3, _ := http.NewRequest("POST", "/api/v1/items/1/override", bytes.NewBuffer(bodyBytes))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	server.router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w3.Code, w3.Body.String())
	}
	var overrideResp ItemResponse
	if err := json.Unmarshal(w3.Body.Bytes(), &overrideResp); err != nil {
		t.Fatalf("failed to unmarshal override response: %v", err)
	}
	if overrideResp.ContentType != models.ContentTypeMovies || overrideResp.Movie == nil || overrideResp.Movie.TMDBID != 101 {
		t.Errorf("unexpected override movie response: %+v", overrideResp)
	}
	if overrideResp.OverrideBy == nil || *overrideResp.OverrideBy != "manual" || overrideResp.OverrideAt == nil {
		t.Errorf("missing override logging fields in response")
	}
	if overrideResp.Movie.PosterPath == nil || *overrideResp.Movie.PosterPath != "/inception-poster.jpg" {
		t.Errorf("expected poster_path=/inception-poster.jpg, got %+v", overrideResp.Movie.PosterPath)
	}
	if overrideResp.Movie.Overview == nil || *overrideResp.Movie.Overview != "A dream within a dream." {
		t.Errorf("expected overview to be persisted, got %+v", overrideResp.Movie.Overview)
	}
	if overrideResp.Movie.IMDBID == nil || *overrideResp.Movie.IMDBID != "tt1375666" {
		t.Errorf("expected imdb_id=tt1375666, got %+v", overrideResp.Movie.IMDBID)
	}
	if overrideResp.Movie.TVDBID == nil || *overrideResp.Movie.TVDBID != 12345 {
		t.Errorf("expected tvdb_id=12345, got %+v", overrideResp.Movie.TVDBID)
	}

	// Verify manual mapping was persisted
	var mapping models.ManualMapping
	if err := db.Where("tvg_name = ? AND group_title = ?", item.TvgName, item.GroupTitle).First(&mapping).Error; err != nil {
		t.Fatalf("failed to find persistent manual mapping in DB: %v", err)
	}
	if mapping.TMDBID != 101 || mapping.ContentType != models.ContentTypeMovies {
		t.Errorf("incorrect manual mapping persisted in DB: %+v", mapping)
	}

	// 4. Test Search Proxy (Disabled Config)
	config.SetConfig(&config.Config{
		TMDB: config.TMDBConfig{
			Enabled: false,
		},
	})
	serverDisabled := NewServer()
	req4, _ := http.NewRequest("GET", "/api/v1/tmdb/search?query=Inception&type=movie", nil)
	w4 := httptest.NewRecorder()
	serverDisabled.router.ServeHTTP(w4, req4)

	if w4.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w4.Code)
	}
}

func TestOverrideItem_TMDBDetailFetchFailureReturnsOverrideFailed(t *testing.T) {
	db := setupTestDB(t)

	// Mock TMDB server that always 404s on detail lookups
	mockTMDB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockTMDB.Close()

	tmdb.SetBaseURL(mockTMDB.URL)
	defer tmdb.SetBaseURL("https://api.themoviedb.org/3")

	config.SetConfig(&config.Config{
		TMDB: config.TMDBConfig{
			Enabled: true,
			APIKey:  "mock-key",
		},
	})

	server := NewServer()

	item := models.ProcessedLine{
		LineContent: "Test content",
		LineHash:    "override-failed-hash",
		TvgName:     "FR: Some Movie (2020)",
		GroupTitle:  "FR: FILMS ACTION",
		ProcessedAt: time.Now(),
		ContentType: models.ContentTypeUncategorized,
		State:       models.StateProcessed,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	db.Create(&item)

	overrideBody := OverrideItemRequest{
		TMDBID: 999,
		Type:   "movie",
	}
	bodyBytes, _ := json.Marshal(overrideBody)
	req, _ := http.NewRequest("POST", "/api/v1/items/1/override", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("Expected status 500, got %d. Body: %s", w.Code, w.Body.String())
	}
	var errResp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}
	if errResp.Error != "override_failed" {
		t.Errorf("Expected error code 'override_failed', got %q", errResp.Error)
	}
}

func TestListItemsFiltering(t *testing.T) {
	db := setupTestDB(t)

	// Seed Movie and TVShow
	movie := models.Movie{
		TMDBID:    111,
		TMDBTitle: "The Matrix",
	}
	db.Create(&movie)

	tvshow := models.TVShow{
		TMDBID:    222,
		TMDBTitle: "Breaking Bad",
	}
	db.Create(&tvshow)

	// Seed items
	item1 := models.ProcessedLine{
		LineContent: "matrix-m3u-line",
		LineHash:    "hash-matrix",
		TvgName:     "The Matrix",
		GroupTitle:  "Sci-Fi",
		ContentType: "movies",
		State:       "processed",
		MovieID:     &movie.ID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	db.Create(&item1)

	item2 := models.ProcessedLine{
		LineContent: "inception-m3u-line",
		LineHash:    "hash-inception",
		TvgName:     "Inception",
		GroupTitle:  "Sci-Fi",
		ContentType: "movies",
		State:       "processed",
		CreatedAt:   time.Now().Add(time.Second),
		UpdatedAt:   time.Now().Add(time.Second),
	}
	db.Create(&item2)

	item3 := models.ProcessedLine{
		LineContent: "breaking-bad-m3u-line",
		LineHash:    "hash-breaking-bad",
		TvgName:     "Breaking Bad",
		GroupTitle:  "Drama",
		ContentType: "tvshows",
		State:       "processed",
		TVShowID:    &tvshow.ID,
		CreatedAt:   time.Now().Add(2 * time.Second),
		UpdatedAt:   time.Now().Add(2 * time.Second),
	}
	db.Create(&item3)

	server := NewServer()

	// 1. Test un-filtered list items
	req, _ := http.NewRequest("GET", "/api/v1/items", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}
	var resp PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Total != 3 {
		t.Errorf("Expected 3 items, got %d", resp.Total)
	}

	// 2. Filter by tvg_name
	reqTvg, _ := http.NewRequest("GET", "/api/v1/items?tvg_name=Matrix", nil)
	wTvg := httptest.NewRecorder()
	server.router.ServeHTTP(wTvg, reqTvg)

	if wTvg.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", wTvg.Code)
	}
	var respTvg PaginatedResponse
	if err := json.Unmarshal(wTvg.Body.Bytes(), &respTvg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if respTvg.Total != 1 {
		t.Errorf("Expected 1 item, got %d", respTvg.Total)
	}

	// 3. Filter by tmdb_enriched=yes
	reqEnriched, _ := http.NewRequest("GET", "/api/v1/items?tmdb_enriched=yes", nil)
	wEnriched := httptest.NewRecorder()
	server.router.ServeHTTP(wEnriched, reqEnriched)

	if wEnriched.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", wEnriched.Code)
	}
	var respEnriched PaginatedResponse
	if err := json.Unmarshal(wEnriched.Body.Bytes(), &respEnriched); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if respEnriched.Total != 2 {
		t.Errorf("Expected 2 items, got %d", respEnriched.Total)
	}

	// 4. Filter by tmdb_enriched=no
	reqNotEnriched, _ := http.NewRequest("GET", "/api/v1/items?tmdb_enriched=no", nil)
	wNotEnriched := httptest.NewRecorder()
	server.router.ServeHTTP(wNotEnriched, reqNotEnriched)

	if wNotEnriched.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", wNotEnriched.Code)
	}
	var respNotEnriched PaginatedResponse
	if err := json.Unmarshal(wNotEnriched.Body.Bytes(), &respNotEnriched); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if respNotEnriched.Total != 1 {
		t.Errorf("Expected 1 item, got %d", respNotEnriched.Total)
	}
}

func TestGetItem_RemoteFileSizePresenceAndAbsence(t *testing.T) {
	db := setupTestDB(t)

	size := int64(4823123456)
	checkedAt := time.Now()

	withSize := models.ProcessedLine{
		LineContent:             "with-size-m3u-line",
		LineHash:                "hash-with-size",
		TvgName:                 "Movie With Known Size",
		GroupTitle:              "Movies",
		ContentType:             "movies",
		State:                   "processed",
		RemoteFileSize:          &size,
		RemoteFileSizeCheckedAt: &checkedAt,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}
	db.Create(&withSize)

	withoutSize := models.ProcessedLine{
		LineContent: "without-size-m3u-line",
		LineHash:    "hash-without-size",
		TvgName:     "Movie Never Checked",
		GroupTitle:  "Movies",
		ContentType: "movies",
		State:       "processed",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	db.Create(&withoutSize)

	server := NewServer()

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/items/%d", withSize.ID), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	var respWithSize ItemResponse
	if err := json.Unmarshal(w.Body.Bytes(), &respWithSize); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if respWithSize.RemoteFileSize == nil || *respWithSize.RemoteFileSize != size {
		t.Errorf("expected remote_file_size=%d, got %v", size, respWithSize.RemoteFileSize)
	}

	reqNoSize, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/items/%d", withoutSize.ID), nil)
	wNoSize := httptest.NewRecorder()
	server.router.ServeHTTP(wNoSize, reqNoSize)
	if wNoSize.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", wNoSize.Code)
	}
	if strings.Contains(wNoSize.Body.String(), "remote_file_size") {
		t.Errorf("expected remote_file_size to be omitted for an unchecked item, got body: %s", wNoSize.Body.String())
	}
	var respNoSize ItemResponse
	if err := json.Unmarshal(wNoSize.Body.Bytes(), &respNoSize); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if respNoSize.RemoteFileSize != nil {
		t.Errorf("expected remote_file_size=nil, got %v", *respNoSize.RemoteFileSize)
	}
}

// typedPaginatedResponse mirrors PaginatedResponse but decodes Data into typed ItemResponse values.
type typedPaginatedResponse struct {
	Data       []ItemResponse `json:"data"`
	Total      int64          `json:"total"`
	Limit      int            `json:"limit"`
	Offset     int            `json:"offset"`
	TotalPages int            `json:"total_pages"`
}

func listItemsSorted(t *testing.T, server *Server, query string) typedPaginatedResponse {
	t.Helper()
	req, _ := http.NewRequest("GET", "/api/v1/items"+query, nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for query %q, got %d: %s", query, w.Code, w.Body.String())
	}

	var resp typedPaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	return resp
}

func TestListItemsSorting(t *testing.T) {
	db := setupTestDB(t)

	movie := models.Movie{TMDBID: 111, TMDBTitle: "Zebra"}
	db.Create(&movie)

	tvshow := models.TVShow{TMDBID: 222, TMDBTitle: "Alpha"}
	db.Create(&tvshow)

	downloadedAt := time.Now()
	items := []models.ProcessedLine{
		{
			LineContent: "line-a", LineHash: "hash-a", TvgName: "Bravo", GroupTitle: "Zeta",
			ContentType: "movies", State: "downloaded", MovieID: &movie.ID,
			DownloadedAt: &downloadedAt, CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		{
			LineContent: "line-b", LineHash: "hash-b", TvgName: "Alfa", GroupTitle: "Alpha",
			ContentType: "tvshows", State: "processed", TVShowID: &tvshow.ID,
			CreatedAt: time.Now().Add(time.Second), UpdatedAt: time.Now().Add(time.Second),
		},
		{
			LineContent: "line-c", LineHash: "hash-c", TvgName: "Charlie", GroupTitle: "Beta",
			ContentType: "movies", State: "failed",
			CreatedAt: time.Now().Add(2 * time.Second), UpdatedAt: time.Now().Add(2 * time.Second),
		},
		{
			LineContent: "line-d", LineHash: "hash-d", TvgName: "Delta", GroupTitle: "Gamma",
			ContentType: "movies", State: "downloaded",
			CreatedAt: time.Now().Add(3 * time.Second), UpdatedAt: time.Now().Add(3 * time.Second),
		},
	}
	for i := range items {
		if err := db.Create(&items[i]).Error; err != nil {
			t.Fatalf("failed to seed item: %v", err)
		}
	}

	server := NewServer()

	// Default sort (no sort/order params) matches prior behavior: created_at desc.
	respDefault := listItemsSorted(t, server, "")
	if len(respDefault.Data) != 4 || respDefault.Data[0].LineHash != "hash-d" {
		t.Fatalf("expected default sort by created_at desc with hash-d first, got %+v", respDefault.Data)
	}

	// Sort by tvg_name ascending.
	respTvgName := listItemsSorted(t, server, "?sort=tvg_name&order=asc")
	if len(respTvgName.Data) != 4 || respTvgName.Data[0].LineHash != "hash-b" {
		t.Fatalf("expected tvg_name asc with hash-b first, got %+v", respTvgName.Data)
	}

	// Sort by group_title ascending.
	respGroup := listItemsSorted(t, server, "?sort=group_title&order=asc")
	if len(respGroup.Data) != 4 || respGroup.Data[0].LineHash != "hash-b" {
		t.Fatalf("expected group_title asc with hash-b first, got %+v", respGroup.Data)
	}

	// Sort by state ascending (alphabetical: downloaded < failed < processed).
	respState := listItemsSorted(t, server, "?sort=state&order=asc")
	if len(respState.Data) != 4 || respState.Data[0].State != "downloaded" || respState.Data[3].State != "processed" {
		t.Fatalf("expected state asc ordered downloaded..processed, got %+v", respState.Data)
	}

	// Sort by downloaded_at across both "downloaded" items: hash-a has a real
	// timestamp, hash-d (never re-processed since this field shipped) is NULL.
	// Regardless of direction, the NULL must sort last so it never buries a
	// genuinely recent download underneath legacy, unset rows.
	respDownloadedAtDesc := listItemsSorted(t, server, "?state=downloaded&sort=downloaded_at&order=desc")
	if len(respDownloadedAtDesc.Data) != 2 || respDownloadedAtDesc.Data[0].LineHash != "hash-a" || respDownloadedAtDesc.Data[1].LineHash != "hash-d" {
		t.Fatalf("expected downloaded_at desc order [hash-a, hash-d], got %+v", respDownloadedAtDesc.Data)
	}
	if respDownloadedAtDesc.Data[0].DownloadedAt == nil {
		t.Errorf("expected downloaded_at to be populated for hash-a")
	}
	if respDownloadedAtDesc.Data[1].DownloadedAt != nil {
		t.Errorf("expected downloaded_at to be null for hash-d")
	}

	respDownloadedAtAsc := listItemsSorted(t, server, "?state=downloaded&sort=downloaded_at&order=asc")
	if len(respDownloadedAtAsc.Data) != 2 || respDownloadedAtAsc.Data[0].LineHash != "hash-a" || respDownloadedAtAsc.Data[1].LineHash != "hash-d" {
		t.Fatalf("expected downloaded_at asc order [hash-a, hash-d] (NULL still last), got %+v", respDownloadedAtAsc.Data)
	}

	// tmdb_title ascending: enriched items ordered Alpha, Zebra, then the two
	// non-enriched items (hash-c, hash-d) last, in either relative order.
	respTmdbAsc := listItemsSorted(t, server, "?sort=tmdb_title&order=asc")
	if len(respTmdbAsc.Data) != 4 {
		t.Fatalf("expected 4 items, got %d", len(respTmdbAsc.Data))
	}
	if respTmdbAsc.Data[0].LineHash != "hash-b" || respTmdbAsc.Data[1].LineHash != "hash-a" {
		t.Fatalf("expected tmdb_title asc order to start with [hash-b, hash-a], got %+v", respTmdbAsc.Data)
	}
	if !isNonEnrichedPair(respTmdbAsc.Data[2].LineHash, respTmdbAsc.Data[3].LineHash) {
		t.Fatalf("expected non-enriched items [hash-c, hash-d] last, got %+v", respTmdbAsc.Data)
	}

	// tmdb_title descending: enriched items ordered Zebra, Alpha, then the two
	// non-enriched items still last.
	respTmdbDesc := listItemsSorted(t, server, "?sort=tmdb_title&order=desc")
	if len(respTmdbDesc.Data) != 4 {
		t.Fatalf("expected 4 items, got %d", len(respTmdbDesc.Data))
	}
	if respTmdbDesc.Data[0].LineHash != "hash-a" || respTmdbDesc.Data[1].LineHash != "hash-b" {
		t.Fatalf("expected tmdb_title desc order to start with [hash-a, hash-b], got %+v", respTmdbDesc.Data)
	}
	if !isNonEnrichedPair(respTmdbDesc.Data[2].LineHash, respTmdbDesc.Data[3].LineHash) {
		t.Fatalf("expected non-enriched items [hash-c, hash-d] last, got %+v", respTmdbDesc.Data)
	}

	// Unknown sort field is rejected.
	reqInvalidField, _ := http.NewRequest("GET", "/api/v1/items?sort=line_hash", nil)
	wInvalidField := httptest.NewRecorder()
	server.router.ServeHTTP(wInvalidField, reqInvalidField)
	if wInvalidField.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid sort field, got %d", wInvalidField.Code)
	}
	var errInvalidField ErrorResponse
	if err := json.Unmarshal(wInvalidField.Body.Bytes(), &errInvalidField); err != nil {
		t.Fatalf("failed to unmarshal error: %v", err)
	}
	if errInvalidField.Error != "invalid_sort_field" {
		t.Errorf("expected invalid_sort_field, got %q", errInvalidField.Error)
	}

	// Malformed / injection-style order value is rejected and never reaches the query.
	reqInvalidOrder, _ := http.NewRequest("GET", "/api/v1/items?sort=created_at&order="+url.QueryEscape("created_at;DROP TABLE processed_lines"), nil)
	wInvalidOrder := httptest.NewRecorder()
	server.router.ServeHTTP(wInvalidOrder, reqInvalidOrder)
	if wInvalidOrder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid sort order, got %d", wInvalidOrder.Code)
	}
	var errInvalidOrder ErrorResponse
	if err := json.Unmarshal(wInvalidOrder.Body.Bytes(), &errInvalidOrder); err != nil {
		t.Fatalf("failed to unmarshal error: %v", err)
	}
	if errInvalidOrder.Error != "invalid_sort_order" {
		t.Errorf("expected invalid_sort_order, got %q", errInvalidOrder.Error)
	}

	// Confirm the table survived the injection attempt.
	var total int64
	db.Model(&models.ProcessedLine{}).Count(&total)
	if total != 4 {
		t.Fatalf("expected processed_lines table to still have 4 rows, got %d", total)
	}
}

func isNonEnrichedPair(a, b string) bool {
	return (a == "hash-c" && b == "hash-d") || (a == "hash-d" && b == "hash-c")
}
