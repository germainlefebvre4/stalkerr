package scheduler

import "github.com/glefebvre/stalkeer/internal/models"

// candidateRank returns a lower-is-better ordinal for a ProcessedLine's
// (language, resolution) pair, mirroring matcher.resolutionOrderSQL's
// language-then-resolution CASE ordering so tier-2 gating and the SQL
// candidate ordering never disagree.
func candidateRank(line *models.ProcessedLine) int {
	return languageRank(line.Language)*10 + resolutionRank(line.Resolution)
}

// languageRank maps a language string to sort priority: VF (1) is preferred
// first, then MULTI (2), unspecified/nil (3), then VOSTFR (4).
func languageRank(language *string) int {
	if language == nil {
		return 3
	}
	switch *language {
	case "VF":
		return 1
	case "MULTI":
		return 2
	case "VOSTFR":
		return 4
	default:
		return 3
	}
}

// resolutionRank maps a resolution string to sort priority: 720p (1) is
// preferred first, then 1080p, 4K, 480p, and unknown/nil last (5).
func resolutionRank(resolution *string) int {
	if resolution == nil {
		return 5
	}
	switch *resolution {
	case "720p":
		return 1
	case "1080p":
		return 2
	case "4K":
		return 3
	case "480p":
		return 4
	default:
		return 5
	}
}
