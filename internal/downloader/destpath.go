package downloader

import (
	"fmt"
	"path/filepath"
)

// BuildRadarrDestPath constructs the base destination path for a movie download.
// It uses moviePath (from the Radarr API) as the authoritative root directory.
// When moviePath is empty it falls back to joining fallbackBase with the standard
// movie directory name.
// The second return value is true when the fallback was used.
func BuildRadarrDestPath(moviePath, fallbackBase, movieTitle string, movieYear int) (string, bool) {
	fileBase := fmt.Sprintf("%s (%d)", SanitizeFilename(movieTitle), movieYear)
	root := moviePath
	usedFallback := false
	if root == "" {
		root = filepath.Join(fallbackBase, fileBase)
		usedFallback = true
	}
	return filepath.Join(root, fileBase), usedFallback
}

// BuildSonarrDestPath constructs the base destination path for a TV show episode download.
// It uses seriesPath (from the Sonarr API) as the authoritative root directory, which
// already encodes the correct Sonarr root folder. When seriesPath is empty it falls back
// to joining fallbackBase with a sanitised seriesTitle and year when available.
// The second return value is true when the fallback was used.
func BuildSonarrDestPath(seriesPath, fallbackBase, seriesTitle string, seriesYear, seasonNum, episodeNum int) (string, bool) {
	root := seriesPath
	usedFallback := false
	if root == "" {
		seriesDir := SanitizeFilename(seriesTitle)
		if seriesYear > 0 {
			seriesDir = fmt.Sprintf("%s (%d)", seriesDir, seriesYear)
		}
		root = filepath.Join(fallbackBase, seriesDir)
		usedFallback = true
	}
	return filepath.Join(
		root,
		fmt.Sprintf("Season %02d", seasonNum),
		fmt.Sprintf("%s - S%02dE%02d", SanitizeFilename(seriesTitle), seasonNum, episodeNum),
	), usedFallback
}

// resolutionSuffix returns the anti-collision suffix appended to a forced download's
// destination base path: the occurrence's resolution when known, or fallbackMarker
// (expected to be unique to the occurrence, e.g. derived from its ProcessedLine id)
// when resolution is nil or empty.
func resolutionSuffix(resolution *string, fallbackMarker string) string {
	if resolution != nil && *resolution != "" {
		return fmt.Sprintf(" [%s]", *resolution)
	}
	return fmt.Sprintf(" [%s]", fallbackMarker)
}

// BuildRadarrDestPathWithResolution builds the destination base path for a forced
// download of a movie occurrence, extending BuildRadarrDestPath with a resolution
// suffix (or fallbackMarker when resolution is unknown) so the resulting file can
// never silently collide with a sibling occurrence's file in the same folder. This
// variant is used only by the force-download path; the automatic pipeline keeps
// calling BuildRadarrDestPath directly.
func BuildRadarrDestPathWithResolution(moviePath, fallbackBase, movieTitle string, movieYear int, resolution *string, fallbackMarker string) (string, bool) {
	base, usedFallback := BuildRadarrDestPath(moviePath, fallbackBase, movieTitle, movieYear)
	return base + resolutionSuffix(resolution, fallbackMarker), usedFallback
}

// BuildSonarrDestPathWithResolution builds the destination base path for a forced
// download of a TV episode occurrence, extending BuildSonarrDestPath with a
// resolution suffix (or fallbackMarker when resolution is unknown) so the resulting
// file can never silently collide with a sibling occurrence's file in the same
// folder. This variant is used only by the force-download path; the automatic
// pipeline keeps calling BuildSonarrDestPath directly.
func BuildSonarrDestPathWithResolution(seriesPath, fallbackBase, seriesTitle string, seriesYear, seasonNum, episodeNum int, resolution *string, fallbackMarker string) (string, bool) {
	base, usedFallback := BuildSonarrDestPath(seriesPath, fallbackBase, seriesTitle, seriesYear, seasonNum, episodeNum)
	return base + resolutionSuffix(resolution, fallbackMarker), usedFallback
}
