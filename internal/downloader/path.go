package downloader

import (
	"fmt"
	"path/filepath"
)

func buildMovieBasePath(basePath, title string, year int) string {
	dir := fmt.Sprintf("%s (%d)", SanitizeFilename(title), year)
	return filepath.Join(basePath, dir, dir)
}

func buildTVShowBasePath(basePath, seriesTitle string, year, season, episode int) string {
	seriesDir := fmt.Sprintf("%s (%d)", SanitizeFilename(seriesTitle), year)
	seasonDir := fmt.Sprintf("Season %02d", season)
	fileName := fmt.Sprintf("%s (%d) - S%02dE%02d", SanitizeFilename(seriesTitle), year, season, episode)
	return filepath.Join(basePath, seriesDir, seasonDir, fileName)
}

// SanitizeFilename replaces filesystem-unsafe characters in name with underscores.
func SanitizeFilename(name string) string {
	replacer := map[rune]rune{
		'/':  '_',
		'\\': '_',
		':':  '_',
		'*':  '_',
		'?':  '_',
		'"':  '_',
		'<':  '_',
		'>':  '_',
		'|':  '_',
	}

	result := []rune(name)
	for i, r := range result {
		if replacement, ok := replacer[r]; ok {
			result[i] = replacement
		}
	}
	return string(result)
}
