package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/downloader"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/glefebvre/stalkeer/internal/notifier"
	"github.com/glefebvre/stalkeer/internal/scheduler"
	"github.com/stretchr/testify/require"
)

// fakeMovieRescanner is a movieRescanner test double recording every call it
// receives and returning a configurable error.
type fakeMovieRescanner struct {
	mu    sync.Mutex
	calls []int
	err   error
}

func (f *fakeMovieRescanner) RescanMovie(ctx context.Context, movieID int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, movieID)
	return f.err
}

// fakeSeriesRescanner is a seriesRescanner test double recording every call
// it receives and returning a configurable error.
type fakeSeriesRescanner struct {
	mu    sync.Mutex
	calls []int
	err   error
}

func (f *fakeSeriesRescanner) RescanSeries(ctx context.Context, seriesID int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, seriesID)
	return f.err
}

func newFakeDownloadServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		w.Write([]byte("fake-video-content"))
	}))
	t.Cleanup(server.Close)
	return server
}

// 4.3 A Radarr rescan failure for a successfully-completed movie must not
// affect the run's reported statistics or exit status, and must be logged.
func TestRunDownloadWorkerPool_RescanFailureDoesNotAffectStats(t *testing.T) {
	db := setupDownloadTestDB(t)
	tempDir := t.TempDir()
	server := newFakeDownloadServer(t)
	buf := captureAppLoggerWarnings(t)

	movieLine := createTestProcessedLine(t, db, server.URL+"/movie.mp4")
	streams := []*scheduler.Stream{{
		Tier:          scheduler.Tier1,
		SourceKey:     "movie:1",
		RadarrMovieID: 501,
		Items: []scheduler.Item{{
			DisplayName: "Test Movie",
			BaseDestDir: filepath.Join(tempDir, "Test Movie"),
			Candidates:  []models.ProcessedLine{movieLine},
		}},
	}}

	sched := scheduler.NewScheduler(streams, 0)
	dl := downloader.New(5*time.Second, 1, 0)
	cfg := &config.Config{Downloads: config.DownloadsConfig{TempDir: tempDir}}

	radarr := &fakeMovieRescanner{err: errors.New("radarr: connection refused")}

	stats := runDownloadWorkerPool(context.Background(), sched, dl, cfg, 1, false, radarr, nil)

	require.Equal(t, 1, stats.Total)
	require.Equal(t, 1, stats.Downloaded)
	require.Equal(t, 0, stats.Failed, "a rescan notification failure must not count as a download failure")
	require.Equal(t, []int{501}, radarr.calls)
	require.Contains(t, buf.String(), "failed to notify radarr", "a rescan failure must be logged")
}

// 4.4 Radarr and Sonarr rescans are attempted independently: a failing Radarr
// rescan must not prevent the Sonarr rescan for a completed episode in the
// same run, and vice versa.
func TestRunDownloadWorkerPool_RadarrAndSonarrRescansAreIndependent(t *testing.T) {
	db := setupDownloadTestDB(t)
	tempDir := t.TempDir()
	server := newFakeDownloadServer(t)

	movieLine := createTestProcessedLine(t, db, server.URL+"/movie.mp4")
	epLine := createTestProcessedLine(t, db, server.URL+"/ep1.mp4")

	streams := []*scheduler.Stream{
		{
			Tier:          scheduler.Tier1,
			SourceKey:     "movie:1",
			RadarrMovieID: 601,
			Items: []scheduler.Item{{
				DisplayName: "Test Movie",
				BaseDestDir: filepath.Join(tempDir, "Test Movie"),
				Candidates:  []models.ProcessedLine{movieLine},
			}},
		},
		{
			Tier:      scheduler.Tier1,
			SourceKey: "series:602",
			SeriesID:  602,
			Season:    1,
			Items: []scheduler.Item{{
				DisplayName: "Show S01E01",
				BaseDestDir: filepath.Join(tempDir, "Show S01E01"),
				Candidates:  []models.ProcessedLine{epLine},
				Episode:     1,
			}},
		},
	}

	sched := scheduler.NewScheduler(streams, 0)
	dl := downloader.New(5*time.Second, 1, 0)
	cfg := &config.Config{Downloads: config.DownloadsConfig{TempDir: tempDir}}

	radarr := &fakeMovieRescanner{err: errors.New("radarr unreachable")}
	sonarr := &fakeSeriesRescanner{}

	stats := runDownloadWorkerPool(context.Background(), sched, dl, cfg, 2, false, radarr, sonarr)

	require.Equal(t, 2, stats.Downloaded)
	require.Equal(t, 0, stats.Failed)
	require.Equal(t, []int{601}, radarr.calls, "radarr rescan should still be attempted despite failing")
	require.Equal(t, []int{602}, sonarr.calls, "sonarr rescan must not be blocked by radarr's failure")
}

// 4.5 Multiple episodes of the same series completing in one run must
// trigger exactly one Sonarr RescanSeries call for that series.
func TestRunDownloadWorkerPool_MultipleEpisodesSameSeries_SingleRescanCall(t *testing.T) {
	db := setupDownloadTestDB(t)
	tempDir := t.TempDir()
	server := newFakeDownloadServer(t)

	ep1Line := createTestProcessedLine(t, db, server.URL+"/ep1.mp4")
	ep2Line := createTestProcessedLine(t, db, server.URL+"/ep2.mp4")

	streams := []*scheduler.Stream{{
		Tier:      scheduler.Tier1,
		SourceKey: "series:701",
		SeriesID:  701,
		Season:    1,
		Items: []scheduler.Item{
			{DisplayName: "Show S01E01", BaseDestDir: filepath.Join(tempDir, "Show S01E01"), Candidates: []models.ProcessedLine{ep1Line}, Episode: 1},
			{DisplayName: "Show S01E02", BaseDestDir: filepath.Join(tempDir, "Show S01E02"), Candidates: []models.ProcessedLine{ep2Line}, Episode: 2},
		},
	}}

	sched := scheduler.NewScheduler(streams, 0)
	dl := downloader.New(5*time.Second, 1, 0)
	cfg := &config.Config{Downloads: config.DownloadsConfig{TempDir: tempDir}}

	sonarr := &fakeSeriesRescanner{}

	stats := runDownloadWorkerPool(context.Background(), sched, dl, cfg, 1, false, nil, sonarr)

	require.Equal(t, 2, stats.Downloaded)
	require.Equal(t, []int{701}, sonarr.calls, "a series with multiple completed episodes must be rescanned exactly once per run")
}

// 4.6 No rescan call is attempted for a service that isn't configured for
// this run (represented by a nil rescanner), even when a stream carries a
// non-zero RadarrMovieID/SeriesID.
func TestRunDownloadWorkerPool_UnconfiguredServiceSkipsRescan(t *testing.T) {
	db := setupDownloadTestDB(t)
	tempDir := t.TempDir()
	server := newFakeDownloadServer(t)

	epLine := createTestProcessedLine(t, db, server.URL+"/ep1.mp4")
	streams := []*scheduler.Stream{{
		Tier:      scheduler.Tier1,
		SourceKey: "series:801",
		SeriesID:  801,
		Season:    1,
		Items: []scheduler.Item{{
			DisplayName: "Show S01E01",
			BaseDestDir: filepath.Join(tempDir, "Show S01E01"),
			Candidates:  []models.ProcessedLine{epLine},
			Episode:     1,
		}},
	}}

	sched := scheduler.NewScheduler(streams, 0)
	dl := downloader.New(5*time.Second, 1, 0)
	cfg := &config.Config{Downloads: config.DownloadsConfig{TempDir: tempDir}}

	radarr := &fakeMovieRescanner{}

	// Sonarr is not configured for this run: passed as a nil interface, as
	// cmd/download.go's Run does when cfg.Sonarr is unset.
	stats := runDownloadWorkerPool(context.Background(), sched, dl, cfg, 1, false, radarr, nil)

	require.Equal(t, 1, stats.Downloaded)
	require.Empty(t, radarr.calls, "radarr must not be called for a series completion")
}

func TestNotifyBuildStreamsFailure_BuildsExpectedEvent(t *testing.T) {
	fake := &fakeNotifier{}

	notifyBuildStreamsFailure(fake, errors.New("radarr: connection refused"))

	require.Len(t, fake.events, 1)
	event := fake.events[0]
	require.Equal(t, notifier.Critical, event.Severity)
	require.Contains(t, event.Message, "radarr: connection refused")
}

func TestNotifyDownloadRunResult_NoFailuresSkipsNotification(t *testing.T) {
	fake := &fakeNotifier{}

	notifyDownloadRunResult(fake, &downloadStats{Total: 5, Downloaded: 5, Failed: 0})

	require.Empty(t, fake.events, "a run with zero failures must not send a notification")
}

func TestNotifyDownloadRunResult_WithFailuresSendsEventWithCounts(t *testing.T) {
	fake := &fakeNotifier{}

	notifyDownloadRunResult(fake, &downloadStats{Total: 5, Downloaded: 3, Failed: 2})

	require.Len(t, fake.events, 1)
	event := fake.events[0]
	require.Equal(t, notifier.Warning, event.Severity)
	require.Contains(t, event.Message, "Total: 5")
	require.Contains(t, event.Message, "Downloaded: 3")
	require.Contains(t, event.Message, "Failed: 2")
}
