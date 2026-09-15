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

	if req.Attribute != "group_title" && req.Attribute != "tvg_name" {
		respondError(c, http.StatusBadRequest, "invalid_attribute", "attribute must be 'group_title' or 'tvg_name'")
		return
	}

	src, ok := findEffectiveSource(req.SourceName)
	if !ok {
		respondError(c, http.StatusNotFound, "not_found", "unknown source")
		return
	}

	includePatterns := splitFilterPatterns(req.IncludePatterns)
	excludePatterns := splitFilterPatterns(req.ExcludePatterns)
	cf, err := filter.CompilePatterns(includePatterns, excludePatterns)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_pattern", err.Error())
		return
	}

	_, archiveDir := m3udownloader.SourcePaths(src.FilePath, src.Download.ArchiveDir, src.Name)
	latest, err := m3udownloader.NewArchiveManager(archiveDir, logger.AppLogger()).GetLatestArchive()
	if err != nil {
		if req.Search != "" {
			c.JSON(http.StatusOK, FilterDryRunSearchResponse{NoArchive: true})
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

	if req.Search != "" {
		c.JSON(http.StatusOK, buildFilterDryRunSearchResponse(lines, req.Attribute, req.Search, cf))
		return
	}

	c.JSON(http.StatusOK, buildFilterDryRunSummaryResponse(lines, req.Attribute, cf))
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
