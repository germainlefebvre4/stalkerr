package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/models"
	"gorm.io/gorm"
)

type keepOneSeed struct {
	db       *gorm.DB
	movie    models.Movie
	dlIDs    []uint
	paths    []string
	lineHash int
}

// seedDuplicateMovie creates a movie with one completed occurrence per path
// (no file is created; callers write the files they want present). Later
// paths complete later.
func seedDuplicateMovie(t *testing.T, db *gorm.DB, tmdbID int, paths ...string) ([]uint, models.Movie) {
	t.Helper()
	movie := models.Movie{TMDBID: tmdbID, TMDBTitle: fmt.Sprintf("Star Wars %d", tmdbID), TMDBYear: 1977}
	if err := db.Create(&movie).Error; err != nil {
		t.Fatal(err)
	}
	base := time.Now().Add(-time.Hour)
	var ids []uint
	for i, p := range paths {
		p := p
		done := base.Add(time.Duration(i) * time.Minute)
		dl := models.DownloadInfo{URL: fmt.Sprintf("http://example.com/%d-%d", tmdbID, i), Status: string(models.DownloadStatusCompleted), DownloadPath: &p, CompletedAt: &done}
		if err := db.Create(&dl).Error; err != nil {
			t.Fatal(err)
		}
		line := models.ProcessedLine{
			LineContent: "x", LineHash: fmt.Sprintf("hash-%d-%d", tmdbID, i), TvgName: "Star Wars",
			ContentType: models.ContentTypeMovies, MovieID: &movie.ID, DownloadInfoID: &dl.ID,
			State: models.StateDownloaded, ProcessedAt: time.Now(),
		}
		if err := db.Create(&line).Error; err != nil {
			t.Fatal(err)
		}
		ids = append(ids, dl.ID)
	}
	return ids, movie
}

func writeMedia(t *testing.T, path, content string) string {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func doReq(server *Server, method, url string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, url, nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)
	return w
}

func downloadStatus(db *gorm.DB, id uint) string {
	var dl models.DownloadInfo
	db.First(&dl, id)
	return dl.Status
}

func decodeCandidates(t *testing.T, w *httptest.ResponseRecorder) []DuplicateCandidate {
	t.Helper()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp DuplicateCandidatesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	return resp.Candidates
}

// radarrFake serves the movie lookup and counts command requests.
func radarrFake(t *testing.T, title string, year int, commandStatus int, commands *int32) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v3/command" {
			atomic.AddInt32(commands, 1)
			w.WriteHeader(commandStatus)
			w.Write([]byte(`{}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{{"id": 7, "tmdbId": 11, "title": title, "year": year}})
	}))
	t.Cleanup(srv.Close)
	return srv
}

// ---- candidates (2.1-2.4) ----

func TestDuplicateCandidates_NotADuplicate(t *testing.T) {
	db := setupTestDB(t)
	newForceDownloadTestConfig(t, "", "")
	p := writeMedia(t, filepath.Join(t.TempDir(), "a.mkv"), "a")
	ids, _ := seedDuplicateMovie(t, db, 11, p)
	w := doReq(NewServer(), "GET", fmt.Sprintf("/api/v1/downloads/%d/duplicate-candidates", ids[0]))
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestDuplicateCandidates_TwoFilesSharedPathAndMissing(t *testing.T) {
	db := setupTestDB(t)
	newForceDownloadTestConfig(t, "", "")
	dir := t.TempDir()
	a := writeMedia(t, filepath.Join(dir, "a.mkv"), "aaaa")
	b := writeMedia(t, filepath.Join(dir, "b.mp4"), "bb")
	missing := filepath.Join(dir, "gone.mkv")
	// b is shared by two occurrences; missing is the most recent.
	ids, _ := seedDuplicateMovie(t, db, 11, a, b, b, missing)

	cands := decodeCandidates(t, doReq(NewServer(), "GET", fmt.Sprintf("/api/v1/downloads/%d/duplicate-candidates", ids[0])))
	if len(cands) != 3 {
		t.Fatalf("expected 3 distinct candidates, got %d", len(cands))
	}
	byPath := map[string]DuplicateCandidate{}
	for _, c := range cands {
		byPath[c.DownloadPath] = c
	}
	if got := byPath[b]; len(got.DownloadInfoIDs) != 2 || got.SizeBytes == nil || *got.SizeBytes != 2 || !got.Exists {
		t.Errorf("unexpected shared-path candidate: %+v", got)
	}
	if got := byPath[missing]; got.Exists || got.SizeBytes != nil || got.Suggested {
		t.Errorf("missing file must not exist/size/suggested: %+v", got)
	}
	// Library unavailable: latest existing completion (b) is suggested, names omitted.
	if !byPath[b].Suggested || byPath[a].Suggested {
		t.Errorf("expected b suggested only: a=%+v b=%+v", byPath[a], byPath[b])
	}
	if byPath[a].MatchesLibraryName != nil {
		t.Error("matches_library_name must be omitted when lookup unavailable")
	}
}

func TestDuplicateCandidates_CanonicalNameWins(t *testing.T) {
	db := setupTestDB(t)
	dir := t.TempDir()
	canonical := writeMedia(t, filepath.Join(dir, "Star Wars (1977).mkv"), "x")
	french := writeMedia(t, filepath.Join(dir, "La Guerre des étoiles (1977).mp4"), "y") // more recent
	ids, _ := seedDuplicateMovie(t, db, 11, canonical, french)
	var cmds int32
	srv := radarrFake(t, "Star Wars", 1977, http.StatusCreated, &cmds)
	newForceDownloadTestConfig(t, srv.URL, "")

	cands := decodeCandidates(t, doReq(NewServer(), "GET", fmt.Sprintf("/api/v1/downloads/%d/duplicate-candidates", ids[0])))
	for _, c := range cands {
		if c.MatchesLibraryName == nil {
			t.Fatalf("expected matches_library_name set: %+v", c)
		}
		if (c.DownloadPath == canonical) != *c.MatchesLibraryName || (c.DownloadPath == canonical) != c.Suggested {
			t.Errorf("unexpected flags for %s: %+v", c.DownloadPath, c)
		}
	}
	if cmds != 0 {
		t.Error("candidates must not issue any command")
	}
}

func TestDuplicateCandidates_LookupErrorStill200AndNoSuggestionWhenNoneExist(t *testing.T) {
	db := setupTestDB(t)
	dir := t.TempDir()
	ids, _ := seedDuplicateMovie(t, db, 11, filepath.Join(dir, "a.mkv"), filepath.Join(dir, "b.mkv"))
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }))
	defer bad.Close()
	newForceDownloadTestConfig(t, bad.URL, "")

	cands := decodeCandidates(t, doReq(NewServer(), "GET", fmt.Sprintf("/api/v1/downloads/%d/duplicate-candidates", ids[0])))
	for _, c := range cands {
		if c.Suggested || c.MatchesLibraryName != nil {
			t.Errorf("unexpected flags: %+v", c)
		}
	}
}

func TestDuplicateCandidates_ItemNotFoundInLibrary(t *testing.T) {
	db := setupTestDB(t)
	dir := t.TempDir()
	a := writeMedia(t, filepath.Join(dir, "a.mkv"), "a")
	b := writeMedia(t, filepath.Join(dir, "b.mkv"), "b")
	ids, _ := seedDuplicateMovie(t, db, 11, a, b)
	empty := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`[]`)) }))
	defer empty.Close()
	newForceDownloadTestConfig(t, empty.URL, "")

	cands := decodeCandidates(t, doReq(NewServer(), "GET", fmt.Sprintf("/api/v1/downloads/%d/duplicate-candidates", ids[0])))
	if cands[0].MatchesLibraryName != nil || !cands[1].Suggested {
		t.Errorf("unexpected: %+v", cands)
	}
}

func TestDuplicateCandidates_ReadOnly(t *testing.T) {
	db := setupTestDB(t)
	newForceDownloadTestConfig(t, "", "")
	dir := t.TempDir()
	a := writeMedia(t, filepath.Join(dir, "a.mkv"), "a")
	b := writeMedia(t, filepath.Join(dir, "b.mkv"), "b")
	ids, _ := seedDuplicateMovie(t, db, 11, a, b)
	doReq(NewServer(), "GET", fmt.Sprintf("/api/v1/downloads/%d/duplicate-candidates", ids[0]))
	for _, id := range ids {
		if downloadStatus(db, id) != string(models.DownloadStatusCompleted) {
			t.Error("candidates must not modify records")
		}
	}
	for _, p := range []string{a, b} {
		if _, err := os.Stat(p); err != nil {
			t.Error("candidates must not modify files")
		}
	}
}

func TestMarkSuggestedCandidate_Precedence(t *testing.T) {
	now := time.Now()
	old := now.Add(-time.Hour)
	yes, no := true, false
	cands := []DuplicateCandidate{
		{DownloadPath: "new-missing", Exists: false, CompletedAt: &now},
		{DownloadPath: "old-exists", Exists: true, CompletedAt: &old},
	}
	markSuggestedCandidate(cands)
	if cands[0].Suggested || !cands[1].Suggested {
		t.Errorf("most recent missing, older existing: %+v", cands)
	}
	cands = []DuplicateCandidate{
		{DownloadPath: "match-old", Exists: true, CompletedAt: &old, MatchesLibraryName: &yes},
		{DownloadPath: "nomatch-new", Exists: true, CompletedAt: &now, MatchesLibraryName: &no},
	}
	markSuggestedCandidate(cands)
	if !cands[0].Suggested || cands[1].Suggested {
		t.Errorf("canonical match must win: %+v", cands)
	}
	cands = []DuplicateCandidate{{DownloadPath: "x", Exists: false}}
	markSuggestedCandidate(cands)
	if cands[0].Suggested {
		t.Error("nothing exists: nothing suggested")
	}
}

// ---- keep-one (3.x, 4.x) ----

func keepURL(id uint) string { return fmt.Sprintf("/api/v1/downloads/%d/keep-duplicate", id) }

func decodeKeep(t *testing.T, w *httptest.ResponseRecorder) KeepDuplicateResponse {
	t.Helper()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp KeepDuplicateResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	return resp
}

func TestKeepDuplicate_NotADuplicateAndInvalidID(t *testing.T) {
	db := setupTestDB(t)
	newForceDownloadTestConfig(t, "", "")
	p := writeMedia(t, filepath.Join(t.TempDir(), "a.mkv"), "a")
	ids, _ := seedDuplicateMovie(t, db, 11, p)
	if w := doReq(NewServer(), "POST", keepURL(ids[0])); w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
	if w := doReq(NewServer(), "POST", "/api/v1/downloads/abc/keep-duplicate"); w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	if _, err := os.Stat(p); err != nil {
		t.Error("file must be untouched")
	}
}

func TestKeepDuplicate_KeepsOneDeletesOther_SharedPaths_AndRefreshOnce(t *testing.T) {
	db := setupTestDB(t)
	dir := t.TempDir()
	a := writeMedia(t, filepath.Join(dir, "Star Wars (1977).mkv"), "a")
	b := writeMedia(t, filepath.Join(dir, "La Guerre.mp4"), "b")
	// unrelated completed download with its own file
	other := writeMedia(t, filepath.Join(dir, "unrelated.mkv"), "u")
	otherIDs, _ := seedDuplicateMovie(t, db, 99, other)
	ids, _ := seedDuplicateMovie(t, db, 11, a, a, b, b) // 4 occurrences, 2 paths

	var cmds int32
	srv := radarrFake(t, "Star Wars", 1977, http.StatusCreated, &cmds)
	newForceDownloadTestConfig(t, srv.URL, "")

	resp := decodeKeep(t, doReq(NewServer(), "POST", keepURL(ids[0])))
	if len(resp.Occurrences) != 4 || resp.Refresh != refreshRequested || cmds != 1 {
		t.Fatalf("unexpected response %+v (commands=%d)", resp, cmds)
	}
	want := map[uint]string{ids[0]: "kept", ids[1]: "kept", ids[2]: "deleted", ids[3]: "deleted"}
	for _, o := range resp.Occurrences {
		if want[o.DownloadInfoID] != o.Outcome || o.DownloadPath == "" {
			t.Errorf("unexpected occurrence %+v", o)
		}
	}
	if _, err := os.Stat(a); err != nil {
		t.Error("kept file must survive")
	}
	if _, err := os.Stat(b); !os.IsNotExist(err) {
		t.Error("other file must be deleted")
	}
	if _, err := os.Stat(other); err != nil || downloadStatus(db, otherIDs[0]) != string(models.DownloadStatusCompleted) {
		t.Error("unrelated download must be intact")
	}
	for id, o := range want {
		st := downloadStatus(db, id)
		if (o == "kept") != (st == string(models.DownloadStatusCompleted)) {
			t.Errorf("download %d outcome %s but status %s", id, o, st)
		}
	}
	var line models.ProcessedLine
	db.Where("download_info_id = ?", ids[2]).First(&line)
	if line.State != models.StateCancelled {
		t.Errorf("processed line must be cancelled, got %s", line.State)
	}
	// group no longer reported
	if w := doReq(NewServer(), "POST", keepURL(ids[0])); w.Code != http.StatusNotFound {
		t.Errorf("group should be resolved, got %d", w.Code)
	}
}

func TestKeepDuplicate_AlreadyMissingPathCancelledWithoutDeletion(t *testing.T) {
	db := setupTestDB(t)
	newForceDownloadTestConfig(t, "", "")
	dir := t.TempDir()
	a := writeMedia(t, filepath.Join(dir, "a.mkv"), "a")
	b := writeMedia(t, filepath.Join(dir, "b.mkv"), "b")
	gone := filepath.Join(dir, "gone.mkv")
	ids, _ := seedDuplicateMovie(t, db, 11, a, b, gone)

	resp := decodeKeep(t, doReq(NewServer(), "POST", keepURL(ids[0])))
	got := map[uint]string{}
	for _, o := range resp.Occurrences {
		got[o.DownloadInfoID] = o.Outcome
	}
	if got[ids[0]] != "kept" || got[ids[1]] != "deleted" || got[ids[2]] != "cancelled_file_missing" {
		t.Errorf("unexpected outcomes %v", got)
	}
	if resp.Refresh != refreshNotConfigured {
		t.Errorf("expected not_configured, got %s", resp.Refresh)
	}
}

func TestKeepDuplicate_KeptFileMissingRefusedNothingChanged(t *testing.T) {
	db := setupTestDB(t)
	var cmds int32
	srv := radarrFake(t, "Star Wars", 1977, http.StatusCreated, &cmds)
	newForceDownloadTestConfig(t, srv.URL, "")
	dir := t.TempDir()
	a := filepath.Join(dir, "ghost.mkv") // missing
	b := writeMedia(t, filepath.Join(dir, "b.mkv"), "b")
	ids, _ := seedDuplicateMovie(t, db, 11, a, b)

	if w := doReq(NewServer(), "POST", keepURL(ids[0])); w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
	if _, err := os.Stat(b); err != nil {
		t.Error("nothing may be deleted")
	}
	for _, id := range ids {
		if downloadStatus(db, id) != string(models.DownloadStatusCompleted) {
			t.Error("no record may change")
		}
	}
	if cmds != 0 {
		t.Error("no refresh when nothing changed")
	}
}

func TestKeepDuplicate_DirectoryOrSymlinkRefusedWholeRequest(t *testing.T) {
	for _, kind := range []string{"directory", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			db := setupTestDB(t)
			newForceDownloadTestConfig(t, "", "")
			dir := t.TempDir()
			a := writeMedia(t, filepath.Join(dir, "a.mkv"), "a")
			ok := writeMedia(t, filepath.Join(dir, "ok.mkv"), "ok")
			bad := filepath.Join(dir, "bad.mkv")
			if kind == "directory" {
				os.Mkdir(bad, 0755)
			} else if err := os.Symlink(a, bad); err != nil {
				t.Skip("symlinks unsupported")
			}
			ids, _ := seedDuplicateMovie(t, db, 11, a, ok, bad)

			if w := doReq(NewServer(), "POST", keepURL(ids[0])); w.Code != http.StatusConflict {
				t.Fatalf("expected 409, got %d", w.Code)
			}
			if _, err := os.Stat(ok); err != nil {
				t.Error("preflight must refuse before any deletion")
			}
			if _, err := os.Lstat(bad); err != nil {
				t.Error("non-regular path must survive")
			}
			for _, id := range ids {
				if downloadStatus(db, id) != string(models.DownloadStatusCompleted) {
					t.Error("no record may change")
				}
			}
		})
	}
}

func TestKeepDuplicate_UndeletableFileLeavesOccurrencesUnchanged(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	db := setupTestDB(t)
	newForceDownloadTestConfig(t, "", "")
	dir := t.TempDir()
	a := writeMedia(t, filepath.Join(dir, "a.mkv"), "a")
	ro := filepath.Join(dir, "ro")
	os.Mkdir(ro, 0755)
	b := writeMedia(t, filepath.Join(ro, "b.mkv"), "b")
	os.Chmod(ro, 0555)
	t.Cleanup(func() { os.Chmod(ro, 0755) })
	ids, _ := seedDuplicateMovie(t, db, 11, a, b)

	if w := doReq(NewServer(), "POST", keepURL(ids[0])); w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if downloadStatus(db, ids[1]) != string(models.DownloadStatusCompleted) {
		t.Error("occurrence of the undeletable file must be unchanged")
	}
	if _, err := os.Stat(a); err != nil {
		t.Error("kept file must be intact")
	}
}

func TestKeepDuplicate_RefreshOutcomes(t *testing.T) {
	cases := []struct {
		name string
		want string
		cfg  func(t *testing.T, cmds *int32)
	}{
		{"failed", refreshFailed, func(t *testing.T, cmds *int32) {
			srv := radarrFake(t, "Star Wars", 1977, http.StatusInternalServerError, cmds)
			newForceDownloadTestConfig(t, srv.URL, "")
		}},
		{"not_found", refreshNotFoundInLibrary, func(t *testing.T, cmds *int32) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`[]`)) }))
			t.Cleanup(srv.Close)
			newForceDownloadTestConfig(t, srv.URL, "")
		}},
		{"not_configured", refreshNotConfigured, func(t *testing.T, cmds *int32) { newForceDownloadTestConfig(t, "", "") }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			var cmds int32
			tc.cfg(t, &cmds)
			dir := t.TempDir()
			a := writeMedia(t, filepath.Join(dir, "a.mkv"), "a")
			b := writeMedia(t, filepath.Join(dir, "b.mkv"), "b")
			ids, _ := seedDuplicateMovie(t, db, 11, a, b)

			resp := decodeKeep(t, doReq(NewServer(), "POST", keepURL(ids[0])))
			if resp.Refresh != tc.want {
				t.Errorf("expected %s, got %s", tc.want, resp.Refresh)
			}
			if _, err := os.Stat(b); !os.IsNotExist(err) || downloadStatus(db, ids[1]) != string(models.DownloadStatusCancelled) {
				t.Error("deletion and cancellation must stay in effect")
			}
			if tc.name == "failed" && cmds != 1 {
				t.Errorf("expected exactly one command attempt, got %d", cmds)
			}
			if tc.name == "not_configured" && cmds != 0 {
				t.Error("no call when not configured")
			}
		})
	}
}

// 3.5 / 4.2: resolve-duplicate never deletes and never issues a refresh.
func TestResolveDuplicate_DeletesNothingAndNeverRefreshes(t *testing.T) {
	db := setupTestDB(t)
	var cmds int32
	srv := radarrFake(t, "Star Wars", 1977, http.StatusCreated, &cmds)
	newForceDownloadTestConfig(t, srv.URL, "")
	dir := t.TempDir()
	a := writeMedia(t, filepath.Join(dir, "a.mkv"), "a")
	b := writeMedia(t, filepath.Join(dir, "b.mkv"), "b")
	ids, _ := seedDuplicateMovie(t, db, 11, a, b)

	if w := doReq(NewServer(), "POST", fmt.Sprintf("/api/v1/downloads/%d/resolve-duplicate", ids[0])); w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	for _, p := range []string{a, b} {
		if _, err := os.Stat(p); err != nil {
			t.Error("resolve must not delete files")
		}
	}
	if cmds != 0 {
		t.Errorf("resolve must not issue Refresh & Scan, got %d command calls", cmds)
	}
}
