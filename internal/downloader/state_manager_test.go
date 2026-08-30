package downloader

import (
	"context"
	"testing"

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
