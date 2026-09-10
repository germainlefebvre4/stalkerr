package main

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// lineHash mirrors parser.Parser.calculateHash (tvgName + url, SHA-256 hex),
// so tests can pre-seed processed_lines rows whose hash matches what the
// parser would compute for a given M3U entry.
func lineHash(tvgName, url string) string {
	sum := sha256.Sum256([]byte(tvgName + url))
	return hex.EncodeToString(sum[:])
}

func TestPruneAllSources_MultipleSourcesPrunedIndependently(t *testing.T) {
	db := setupProcessTestDB(t)
	tmpDir := t.TempDir()

	// The shared hash is still present in provider-b's current file, but no
	// longer present in provider-a's current file.
	sharedHash := lineHash("Shared Title", "http://example.com/shared.mkv")

	fileA := writeTestM3U(t, tmpDir, "provider-a.m3u", "#EXTM3U\n#EXTINF:-1,Still Active\nhttp://example.com/active-a.mkv")
	fileB := writeTestM3U(t, tmpDir, "provider-b.m3u", "#EXTM3U\n#EXTINF:-1,Shared Title\nhttp://example.com/shared.mkv")

	seedProcessedLine(t, db, "provider-a", sharedHash, models.StateProcessed)
	seedProcessedLine(t, db, "provider-b", sharedHash, models.StateProcessed)

	sources := []config.M3USourceConfig{
		{Name: "provider-a", FilePath: fileA},
		{Name: "provider-b", FilePath: fileB},
	}

	log := logger.NewWithLevelAndFormat("info", "text")
	result, unprunable := pruneAllSources(db, log, sources, false, false)

	if unprunable != 0 {
		t.Errorf("expected 0 unprunable sources, got %d", unprunable)
	}
	if result.linesPruned != 1 {
		t.Errorf("expected exactly 1 pruned line, got %d", result.linesPruned)
	}

	var aCount int64
	require.NoError(t, db.Model(&models.ProcessedLine{}).
		Where("source_name = ? AND line_hash = ?", "provider-a", sharedHash).
		Count(&aCount).Error)
	if aCount != 0 {
		t.Errorf("expected provider-a's stale line to be pruned, but it still exists")
	}

	var bCount int64
	require.NoError(t, db.Model(&models.ProcessedLine{}).
		Where("source_name = ? AND line_hash = ?", "provider-b", sharedHash).
		Count(&bCount).Error)
	if bCount != 1 {
		t.Errorf("expected provider-b's line with the same hash to be retained, got count %d", bCount)
	}
}

func TestPruneAllSources_UnavailableSourceSkippedNotAborted(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		db := setupProcessTestDB(t)
		tmpDir := t.TempDir()

		staleHash := lineHash("Stale", "http://example.com/stale.mkv")

		presentFile := writeTestM3U(t, tmpDir, "provider-present.m3u", "#EXTM3U\n#EXTINF:-1,Active\nhttp://example.com/active.mkv")
		missingFile := filepath.Join(tmpDir, "provider-missing.m3u") // never written

		seedProcessedLine(t, db, "provider-present", staleHash, models.StateProcessed)
		seedProcessedLine(t, db, "provider-missing", staleHash, models.StateProcessed)

		sources := []config.M3USourceConfig{
			{Name: "provider-missing", FilePath: missingFile},
			{Name: "provider-present", FilePath: presentFile},
		}

		log := logger.NewWithLevelAndFormat("info", "text")
		result, unprunable := pruneAllSources(db, log, sources, false, false)

		if unprunable != 1 {
			t.Errorf("expected exactly 1 unprunable source, got %d", unprunable)
		}
		if result.linesPruned != 1 {
			t.Errorf("expected exactly 1 pruned line (from provider-present only), got %d", result.linesPruned)
		}

		var missingCount int64
		require.NoError(t, db.Model(&models.ProcessedLine{}).
			Where("source_name = ?", "provider-missing").
			Count(&missingCount).Error)
		if missingCount != 1 {
			t.Errorf("expected provider-missing's line to be untouched (source skipped), got count %d", missingCount)
		}

		var presentCount int64
		require.NoError(t, db.Model(&models.ProcessedLine{}).
			Where("source_name = ?", "provider-present").
			Count(&presentCount).Error)
		if presentCount != 0 {
			t.Errorf("expected provider-present's stale line to be pruned, got count %d", presentCount)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		db := setupProcessTestDB(t)
		tmpDir := t.TempDir()

		staleHash := lineHash("Stale", "http://example.com/stale.mkv")
		emptyFile := writeTestM3U(t, tmpDir, "provider-empty.m3u", "#EXTM3U\n")

		seedProcessedLine(t, db, "provider-empty", staleHash, models.StateProcessed)

		sources := []config.M3USourceConfig{
			{Name: "provider-empty", FilePath: emptyFile},
		}

		log := logger.NewWithLevelAndFormat("info", "text")
		result, unprunable := pruneAllSources(db, log, sources, false, false)

		if unprunable != 1 {
			t.Errorf("expected exactly 1 unprunable source, got %d", unprunable)
		}
		if result.linesPruned != 0 {
			t.Errorf("expected 0 pruned lines (source skipped, not aborted), got %d", result.linesPruned)
		}

		var count int64
		require.NoError(t, db.Model(&models.ProcessedLine{}).
			Where("source_name = ?", "provider-empty").
			Count(&count).Error)
		if count != 1 {
			t.Errorf("expected provider-empty's line to be untouched (source skipped), got count %d", count)
		}
	})
}

// seedProcessedLine inserts a minimal processed_lines row for test setup.
func seedProcessedLine(t *testing.T, db *gorm.DB, sourceName, hash string, state models.ProcessingState) {
	t.Helper()
	line := models.ProcessedLine{
		LineContent: "seed",
		SourceName:  sourceName,
		LineHash:    hash,
		TvgName:     "seed",
		GroupTitle:  "seed",
		ProcessedAt: time.Now(),
		ContentType: models.ContentTypeUncategorized,
		State:       state,
	}
	require.NoError(t, db.Create(&line).Error)
}
