package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/filter"
	"github.com/glefebvre/stalkeer/internal/models"
)

// buildM3UFixture writes a minimal M3U playlist with n entries into dir,
// naming it the way ArchiveManager.ListArchiveFiles expects
// ("playlist_*.m3u"), so GetLatestArchive picks it up. Each entry gets a
// unique tvg-name/url pair (the parser dedups on their hash) and the given
// group-title.
func buildM3UFixture(t *testing.T, dir string, entries []struct{ tvgName, groupTitle string }) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create archive dir: %v", err)
	}

	var buf bytes.Buffer
	buf.WriteString("#EXTM3U\n")
	for i, e := range entries {
		fmt.Fprintf(&buf, "#EXTINF:-1 tvg-name=\"%s\" group-title=\"%s\",%s\n", e.tvgName, e.groupTitle, e.tvgName)
		fmt.Fprintf(&buf, "http://example.com/stream/%d\n", i)
	}

	path := filepath.Join(dir, "playlist_20260101_000000.000000.m3u")
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}
	return path
}

func setupDryRunSource(t *testing.T, name string) (archiveDir string) {
	t.Helper()
	base := t.TempDir()
	config.SetConfig(&config.Config{M3U: config.M3UConfig{Sources: []config.M3USourceConfig{
		{
			Name:     name,
			FilePath: filepath.Join(base, "downloads", "playlist.m3u"),
			Download: config.M3UDownloadConfig{ArchiveDir: filepath.Join(base, "archives")},
		},
	}}})
	// SourcePaths namespaces the archive dir by source name.
	return filepath.Join(base, "archives", name)
}

func dryRunRequest(t *testing.T, server *Server, req FilterDryRunRequest) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", "/api/v1/filters/dryrun", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, httpReq)
	return w
}

func TestFilterDryRun_UnknownSourceRejected(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(&config.Config{})
	server := NewServer()

	w := dryRunRequest(t, server, FilterDryRunRequest{SourceName: "nope", Attributes: []string{"group_title"}})

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestFilterDryRun_NoAttributesRejected(t *testing.T) {
	setupTestDB(t)
	setupDryRunSource(t, "src")
	server := NewServer()

	w := dryRunRequest(t, server, FilterDryRunRequest{SourceName: "src", Attributes: []string{}})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	var errResp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &errResp)
	if errResp.Error != "invalid_attribute" {
		t.Errorf("expected invalid_attribute, got %q", errResp.Error)
	}
}

func TestFilterDryRun_UnsupportedAttributeRejected(t *testing.T) {
	setupTestDB(t)
	setupDryRunSource(t, "src")
	server := NewServer()

	w := dryRunRequest(t, server, FilterDryRunRequest{SourceName: "src", Attributes: []string{"bogus"}})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	var errResp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &errResp)
	if errResp.Error != "invalid_attribute" {
		t.Errorf("expected invalid_attribute, got %q", errResp.Error)
	}
}

func TestFilterDryRun_DuplicateAttributeRejected(t *testing.T) {
	setupTestDB(t)
	setupDryRunSource(t, "src")
	server := NewServer()

	w := dryRunRequest(t, server, FilterDryRunRequest{SourceName: "src", Attributes: []string{"group_title", "group_title"}})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	var errResp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &errResp)
	if errResp.Error != "invalid_attribute" {
		t.Errorf("expected invalid_attribute, got %q", errResp.Error)
	}
}

func TestFilterDryRun_SearchAttributeRequiredForCombined(t *testing.T) {
	setupTestDB(t)
	setupDryRunSource(t, "src")
	server := NewServer()

	w := dryRunRequest(t, server, FilterDryRunRequest{
		SourceName: "src",
		Attributes: []string{"group_title", "tvg_name"},
		Search:     "Sport",
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	var errResp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &errResp)
	if errResp.Error != "invalid_search_attribute" {
		t.Errorf("expected invalid_search_attribute, got %q", errResp.Error)
	}
}

func TestFilterDryRun_SearchAttributeMustBeSupplied(t *testing.T) {
	setupTestDB(t)
	setupDryRunSource(t, "src")
	server := NewServer()

	w := dryRunRequest(t, server, FilterDryRunRequest{
		SourceName:      "src",
		Attributes:      []string{"group_title"},
		Search:          "Sport",
		SearchAttribute: "tvg_name",
	})

	// Search-attribute mismatches are only enforced in combined mode; in
	// single-attribute mode the tested attribute is always the search
	// attribute, so this succeeds (the request just isn't rejected).
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestFilterDryRun_InvalidRegexRejected(t *testing.T) {
	setupTestDB(t)
	setupDryRunSource(t, "src")
	server := NewServer()

	w := dryRunRequest(t, server, FilterDryRunRequest{
		SourceName:        "src",
		Attributes:        []string{"group_title"},
		GroupTitleInclude: "^(Movies",
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	var errResp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &errResp)
	if errResp.Error != "invalid_pattern" {
		t.Errorf("expected invalid_pattern, got %q", errResp.Error)
	}
}

func TestFilterDryRun_NoArchiveCondition(t *testing.T) {
	setupTestDB(t)
	setupDryRunSource(t, "src")
	server := NewServer()

	w := dryRunRequest(t, server, FilterDryRunRequest{SourceName: "src", Attributes: []string{"group_title"}})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp FilterDryRunSummaryResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.NoArchive {
		t.Errorf("expected no_archive=true, got %+v", resp)
	}
}

func TestFilterDryRun_CombinedNoArchiveCondition(t *testing.T) {
	setupTestDB(t)
	setupDryRunSource(t, "src")
	server := NewServer()

	w := dryRunRequest(t, server, FilterDryRunRequest{SourceName: "src", Attributes: []string{"group_title", "tvg_name"}})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp FilterDryRunCombinedSummaryResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.NoArchive {
		t.Errorf("expected no_archive=true, got %+v", resp)
	}
}

func TestFilterDryRun_SuccessfulSummary(t *testing.T) {
	setupTestDB(t)
	archiveDir := setupDryRunSource(t, "src")
	buildM3UFixture(t, archiveDir, []struct{ tvgName, groupTitle string }{
		{"Channel A", "Movies HD"},
		{"Channel B", "Movies HD"},
		{"Channel C", "Adult XXX"},
	})
	server := NewServer()

	w := dryRunRequest(t, server, FilterDryRunRequest{
		SourceName:        "src",
		Attributes:        []string{"group_title"},
		GroupTitleExclude: "XXX",
	})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp FilterDryRunSummaryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.NoArchive {
		t.Fatalf("expected no_archive=false, got true")
	}
	if resp.TotalLines != 3 {
		t.Errorf("expected 3 total lines, got %d", resp.TotalLines)
	}
	if resp.MatchedCount != 2 || resp.ExcludedCount != 1 {
		t.Errorf("expected matched=2 excluded=1, got matched=%d excluded=%d", resp.MatchedCount, resp.ExcludedCount)
	}
	if len(resp.TopMatched) != 1 || resp.TopMatched[0].Value != "Movies HD" || resp.TopMatched[0].Count != 2 {
		t.Errorf("unexpected top matched values: %+v", resp.TopMatched)
	}
	if len(resp.TopExcluded) != 1 || resp.TopExcluded[0].Value != "Adult XXX" || resp.TopExcluded[0].Count != 1 {
		t.Errorf("unexpected top excluded values: %+v", resp.TopExcluded)
	}
}

func TestFilterDryRun_SuccessfulSearchWithTruncation(t *testing.T) {
	setupTestDB(t)
	archiveDir := setupDryRunSource(t, "src")

	entries := make([]struct{ tvgName, groupTitle string }, 0, 150)
	for i := 0; i < 150; i++ {
		entries = append(entries, struct{ tvgName, groupTitle string }{
			tvgName:    fmt.Sprintf("Sport Channel %d", i),
			groupTitle: "Sports",
		})
	}
	buildM3UFixture(t, archiveDir, entries)
	server := NewServer()

	w := dryRunRequest(t, server, FilterDryRunRequest{
		SourceName: "src",
		Attributes: []string{"tvg_name"},
		Search:     "Sport",
	})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp FilterDryRunSearchResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(resp.Results) != 100 {
		t.Errorf("expected 100 results, got %d", len(resp.Results))
	}
	if !resp.Truncated {
		t.Errorf("expected truncated=true")
	}
	for _, r := range resp.Results {
		if !r.Matched {
			t.Errorf("expected every result matched (no patterns supplied), got %+v", r)
		}
	}
}

func TestFilterDryRun_CombinedSummary(t *testing.T) {
	setupTestDB(t)
	archiveDir := setupDryRunSource(t, "src")
	buildM3UFixture(t, archiveDir, []struct{ tvgName, groupTitle string }{
		{"Good Channel", "Movies HD"}, // kept: passes both
		{"Bad Channel", "Movies HD"},  // excluded by tvg_name only
		{"Good Channel", "Adult XXX"}, // excluded by group_title only
		{"Bad Channel", "Adult XXX"},  // excluded by both
	})
	server := NewServer()

	w := dryRunRequest(t, server, FilterDryRunRequest{
		SourceName:        "src",
		Attributes:        []string{"group_title", "tvg_name"},
		GroupTitleExclude: "XXX",
		TvgNameExclude:    "Bad",
	})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp FilterDryRunCombinedSummaryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.TotalLines != 4 {
		t.Errorf("expected 4 total lines, got %d", resp.TotalLines)
	}
	if resp.KeptCount != 1 {
		t.Errorf("expected kept=1, got %d", resp.KeptCount)
	}
	if resp.ExcludedByGroupTitleOnly != 1 {
		t.Errorf("expected excluded_by_group_title_only=1, got %d", resp.ExcludedByGroupTitleOnly)
	}
	if resp.ExcludedByTvgNameOnly != 1 {
		t.Errorf("expected excluded_by_tvg_name_only=1, got %d", resp.ExcludedByTvgNameOnly)
	}
	if resp.ExcludedByBoth != 1 {
		t.Errorf("expected excluded_by_both=1, got %d", resp.ExcludedByBoth)
	}
	sum := resp.KeptCount + resp.ExcludedByGroupTitleOnly + resp.ExcludedByTvgNameOnly + resp.ExcludedByBoth
	if sum != resp.TotalLines {
		t.Errorf("expected counts to sum to total_lines=%d, got %d", resp.TotalLines, sum)
	}
	if len(resp.GroupTitleTopMatched) != 1 || resp.GroupTitleTopMatched[0].Value != "Movies HD" || resp.GroupTitleTopMatched[0].Count != 2 {
		t.Errorf("unexpected group_title top matched: %+v", resp.GroupTitleTopMatched)
	}
	if len(resp.TvgNameTopExcluded) != 1 || resp.TvgNameTopExcluded[0].Value != "Bad Channel" || resp.TvgNameTopExcluded[0].Count != 2 {
		t.Errorf("unexpected tvg_name top excluded: %+v", resp.TvgNameTopExcluded)
	}
}

func TestFilterDryRun_CombinedSearchWithVerdict(t *testing.T) {
	setupTestDB(t)
	archiveDir := setupDryRunSource(t, "src")
	buildM3UFixture(t, archiveDir, []struct{ tvgName, groupTitle string }{
		{"Good Channel", "Movies HD"},
		{"Bad Channel", "Movies HD"},
	})
	server := NewServer()

	w := dryRunRequest(t, server, FilterDryRunRequest{
		SourceName:      "src",
		Attributes:      []string{"group_title", "tvg_name"},
		TvgNameExclude:  "Bad",
		Search:          "Channel",
		SearchAttribute: "tvg_name",
	})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp FilterDryRunSearchResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(resp.Results) != 2 {
		t.Fatalf("expected 2 results, got %d: %+v", len(resp.Results), resp.Results)
	}
	verdicts := map[string]string{}
	for _, r := range resp.Results {
		verdicts[r.TvgName] = r.Verdict
	}
	if verdicts["Good Channel"] != "kept" {
		t.Errorf("expected 'Good Channel' verdict=kept, got %q", verdicts["Good Channel"])
	}
	if verdicts["Bad Channel"] != "excluded_by_tvg_name" {
		t.Errorf("expected 'Bad Channel' verdict=excluded_by_tvg_name, got %q", verdicts["Bad Channel"])
	}
}

func TestBuildFilterDryRunSummaryResponse_DistinctValueCounts(t *testing.T) {
	lines := []models.ProcessedLine{
		{GroupTitle: "Movies HD", TvgName: "A"},
		{GroupTitle: "Movies HD", TvgName: "B"},
		{GroupTitle: "Movies 4K", TvgName: "C"},
		{GroupTitle: "Adult XXX", TvgName: "D"},
	}
	cf, err := filter.CompilePatterns(nil, []string{"XXX"})
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	resp := buildFilterDryRunSummaryResponse(lines, "group_title", cf)

	if resp.TotalLines != 4 {
		t.Errorf("expected 4 total lines, got %d", resp.TotalLines)
	}
	if resp.MatchedCount != 3 || resp.ExcludedCount != 1 {
		t.Errorf("expected matched=3 excluded=1, got matched=%d excluded=%d", resp.MatchedCount, resp.ExcludedCount)
	}
	if len(resp.TopMatched) != 2 {
		t.Fatalf("expected 2 distinct matched values, got %d: %+v", len(resp.TopMatched), resp.TopMatched)
	}
	if resp.TopMatched[0].Value != "Movies HD" || resp.TopMatched[0].Count != 2 {
		t.Errorf("expected 'Movies HD' first with count 2, got %+v", resp.TopMatched[0])
	}
	if len(resp.TopExcluded) != 1 || resp.TopExcluded[0].Value != "Adult XXX" || resp.TopExcluded[0].Count != 1 {
		t.Errorf("unexpected top excluded values: %+v", resp.TopExcluded)
	}
}

func TestBuildFilterDryRunSearchResponse_NoMatches(t *testing.T) {
	lines := []models.ProcessedLine{
		{GroupTitle: "Movies HD", TvgName: "A"},
		{GroupTitle: "News", TvgName: "B"},
	}
	cf, err := filter.CompilePatterns(nil, nil)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	resp := buildFilterDryRunSearchResponse(lines, "group_title", "does-not-exist", cf)
	if len(resp.Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(resp.Results))
	}
	if resp.Truncated {
		t.Errorf("expected truncated=false")
	}
}
