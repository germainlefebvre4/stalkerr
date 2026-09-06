package downloader

import (
	"context"
	"testing"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateState_ClearsStaleErrorMessageOnDownloading(t *testing.T) {
	db := setupTestDB(t)
	sm := NewStateManager(DefaultStateManagerConfig())

	errMsg := "HTTP 429: rate limited"
	download := &models.DownloadInfo{
		Status:       string(models.DownloadStatusFailed),
		ErrorMessage: &errMsg,
	}
	require.NoError(t, db.Create(download).Error)

	err := sm.UpdateState(context.Background(), download.ID, models.DownloadStatusDownloading, nil)
	require.NoError(t, err)

	var updated models.DownloadInfo
	require.NoError(t, db.First(&updated, download.ID).Error)
	assert.Equal(t, string(models.DownloadStatusDownloading), updated.Status)
	assert.Nil(t, updated.ErrorMessage)
}

func TestUpdateState_FailedThenRetriedSuccessfully(t *testing.T) {
	db := setupTestDB(t)
	sm := NewStateManager(DefaultStateManagerConfig())
	ctx := context.Background()

	download := &models.DownloadInfo{
		Status: string(models.DownloadStatusDownloading),
	}
	require.NoError(t, db.Create(download).Error)

	errMsg := "HTTP 429: rate limited"
	require.NoError(t, sm.UpdateState(ctx, download.ID, models.DownloadStatusFailed, &errMsg))

	var afterFailure models.DownloadInfo
	require.NoError(t, db.First(&afterFailure, download.ID).Error)
	require.NotNil(t, afterFailure.ErrorMessage)
	assert.Equal(t, errMsg, *afterFailure.ErrorMessage)

	require.NoError(t, sm.UpdateState(ctx, download.ID, models.DownloadStatusDownloading, nil))
	require.NoError(t, sm.UpdateState(ctx, download.ID, models.DownloadStatusCompleted, nil))

	var final models.DownloadInfo
	require.NoError(t, db.First(&final, download.ID).Error)
	assert.Equal(t, string(models.DownloadStatusCompleted), final.Status)
	assert.Nil(t, final.ErrorMessage)
}

func TestUpdateState_FailedIncrementsRetryCountOnce(t *testing.T) {
	db := setupTestDB(t)
	config.SetConfig(&config.Config{})
	sm := NewStateManager(DefaultStateManagerConfig())
	ctx := context.Background()

	download := &models.DownloadInfo{Status: string(models.DownloadStatusDownloading)}
	require.NoError(t, db.Create(download).Error)

	errMsg := "unexpected EOF"
	require.NoError(t, sm.UpdateState(ctx, download.ID, models.DownloadStatusFailed, &errMsg))

	var updated models.DownloadInfo
	require.NoError(t, db.First(&updated, download.ID).Error)
	assert.Equal(t, 1, updated.RetryCount)
	assert.Equal(t, string(models.DownloadStatusFailed), updated.Status)
}

func TestUpdateState_FailedReachingMaxRetriesTransitionsToCancelled(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.Config{}
	cfg.Downloads.MaxRetryAttempts = 3
	config.SetConfig(cfg)
	sm := NewStateManager(DefaultStateManagerConfig())
	ctx := context.Background()

	download := &models.DownloadInfo{Status: string(models.DownloadStatusDownloading), RetryCount: 2}
	require.NoError(t, db.Create(download).Error)

	line := &models.ProcessedLine{
		LineContent:    "#EXTINF:-1,Test Movie",
		LineHash:       "testhash-cancel-transition",
		TvgName:        "Test Movie",
		GroupTitle:     "Movies",
		ContentType:    models.ContentTypeMovies,
		State:          models.StateDownloading,
		DownloadInfoID: &download.ID,
	}
	require.NoError(t, db.Create(line).Error)

	errMsg := "unexpected EOF"
	require.NoError(t, sm.UpdateState(ctx, download.ID, models.DownloadStatusFailed, &errMsg))

	var updatedDownload models.DownloadInfo
	require.NoError(t, db.First(&updatedDownload, download.ID).Error)
	assert.Equal(t, 3, updatedDownload.RetryCount)
	assert.Equal(t, string(models.DownloadStatusCancelled), updatedDownload.Status)

	var updatedLine models.ProcessedLine
	require.NoError(t, db.First(&updatedLine, line.ID).Error)
	assert.Equal(t, models.StateCancelled, updatedLine.State)
}

func TestUpdateState_FailedBelowMaxRetriesStaysFailed(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.Config{}
	cfg.Downloads.MaxRetryAttempts = 3
	config.SetConfig(cfg)
	sm := NewStateManager(DefaultStateManagerConfig())
	ctx := context.Background()

	download := &models.DownloadInfo{Status: string(models.DownloadStatusDownloading), RetryCount: 1}
	require.NoError(t, db.Create(download).Error)

	errMsg := "unexpected EOF"
	require.NoError(t, sm.UpdateState(ctx, download.ID, models.DownloadStatusFailed, &errMsg))

	var updated models.DownloadInfo
	require.NoError(t, db.First(&updated, download.ID).Error)
	assert.Equal(t, 2, updated.RetryCount)
	assert.Equal(t, string(models.DownloadStatusFailed), updated.Status)
}

// 3.2: GetIncompleteDownloads relies on its existing status IN (...)
// allow-list (which omits "cancelled") to exclude a cancelled/retry-exhausted
// download; this guards that invariant.
func TestGetIncompleteDownloads_ExcludesCancelled(t *testing.T) {
	db := setupTestDB(t)
	sm := NewStateManager(DefaultStateManagerConfig())
	ctx := context.Background()

	cancelled := &models.DownloadInfo{Status: string(models.DownloadStatusCancelled)}
	require.NoError(t, db.Create(cancelled).Error)

	pending := &models.DownloadInfo{Status: string(models.DownloadStatusPending)}
	require.NoError(t, db.Create(pending).Error)

	downloads, err := sm.GetIncompleteDownloads(ctx, 0, 0)
	require.NoError(t, err)

	ids := make([]uint, 0, len(downloads))
	for _, d := range downloads {
		ids = append(ids, d.ID)
	}

	assert.NotContains(t, ids, cancelled.ID)
	assert.Contains(t, ids, pending.ID)
}

func TestUpdateState_FirstAttemptLeavesErrorMessageNil(t *testing.T) {
	db := setupTestDB(t)
	sm := NewStateManager(DefaultStateManagerConfig())

	download := &models.DownloadInfo{
		Status: string(models.DownloadStatusPending),
	}
	require.NoError(t, db.Create(download).Error)
	require.Nil(t, download.ErrorMessage)

	err := sm.UpdateState(context.Background(), download.ID, models.DownloadStatusDownloading, nil)
	require.NoError(t, err)

	var updated models.DownloadInfo
	require.NoError(t, db.First(&updated, download.ID).Error)
	assert.Equal(t, string(models.DownloadStatusDownloading), updated.Status)
	assert.Nil(t, updated.ErrorMessage)
}
