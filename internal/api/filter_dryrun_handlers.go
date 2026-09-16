package api

import (
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/filter"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/m3udownloader"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/glefebvre/stalkeer/internal/parser"
	"github.com/glefebvre/stalkeer/internal/settings"
)

// filterDryRunSearchCap is the maximum number of lines the content-search
// mode returns before reporting the result as truncated.
const filterDryRunSearchCap = 100

// filterDryRunTopValues is how many distinct attribute values the aggregate
// summary reports on each side (matched/excluded).
const filterDryRunTopValues = 20

// filterDryRun handles POST /api/v1/filters/dryrun: evaluates a
// caller-supplied (not necessarily saved) attribute/pattern combination
// against the identified source's most recently downloaded archive, without
// touching the database, config.yml, or triggering a live download. See
// design.md's "Reuse settings.EffectiveSources() to resolve and validate
// source_name" and "New exported, Manager-independent pattern evaluator".
func (s *Server) filterDryRun(c *gin.Context) {
	var req FilterDryRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	if errCode, msg := validateFilterDryRunRequest(req); errCode != "" {
		respondError(c, http.StatusBadRequest, errCode, msg)
		return
	}

	src, ok := findEffectiveSource(req.SourceName)
	if !ok {
		respondError(c, http.StatusNotFound, "not_found", "unknown source")
		return
	}

	compiled := make(map[string]*filter.CompiledFilter, len(req.Attributes))
	for _, attribute := range req.Attributes {
		cf, err := compileAttributeFilter(req, attribute)
		if err != nil {
			respondError(c, http.StatusBadRequest, "invalid_pattern", err.Error())
			return
		}
		compiled[attribute] = cf
	}

	_, archiveDir := m3udownloader.SourcePaths(src.FilePath, src.Download.ArchiveDir, src.Name)
	latest, err := m3udownloader.NewArchiveManager(archiveDir, logger.AppLogger()).GetLatestArchive()
	if err != nil {
		if req.Search != "" {
			c.JSON(http.StatusOK, FilterDryRunSearchResponse{NoArchive: true})
		} else if len(req.Attributes) == 2 {
			c.JSON(http.StatusOK, FilterDryRunCombinedSummaryResponse{NoArchive: true})
		} else {
			c.JSON(http.StatusOK, FilterDryRunSummaryResponse{NoArchive: true})
		}
		return
	}

	lines, err := parser.NewParser(latest.Path, req.SourceName).Parse()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "parse_failed", "failed to parse the source's archive")
		return
	}

	if len(req.Attributes) == 2 {
		cfGroupTitle, cfTvgName := compiled["group_title"], compiled["tvg_name"]
		if req.Search != "" {
			c.JSON(http.StatusOK, buildFilterDryRunCombinedSearchResponse(lines, req.SearchAttribute, req.Search, cfGroupTitle, cfTvgName))
			return
		}
		c.JSON(http.StatusOK, buildFilterDryRunCombinedSummaryResponse(lines, cfGroupTitle, cfTvgName))
		return
	}

	attribute := req.Attributes[0]
	cf := compiled[attribute]
	if req.Search != "" {
		c.JSON(http.StatusOK, buildFilterDryRunSearchResponse(lines, attribute, req.Search, cf))
		return
	}

	c.JSON(http.StatusOK, buildFilterDryRunSummaryResponse(lines, attribute, cf))
}

// validateFilterDryRunRequest checks req against the filter-dry-run-test
// spec's Input Validation requirement (source name is checked separately,
// once the effective source list is resolved). Returns a non-empty error
// code and message when the request must be rejected.
func validateFilterDryRunRequest(req FilterDryRunRequest) (code, message string) {
	if len(req.Attributes) == 0 {
		return "invalid_attribute", "at least one of 'group_title' or 'tvg_name' must be supplied in attributes"
	}
	if len(req.Attributes) > 2 {
		return "invalid_attribute", "attributes must contain at most 'group_title' and 'tvg_name'"
	}

	seen := make(map[string]bool, len(req.Attributes))
	for _, attribute := range req.Attributes {
		if attribute != "group_title" && attribute != "tvg_name" {
			return "invalid_attribute", "attributes must be 'group_title' or 'tvg_name'"
		}
		if seen[attribute] {
			return "invalid_attribute", "attributes must not contain duplicates"
		}
		seen[attribute] = true
	}

	if req.Search == "" {
		return "", ""
	}

	if len(req.Attributes) == 2 {
		if req.SearchAttribute != "group_title" && req.SearchAttribute != "tvg_name" {
			return "invalid_search_attribute", "search_attribute must be 'group_title' or 'tvg_name'"
		}
		if !seen[req.SearchAttribute] {
			return "invalid_search_attribute", "search_attribute must name an attribute supplied in attributes"
		}
	}

	return "", ""
}

// compileAttributeFilter compiles the include/exclude patterns req supplied
// for the given attribute ("group_title" or "tvg_name").
func compileAttributeFilter(req FilterDryRunRequest, attribute string) (*filter.CompiledFilter, error) {
	if attribute == "tvg_name" {
		return filter.CompilePatterns(splitFilterPatterns(req.TvgNameInclude), splitFilterPatterns(req.TvgNameExclude))
	}
	return filter.CompilePatterns(splitFilterPatterns(req.GroupTitleInclude), splitFilterPatterns(req.GroupTitleExclude))
}

// findEffectiveSource resolves name against the effective M3U source list
// (runtime override if present, otherwise origin config.yml).
func findEffectiveSource(name string) (config.M3USourceConfig, bool) {
	for _, src := range settings.EffectiveSources() {
		if src.Name == name {
			return src, true
		}
	}
	return config.M3USourceConfig{}, false
}

// splitFilterPatterns splits a comma-separated pattern string the way
// CreateFilterDialog's fields are already edited: trimmed, empties dropped.
func splitFilterPatterns(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	patterns := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		patterns = append(patterns, p)
	}
	return patterns
}

// attributeValue returns the value of the tested attribute for a parsed line.
func attributeValue(line models.ProcessedLine, attribute string) string {
	if attribute == "tvg_name" {
		return line.TvgName
	}
	return line.GroupTitle
}

// buildFilterDryRunSummaryResponse evaluates every line against cf and
// aggregates the total/matched/excluded counts plus the top distinct values
// on each side, ordered by descending count.
func buildFilterDryRunSummaryResponse(lines []models.ProcessedLine, attribute string, cf *filter.CompiledFilter) FilterDryRunSummaryResponse {
	matchedCounts := make(map[string]int)
	excludedCounts := make(map[string]int)
	matched, excluded := 0, 0

	for _, line := range lines {
		value := attributeValue(line, attribute)
		if cf.Matches(value) {
			matched++
			matchedCounts[value]++
		} else {
			excluded++
			excludedCounts[value]++
		}
	}

	return FilterDryRunSummaryResponse{
		TotalLines:    len(lines),
		MatchedCount:  matched,
		ExcludedCount: excluded,
		TopMatched:    topValueCounts(matchedCounts, filterDryRunTopValues),
		TopExcluded:   topValueCounts(excludedCounts, filterDryRunTopValues),
	}
}

// buildFilterDryRunSearchResponse returns every line whose tested-attribute
// value contains search (case-insensitive), each tagged matched/excluded,
// capped at filterDryRunSearchCap results.
func buildFilterDryRunSearchResponse(lines []models.ProcessedLine, attribute, search string, cf *filter.CompiledFilter) FilterDryRunSearchResponse {
	searchLower := strings.ToLower(search)
	results := make([]FilterDryRunResultLine, 0)
	truncated := false

	for _, line := range lines {
		value := attributeValue(line, attribute)
		if !strings.Contains(strings.ToLower(value), searchLower) {
			continue
		}
		if len(results) >= filterDryRunSearchCap {
			truncated = true
			break
		}
		results = append(results, FilterDryRunResultLine{
			GroupTitle: line.GroupTitle,
			TvgName:    line.TvgName,
			Matched:    cf.Matches(value),
		})
	}

	return FilterDryRunSearchResponse{Results: results, Truncated: truncated}
}

// buildFilterDryRunCombinedSummaryResponse evaluates every line against both
// attributes' compiled filters, mirroring filter.Manager.ShouldProcess's
// AND-logic, and reports the cause-ventilated kept/excluded counts plus each
// attribute's own top distinct matched/excluded values.
func buildFilterDryRunCombinedSummaryResponse(lines []models.ProcessedLine, cfGroupTitle, cfTvgName *filter.CompiledFilter) FilterDryRunCombinedSummaryResponse {
	groupMatchedCounts := make(map[string]int)
	groupExcludedCounts := make(map[string]int)
	tvgMatchedCounts := make(map[string]int)
	tvgExcludedCounts := make(map[string]int)
	kept, excludedByGroupOnly, excludedByTvgOnly, excludedByBoth := 0, 0, 0, 0

	for _, line := range lines {
		groupMatches := cfGroupTitle.Matches(line.GroupTitle)
		tvgMatches := cfTvgName.Matches(line.TvgName)

		if groupMatches {
			groupMatchedCounts[line.GroupTitle]++
		} else {
			groupExcludedCounts[line.GroupTitle]++
		}
		if tvgMatches {
			tvgMatchedCounts[line.TvgName]++
		} else {
			tvgExcludedCounts[line.TvgName]++
		}

		switch {
		case groupMatches && tvgMatches:
			kept++
		case !groupMatches && tvgMatches:
			excludedByGroupOnly++
		case groupMatches && !tvgMatches:
			excludedByTvgOnly++
		default:
			excludedByBoth++
		}
	}

	return FilterDryRunCombinedSummaryResponse{
		TotalLines:               len(lines),
		KeptCount:                kept,
		ExcludedByGroupTitleOnly: excludedByGroupOnly,
		ExcludedByTvgNameOnly:    excludedByTvgOnly,
		ExcludedByBoth:           excludedByBoth,
		GroupTitleTopMatched:     topValueCounts(groupMatchedCounts, filterDryRunTopValues),
		GroupTitleTopExcluded:    topValueCounts(groupExcludedCounts, filterDryRunTopValues),
		TvgNameTopMatched:        topValueCounts(tvgMatchedCounts, filterDryRunTopValues),
		TvgNameTopExcluded:       topValueCounts(tvgExcludedCounts, filterDryRunTopValues),
	}
}

// combinedVerdict reports the combined kept/excluded verdict string for a
// line already evaluated against both attributes' compiled filters.
func combinedVerdict(groupMatches, tvgMatches bool) string {
	switch {
	case groupMatches && tvgMatches:
		return "kept"
	case !groupMatches && tvgMatches:
		return "excluded_by_group_title"
	case groupMatches && !tvgMatches:
		return "excluded_by_tvg_name"
	default:
		return "excluded_by_both"
	}
}

// buildFilterDryRunCombinedSearchResponse returns every line whose value for
// searchAttribute contains search (case-insensitive), each tagged with its
// combined verdict against both attributes' patterns, capped at
// filterDryRunSearchCap results.
func buildFilterDryRunCombinedSearchResponse(lines []models.ProcessedLine, searchAttribute, search string, cfGroupTitle, cfTvgName *filter.CompiledFilter) FilterDryRunSearchResponse {
	searchLower := strings.ToLower(search)
	results := make([]FilterDryRunResultLine, 0)
	truncated := false

	for _, line := range lines {
		value := attributeValue(line, searchAttribute)
		if !strings.Contains(strings.ToLower(value), searchLower) {
			continue
		}
		if len(results) >= filterDryRunSearchCap {
			truncated = true
			break
		}
		results = append(results, FilterDryRunResultLine{
			GroupTitle: line.GroupTitle,
			TvgName:    line.TvgName,
			Verdict:    combinedVerdict(cfGroupTitle.Matches(line.GroupTitle), cfTvgName.Matches(line.TvgName)),
		})
	}

	return FilterDryRunSearchResponse{Results: results, Truncated: truncated}
}

// topValueCounts sorts counts by descending count (ties broken
// alphabetically by value for deterministic output) and returns the top n.
func topValueCounts(counts map[string]int, n int) []FilterDryRunValueCount {
	values := make([]FilterDryRunValueCount, 0, len(counts))
	for value, count := range counts {
		values = append(values, FilterDryRunValueCount{Value: value, Count: count})
	}
	sort.Slice(values, func(i, j int) bool {
		if values[i].Count != values[j].Count {
			return values[i].Count > values[j].Count
		}
		return values[i].Value < values[j].Value
	})
	if len(values) > n {
		values = values[:n]
	}
	return values
}
