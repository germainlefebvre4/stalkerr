package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/models"
)

type typedGroupedResponse struct {
	Data       []ItemGroupResponse `json:"data"`
	Total      int64                `json:"total"`
	Limit      int                  `json:"limit"`
	Offset     int                  `json:"offset"`
	TotalPages int                  `json:"total_pages"`
}

func listItemGroupsRequest(t *testing.T, server *Server, query string) typedGroupedResponse {
	t.Helper()
	req, _ := http.NewRequest("GET", "/api/v1/items/grouped"+query, nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for query %q, got %d: %s", query, w.Code, w.Body.String())
	}

	var resp typedGroupedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	return resp
}

func TestListItemGroups_MovieAggregation(t *testing.T) {
	db := setupTestDB(t)

	movie := models.Movie{TMDBID: 111, TMDBTitle: "The Matrix", TMDBYear: 1999}
	db.Create(&movie)

	base := time.Now()
	lines := []models.ProcessedLine{
		{LineContent: "l1", LineHash: "h1", TvgName: "Matrix 1", GroupTitle: "g", ContentType: "movies", State: "processed", MovieID: &movie.ID, CreatedAt: base, UpdatedAt: base},
		{LineContent: "l2", LineHash: "h2", TvgName: "Matrix 2", GroupTitle: "g", ContentType: "movies", State: "processed", MovieID: &movie.ID, CreatedAt: base.Add(time.Hour), UpdatedAt: base},
		{LineContent: "l3", LineHash: "h3", TvgName: "Matrix 3", GroupTitle: "g", ContentType: "movies", State: "processed", MovieID: &movie.ID, CreatedAt: base.Add(2 * time.Hour), UpdatedAt: base},
	}
	for i := range lines {
		if err := db.Create(&lines[i]).Error; err != nil {
			t.Fatalf("seed error: %v", err)
		}
	}

	server := NewServer()
	resp := listItemGroupsRequest(t, server, "")

	if resp.Total != 1 || len(resp.Data) != 1 {
		t.Fatalf("expected a single movie group, got total=%d data=%+v", resp.Total, resp.Data)
	}
	group := resp.Data[0]
	if group.Type != "movie" || group.MovieID == nil || *group.MovieID != movie.ID {
		t.Fatalf("expected movie group for movie id %d, got %+v", movie.ID, group)
	}
	if group.Title == nil || *group.Title != "The Matrix" {
		t.Errorf("expected title 'The Matrix', got %+v", group.Title)
	}
	expectedLatest := base.Add(2 * time.Hour).Format("2006-01-02T15:04:05Z07:00")
	if group.LatestActivity[:19] != expectedLatest[:19] {
		t.Errorf("expected latest_activity ~%s, got %s", expectedLatest, group.LatestActivity)
	}
}

func TestListItemGroups_TVShowAggregation(t *testing.T) {
	db := setupTestDB(t)

	// Three TVShow rows for the same show (shared tmdb_id), across seasons/episodes.
	ep1 := models.TVShow{TMDBID: 222, TMDBTitle: "Breaking Bad", TMDBYear: 2008, Season: intPtr(1), Episode: intPtr(1)}
	db.Create(&ep1)
	ep2 := models.TVShow{TMDBID: 222, TMDBTitle: "Breaking Bad", TMDBYear: 2008, Season: intPtr(3), Episode: intPtr(5)}
	db.Create(&ep2)
	ep3 := models.TVShow{TMDBID: 222, TMDBTitle: "Breaking Bad", TMDBYear: 2008, Season: intPtr(5), Episode: intPtr(16)}
	db.Create(&ep3)

	base := time.Now()
	lines := []models.ProcessedLine{
		{LineContent: "l1", LineHash: "h1", TvgName: "BB S01E01", GroupTitle: "g", ContentType: "tvshows", State: "processed", TVShowID: &ep1.ID, CreatedAt: base, UpdatedAt: base},
		{LineContent: "l2", LineHash: "h2", TvgName: "BB S03E05", GroupTitle: "g", ContentType: "tvshows", State: "processed", TVShowID: &ep2.ID, CreatedAt: base.Add(time.Hour), UpdatedAt: base},
		{LineContent: "l3", LineHash: "h3", TvgName: "BB S05E16", GroupTitle: "g", ContentType: "tvshows", State: "processed", TVShowID: &ep3.ID, CreatedAt: base.Add(2 * time.Hour), UpdatedAt: base},
	}
	for i := range lines {
		if err := db.Create(&lines[i]).Error; err != nil {
			t.Fatalf("seed error: %v", err)
		}
	}

	server := NewServer()
	resp := listItemGroupsRequest(t, server, "")

	if resp.Total != 1 || len(resp.Data) != 1 {
		t.Fatalf("expected a single tvshow group, got total=%d data=%+v", resp.Total, resp.Data)
	}
	group := resp.Data[0]
	if group.Type != "tvshow" || group.TMDBID == nil || *group.TMDBID != 222 {
		t.Fatalf("expected tvshow group for tmdb_id 222, got %+v", group)
	}
	if group.SeasonStart == nil || *group.SeasonStart != 1 {
		t.Errorf("expected season_start=1, got %+v", group.SeasonStart)
	}
	if group.SeasonEnd == nil || *group.SeasonEnd != 5 {
		t.Errorf("expected season_end=5, got %+v", group.SeasonEnd)
	}
}

func TestListItemGroups_UnmatchedPseudoGroups(t *testing.T) {
	db := setupTestDB(t)

	movie := models.Movie{TMDBID: 111, TMDBTitle: "The Matrix", TMDBYear: 1999}
	db.Create(&movie)

	base := time.Now()
	lines := []models.ProcessedLine{
		// Matched movie: should not affect the unmatched pseudo-group.
		{LineContent: "l1", LineHash: "h1", TvgName: "Matrix", GroupTitle: "g", ContentType: "movies", State: "processed", MovieID: &movie.ID, CreatedAt: base, UpdatedAt: base},
		// Unmatched movies.
		{LineContent: "l2", LineHash: "h2", TvgName: "Unknown Movie 1", GroupTitle: "g", ContentType: "movies", State: "processed", CreatedAt: base.Add(time.Hour), UpdatedAt: base},
		{LineContent: "l3", LineHash: "h3", TvgName: "Unknown Movie 2", GroupTitle: "g", ContentType: "movies", State: "processed", CreatedAt: base.Add(2 * time.Hour), UpdatedAt: base},
	}
	for i := range lines {
		if err := db.Create(&lines[i]).Error; err != nil {
			t.Fatalf("seed error: %v", err)
		}
	}

	server := NewServer()
	resp := listItemGroupsRequest(t, server, "")

	// One movie group + one unmatched_movies pseudo-group; no unmatched_tvshows
	// pseudo-group since there are zero unmatched tvshows items.
	if resp.Total != 2 {
		t.Fatalf("expected 2 groups (movie + unmatched_movies), got total=%d data=%+v", resp.Total, resp.Data)
	}
	var sawMovie, sawUnmatchedMovies, sawUnmatchedTVShows bool
	for _, g := range resp.Data {
		switch g.Type {
		case "movie":
			sawMovie = true
		case "unmatched_movies":
			sawUnmatchedMovies = true
		case "unmatched_tvshows":
			sawUnmatchedTVShows = true
		}
	}
	if !sawMovie {
		t.Error("expected a movie group to be present")
	}
	if !sawUnmatchedMovies {
		t.Error("expected an unmatched_movies pseudo-group to be present")
	}
	if sawUnmatchedTVShows {
		t.Error("expected no unmatched_tvshows pseudo-group when there are zero unmatched tvshows items")
	}
}

func TestListItemGroups_PaginationCountsDistinctGroups(t *testing.T) {
	db := setupTestDB(t)

	tvshow := models.TVShow{TMDBID: 222, TMDBTitle: "Breaking Bad", TMDBYear: 2008}
	db.Create(&tvshow)

	base := time.Now()
	// 100 episodes belonging to the same show should count as exactly 1 group.
	for i := 0; i < 100; i++ {
		ep := models.TVShow{TMDBID: 222, TMDBTitle: "Breaking Bad", TMDBYear: 2008, Season: intPtr(1), Episode: intPtr(i + 1)}
		if err := db.Create(&ep).Error; err != nil {
			t.Fatalf("seed error: %v", err)
		}
		line := models.ProcessedLine{
			LineContent: fmt.Sprintf("l%d", i), LineHash: fmt.Sprintf("h%d", i), TvgName: fmt.Sprintf("BB E%d", i),
			GroupTitle: "g", ContentType: "tvshows", State: "processed", TVShowID: &ep.ID,
			CreatedAt: base.Add(time.Duration(i) * time.Minute), UpdatedAt: base,
		}
		if err := db.Create(&line).Error; err != nil {
			t.Fatalf("seed error: %v", err)
		}
	}

	server := NewServer()
	resp := listItemGroupsRequest(t, server, "?limit=10&offset=0")

	if resp.Total != 1 {
		t.Fatalf("expected total=1 distinct group for a 100-episode show, got %d", resp.Total)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 group unit toward limit, got %d entries", len(resp.Data))
	}
	if resp.TotalPages != 1 {
		t.Errorf("expected total_pages=1, got %d", resp.TotalPages)
	}
}

func TestListItemGroups_EndToEndFiltersAndExclusions(t *testing.T) {
	db := setupTestDB(t)

	movie := models.Movie{TMDBID: 111, TMDBTitle: "The Matrix", TMDBYear: 1999}
	db.Create(&movie)

	// A show with episodes across 3 seasons, only some of which are "downloaded".
	epS1 := models.TVShow{TMDBID: 222, TMDBTitle: "Breaking Bad", TMDBYear: 2008, Season: intPtr(1), Episode: intPtr(1)}
	db.Create(&epS1)
	epS2 := models.TVShow{TMDBID: 222, TMDBTitle: "Breaking Bad", TMDBYear: 2008, Season: intPtr(2), Episode: intPtr(1)}
	db.Create(&epS2)
	epS3 := models.TVShow{TMDBID: 222, TMDBTitle: "Breaking Bad", TMDBYear: 2008, Season: intPtr(3), Episode: intPtr(1)}
	db.Create(&epS3)

	base := time.Now()
	channel := models.Channel{Name: "News 24", GroupTitle: "g"}
	db.Create(&channel)

	lines := []models.ProcessedLine{
		{LineContent: "movie-1", LineHash: "hm1", TvgName: "Matrix", GroupTitle: "g", ContentType: "movies", State: "processed", MovieID: &movie.ID, CreatedAt: base, UpdatedAt: base},
		{LineContent: "unmatched-movie", LineHash: "hum1", TvgName: "Unknown", GroupTitle: "g", ContentType: "movies", State: "processed", CreatedAt: base.Add(time.Minute), UpdatedAt: base},
		{LineContent: "bb-s1", LineHash: "hbb1", TvgName: "BB S01", GroupTitle: "g", ContentType: "tvshows", State: "processed", TVShowID: &epS1.ID, CreatedAt: base.Add(2 * time.Minute), UpdatedAt: base},
		{LineContent: "bb-s2", LineHash: "hbb2", TvgName: "BB S02", GroupTitle: "g", ContentType: "tvshows", State: "downloaded", TVShowID: &epS2.ID, CreatedAt: base.Add(3 * time.Minute), UpdatedAt: base},
		{LineContent: "bb-s3", LineHash: "hbb3", TvgName: "BB S03", GroupTitle: "g", ContentType: "tvshows", State: "processed", TVShowID: &epS3.ID, CreatedAt: base.Add(4 * time.Minute), UpdatedAt: base},
		{LineContent: "channel-1", LineHash: "hc1", TvgName: "News", GroupTitle: "g", ContentType: "channels", State: "processed", ChannelID: &channel.ID, CreatedAt: base.Add(5 * time.Minute), UpdatedAt: base},
	}
	for i := range lines {
		if err := db.Create(&lines[i]).Error; err != nil {
			t.Fatalf("seed error: %v", err)
		}
	}

	server := NewServer()

	// Mixed movies+shows+unmatched, no filter: channel must never appear.
	respAll := listItemGroupsRequest(t, server, "")
	for _, g := range respAll.Data {
		if g.Type != "movie" && g.Type != "tvshow" && g.Type != "unmatched_movies" && g.Type != "unmatched_tvshows" {
			t.Errorf("unexpected group type %q in unfiltered response", g.Type)
		}
	}
	if respAll.Total != 3 {
		t.Fatalf("expected 3 groups (movie, unmatched_movies, tvshow), got %d: %+v", respAll.Total, respAll.Data)
	}

	// state=downloaded empties the show down to its season-2 episode only: the
	// season range must reflect only the matching episode.
	respDownloaded := listItemGroupsRequest(t, server, "?state=downloaded")
	if respDownloaded.Total != 1 {
		t.Fatalf("expected 1 group for state=downloaded, got %d: %+v", respDownloaded.Total, respDownloaded.Data)
	}
	showGroup := respDownloaded.Data[0]
	if showGroup.Type != "tvshow" {
		t.Fatalf("expected the downloaded group to be the tvshow, got %+v", showGroup)
	}
	if showGroup.SeasonStart == nil || *showGroup.SeasonStart != 2 || showGroup.SeasonEnd == nil || *showGroup.SeasonEnd != 2 {
		t.Errorf("expected season range S02-S02 for state=downloaded filter, got start=%v end=%v", showGroup.SeasonStart, showGroup.SeasonEnd)
	}

	// content_type=movies excludes tvshows/unmatched-tvshows entries entirely.
	respMovies := listItemGroupsRequest(t, server, "?content_type=movies")
	if respMovies.Total != 2 {
		t.Fatalf("expected 2 groups (movie + unmatched_movies) for content_type=movies, got %d: %+v", respMovies.Total, respMovies.Data)
	}
	for _, g := range respMovies.Data {
		if g.Type == "tvshow" || g.Type == "unmatched_tvshows" {
			t.Errorf("content_type=movies leaked a tvshow-related group: %+v", g)
		}
	}
}

func TestListItemGroups_ReportsLatestProcessingLogID(t *testing.T) {
	db := setupTestDB(t)

	movie := models.Movie{TMDBID: 111, TMDBTitle: "The Matrix", TMDBYear: 1999}
	db.Create(&movie)

	olderRun := models.ProcessingLog{Status: "completed"}
	db.Create(&olderRun)
	newerRun := models.ProcessingLog{Status: "completed"}
	db.Create(&newerRun)

	base := time.Now()
	lines := []models.ProcessedLine{
		{LineContent: "l1", LineHash: "h1", TvgName: "Matrix 1", GroupTitle: "g", ContentType: "movies", State: "processed", MovieID: &movie.ID, ProcessingLogID: &olderRun.ID, CreatedAt: base, UpdatedAt: base},
		{LineContent: "l2", LineHash: "h2", TvgName: "Matrix 2", GroupTitle: "g", ContentType: "movies", State: "processed", MovieID: &movie.ID, ProcessingLogID: &newerRun.ID, CreatedAt: base.Add(time.Hour), UpdatedAt: base},
	}
	for i := range lines {
		if err := db.Create(&lines[i]).Error; err != nil {
			t.Fatalf("seed error: %v", err)
		}
	}

	server := NewServer()
	resp := listItemGroupsRequest(t, server, "")

	if resp.Total != 1 || len(resp.Data) != 1 {
		t.Fatalf("expected a single movie group, got total=%d data=%+v", resp.Total, resp.Data)
	}
	group := resp.Data[0]
	if group.LatestProcessingLogID == nil || *group.LatestProcessingLogID != newerRun.ID {
		t.Fatalf("expected latest_processing_log_id=%d (matching latest_activity's run), got %+v", newerRun.ID, group.LatestProcessingLogID)
	}
}

func TestListItemGroups_NullProcessingLogIDReportsNoValue(t *testing.T) {
	db := setupTestDB(t)

	movie := models.Movie{TMDBID: 111, TMDBTitle: "The Matrix", TMDBYear: 1999}
	db.Create(&movie)

	base := time.Now()
	line := models.ProcessedLine{
		LineContent: "l1", LineHash: "h1", TvgName: "Matrix 1", GroupTitle: "g",
		ContentType: "movies", State: "processed", MovieID: &movie.ID,
		ProcessingLogID: nil, CreatedAt: base, UpdatedAt: base,
	}
	if err := db.Create(&line).Error; err != nil {
		t.Fatalf("seed error: %v", err)
	}

	server := NewServer()
	resp := listItemGroupsRequest(t, server, "")

	if resp.Total != 1 || len(resp.Data) != 1 {
		t.Fatalf("expected a single movie group, got total=%d data=%+v", resp.Total, resp.Data)
	}
	group := resp.Data[0]
	if group.LatestProcessingLogID != nil {
		t.Fatalf("expected latest_processing_log_id to be absent for a group whose only item predates run attribution, got %v", *group.LatestProcessingLogID)
	}
}

func TestExpandGroup_ScopesToLatestProcessingRun(t *testing.T) {
	db := setupTestDB(t)

	movie := models.Movie{TMDBID: 111, TMDBTitle: "The Matrix", TMDBYear: 1999}
	db.Create(&movie)

	olderRun := models.ProcessingLog{Action: "process_m3u", Status: "success", StartedAt: time.Now()}
	db.Create(&olderRun)
	newerRun := models.ProcessingLog{Action: "process_m3u", Status: "success", StartedAt: time.Now()}
	db.Create(&newerRun)

	base := time.Now()
	lines := []models.ProcessedLine{
		{LineContent: "old-1", LineHash: "hold1", TvgName: "Matrix (old run)", GroupTitle: "g", ContentType: "movies", State: "processed", MovieID: &movie.ID, ProcessingLogID: &olderRun.ID, CreatedAt: base, UpdatedAt: base},
		{LineContent: "new-1", LineHash: "hnew1", TvgName: "Matrix (new run)", GroupTitle: "g", ContentType: "movies", State: "processed", MovieID: &movie.ID, ProcessingLogID: &newerRun.ID, CreatedAt: base.Add(time.Hour), UpdatedAt: base},
	}
	for i := range lines {
		if err := db.Create(&lines[i]).Error; err != nil {
			t.Fatalf("seed error: %v", err)
		}
	}

	server := NewServer()

	// Step 1: list groups, as the "Films & Séries" table does.
	groupsResp := listItemGroupsRequest(t, server, "")
	if groupsResp.Total != 1 || len(groupsResp.Data) != 1 {
		t.Fatalf("expected a single movie group, got total=%d data=%+v", groupsResp.Total, groupsResp.Data)
	}
	group := groupsResp.Data[0]
	if group.LatestProcessingLogID == nil || *group.LatestProcessingLogID != newerRun.ID {
		t.Fatalf("expected group's latest_processing_log_id=%d, got %+v", newerRun.ID, group.LatestProcessingLogID)
	}

	// Step 2: expand it, as the frontend does, scoped by movie_id + the group's
	// reported latest_processing_log_id.
	itemsResp := listItemsSorted(t, server, fmt.Sprintf("?movie_id=%d&processing_log_id=%d", *group.MovieID, *group.LatestProcessingLogID))
	if itemsResp.Total != 1 {
		t.Fatalf("expected only the newer run's item, got total=%d: %+v", itemsResp.Total, itemsResp.Data)
	}
	if itemsResp.Data[0].LineHash != "hnew1" {
		t.Errorf("expected the newer run's item (hnew1), got %+v", itemsResp.Data[0])
	}
}

func TestListItemGroups_ChannelsNeverAppear(t *testing.T) {
	db := setupTestDB(t)

	channel := models.Channel{Name: "News 24", GroupTitle: "g"}
	db.Create(&channel)

	line := models.ProcessedLine{
		LineContent: "channel-1", LineHash: "hc1", TvgName: "News", GroupTitle: "g",
		ContentType: "channels", State: "processed", ChannelID: &channel.ID,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := db.Create(&line).Error; err != nil {
		t.Fatalf("seed error: %v", err)
	}

	server := NewServer()
	resp := listItemGroupsRequest(t, server, "")

	if resp.Total != 0 || len(resp.Data) != 0 {
		t.Fatalf("expected no groups derived from channels, got total=%d data=%+v", resp.Total, resp.Data)
	}
}
