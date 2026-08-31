package downloader

import (
	"path/filepath"
	"testing"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/models"
)

func TestNormalizeContentType(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"movies", string(models.ContentTypeMovies)},
		{"radarr", string(models.ContentTypeMovies)},
		{"tvshows", string(models.ContentTypeTVShows)},
		{"sonarr", string(models.ContentTypeTVShows)},
		{"unknown", ""},
	}

	for _, tc := range cases {
		if got := normalizeContentType(tc.input); got != tc.expected {
			t.Fatalf("expected %q, got %q for input %q", tc.expected, got, tc.input)
		}
	}
}

func TestHasContentType(t *testing.T) {
	download := &models.DownloadInfo{
		ProcessedLines: []models.ProcessedLine{
			{ContentType: models.ContentTypeMovies},
			{ContentType: models.ContentTypeTVShows},
		},
	}

	if !hasContentType(download, string(models.ContentTypeTVShows)) {
		t.Fatalf("expected tvshows content type to match")
	}

	if hasContentType(download, "channels") {
		t.Fatalf("expected channels content type to not match")
	}
}

func TestBuildBaseDestPath_PrefersRecordedDownloadPathOverRecomputing(t *testing.T) {
	rh := &ResumeHelper{}
	cfg := &config.Config{}
	cfg.Downloads.MoviesPath = "./data/movies"

	suffixedPath := filepath.Join("/downloads/radarr/Dune (2021)", "Dune (2021) [1080p]") + ".mkv"
	movieID := uint(1)
	line := &models.ProcessedLine{
		ContentType: models.ContentTypeMovies,
		MovieID:     &movieID,
		Movie:       &models.Movie{TMDBTitle: "Dune", TMDBYear: 2021},
	}
	download := &models.DownloadInfo{DownloadPath: &suffixedPath}

	got, _, err := rh.buildBaseDestPath(cfg, line, download)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := filepath.Join("/downloads/radarr/Dune (2021)", "Dune (2021) [1080p]")
	if got != want {
		t.Errorf("expected resume to reuse the recorded (suffixed) path %q, got %q", want, got)
	}

	recomputed := buildMovieBasePath(cfg.Downloads.MoviesPath, line.Movie.TMDBTitle, line.Movie.TMDBYear)
	if got == recomputed {
		t.Errorf("expected recorded path to differ from the recomputed (unsuffixed) path %q", recomputed)
	}
}

func TestBuildBaseDestPath_FallsBackToRecomputingWhenNoPathRecorded(t *testing.T) {
	rh := &ResumeHelper{}
	cfg := &config.Config{}
	cfg.Downloads.MoviesPath = "./data/movies"

	movieID := uint(1)
	line := &models.ProcessedLine{
		ContentType: models.ContentTypeMovies,
		MovieID:     &movieID,
		Movie:       &models.Movie{TMDBTitle: "Dune", TMDBYear: 2021},
	}
	download := &models.DownloadInfo{}

	got, _, err := rh.buildBaseDestPath(cfg, line, download)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := buildMovieBasePath(cfg.Downloads.MoviesPath, line.Movie.TMDBTitle, line.Movie.TMDBYear)
	if got != want {
		t.Errorf("expected fallback recomputed path %q, got %q", want, got)
	}
}
