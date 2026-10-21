package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/models"
	"gorm.io/gorm"
)

type downloadsListResp struct {
	Data []struct {
		ID     uint   `json:"id"`
		Status string `json:"status"`
	} `json:"data"`
	Total           int64 `json:"total"`
	TotalPages      int   `json:"total_pages"`
	CancelledHidden int64 `json:"cancelled_hidden"`
}

func getDownloadsList(t *testing.T, server *Server, query string) downloadsListResp {
	t.Helper()
	req, _ := http.NewRequest("GET", "/api/v1/downloads"+query, nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET %s: status %d: %s", query, w.Code, w.Body.String())
	}
	var resp downloadsListResp
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return resp
}

func (r downloadsListResp) ids() []uint {
	out := make([]uint, 0, len(r.Data))
	for _, d := range r.Data {
		out = append(out, d.ID)
	}
	return out
}

func seedDL(t *testing.T, db *gorm.DB, status, path string, age time.Duration) models.DownloadInfo {
	t.Helper()
	p := path
	dl := models.DownloadInfo{
		URL: "http://x/" + path, Status: status, DownloadPath: &p,
		CreatedAt: time.Now().Add(-age), UpdatedAt: time.Now().Add(-age),
	}
	if err := db.Create(&dl).Error; err != nil {
		t.Fatal(err)
	}
	return dl
}

func seedLine(t *testing.T, db *gorm.DB, dl models.DownloadInfo, tvg string, ct models.ContentType, movieID, showID *uint) {
	t.Helper()
	id := dl.ID
	pl := models.ProcessedLine{TvgName: tvg, ContentType: ct, DownloadInfoID: &id, MovieID: movieID, TVShowID: showID, LineHash: tvg + dl.URL}
	if err := db.Create(&pl).Error; err != nil {
		t.Fatal(err)
	}
}

func TestListDownloadsEnriched_StatusActive(t *testing.T) {
	db := setupTestDB(t)
	server := NewServer()
	seedDL(t, db, "completed", "a.mkv", 3*time.Hour)
	seedDL(t, db, "cancelled", "b.mkv", time.Hour)
	seedDL(t, db, "failed", "c.mkv", 2*time.Hour)

	active := getDownloadsList(t, server, "?status=active")
	if active.Total != 2 || active.CancelledHidden != 1 {
		t.Errorf("active: total=%d hidden=%d", active.Total, active.CancelledHidden)
	}
	for _, d := range active.Data {
		if d.Status == "cancelled" {
			t.Error("active list must not contain cancelled")
		}
	}

	if r := getDownloadsList(t, server, "?status=failed"); r.Total != 1 || r.CancelledHidden != 0 {
		t.Errorf("failed: total=%d hidden=%d", r.Total, r.CancelledHidden)
	}
	if r := getDownloadsList(t, server, "?status=cancelled"); r.Total != 1 || r.CancelledHidden != 0 {
		t.Errorf("cancelled: total=%d hidden=%d", r.Total, r.CancelledHidden)
	}
	if r := getDownloadsList(t, server, ""); r.Total != 3 || r.CancelledHidden != 0 {
		t.Errorf("none: total=%d hidden=%d", r.Total, r.CancelledHidden)
	}
}

func TestListDownloadsEnriched_StatusRankOrdering(t *testing.T) {
	db := setupTestDB(t)
	server := NewServer()
	cancelled := seedDL(t, db, "cancelled", "a.mkv", time.Minute)
	completedOld := seedDL(t, db, "completed", "b.mkv", 5*time.Hour)
	completedNew := seedDL(t, db, "completed", "c.mkv", 2*time.Hour)
	failed := seedDL(t, db, "failed", "d.mkv", 30*time.Second)
	downloading := seedDL(t, db, "downloading", "e.mkv", 10*time.Hour)

	want := []uint{downloading.ID, completedNew.ID, completedOld.ID, failed.ID, cancelled.ID}
	for _, q := range []string{"", "?problem=missing_year,unknown_format,year_mismatch,low_quality"} {
		got := getDownloadsList(t, server, q).ids()
		if q != "" {
			// problem path keeps the same relative order for the rows it returns
			filtered := map[uint]bool{}
			for _, id := range got {
				filtered[id] = true
			}
			var w2 []uint
			for _, id := range want {
				if filtered[id] {
					w2 = append(w2, id)
				}
			}
			if len(w2) != len(got) {
				t.Fatalf("problem path returned unexpected rows: %v", got)
			}
			want2 := w2
			for i := range got {
				if got[i] != want2[i] {
					t.Errorf("problem path order = %v, want %v", got, want2)
					break
				}
			}
			continue
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("order = %v, want %v", got, want)
			}
		}
	}
}

func TestDownloadStatusRank_CoversEveryConstant(t *testing.T) {
	for _, st := range []models.DownloadStatus{
		models.DownloadStatusPending, models.DownloadStatusDownloading, models.DownloadStatusPaused,
		models.DownloadStatusCompleted, models.DownloadStatusFailed, models.DownloadStatusRetrying,
		models.DownloadStatusCancelled, models.DownloadStatusPolicyStopped,
	} {
		if _, ok := models.DownloadStatusRank[st]; !ok {
			t.Errorf("status %q has no rank", st)
		}
	}
}

func TestListDownloadsEnriched_QueryHasNoDistinct(t *testing.T) {
	db := setupTestDB(t)
	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		q := applyDownloadsContentFilters(db, tx.Model(&models.DownloadInfo{}), "movies", "peace")
		var out []models.DownloadInfo
		return q.Order(downloadStatusRankOrder()).Find(&out)
	})
	if strings.Contains(strings.ToUpper(sql), "DISTINCT") {
		t.Errorf("generated SQL must not contain DISTINCT: %s", sql)
	}
}

func TestListDownloadsEnriched_Search(t *testing.T) {
	db := setupTestDB(t)
	server := NewServer()

	show := models.TVShow{TMDBTitle: "Peacemaker", TMDBYear: 2022}
	if err := db.Create(&show).Error; err != nil {
		t.Fatal(err)
	}
	movie := models.Movie{TMDBTitle: "Other Film", TMDBYear: 2020}
	if err := db.Create(&movie).Error; err != nil {
		t.Fatal(err)
	}

	byTitle := seedDL(t, db, "completed", "x1.mkv", time.Hour)
	seedLine(t, db, byTitle, "ep one", models.ContentTypeTVShows, nil, &show.ID)
	seedLine(t, db, byTitle, "ep one bis", models.ContentTypeTVShows, nil, &show.ID) // duplicate match
	byTitleCancelled := seedDL(t, db, "cancelled", "x1b.mkv", time.Hour)
	seedLine(t, db, byTitleCancelled, "ep one", models.ContentTypeTVShows, nil, &show.ID)
	byTvg := seedDL(t, db, "completed", "x2.mkv", 2*time.Hour)
	seedLine(t, db, byTvg, "FR: PEACEMAKER S01", models.ContentTypeTVShows, nil, nil)
	byPath := seedDL(t, db, "completed", "/dl/peacemaker.s01.mkv", 3*time.Hour)
	other := seedDL(t, db, "completed", "o.mkv", 4*time.Hour)
	seedLine(t, db, other, "nothing", models.ContentTypeMovies, &movie.ID, nil)

	r := getDownloadsList(t, server, "?search=peacemaker")
	if r.Total != 4 {
		t.Fatalf("total = %d, want 4 (ids %v)", r.Total, r.ids())
	}
	seen := map[uint]int{}
	for _, id := range r.ids() {
		seen[id]++
	}
	if seen[byTitle.ID] != 1 || seen[byTvg.ID] != 1 || seen[byPath.ID] != 1 || seen[other.ID] != 0 {
		t.Errorf("unexpected rows: %v", seen)
	}

	if up := getDownloadsList(t, server, "?search=PEACEMAKER"); up.Total != r.Total {
		t.Errorf("case-insensitive total = %d, want %d", up.Total, r.Total)
	}

	combined := getDownloadsList(t, server, "?search=peacemaker&status=active&limit=2&offset=0")
	if combined.Total != 3 || len(combined.Data) != 2 || combined.TotalPages != 2 || combined.CancelledHidden != 1 {
		t.Errorf("combined: total=%d len=%d pages=%d hidden=%d", combined.Total, len(combined.Data), combined.TotalPages, combined.CancelledHidden)
	}

	if tv := getDownloadsList(t, server, "?search=peacemaker&type=tvshows"); tv.Total != 3 {
		t.Errorf("search+type total = %d, want 3", tv.Total)
	}

	if none := getDownloadsList(t, server, "?search=100%25"); none.Total != 0 {
		t.Errorf("escaped %% should match nothing, got %d", none.Total)
	}
	if none := getDownloadsList(t, server, "?search=p_acemaker"); none.Total != 0 {
		t.Errorf("escaped _ should match nothing, got %d", none.Total)
	}

	// problem path: cancelled_hidden counts only cancelled rows matching problem
	pr := getDownloadsList(t, server, "?search=peacemaker&status=active&problem=missing_year")
	if pr.CancelledHidden > 1 {
		t.Errorf("problem hidden = %d", pr.CancelledHidden)
	}
}

func TestListDownloadsEnriched_TypeFilterNoDuplicates(t *testing.T) {
	db := setupTestDB(t)
	server := NewServer()
	dl := seedDL(t, db, "completed", "m.mkv", time.Hour)
	seedLine(t, db, dl, "a", models.ContentTypeMovies, nil, nil)
	seedLine(t, db, dl, "b", models.ContentTypeMovies, nil, nil)
	if r := getDownloadsList(t, server, "?type=movies"); r.Total != 1 || len(r.Data) != 1 {
		t.Errorf("total=%d len=%d", r.Total, len(r.Data))
	}
}
