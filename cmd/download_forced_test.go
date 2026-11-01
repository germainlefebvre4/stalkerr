package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/downloader"
	"github.com/glefebvre/stalkeer/internal/external/radarr"
	"github.com/glefebvre/stalkeer/internal/external/sonarr"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/glefebvre/stalkeer/internal/scheduler"
	"github.com/stretchr/testify/require"
)

type emptyRadarr struct{}

func (emptyRadarr) GetMissingMovies(context.Context, radarr.FetchOptions) ([]radarr.Movie, error) {
	return nil, nil
}
func (emptyRadarr) GetMovieByTMDBID(context.Context, int) (*radarr.Movie, error) { return nil, nil }

type emptySonarr struct{}

func (emptySonarr) GetMissingEpisodes(context.Context, sonarr.FetchOptions) ([]sonarr.Episode, error) {
	return nil, nil
}
func (emptySonarr) GetSeriesDetails(context.Context, int) (*sonarr.Series, error) { return nil, nil }
func (emptySonarr) GetSeriesByTVDBID(context.Context, int) (*sonarr.Series, error) {
	return nil, nil
}

// 3.1 A forced 1080p is downloaded to its recorded path even when a better-ranked
// 720p sibling candidate is present, and the sibling is never attempted.
func TestDownloadItem_ForcedOccurrenceAndDestination(t *testing.T) {
	db := setupDownloadTestDB(t)
	tempDir := t.TempDir()

	var hits []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits = append(hits, r.URL.Path)
		w.Header().Set("Content-Type", "video/mp4")
		w.Write([]byte("fake-content"))
	}))
	defer server.Close()

	forced := createTestProcessedLine(t, db, server.URL+"/1080p.mp4")
	forced.Resolution = strPtrDL("1080p")
	forced.Language = strPtrDL("MULTI")
	better := createTestProcessedLine(t, db, server.URL+"/720p.mp4")
	better.Resolution = strPtrDL("720p")
	better.Language = strPtrDL("VF")

	item := &scheduler.Item{
		DisplayName:    "Luca",
		BaseDestDir:    filepath.Join(tempDir, "Luca"),
		Candidates:     []models.ProcessedLine{better, forced},
		ForcedLine:     &forced,
		ForcedDestPath: filepath.Join(tempDir, "Luca (2021) [1080p][MULTI]"),
	}

	dl := downloader.New(5*time.Second, 1, 0)
	cfg := &config.Config{Downloads: config.DownloadsConfig{TempDir: tempDir}}

	outcome, _ := downloadItem(context.Background(), dl, cfg, item, false)

	require.Equal(t, itemDownloaded, outcome)
	require.FileExists(t, filepath.Join(tempDir, "Luca (2021) [1080p][MULTI].mp4"))
	require.Equal(t, []string{"/1080p.mp4"}, hits, "the 720p sibling must not be requested")
}

// A failing forced occurrence does not fall through to siblings.
func TestDownloadItem_ForcedFailureDoesNotFallThrough(t *testing.T) {
	db := setupDownloadTestDB(t)
	tempDir := t.TempDir()

	var hits []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits = append(hits, r.URL.Path)
		if r.URL.Path == "/1080p.mp4" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "video/mp4")
		w.Write([]byte("fake-content"))
	}))
	defer server.Close()

	forced := createTestProcessedLine(t, db, server.URL+"/1080p.mp4")
	sibling := createTestProcessedLine(t, db, server.URL+"/720p.mp4")

	item := &scheduler.Item{
		DisplayName:    "Luca",
		BaseDestDir:    filepath.Join(tempDir, "Luca"),
		Candidates:     []models.ProcessedLine{sibling, forced},
		ForcedLine:     &forced,
		ForcedDestPath: filepath.Join(tempDir, "Luca [1080p]"),
	}

	dl := downloader.New(5*time.Second, 1, 0)
	cfg := &config.Config{Downloads: config.DownloadsConfig{TempDir: tempDir}}

	outcome, _ := downloadItem(context.Background(), dl, cfg, item, false)

	require.Equal(t, itemFailed, outcome)
	require.Equal(t, []string{"/1080p.mp4"}, hits)
}

// 3.2 End to end: scheduler build + executor for a forced occurrence whose
// sibling is already downloaded.
func TestForcedDownload_EndToEndWithDownloadedSibling(t *testing.T) {
	db := setupDownloadTestDB(t)
	tempDir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		w.Write([]byte("fake-content"))
	}))
	defer server.Close()

	movie := models.Movie{TMDBID: 220, TVDBID: intPtrDL(1220), TMDBTitle: "Luca", TMDBYear: 2021}
	require.NoError(t, db.Create(&movie).Error)

	siblingURL := server.URL + "/720p.mp4"
	sibling := models.ProcessedLine{
		LineContent: "s", LineHash: "sib", TvgName: "x", GroupTitle: "g", ProcessedAt: time.Now(),
		ContentType: models.ContentTypeMovies, State: models.StateDownloaded, MovieID: &movie.ID,
		LineURL: &siblingURL, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&sibling).Error)

	forcedURL := server.URL + "/1080p.mp4"
	forced := models.ProcessedLine{
		LineContent: "f", LineHash: "frc", TvgName: "x", GroupTitle: "g", ProcessedAt: time.Now(),
		ContentType: models.ContentTypeMovies, State: models.StateProcessed, MovieID: &movie.ID,
		Resolution: strPtrDL("1080p"), Language: strPtrDL("MULTI"),
		LineURL: &forcedURL, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&forced).Error)

	recorded := filepath.Join(tempDir, "movies", "Luca (2021)", "Luca (2021) [1080p][MULTI]")
	info := models.DownloadInfo{URL: forcedURL, Status: string(models.DownloadStatusPending), IsForced: true,
		DownloadPath: &recorded, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	require.NoError(t, db.Create(&info).Error)
	require.NoError(t, db.Model(&forced).Update("download_info_id", info.ID).Error)

	cfg := &config.Config{Downloads: config.DownloadsConfig{
		MoviesPath: filepath.Join(tempDir, "movies"), TVShowsPath: filepath.Join(tempDir, "tv"),
		TempDir: filepath.Join(tempDir, "tmp"), MaxRetryAttempts: 5, LockTimeoutMinutes: 5,
	}}
	streams, err := scheduler.BuildStreams(context.Background(), scheduler.BuildDeps{
		Config: cfg, Radarr: emptyRadarr{}, Sonarr: emptySonarr{}, DB: db,
		StateManager: downloader.NewStateManager(downloader.DefaultStateManagerConfig()),
	})
	require.NoError(t, err)
	require.Len(t, streams, 1)
	require.Len(t, streams[0].Items, 1)

	dl := downloader.New(5*time.Second, 1, 0)
	outcome, _ := downloadItem(context.Background(), dl, cfg, &streams[0].Items[0], false)
	require.Equal(t, itemDownloaded, outcome)

	require.FileExists(t, recorded+".mp4")

	var infoAfter models.DownloadInfo
	require.NoError(t, db.First(&infoAfter, info.ID).Error)
	require.Equal(t, string(models.DownloadStatusCompleted), infoAfter.Status)

	var forcedAfter, siblingAfter models.ProcessedLine
	require.NoError(t, db.First(&forcedAfter, forced.ID).Error)
	require.NoError(t, db.First(&siblingAfter, sibling.ID).Error)
	require.Equal(t, models.StateDownloaded, forcedAfter.State)
	require.Equal(t, models.StateDownloaded, siblingAfter.State)
	require.Nil(t, siblingAfter.DownloadInfoID, "sibling untouched")
}

func intPtrDL(i int) *int { return &i }
