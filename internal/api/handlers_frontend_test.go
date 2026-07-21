package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
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
