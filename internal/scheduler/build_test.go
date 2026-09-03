package scheduler

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/downloader"
	"github.com/glefebvre/stalkeer/internal/external/radarr"
	"github.com/glefebvre/stalkeer/internal/external/sonarr"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.ProcessedLine{},
		&models.Movie{},
		&models.TVShow{},
		&models.DownloadInfo{},
	))
	database.SetDB(db)
	return db
}

func testDeps(db *gorm.DB, radarrClient RadarrClient, sonarrClient SonarrClient) BuildDeps {
	return BuildDeps{
		Config: &config.Config{
			Downloads: config.DownloadsConfig{
				MoviesPath:         "/media/movies",
				TVShowsPath:        "/media/tv",
				MaxRetryAttempts:   5,
				LockTimeoutMinutes: 5,
			},
		},
		Radarr:       radarrClient,
		Sonarr:       sonarrClient,
		DB:           db,
		StateManager: downloader.NewStateManager(downloader.DefaultStateManagerConfig()),
	}
}

type fakeRadarr struct {
	missing []radarr.Movie
}

func (f *fakeRadarr) GetMissingMovies(ctx context.Context, opts radarr.FetchOptions) ([]radarr.Movie, error) {
	return f.missing, nil
}

type fakeSonarr struct {
	missing []sonarr.Episode
	series  map[int]*sonarr.Series
}

func (f *fakeSonarr) GetMissingEpisodes(ctx context.Context, opts sonarr.FetchOptions) ([]sonarr.Episode, error) {
	return f.missing, nil
}

func (f *fakeSonarr) GetSeriesDetails(ctx context.Context, id int) (*sonarr.Series, error) {
	s, ok := f.series[id]
	if !ok {
		return nil, fmt.Errorf("series %d not found", id)
	}
	return s, nil
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

// 2.1 Fetch-and-match for missing movies and episodes, using mocked API responses.
func TestBuildTier1Streams_MockedAPI(t *testing.T) {
	db := setupTestDB(t)

	movie := models.Movie{TMDBID: 101, TVDBID: intPtr(12345), TMDBTitle: "Test Movie", TMDBYear: 2020}
	require.NoError(t, db.Create(&movie).Error)
	require.NoError(t, db.Create(&models.ProcessedLine{
		LineContent: "movie line", LineHash: "h1", TvgName: "x", GroupTitle: "g",
		ProcessedAt: time.Now(), ContentType: models.ContentTypeMovies, State: models.StateProcessed,
		MovieID: &movie.ID, LineURL: strPtr("http://example.com/movie.mkv"), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error)

	season := 1
	episode := 1
	show := models.TVShow{TMDBID: 201, TVDBID: intPtr(55555), TMDBTitle: "Test Show", TMDBYear: 2019, Season: &season, Episode: &episode}
	require.NoError(t, db.Create(&show).Error)
	require.NoError(t, db.Create(&models.ProcessedLine{
		LineContent: "ep line", LineHash: "h2", TvgName: "x", GroupTitle: "g",
		ProcessedAt: time.Now(), ContentType: models.ContentTypeTVShows, State: models.StateProcessed,
		TVShowID: &show.ID, LineURL: strPtr("http://example.com/ep.mkv"), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error)

	fr := &fakeRadarr{missing: []radarr.Movie{{ID: 1, Title: "Test Movie", Year: 2020, TvdbID: 12345, TMDBID: 101, Path: "/media/movies/Test Movie (2020)"}}}
	fs := &fakeSonarr{
		missing: []sonarr.Episode{{ID: 1, SeriesID: 10, SeasonNumber: 1, EpisodeNumber: 1, Title: "Pilot"}},
		series:  map[int]*sonarr.Series{10: {ID: 10, Title: "Test Show", Year: 2019, TvdbID: 55555, Path: "/media/tv/Test Show"}},
	}

	streams, err := BuildStreams(context.Background(), testDeps(db, fr, fs))
	require.NoError(t, err)
	require.Len(t, streams, 2)

	var movieStream, seriesStream *Stream
	for _, s := range streams {
		if s.SeriesID != 0 {
			seriesStream = s
		} else {
			movieStream = s
		}
	}

	require.NotNil(t, movieStream)
	require.Equal(t, Tier1, movieStream.Tier)
	require.Len(t, movieStream.Items, 1)
	require.Contains(t, movieStream.Items[0].BaseDestPath, "Test Movie (2020)")

	require.NotNil(t, seriesStream)
	require.Equal(t, Tier1, seriesStream.Tier)
	require.Equal(t, 1, seriesStream.Season)
	require.Len(t, seriesStream.Items, 1)
	require.Equal(t, 1, seriesStream.Items[0].Episode)
}

// 2.2 Tier classification: already-downloaded -> tier2, never-downloaded -> tier1.
func TestBuildStreams_TierClassification(t *testing.T) {
	db := setupTestDB(t)

	// Movie 1: never downloaded, missing per Radarr -> tier1.
	freshMovie := models.Movie{TMDBID: 1, TVDBID: intPtr(1001), TMDBTitle: "Fresh Movie", TMDBYear: 2021}
	require.NoError(t, db.Create(&freshMovie).Error)
	require.NoError(t, db.Create(&models.ProcessedLine{
		LineContent: "l1", LineHash: "l1", TvgName: "x", GroupTitle: "g", ProcessedAt: time.Now(),
		ContentType: models.ContentTypeMovies, State: models.StateProcessed, MovieID: &freshMovie.ID,
		LineURL: strPtr("http://example.com/fresh.mkv"), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error)

	// Movie 2: already has a downloaded line, plus an alternate candidate -> tier2.
	upgradeMovie := models.Movie{TMDBID: 2, TVDBID: intPtr(1002), TMDBTitle: "Upgrade Movie", TMDBYear: 2022}
	require.NoError(t, db.Create(&upgradeMovie).Error)
	require.NoError(t, db.Create(&models.ProcessedLine{
		LineContent: "l2a", LineHash: "l2a", TvgName: "x", GroupTitle: "g", ProcessedAt: time.Now(),
		ContentType: models.ContentTypeMovies, State: models.StateDownloaded, MovieID: &upgradeMovie.ID,
		LineURL: strPtr("http://example.com/upgrade-720p.mkv"), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error)
	require.NoError(t, db.Create(&models.ProcessedLine{
		LineContent: "l2b", LineHash: "l2b", TvgName: "x", GroupTitle: "g", ProcessedAt: time.Now(),
		ContentType: models.ContentTypeMovies, State: models.StateProcessed, MovieID: &upgradeMovie.ID,
		LineURL: strPtr("http://example.com/upgrade-1080p.mkv"), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error)

	fr := &fakeRadarr{missing: []radarr.Movie{{ID: 1, Title: "Fresh Movie", Year: 2021, TvdbID: 1001, TMDBID: 1}}}
	fs := &fakeSonarr{series: map[int]*sonarr.Series{}}

	streams, err := BuildStreams(context.Background(), testDeps(db, fr, fs))
	require.NoError(t, err)
	require.Len(t, streams, 2)

	tiers := map[string]Tier{}
	for _, s := range streams {
		tiers[s.SourceKey] = s.Tier
	}
	require.Equal(t, Tier1, tiers[fmt.Sprintf("movie:%d", freshMovie.ID)])
	require.Equal(t, Tier2, tiers[fmt.Sprintf("movie:%d", upgradeMovie.ID)])
}

// A movie still reported missing by Radarr, but already downloaded locally
// with another eligible candidate, must yield only its tier-1 stream - not a
// second tier-2 stream for the same SourceKey (see fix-dedupe-tier-streams).
func TestBuildStreams_DedupTier2AgainstTier1_Movie(t *testing.T) {
	db := setupTestDB(t)

	movie := models.Movie{TMDBID: 1, TVDBID: intPtr(1001), TMDBTitle: "Dune", TMDBYear: 2021}
	require.NoError(t, db.Create(&movie).Error)
	require.NoError(t, db.Create(&models.ProcessedLine{
		LineContent: "downloaded", LineHash: "d1", TvgName: "x", GroupTitle: "g", ProcessedAt: time.Now(),
		ContentType: models.ContentTypeMovies, State: models.StateDownloaded, MovieID: &movie.ID,
		LineURL: strPtr("http://example.com/dune-720p.mkv"), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error)
	require.NoError(t, db.Create(&models.ProcessedLine{
		LineContent: "alt", LineHash: "d2", TvgName: "x", GroupTitle: "g", ProcessedAt: time.Now(),
		ContentType: models.ContentTypeMovies, State: models.StateProcessed, MovieID: &movie.ID,
		LineURL: strPtr("http://example.com/dune-1080p.mkv"), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error)

	fr := &fakeRadarr{missing: []radarr.Movie{{ID: 1, Title: "Dune", Year: 2021, TvdbID: 1001, TMDBID: 1}}}
	fs := &fakeSonarr{series: map[int]*sonarr.Series{}}

	streams, err := BuildStreams(context.Background(), testDeps(db, fr, fs))
	require.NoError(t, err)
	require.Len(t, streams, 1, "movie missing per Radarr and already downloaded should yield a single stream")
	require.Equal(t, fmt.Sprintf("movie:%d", movie.ID), streams[0].SourceKey)
	require.Equal(t, Tier1, streams[0].Tier)
}

// A series-season with a missing episode per Sonarr, but with another episode
// already downloaded locally plus an eligible candidate, must yield only the
// tier-1 season stream - not also a tier-2 stream for the same season.
func TestBuildStreams_DedupTier2AgainstTier1_Series(t *testing.T) {
	db := setupTestDB(t)

	tvdbID := 777

	e1 := models.TVShow{TMDBID: 300, TVDBID: intPtr(tvdbID), TMDBTitle: "Show", TMDBYear: 2020, Season: intPtr(1), Episode: intPtr(1)}
	require.NoError(t, db.Create(&e1).Error)
	require.NoError(t, db.Create(&models.ProcessedLine{
		LineContent: "e1", LineHash: "e1", TvgName: "x", GroupTitle: "g", ProcessedAt: time.Now(),
		ContentType: models.ContentTypeTVShows, State: models.StateProcessed, TVShowID: &e1.ID,
		LineURL: strPtr("http://example.com/e1.mkv"), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error)

	e2 := models.TVShow{TMDBID: 300, TVDBID: intPtr(tvdbID), TMDBTitle: "Show", TMDBYear: 2020, Season: intPtr(1), Episode: intPtr(2)}
	require.NoError(t, db.Create(&e2).Error)
	require.NoError(t, db.Create(&models.ProcessedLine{
		LineContent: "e2-downloaded", LineHash: "e2a", TvgName: "x", GroupTitle: "g", ProcessedAt: time.Now(),
		ContentType: models.ContentTypeTVShows, State: models.StateDownloaded, TVShowID: &e2.ID,
		LineURL: strPtr("http://example.com/e2-720p.mkv"), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error)
	require.NoError(t, db.Create(&models.ProcessedLine{
		LineContent: "e2-alt", LineHash: "e2b", TvgName: "x", GroupTitle: "g", ProcessedAt: time.Now(),
		ContentType: models.ContentTypeTVShows, State: models.StateProcessed, TVShowID: &e2.ID,
		LineURL: strPtr("http://example.com/e2-1080p.mkv"), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error)

	fr := &fakeRadarr{}
	fs := &fakeSonarr{
		missing: []sonarr.Episode{{ID: 1, SeriesID: 20, SeasonNumber: 1, EpisodeNumber: 1}},
		series:  map[int]*sonarr.Series{20: {ID: 20, Title: "Show", Year: 2020, TvdbID: tvdbID}},
	}

	streams, err := BuildStreams(context.Background(), testDeps(db, fr, fs))
	require.NoError(t, err)
	require.Len(t, streams, 1, "series-season with a missing episode and an already-downloaded episode should yield a single stream")
	require.Equal(t, "series:20", streams[0].SourceKey)
	require.Equal(t, Tier1, streams[0].Tier)
	require.Equal(t, 1, streams[0].Season)
	require.Len(t, streams[0].Items, 1)
	require.Equal(t, 1, streams[0].Items[0].Episode)
}

// mergeIncompleteDownloads can synthesize a fresh tier-1 stream for a movie
// after tier-2 streams have already been built (a stuck download for a
// candidate outside the normal tier-1/tier-2 fetch). The synthesized tier-1
// stream must still win over the pre-existing tier-2 stream for that movie.
func TestBuildStreams_DedupTier2AgainstTier1_SynthesizedWins(t *testing.T) {
	db := setupTestDB(t)

	movie := models.Movie{TMDBID: 5, TVDBID: intPtr(5005), TMDBTitle: "Stuck Movie", TMDBYear: 2019}
	require.NoError(t, db.Create(&movie).Error)

	// Tier-2 eligibility: a downloaded line plus another eligible candidate.
	require.NoError(t, db.Create(&models.ProcessedLine{
		LineContent: "downloaded", LineHash: "s1", TvgName: "x", GroupTitle: "g", ProcessedAt: time.Now(),
		ContentType: models.ContentTypeMovies, State: models.StateDownloaded, MovieID: &movie.ID,
		LineURL: strPtr("http://example.com/stuck-720p.mkv"), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error)
	require.NoError(t, db.Create(&models.ProcessedLine{
		LineContent: "alt", LineHash: "s2", TvgName: "x", GroupTitle: "g", ProcessedAt: time.Now(),
		ContentType: models.ContentTypeMovies, State: models.StateProcessed, MovieID: &movie.ID,
		LineURL: strPtr("http://example.com/stuck-1080p.mkv"), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error)

	// A stuck/incomplete download for a third, different candidate - outside
	// tier-1 (Radarr doesn't report the movie missing) and tier-2 fetch
	// (state=downloading falls outside the processed/failed candidate filter).
	stuckLine := models.ProcessedLine{
		LineContent: "stuck", LineHash: "s3", TvgName: "x", GroupTitle: "g", ProcessedAt: time.Now(),
		ContentType: models.ContentTypeMovies, State: models.StateDownloading, MovieID: &movie.ID,
		LineURL: strPtr("http://example.com/stuck-4k.mkv"), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&stuckLine).Error)
	dlInfo := models.DownloadInfo{Status: string(models.DownloadStatusDownloading)}
	require.NoError(t, db.Create(&dlInfo).Error)
	require.NoError(t, db.Model(&stuckLine).Update("download_info_id", dlInfo.ID).Error)

	fr := &fakeRadarr{} // movie not reported missing by Radarr
	fs := &fakeSonarr{series: map[int]*sonarr.Series{}}

	streams, err := BuildStreams(context.Background(), testDeps(db, fr, fs))
	require.NoError(t, err)
	require.Len(t, streams, 1, "synthesized tier-1 stream from the stuck download should win over the pre-existing tier-2 stream")
	require.Equal(t, Tier1, streams[0].Tier)
	require.Equal(t, fmt.Sprintf("movie:%d", movie.ID), streams[0].SourceKey)
	require.NotNil(t, streams[0].Items[0].ResumeInfo)
}

// 2.3 A resumed item is attempted before a fresh item in the same stream.
func TestBuildStreams_ResumedItemBeforeFreshItem(t *testing.T) {
	db := setupTestDB(t)

	seriesTvdbID := 777
	// Episode 3: fresh, never attempted.
	e3 := models.TVShow{TMDBID: 300, TVDBID: intPtr(seriesTvdbID), TMDBTitle: "Show", TMDBYear: 2020, Season: intPtr(1), Episode: intPtr(3)}
	require.NoError(t, db.Create(&e3).Error)
	require.NoError(t, db.Create(&models.ProcessedLine{
		LineContent: "e3", LineHash: "e3", TvgName: "x", GroupTitle: "g", ProcessedAt: time.Now(),
		ContentType: models.ContentTypeTVShows, State: models.StateProcessed, TVShowID: &e3.ID,
		LineURL: strPtr("http://example.com/e3.mkv"), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error)

	// Episode 5: has an incomplete/interrupted download in progress.
	e5 := models.TVShow{TMDBID: 300, TVDBID: intPtr(seriesTvdbID), TMDBTitle: "Show", TMDBYear: 2020, Season: intPtr(1), Episode: intPtr(5)}
	require.NoError(t, db.Create(&e5).Error)
	e5Line := models.ProcessedLine{
		LineContent: "e5", LineHash: "e5", TvgName: "x", GroupTitle: "g", ProcessedAt: time.Now(),
		ContentType: models.ContentTypeTVShows, State: models.StateDownloading, TVShowID: &e5.ID,
		LineURL: strPtr("http://example.com/e5.mkv"), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&e5Line).Error)

	dlInfo := models.DownloadInfo{Status: string(models.DownloadStatusDownloading)}
	require.NoError(t, db.Create(&dlInfo).Error)
	require.NoError(t, db.Model(&e5Line).Update("download_info_id", dlInfo.ID).Error)

	fr := &fakeRadarr{}
	fs := &fakeSonarr{
		missing: []sonarr.Episode{
			{ID: 1, SeriesID: 20, SeasonNumber: 1, EpisodeNumber: 3},
			{ID: 2, SeriesID: 20, SeasonNumber: 1, EpisodeNumber: 5},
		},
		series: map[int]*sonarr.Series{20: {ID: 20, Title: "Show", Year: 2020, TvdbID: seriesTvdbID}},
	}

	streams, err := BuildStreams(context.Background(), testDeps(db, fr, fs))
	require.NoError(t, err)
	require.Len(t, streams, 1)
	require.Len(t, streams[0].Items, 2)

	require.Equal(t, 5, streams[0].Items[0].Episode, "resumed episode should be attempted first")
	require.NotNil(t, streams[0].Items[0].ResumeInfo)
	require.Equal(t, 3, streams[0].Items[1].Episode)
	require.Nil(t, streams[0].Items[1].ResumeInfo)
}

// 2.4 A stale lock is cleared before stream construction proceeds.
func TestBuildStreams_ClearsStaleLocksBeforeConstruction(t *testing.T) {
	db := setupTestDB(t)

	staleLockedAt := time.Now().Add(-1 * time.Hour)
	dlInfo := models.DownloadInfo{
		Status:   string(models.DownloadStatusDownloading),
		LockedAt: &staleLockedAt,
		LockedBy: strPtr("crashed-instance"),
	}
	require.NoError(t, db.Create(&dlInfo).Error)

	fr := &fakeRadarr{}
	fs := &fakeSonarr{series: map[int]*sonarr.Series{}}

	_, err := BuildStreams(context.Background(), testDeps(db, fr, fs))
	require.NoError(t, err)

	var updated models.DownloadInfo
	require.NoError(t, db.First(&updated, dlInfo.ID).Error)
	require.Nil(t, updated.LockedAt, "stale lock should have been cleared before construction")
	require.Nil(t, updated.LockedBy)
}

// ApplyLimit caps distinct work units while keeping a series' season chain intact.
func TestApplyLimit_KeepsSeasonChainsTogether(t *testing.T) {
	streams := []*Stream{
		movieStream(1, Tier1),
		movieStream(2, Tier1),
		seasonStream(10, 1, Tier1, 1, 2),
		seasonStream(10, 2, Tier1, 1, 2),
	}

	limited := ApplyLimit(streams, 2)

	sourceKeys := map[string]bool{}
	for _, s := range limited {
		sourceKeys[s.SourceKey] = true
	}
	require.Len(t, sourceKeys, 2)

	// If the series (source key "series:10") was kept, both its seasons must be present.
	if sourceKeys["series:10"] {
		count := 0
		for _, s := range limited {
			if s.SourceKey == "series:10" {
				count++
			}
		}
		require.Equal(t, 2, count)
	}
}

// 2.5 Integration-style test combining 2.1-2.4 against a seeded test DB.
func TestBuildStreams_FullIntegration(t *testing.T) {
	db := setupTestDB(t)

	movie := models.Movie{TMDBID: 1, TVDBID: intPtr(1), TMDBTitle: "Movie", TMDBYear: 2020}
	require.NoError(t, db.Create(&movie).Error)
	require.NoError(t, db.Create(&models.ProcessedLine{
		LineContent: "m", LineHash: "m", TvgName: "x", GroupTitle: "g", ProcessedAt: time.Now(),
		ContentType: models.ContentTypeMovies, State: models.StateProcessed, MovieID: &movie.ID,
		LineURL: strPtr("http://example.com/m.mkv"), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error)

	s1e1 := models.TVShow{TMDBID: 2, TVDBID: intPtr(2), TMDBTitle: "Show", TMDBYear: 2021, Season: intPtr(1), Episode: intPtr(1)}
	require.NoError(t, db.Create(&s1e1).Error)
	require.NoError(t, db.Create(&models.ProcessedLine{
		LineContent: "s1e1", LineHash: "s1e1", TvgName: "x", GroupTitle: "g", ProcessedAt: time.Now(),
		ContentType: models.ContentTypeTVShows, State: models.StateProcessed, TVShowID: &s1e1.ID,
		LineURL: strPtr("http://example.com/s1e1.mkv"), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error)

	s2e1 := models.TVShow{TMDBID: 2, TVDBID: intPtr(2), TMDBTitle: "Show", TMDBYear: 2021, Season: intPtr(2), Episode: intPtr(1)}
	require.NoError(t, db.Create(&s2e1).Error)
	require.NoError(t, db.Create(&models.ProcessedLine{
		LineContent: "s2e1", LineHash: "s2e1", TvgName: "x", GroupTitle: "g", ProcessedAt: time.Now(),
		ContentType: models.ContentTypeTVShows, State: models.StateProcessed, TVShowID: &s2e1.ID,
		LineURL: strPtr("http://example.com/s2e1.mkv"), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error)

	staleLockedAt := time.Now().Add(-1 * time.Hour)
	require.NoError(t, db.Create(&models.DownloadInfo{Status: string(models.DownloadStatusFailed), LockedAt: &staleLockedAt, LockedBy: strPtr("dead")}).Error)

	fr := &fakeRadarr{missing: []radarr.Movie{{ID: 1, Title: "Movie", Year: 2020, TvdbID: 1, TMDBID: 1}}}
	fs := &fakeSonarr{
		missing: []sonarr.Episode{
			{ID: 1, SeriesID: 5, SeasonNumber: 1, EpisodeNumber: 1},
			{ID: 2, SeriesID: 5, SeasonNumber: 2, EpisodeNumber: 1},
		},
		series: map[int]*sonarr.Series{5: {ID: 5, Title: "Show", Year: 2021, TvdbID: 2}},
	}

	streams, err := BuildStreams(context.Background(), testDeps(db, fr, fs))
	require.NoError(t, err)
	require.Len(t, streams, 3) // 1 movie stream + 2 season streams

	sched := NewScheduler(streams, 0)
	require.Equal(t, 2, sched.Remaining(), "movie stream + season 1 claimable; season 2 pending")
}
