package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
				w.Write([]byte(`{"tvdb_id":12345}`))
			} else {
				w.Write([]byte(`{"id":101,"title":"Mock Inception","release_date":"2010-07-16","genres":[{"id":28,"name":"Action"}],"runtime":148}`))
			}
		} else if strings.HasPrefix(r.URL.Path, "/tv/202") {
			if strings.HasSuffix(r.URL.Path, "/external_ids") {
				w.Write([]byte(`{"tvdb_id":67890}`))
			} else {
				w.Write([]byte(`{"id":202,"name":"Mock Malcolm","first_air_date":"2000-01-09","genres":[{"id":35,"name":"Comedy"}]}`))
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
