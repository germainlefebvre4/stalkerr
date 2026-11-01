package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type forcedFixture struct {
	movie    models.Movie
	sibling  models.ProcessedLine // downloaded sibling
	forced   models.ProcessedLine
	forcedDL models.DownloadInfo
}

// seedMovieWithDownloadedSibling creates a movie with a downloaded 720p
// sibling, a better-ranked untried candidate is NOT needed: the forced line is
// the 1080p occurrence and the sibling is the 720p downloaded one.
func seedMovieWithDownloadedSibling(t *testing.T, db *gorm.DB, forced bool, destPath string) forcedFixture {
	t.Helper()
	f := forcedFixture{movie: models.Movie{TMDBID: 220, TVDBID: intPtr(1220), TMDBTitle: "Luca", TMDBYear: 2021}}
	require.NoError(t, db.Create(&f.movie).Error)

	res720, res1080 := "720p", "1080p"
	lang := "MULTI"
	f.sibling = models.ProcessedLine{
		LineContent: "luca-720", LineHash: "luca-720", TvgName: "x", GroupTitle: "g", ProcessedAt: time.Now(),
		ContentType: models.ContentTypeMovies, State: models.StateDownloaded, MovieID: &f.movie.ID,
		Resolution: &res720, Language: &lang,
		LineURL: strPtr("http://example.com/luca-720.mkv"), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&f.sibling).Error)

	f.forced = models.ProcessedLine{
		LineContent: "luca-1080", LineHash: "luca-1080", TvgName: "x", GroupTitle: "g", ProcessedAt: time.Now(),
		ContentType: models.ContentTypeMovies, State: models.StateProcessed, MovieID: &f.movie.ID,
		Resolution: &res1080, Language: &lang,
		LineURL: strPtr("http://example.com/luca-1080.mkv"), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&f.forced).Error)

	f.forcedDL = models.DownloadInfo{
		URL: "http://example.com/luca-1080.mkv", Status: string(models.DownloadStatusPending),
		IsForced: forced, DownloadPath: &destPath, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&f.forcedDL).Error)
	require.NoError(t, db.Model(&f.forced).Update("download_info_id", f.forcedDL.ID).Error)
	return f
}

func TestAttachResume_ForcedPinsSingleCandidateAndDestination(t *testing.T) {
	line := models.ProcessedLine{ID: 2}
	other := models.ProcessedLine{ID: 1}
	path := "/media/movies/Luca (2021)/Luca (2021) [1080p][MULTI].mkv"
	item := Item{Candidates: []models.ProcessedLine{other, line}}

	attachResume(&item, &line, &models.DownloadInfo{IsForced: true, DownloadPath: &path})

	require.Len(t, item.Candidates, 1)
	require.Equal(t, uint(2), item.Candidates[0].ID)
	require.NotNil(t, item.ForcedLine)
	require.Equal(t, "/media/movies/Luca (2021)/Luca (2021) [1080p][MULTI]", item.ForcedDestPath)
}

func TestAttachResume_ForcedKeepsDotInTitle(t *testing.T) {
	line := models.ProcessedLine{ID: 2}
	path := "/tv/Mr. Robot (2015)/Mr. Robot (2015) - S01E01 [1080p]"
	item := Item{}
	attachResume(&item, &line, &models.DownloadInfo{IsForced: true, DownloadPath: &path})
	require.Equal(t, path, item.ForcedDestPath)
}

func TestAttachResume_NonForcedKeepsExistingBehavior(t *testing.T) {
	line := models.ProcessedLine{ID: 2}
	other := models.ProcessedLine{ID: 1}
	item := Item{Candidates: []models.ProcessedLine{other}}

	attachResume(&item, &line, &models.DownloadInfo{})

	require.Len(t, item.Candidates, 2)
	require.Equal(t, uint(2), item.Candidates[0].ID, "non-forced resume still prepends its line")
	require.Nil(t, item.ForcedLine)
	require.Empty(t, item.ForcedDestPath)
}

func TestMergeIncomplete_ForcedWithDownloadedSiblingIsKeptAndPinned(t *testing.T) {
	db := setupTestDB(t)
	f := seedMovieWithDownloadedSibling(t, db, true, "/media/movies/Luca (2021)/Luca (2021) [1080p][MULTI]")

	streams, err := BuildStreams(context.Background(), testDeps(db, &fakeRadarr{}, &fakeSonarr{}))
	require.NoError(t, err)
	require.Len(t, streams, 1, "the tier-2 upgrade stream must be deduped away by the synthesized tier-1 stream")
	require.Equal(t, Tier1, streams[0].Tier)
	require.Len(t, streams[0].Items, 1)

	item := streams[0].Items[0]
	require.Len(t, item.Candidates, 1)
	require.Equal(t, f.forced.ID, item.Candidates[0].ID)
	require.Equal(t, "/media/movies/Luca (2021)/Luca (2021) [1080p][MULTI]", item.ForcedDestPath)

	var dl models.DownloadInfo
	require.NoError(t, db.First(&dl, f.forcedDL.ID).Error)
	require.Equal(t, string(models.DownloadStatusPending), dl.Status, "forced download must not be cancelled")
}

func TestMergeIncomplete_NonForcedWithDownloadedSiblingStillCancelled(t *testing.T) {
	db := setupTestDB(t)
	f := seedMovieWithDownloadedSibling(t, db, false, "/media/movies/Luca (2021)/Luca (2021) [1080p][MULTI]")

	streams, err := BuildStreams(context.Background(), testDeps(db, &fakeRadarr{}, &fakeSonarr{}))
	require.NoError(t, err)
	for _, s := range streams {
		require.NotEqual(t, Tier1, s.Tier, "no tier-1 stream for a cancelled duplicate")
	}

	var dl models.DownloadInfo
	require.NoError(t, db.First(&dl, f.forcedDL.ID).Error)
	require.Equal(t, string(models.DownloadStatusCancelled), dl.Status)
}

func TestMergeIncomplete_ForcedUnmonitoredIsSkippedWithoutCancel(t *testing.T) {
	db := setupTestDB(t)
	f := seedMovieWithDownloadedSibling(t, db, true, "/media/movies/Luca (2021)/Luca (2021) [1080p][MULTI]")

	deps := testDeps(db, &fakeRadarr{}, &fakeSonarr{})
	deps.MonitoredMovieTMDBIDs = map[int]bool{f.movie.TMDBID: false}
	streams, err := BuildStreams(context.Background(), deps)
	require.NoError(t, err)
	for _, s := range streams {
		require.NotEqual(t, Tier1, s.Tier)
	}

	var dl models.DownloadInfo
	require.NoError(t, db.Find(&dl, f.forcedDL.ID).Error)
	require.Equal(t, string(models.DownloadStatusPending), dl.Status, "status must be unchanged")
}

func TestMergeIncomplete_TwoForcedSameMovie_OnlyOldestKept(t *testing.T) {
	db := setupTestDB(t)
	f := seedMovieWithDownloadedSibling(t, db, true, "/media/movies/Luca (2021)/Luca (2021) [1080p][MULTI]")
	// Make the first forced record clearly older.
	require.NoError(t, db.Model(&models.DownloadInfo{}).Where("id = ?", f.forcedDL.ID).
		Update("created_at", time.Now().Add(-time.Hour)).Error)

	res4k := "2160p"
	second := models.ProcessedLine{
		LineContent: "luca-4k", LineHash: "luca-4k", TvgName: "x", GroupTitle: "g", ProcessedAt: time.Now(),
		ContentType: models.ContentTypeMovies, State: models.StateProcessed, MovieID: &f.movie.ID, Resolution: &res4k,
		LineURL: strPtr("http://example.com/luca-4k.mkv"), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&second).Error)
	path2 := "/media/movies/Luca (2021)/Luca (2021) [2160p]"
	dl2 := models.DownloadInfo{URL: "http://example.com/luca-4k.mkv", Status: string(models.DownloadStatusPending),
		IsForced: true, DownloadPath: &path2, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	require.NoError(t, db.Create(&dl2).Error)
	require.NoError(t, db.Model(&second).Update("download_info_id", dl2.ID).Error)

	streams, err := BuildStreams(context.Background(), testDeps(db, &fakeRadarr{}, &fakeSonarr{}))
	require.NoError(t, err)
	require.Len(t, streams, 1)
	require.Len(t, streams[0].Items[0].Candidates, 1)
	require.Equal(t, f.forced.ID, streams[0].Items[0].Candidates[0].ID, "oldest forced record wins")

	var after models.DownloadInfo
	require.NoError(t, db.First(&after, dl2.ID).Error)
	require.Equal(t, string(models.DownloadStatusPending), after.Status, "the other forced download stays pending, untouched")
}

func TestMergeIncomplete_TwoForcedSameEpisode_OnlyOldestKept(t *testing.T) {
	db := setupTestDB(t)

	show := models.TVShow{TMDBID: 300, TVDBID: intPtr(777), TMDBTitle: "Show", TMDBYear: 2020, Season: intPtr(1), Episode: intPtr(2)}
	require.NoError(t, db.Create(&show).Error)

	mk := func(hash string, age time.Duration) (models.ProcessedLine, models.DownloadInfo) {
		line := models.ProcessedLine{
			LineContent: hash, LineHash: hash, TvgName: "x", GroupTitle: "g", ProcessedAt: time.Now(),
			ContentType: models.ContentTypeTVShows, State: models.StateProcessed, TVShowID: &show.ID,
			LineURL: strPtr("http://example.com/" + hash), CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}
		require.NoError(t, db.Create(&line).Error)
		p := "/media/tv/Show/Season 01/" + hash
		dl := models.DownloadInfo{URL: "http://example.com/" + hash, Status: string(models.DownloadStatusPending),
			IsForced: true, DownloadPath: &p, CreatedAt: time.Now().Add(-age), UpdatedAt: time.Now()}
		require.NoError(t, db.Create(&dl).Error)
		require.NoError(t, db.Model(&line).Update("download_info_id", dl.ID).Error)
		return line, dl
	}
	oldLine, _ := mk("old", 2*time.Hour)
	_, newDL := mk("new", time.Hour)

	streams, err := BuildStreams(context.Background(), testDeps(db, &fakeRadarr{}, &fakeSonarr{}))
	require.NoError(t, err)
	require.Len(t, streams, 1)
	require.Len(t, streams[0].Items, 1)
	require.Equal(t, oldLine.ID, streams[0].Items[0].Candidates[0].ID)

	var after models.DownloadInfo
	require.NoError(t, db.First(&after, newDL.ID).Error)
	require.Equal(t, string(models.DownloadStatusPending), after.Status)
}
