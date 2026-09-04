package downloader

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildSonarrDestPath_UseSeriesPath(t *testing.T) {
	t.Run("primary root folder", func(t *testing.T) {
		got, fallback := BuildSonarrDestPath("/downloads/sonarr/Breaking Bad", "./data/sonarr", "Breaking Bad", 2008, 1, 1)
		if fallback {
			t.Error("expected no fallback")
		}
		want := filepath.Join("/downloads/sonarr/Breaking Bad", "Season 01", "Breaking Bad - S01E01")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("secondary root folder (sonarr-bis)", func(t *testing.T) {
		got, fallback := BuildSonarrDestPath("/downloads/sonarr-bis/Malcolm in the Middle", "./data/sonarr", "Malcolm in the Middle", 2000, 1, 1)
		if fallback {
			t.Error("expected no fallback")
		}
		if !strings.HasPrefix(got, "/downloads/sonarr-bis/Malcolm in the Middle") {
			t.Errorf("expected path to start with /downloads/sonarr-bis/Malcolm in the Middle, got %q", got)
		}
		want := filepath.Join("/downloads/sonarr-bis/Malcolm in the Middle", "Season 01", "Malcolm in the Middle - S01E01")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("season and episode zero-padding", func(t *testing.T) {
		got, _ := BuildSonarrDestPath("/downloads/sonarr/Show", "./data/sonarr", "Show", 0, 3, 12)
		if !strings.HasSuffix(got, "Season 03"+string(filepath.Separator)+"Show - S03E12") {
			t.Errorf("unexpected path suffix, got %q", got)
		}
	})
}

func TestBuildSonarrDestPath_EmptyPathFallback(t *testing.T) {
	t.Run("includes year when available", func(t *testing.T) {
		got, fallback := BuildSonarrDestPath("", "./data/sonarr", "My Show", 2020, 2, 5)
		if !fallback {
			t.Error("expected fallback=true when seriesPath is empty")
		}
		want := filepath.Join("./data/sonarr", "My Show (2020)", "Season 02", "My Show - S02E05")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("keeps previous behavior when year is unavailable", func(t *testing.T) {
		got, fallback := BuildSonarrDestPath("", "./data/sonarr", "My Show", 0, 2, 5)
		if !fallback {
			t.Error("expected fallback=true when seriesPath is empty")
		}
		want := filepath.Join("./data/sonarr", "My Show", "Season 02", "My Show - S02E05")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestBuildRadarrDestPath_UseMoviePath(t *testing.T) {
	t.Run("primary root folder", func(t *testing.T) {
		got, fallback := BuildRadarrDestPath("/downloads/radarr/The Matrix (1999)", "./data/radarr", "The Matrix", 1999)
		if fallback {
			t.Error("expected no fallback")
		}
		want := filepath.Join("/downloads/radarr/The Matrix (1999)", "The Matrix (1999)")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("secondary root folder (4k)", func(t *testing.T) {
		got, fallback := BuildRadarrDestPath("/downloads/radarr-4k/Inception (2010)", "./data/radarr", "Inception", 2010)
		if fallback {
			t.Error("expected no fallback")
		}
		want := filepath.Join("/downloads/radarr-4k/Inception (2010)", "Inception (2010)")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestBuildRadarrDestPath_EmptyPathFallback(t *testing.T) {
	got, fallback := BuildRadarrDestPath("", "./data/radarr", "Dune", 2021)
	if !fallback {
		t.Error("expected fallback=true when moviePath is empty")
	}
	want := filepath.Join("./data/radarr", "Dune (2021)", "Dune (2021)")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func strPtr(s string) *string { return &s }

func TestBuildRadarrDestPathWithResolution_DifferentResolutionsDiffer(t *testing.T) {
	hd, _ := BuildRadarrDestPathWithResolution("/downloads/radarr/Dune (2021)", "./data/radarr", "Dune", 2021, strPtr("1080p"), nil, nil, "occurrence-1")
	sd, _ := BuildRadarrDestPathWithResolution("/downloads/radarr/Dune (2021)", "./data/radarr", "Dune", 2021, strPtr("480p"), nil, nil, "occurrence-2")

	if hd == sd {
		t.Fatalf("expected different paths for different resolutions, both got %q", hd)
	}

	base, _ := BuildRadarrDestPath("/downloads/radarr/Dune (2021)", "./data/radarr", "Dune", 2021)
	if !strings.HasPrefix(hd, base) || hd == base {
		t.Errorf("expected suffixed path to extend the unsuffixed base %q, got %q", base, hd)
	}
}

func TestBuildRadarrDestPathWithResolution_NilResolutionUsesFallbackMarker(t *testing.T) {
	sibling, _ := BuildRadarrDestPathWithResolution("/downloads/radarr/Dune (2021)", "./data/radarr", "Dune", 2021, strPtr("1080p"), nil, nil, "occurrence-1")
	unknown, _ := BuildRadarrDestPathWithResolution("/downloads/radarr/Dune (2021)", "./data/radarr", "Dune", 2021, nil, nil, nil, "occurrence-2")

	if sibling == unknown {
		t.Fatalf("expected nil-resolution path to differ from a sibling's resolved path, both got %q", sibling)
	}
	if !strings.Contains(unknown, "occurrence-2") {
		t.Errorf("expected fallback marker in path, got %q", unknown)
	}
}

func TestBuildRadarrDestPathWithResolution_LanguageAndVariantTags(t *testing.T) {
	got, _ := BuildRadarrDestPathWithResolution("/downloads/radarr/Dune (2021)", "./data/radarr", "Dune", 2021, strPtr("1080p"), strPtr("MULTI"), strPtr("VFQ"), "occurrence-1")
	if !strings.HasSuffix(got, "[1080p][MULTI][VFQ]") {
		t.Errorf("expected path to end with resolution/language/variant tags, got %q", got)
	}
}

func TestBuildSonarrDestPathWithResolution_DifferentResolutionsDiffer(t *testing.T) {
	hd, _ := BuildSonarrDestPathWithResolution("/downloads/sonarr/Breaking Bad", "./data/sonarr", "Breaking Bad", 2008, 1, 1, strPtr("1080p"), nil, nil, "occurrence-1")
	sd, _ := BuildSonarrDestPathWithResolution("/downloads/sonarr/Breaking Bad", "./data/sonarr", "Breaking Bad", 2008, 1, 1, strPtr("480p"), nil, nil, "occurrence-2")

	if hd == sd {
		t.Fatalf("expected different paths for different resolutions, both got %q", hd)
	}
}

func TestBuildSonarrDestPathWithResolution_NilResolutionUsesFallbackMarker(t *testing.T) {
	sibling, _ := BuildSonarrDestPathWithResolution("/downloads/sonarr/Breaking Bad", "./data/sonarr", "Breaking Bad", 2008, 1, 1, strPtr("1080p"), nil, nil, "occurrence-1")
	unknown, _ := BuildSonarrDestPathWithResolution("/downloads/sonarr/Breaking Bad", "./data/sonarr", "Breaking Bad", 2008, 1, 1, nil, nil, nil, "occurrence-2")

	if sibling == unknown {
		t.Fatalf("expected nil-resolution path to differ from a sibling's resolved path, both got %q", sibling)
	}
	if !strings.Contains(unknown, "occurrence-2") {
		t.Errorf("expected fallback marker in path, got %q", unknown)
	}
}

func TestBuildSonarrDestPathWithResolution_LanguageAndVariantTags(t *testing.T) {
	got, _ := BuildSonarrDestPathWithResolution("/downloads/sonarr/Breaking Bad", "./data/sonarr", "Breaking Bad", 2008, 1, 1, strPtr("1080p"), strPtr("MULTI"), strPtr("VFQ"), "occurrence-1")
	if !strings.HasSuffix(got, "[1080p][MULTI][VFQ]") {
		t.Errorf("expected path to end with resolution/language/variant tags, got %q", got)
	}
}

func TestQualityTags(t *testing.T) {
	tests := []struct {
		name          string
		resolution    *string
		language      *string
		frenchVariant *string
		want          string
	}{
		{name: "resolution only", resolution: strPtr("1080p"), want: "[1080p]"},
		{name: "language only", language: strPtr("MULTI"), want: "[MULTI]"},
		{name: "resolution and language", resolution: strPtr("1080p"), language: strPtr("MULTI"), want: "[1080p][MULTI]"},
		{name: "resolution, language, and VFQ", resolution: strPtr("720p"), language: strPtr("VF"), frenchVariant: strPtr("VFQ"), want: "[720p][VF][VFQ]"},
		{name: "VFQ only, no language", frenchVariant: strPtr("VFQ"), want: "[VFQ]"},
		{name: "all nil produces no tags", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := QualityTags(tt.resolution, tt.language, tt.frenchVariant)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestQualityTagsAppendedToAutomaticPipelineBaseDestDir(t *testing.T) {
	// Mirrors how cmd/download.go's downloadItem appends QualityTags directly to
	// an item's already-known, untagged BaseDestDir once the attempted
	// candidate's quality info is known.
	baseDestDir, _ := BuildRadarrDestPath("/downloads/radarr/Dune (2021)", "./data/radarr", "Dune", 2021)

	t.Run("all quality info known", func(t *testing.T) {
		got := baseDestDir + QualityTags(strPtr("1080p"), strPtr("MULTI"), strPtr("VFQ"))
		want := baseDestDir + "[1080p][MULTI][VFQ]"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("all nil leaves base path unchanged", func(t *testing.T) {
		got := baseDestDir + QualityTags(nil, nil, nil)
		if got != baseDestDir {
			t.Errorf("expected untagged path %q, got %q", baseDestDir, got)
		}
	})
}
