package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/glefebvre/stalkeer/internal/models"
	"gorm.io/gorm"
)

// --- 1.1 extractDownloadRoot ---

func TestExtractDownloadRoot_Movie(t *testing.T) {
	path := filepath.Join("/media/movies", "Interstellar (2014)", "Interstellar (2014).mkv")

	root, ok := extractDownloadRoot(path, models.ContentTypeMovies)
	if !ok {
		t.Fatalf("expected extraction to succeed for a movie path")
	}
	if root.Root != filepath.Join("/media/movies", "Interstellar (2014)") {
		t.Errorf("unexpected root: %s", root.Root)
	}
	if root.SubPath != "Interstellar (2014).mkv" {
		t.Errorf("unexpected sub-path: %s", root.SubPath)
	}
}

func TestExtractDownloadRoot_TVEpisode(t *testing.T) {
	path := filepath.Join("/media/tvshows", "Series Name (2020)", "Season 01", "Series Name (2020) - S01E01.mkv")

	root, ok := extractDownloadRoot(path, models.ContentTypeTVShows)
	if !ok {
		t.Fatalf("expected extraction to succeed for a TV episode path")
	}
	if root.Root != filepath.Join("/media/tvshows", "Series Name (2020)") {
		t.Errorf("unexpected root: %s", root.Root)
	}
	if root.SubPath != filepath.Join("Season 01", "Series Name (2020) - S01E01.mkv") {
		t.Errorf("unexpected sub-path: %s", root.SubPath)
	}
}

func TestExtractDownloadRoot_TVEpisode_UnrecognizedShape(t *testing.T) {
	path := filepath.Join("/media/tvshows", "Series Name (2020)", "Series Name (2020) - S01E01.mkv")

	if _, ok := extractDownloadRoot(path, models.ContentTypeTVShows); ok {
		t.Fatalf("expected extraction to fail when parent dir isn't a Season NN directory")
	}
}

// --- 1.2 computeReconciledPath ---

func TestComputeReconciledPath_SameRootIsNoOp(t *testing.T) {
	root := downloadRoot{Root: "/media/movies/Dune (2021)", SubPath: "Dune (2021).mkv"}

	if _, changed := computeReconciledPath(root, "/media/movies/Dune (2021)"); changed {
		t.Errorf("expected no change when root already matches")
	}
}

func TestComputeReconciledPath_RenamedRootSameParent(t *testing.T) {
	root := downloadRoot{Root: "/media/movies/Dune (2021)", SubPath: "Dune (2021).mkv"}

	newPath, changed := computeReconciledPath(root, "/media/movies/Dune Part Two (2021)")
	if !changed {
		t.Fatalf("expected a change when root basename differs")
	}
	want := filepath.Join("/media/movies/Dune Part Two (2021)", "Dune (2021).mkv")
	if newPath != want {
		t.Errorf("got %q, want %q", newPath, want)
	}
}

func TestComputeReconciledPath_RootMovedToDifferentParent(t *testing.T) {
	root := downloadRoot{Root: "/media/movies/Dune (2021)", SubPath: "Dune (2021).mkv"}

	newPath, changed := computeReconciledPath(root, "/media/movies-4k/Dune (2021)")
	if !changed {
		t.Fatalf("expected a change when parent directory differs")
	}
	want := filepath.Join("/media/movies-4k/Dune (2021)", "Dune (2021).mkv")
	if newPath != want {
		t.Errorf("got %q, want %q", newPath, want)
	}
}

func TestComputeReconciledPath_TargetRootTrailingSeparatorIsNoOp(t *testing.T) {
	root := downloadRoot{Root: "/media/movies/Dune (2021)", SubPath: "Dune (2021).mkv"}

	if _, changed := computeReconciledPath(root, "/media/movies/Dune (2021)/"); changed {
		t.Errorf("expected no change when targetRoot only differs from root by a trailing separator")
	}
}

func TestComputeReconciledPath_TrailingSeparatorDoesNotMaskRealDrift(t *testing.T) {
	root := downloadRoot{Root: "/media/movies/Dune (2021)", SubPath: "Dune (2021).mkv"}

	cases := []struct {
		name       string
		targetRoot string
	}{
		{"without trailing separator", "/media/movies/Dune Part Two (2021)"},
		{"with trailing separator", "/media/movies/Dune Part Two (2021)/"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			newPath, changed := computeReconciledPath(root, tc.targetRoot)
			if !changed {
				t.Fatalf("expected a change for a genuinely different root")
			}
			want := filepath.Join("/media/movies/Dune Part Two (2021)", "Dune (2021).mkv")
			if newPath != want {
				t.Errorf("got %q, want %q", newPath, want)
			}
		})
	}
}

// --- 1.3 applyPathCorrection ---

func TestApplyPathCorrection_MovesFileAndUpdatesDB(t *testing.T) {
	db := setupTestDB(t)
	tempDir := t.TempDir()

	oldPath := filepath.Join(tempDir, "movies", "Dune (2021)", "Dune (2021).mkv")
	if err := os.MkdirAll(filepath.Dir(oldPath), 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.WriteFile(oldPath, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	dl := models.DownloadInfo{URL: "http://example.com/dune", Status: "completed", DownloadPath: &oldPath}
	if err := db.Create(&dl).Error; err != nil {
		t.Fatalf("failed to seed download: %v", err)
	}

	newPath := filepath.Join(tempDir, "movies", "Dune Part Two (2021)", "Dune (2021).mkv")

	outcome := applyPathCorrection(db, dl.ID, oldPath, newPath)
	if outcome != OutcomeCorrected {
		t.Fatalf("expected outcome %q, got %q", OutcomeCorrected, outcome)
	}

	if _, err := os.Stat(newPath); err != nil {
		t.Errorf("expected file to exist at new path: %v", err)
	}
	if _, err := os.Stat(oldPath); err == nil {
		t.Errorf("expected old file to be gone")
	}

	var updated models.DownloadInfo
	db.First(&updated, dl.ID)
	if updated.DownloadPath == nil || *updated.DownloadPath != newPath {
		t.Errorf("expected download_path %q, got %v", newPath, updated.DownloadPath)
	}
}

// TestApplyPathCorrection_WholeFolderAlreadyRenamedOnDisk covers the exact
// scenario the feature exists for: Sonarr/Radarr (or a manual mv) renamed the
// whole series/movie folder as one filesystem operation, so oldPath no
// longer exists and the file is already sitting at newPath. This must be
// treated as "already relocated, just catch the DB up", not as a collision
// that permanently blocks reconciliation (found via manual verification:
// without this, every subsequent run re-logs rename_target_exists forever).
func TestApplyPathCorrection_WholeFolderAlreadyRenamedOnDisk(t *testing.T) {
	db := setupTestDB(t)
	tempDir := t.TempDir()

	oldPath := filepath.Join(tempDir, "tvshows", "Malcolm in the Middle", "Season 01", "Malcolm in the Middle - S01E01.mkv")
	newPath := filepath.Join(tempDir, "tvshows", "Malcolm in the Middle (2000)", "Season 01", "Malcolm in the Middle - S01E01.mkv")

	// Simulate the whole folder having already been renamed externally: only
	// the new path exists on disk, the old one is gone.
	if err := os.MkdirAll(filepath.Dir(newPath), 0755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}
	if err := os.WriteFile(newPath, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write file at new path: %v", err)
	}

	dl := models.DownloadInfo{URL: "http://example.com/malcolm", Status: "completed", DownloadPath: &oldPath}
	if err := db.Create(&dl).Error; err != nil {
		t.Fatalf("failed to seed download: %v", err)
	}

	outcome := applyPathCorrection(db, dl.ID, oldPath, newPath)
	if outcome != OutcomeCorrected {
		t.Fatalf("expected outcome %q, got %q", OutcomeCorrected, outcome)
	}

	var updated models.DownloadInfo
	db.First(&updated, dl.ID)
	if updated.DownloadPath == nil || *updated.DownloadPath != newPath {
		t.Errorf("expected download_path %q, got %v", newPath, updated.DownloadPath)
	}

	// The file at newPath must be left untouched (no move was needed/attempted).
	content, err := os.ReadFile(newPath)
	if err != nil || string(content) != "content" {
		t.Errorf("expected file at new path to remain untouched")
	}
}

func TestApplyPathCorrection_NeitherOldNorNewPathExists(t *testing.T) {
	db := setupTestDB(t)
	tempDir := t.TempDir()

	oldPath := filepath.Join(tempDir, "tvshows", "Gone Show", "Season 01", "Gone Show - S01E01.mkv")
	newPath := filepath.Join(tempDir, "tvshows", "Gone Show (2020)", "Season 01", "Gone Show - S01E01.mkv")

	dl := models.DownloadInfo{URL: "http://example.com/gone", Status: "completed", DownloadPath: &oldPath}
	if err := db.Create(&dl).Error; err != nil {
		t.Fatalf("failed to seed download: %v", err)
	}

	outcome := applyPathCorrection(db, dl.ID, oldPath, newPath)
	if outcome != OutcomeRenameFailed {
		t.Fatalf("expected outcome %q, got %q", OutcomeRenameFailed, outcome)
	}

	var unchanged models.DownloadInfo
	db.First(&unchanged, dl.ID)
	if unchanged.DownloadPath == nil || *unchanged.DownloadPath != oldPath {
		t.Errorf("expected download_path to remain %q, got %v", oldPath, unchanged.DownloadPath)
	}
}

func TestApplyPathCorrection_CollisionBlocksMoveAndDB(t *testing.T) {
	db := setupTestDB(t)
	tempDir := t.TempDir()

	oldPath := filepath.Join(tempDir, "movies", "Dune (2021)", "Dune (2021).mkv")
	if err := os.MkdirAll(filepath.Dir(oldPath), 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.WriteFile(oldPath, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	dl := models.DownloadInfo{URL: "http://example.com/dune", Status: "completed", DownloadPath: &oldPath}
	if err := db.Create(&dl).Error; err != nil {
		t.Fatalf("failed to seed download: %v", err)
	}

	newPath := filepath.Join(tempDir, "movies", "Dune Part Two (2021)", "Dune (2021).mkv")
	if err := os.MkdirAll(filepath.Dir(newPath), 0755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}
	if err := os.WriteFile(newPath, []byte("existing"), 0644); err != nil {
		t.Fatalf("failed to write colliding file: %v", err)
	}

	outcome := applyPathCorrection(db, dl.ID, oldPath, newPath)
	if outcome != OutcomeRenameTargetExists {
		t.Fatalf("expected outcome %q, got %q", OutcomeRenameTargetExists, outcome)
	}

	if _, err := os.Stat(oldPath); err != nil {
		t.Errorf("expected source file to remain untouched: %v", err)
	}
	existing, err := os.ReadFile(newPath)
	if err != nil || string(existing) != "existing" {
		t.Errorf("expected colliding destination file to remain untouched")
	}

	var unchanged models.DownloadInfo
	db.First(&unchanged, dl.ID)
	if unchanged.DownloadPath == nil || *unchanged.DownloadPath != oldPath {
		t.Errorf("expected download_path to remain %q, got %v", oldPath, unchanged.DownloadPath)
	}
}

// --- 2.2 ReconcileScheduledDownloadPaths ---

func TestReconcileScheduledDownloadPaths_CorrectsMatchedRow(t *testing.T) {
	db := setupTestDB(t)
	tempDir := t.TempDir()

	movie := models.Movie{TMDBID: 42, TMDBTitle: "Dune", TMDBYear: 2021}
	db.Create(&movie)

	oldPath := filepath.Join(tempDir, "movies", "Dune (2021)", "Dune (2021).mkv")
	if err := os.MkdirAll(filepath.Dir(oldPath), 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.WriteFile(oldPath, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	dl := models.DownloadInfo{URL: "http://example.com/dune", Status: "completed", DownloadPath: &oldPath}
	db.Create(&dl)
	line := models.ProcessedLine{
		LineContent: "dune", LineHash: "hash-dune", TvgName: "Dune",
		ContentType: models.ContentTypeMovies, MovieID: &movie.ID, DownloadInfoID: &dl.ID,
		State: models.StateDownloaded,
	}
	db.Create(&line)

	// Untouched: no matching entry in the fake Radarr library fixture.
	unmatchedMovie := models.Movie{TMDBID: 99, TMDBTitle: "Unmatched", TMDBYear: 2020}
	db.Create(&unmatchedMovie)
	unmatchedPath := filepath.Join(tempDir, "movies", "Unmatched (2020)", "Unmatched (2020).mkv")
	if err := os.MkdirAll(filepath.Dir(unmatchedPath), 0755); err != nil {
		t.Fatalf("failed to create unmatched dir: %v", err)
	}
	if err := os.WriteFile(unmatchedPath, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write unmatched file: %v", err)
	}
	unmatchedDl := models.DownloadInfo{URL: "http://example.com/unmatched", Status: "completed", DownloadPath: &unmatchedPath}
	db.Create(&unmatchedDl)
	unmatchedLine := models.ProcessedLine{
		LineContent: "unmatched", LineHash: "hash-unmatched", TvgName: "Unmatched",
		ContentType: models.ContentTypeMovies, MovieID: &unmatchedMovie.ID, DownloadInfoID: &unmatchedDl.ID,
		State: models.StateDownloaded,
	}
	db.Create(&unmatchedLine)

	newRoot := filepath.Join(tempDir, "movies", "Dune Part Two (2021)")
	ReconcileScheduledDownloadPaths(db, map[int]string{42: newRoot}, map[int]string{})

	var updated models.DownloadInfo
	db.First(&updated, dl.ID)
	expectedNewPath := filepath.Join(newRoot, "Dune (2021).mkv")
	if updated.DownloadPath == nil || *updated.DownloadPath != expectedNewPath {
		t.Errorf("expected download_path %q, got %v", expectedNewPath, updated.DownloadPath)
	}

	var untouched models.DownloadInfo
	db.First(&untouched, unmatchedDl.ID)
	if untouched.DownloadPath == nil || *untouched.DownloadPath != unmatchedPath {
		t.Errorf("expected unmatched download_path to remain %q, got %v", unmatchedPath, untouched.DownloadPath)
	}
}

func TestReconcileScheduledDownloadPaths_TrailingSeparatorNoOp(t *testing.T) {
	db := setupTestDB(t)
	tempDir := t.TempDir()

	movie := models.Movie{TMDBID: 42, TMDBTitle: "Dune", TMDBYear: 2021}
	db.Create(&movie)

	root := filepath.Join(tempDir, "movies", "Dune (2021)")
	oldPath := filepath.Join(root, "Dune (2021).mkv")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.WriteFile(oldPath, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	dl := models.DownloadInfo{URL: "http://example.com/dune", Status: "completed", DownloadPath: &oldPath}
	db.Create(&dl)
	line := models.ProcessedLine{
		LineContent: "dune", LineHash: "hash-dune", TvgName: "Dune",
		ContentType: models.ContentTypeMovies, MovieID: &movie.ID, DownloadInfoID: &dl.ID,
		State: models.StateDownloaded,
	}
	db.Create(&line)

	// Fake Radarr/Sonarr library fixture reports the same root with a
	// trailing separator: should be a no-op, not a "drift" correction.
	ReconcileScheduledDownloadPaths(db, map[int]string{42: root + string(filepath.Separator)}, map[int]string{})

	var unchanged models.DownloadInfo
	db.First(&unchanged, dl.ID)
	if unchanged.DownloadPath == nil || *unchanged.DownloadPath != oldPath {
		t.Errorf("expected download_path to remain %q, got %v", oldPath, unchanged.DownloadPath)
	}
	if _, err := os.Stat(oldPath); err != nil {
		t.Errorf("expected file to remain at original path untouched: %v", err)
	}
}

// --- 3.x resyncDownloadPath handler ---

func seedResyncMovieDownload(t *testing.T, db *gorm.DB, tmdbID int, path string) models.DownloadInfo {
	t.Helper()

	movie := models.Movie{TMDBID: tmdbID, TMDBTitle: "Dune", TMDBYear: 2021}
	if err := db.Create(&movie).Error; err != nil {
		t.Fatalf("failed to create movie: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	dl := models.DownloadInfo{URL: "http://example.com/dune", Status: "completed", DownloadPath: &path}
	if err := db.Create(&dl).Error; err != nil {
		t.Fatalf("failed to create download: %v", err)
	}

	line := models.ProcessedLine{
		LineContent: "dune", LineHash: fmt.Sprintf("hash-%d", dl.ID), TvgName: "Dune",
		ContentType: models.ContentTypeMovies, MovieID: &movie.ID, DownloadInfoID: &dl.ID,
		State: models.StateDownloaded,
	}
	if err := db.Create(&line).Error; err != nil {
		t.Fatalf("failed to create processed line: %v", err)
	}

	return dl
}

func TestResyncDownloadPath_Corrected(t *testing.T) {
	db := setupTestDB(t)
	tempDir := t.TempDir()

	oldPath := filepath.Join(tempDir, "movies", "Dune (2021)", "Dune (2021).mkv")
	dl := seedResyncMovieDownload(t, db, 42, oldPath)

	newRoot := filepath.Join(tempDir, "movies", "Dune Part Two (2021)")
	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{{"id": 1, "tmdbId": 42, "path": newRoot}})
	}))
	defer radarrServer.Close()
	newForceDownloadTestConfig(t, radarrServer.URL, "")

	server := NewServer()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/downloads/%d/resync-path", dl.ID), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["status"] != string(OutcomeCorrected) {
		t.Errorf("expected status %q, got %v", OutcomeCorrected, resp["status"])
	}

	expectedNewPath := filepath.Join(newRoot, "Dune (2021).mkv")
	if resp["new_path"] != expectedNewPath {
		t.Errorf("expected new_path %q, got %v", expectedNewPath, resp["new_path"])
	}

	var updated models.DownloadInfo
	db.First(&updated, dl.ID)
	if updated.DownloadPath == nil || *updated.DownloadPath != expectedNewPath {
		t.Errorf("expected download_path %q, got %v", expectedNewPath, updated.DownloadPath)
	}
}

func TestResyncDownloadPath_AlreadyUpToDate(t *testing.T) {
	db := setupTestDB(t)
	tempDir := t.TempDir()

	root := filepath.Join(tempDir, "movies", "Dune (2021)")
	oldPath := filepath.Join(root, "Dune (2021).mkv")
	dl := seedResyncMovieDownload(t, db, 42, oldPath)

	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{{"id": 1, "tmdbId": 42, "path": root}})
	}))
	defer radarrServer.Close()
	newForceDownloadTestConfig(t, radarrServer.URL, "")

	server := NewServer()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/downloads/%d/resync-path", dl.ID), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != string(OutcomeAlreadyUpToDate) {
		t.Errorf("expected status %q, got %v", OutcomeAlreadyUpToDate, resp["status"])
	}

	var unchanged models.DownloadInfo
	db.First(&unchanged, dl.ID)
	if unchanged.DownloadPath == nil || *unchanged.DownloadPath != oldPath {
		t.Errorf("expected download_path to remain %q, got %v", oldPath, unchanged.DownloadPath)
	}
}

func TestResyncDownloadPath_NotManagedByRadarrSonarr(t *testing.T) {
	db := setupTestDB(t)
	tempDir := t.TempDir()

	oldPath := filepath.Join(tempDir, "movies", "Dune (2021)", "Dune (2021).mkv")
	dl := seedResyncMovieDownload(t, db, 42, oldPath)

	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{})
	}))
	defer radarrServer.Close()
	newForceDownloadTestConfig(t, radarrServer.URL, "")

	server := NewServer()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/downloads/%d/resync-path", dl.ID), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != string(OutcomeNotManaged) {
		t.Errorf("expected status %q, got %v", OutcomeNotManaged, resp["status"])
	}

	var unchanged models.DownloadInfo
	db.First(&unchanged, dl.ID)
	if unchanged.DownloadPath == nil || *unchanged.DownloadPath != oldPath {
		t.Errorf("expected download_path to remain %q, got %v", oldPath, unchanged.DownloadPath)
	}
}

func TestResyncDownloadPath_IncompleteDownloadRejected(t *testing.T) {
	db := setupTestDB(t)
	newForceDownloadTestConfig(t, "", "")

	dl := models.DownloadInfo{URL: "http://example.com/pending", Status: "downloading"}
	db.Create(&dl)

	server := NewServer()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/downloads/%d/resync-path", dl.ID), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestResyncDownloadPath_CollisionBlocked(t *testing.T) {
	db := setupTestDB(t)
	tempDir := t.TempDir()

	oldPath := filepath.Join(tempDir, "movies", "Dune (2021)", "Dune (2021).mkv")
	dl := seedResyncMovieDownload(t, db, 42, oldPath)

	newRoot := filepath.Join(tempDir, "movies", "Dune Part Two (2021)")
	collidingPath := filepath.Join(newRoot, "Dune (2021).mkv")
	if err := os.MkdirAll(newRoot, 0755); err != nil {
		t.Fatalf("failed to create colliding dir: %v", err)
	}
	if err := os.WriteFile(collidingPath, []byte("existing"), 0644); err != nil {
		t.Fatalf("failed to write colliding file: %v", err)
	}

	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{{"id": 1, "tmdbId": 42, "path": newRoot}})
	}))
	defer radarrServer.Close()
	newForceDownloadTestConfig(t, radarrServer.URL, "")

	server := NewServer()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/downloads/%d/resync-path", dl.ID), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}

	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Error != string(OutcomeRenameTargetExists) {
		t.Errorf("expected error code %q, got %s", OutcomeRenameTargetExists, resp.Error)
	}

	var unchanged models.DownloadInfo
	db.First(&unchanged, dl.ID)
	if unchanged.DownloadPath == nil || *unchanged.DownloadPath != oldPath {
		t.Errorf("expected download_path to remain %q, got %v", oldPath, unchanged.DownloadPath)
	}
}

func TestResyncDownloadPath_NotFound(t *testing.T) {
	setupTestDB(t)
	newForceDownloadTestConfig(t, "", "")
	server := NewServer()

	req, _ := http.NewRequest("POST", "/api/v1/downloads/999/resync-path", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d: %s", w.Code, w.Body.String())
	}
}
