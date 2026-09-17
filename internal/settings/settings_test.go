package settings

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test SQLite DB: %v", err)
	}
	if err := db.AutoMigrate(&models.SettingsOverride{}, &models.M3USourceConfig{}, &models.BandwidthScheduleWindow{}); err != nil {
		t.Fatalf("failed to migrate models: %v", err)
	}
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(1)
	}
	database.SetDB(db)
	return db
}

func baseTestConfig() *config.Config {
	return &config.Config{
		Radarr: config.RadarrConfig{
			URL:              "http://file-radarr.example.com",
			APIKey:           "file-key",
			Enabled:          true,
			SyncInterval:     3600,
			QualityProfileID: 1,
		},
		Downloads: config.DownloadsConfig{
			Timeout: 300,
		},
	}
}

func TestSetOverride_RoundTripsEachKind(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(baseTestConfig())

	cases := []struct {
		key string
		raw string
	}{
		{"radarr.url", `"http://override.example.com"`},
		{"radarr.sync_interval", `7200`},
		{"downloads.progress_interval_mb", `50`},
		{"tmdb.requests_per_second", `2.5`},
		{"radarr.enabled", `false`},
		{"downloads.throttle_rate_kbps", `2048`},
		{"jellyfin.playback_check_enabled", `true`},
		{"jellyfin.playback_action", `"stop"`},
		{"jellyfin.playback_poll_interval_seconds", `45`},
	}

	for _, c := range cases {
		if err := SetOverride(c.key, json.RawMessage(c.raw)); err != nil {
			t.Fatalf("SetOverride(%q, %q) failed: %v", c.key, c.raw, err)
		}
	}

	eff := Effective()
	if eff.Radarr.URL != "http://override.example.com" {
		t.Errorf("radarr.url = %q, want override", eff.Radarr.URL)
	}
	if eff.Radarr.SyncInterval != 7200 {
		t.Errorf("radarr.sync_interval = %d, want 7200", eff.Radarr.SyncInterval)
	}
	if eff.Downloads.ProgressIntervalMB != 50 {
		t.Errorf("downloads.progress_interval_mb = %d, want 50", eff.Downloads.ProgressIntervalMB)
	}
	if eff.TMDB.RequestsPerSecond != 2.5 {
		t.Errorf("tmdb.requests_per_second = %v, want 2.5", eff.TMDB.RequestsPerSecond)
	}
	if eff.Radarr.Enabled != false {
		t.Errorf("radarr.enabled = %v, want false", eff.Radarr.Enabled)
	}
	if eff.Downloads.ThrottleRateKbps != 2048 {
		t.Errorf("downloads.throttle_rate_kbps = %d, want 2048", eff.Downloads.ThrottleRateKbps)
	}
	if eff.Jellyfin.PlaybackCheckEnabled != true {
		t.Errorf("jellyfin.playback_check_enabled = %v, want true", eff.Jellyfin.PlaybackCheckEnabled)
	}
	if eff.Jellyfin.PlaybackAction != "stop" {
		t.Errorf("jellyfin.playback_action = %q, want stop", eff.Jellyfin.PlaybackAction)
	}
	if eff.Jellyfin.PlaybackPollIntervalSeconds != 45 {
		t.Errorf("jellyfin.playback_poll_interval_seconds = %d, want 45", eff.Jellyfin.PlaybackPollIntervalSeconds)
	}
}

func TestClearOverride_RevertsToConfigValue(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(baseTestConfig())

	if err := SetOverride("radarr.api_key", json.RawMessage(`"ui-key"`)); err != nil {
		t.Fatalf("SetOverride failed: %v", err)
	}
	if got := Effective().Radarr.APIKey; got != "ui-key" {
		t.Fatalf("expected override applied, got %q", got)
	}

	if err := ClearOverride("radarr.api_key"); err != nil {
		t.Fatalf("ClearOverride failed: %v", err)
	}

	if got := Effective().Radarr.APIKey; got != "file-key" {
		t.Errorf("radarr.api_key = %q after clear, want file-key value", got)
	}
	if origin := Origin("radarr.api_key"); origin != OriginConfig {
		t.Errorf("Origin() = %q after clear, want %q", origin, OriginConfig)
	}
}

func TestClearOverride_IsIdempotentWhenNoOverrideStored(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(baseTestConfig())

	if err := ClearOverride("radarr.url"); err != nil {
		t.Errorf("ClearOverride on unset field should not error, got %v", err)
	}
}

func TestEffective_OverridingOneFieldLeavesSiblingsUntouched(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(baseTestConfig())

	if err := SetOverride("radarr.api_key", json.RawMessage(`"ui-key"`)); err != nil {
		t.Fatalf("SetOverride failed: %v", err)
	}

	eff := Effective()
	if eff.Radarr.URL != "http://file-radarr.example.com" {
		t.Errorf("radarr.url should be untouched, got %q", eff.Radarr.URL)
	}
	if eff.Radarr.SyncInterval != 3600 {
		t.Errorf("radarr.sync_interval should be untouched, got %d", eff.Radarr.SyncInterval)
	}
}

func TestSetOverride_ReplacesPreviousValue(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(baseTestConfig())

	if err := SetOverride("radarr.url", json.RawMessage(`"http://first.example.com"`)); err != nil {
		t.Fatalf("SetOverride failed: %v", err)
	}
	if err := SetOverride("radarr.url", json.RawMessage(`"http://second.example.com"`)); err != nil {
		t.Fatalf("second SetOverride failed: %v", err)
	}

	if got := Effective().Radarr.URL; got != "http://second.example.com" {
		t.Errorf("radarr.url = %q, want the second override value", got)
	}

	var count int64
	database.Get().Model(&models.SettingsOverride{}).Where("key = ?", "radarr.url").Count(&count)
	if count != 1 {
		t.Errorf("expected exactly 1 stored row for radarr.url, got %d", count)
	}
}

func TestOrigin_ReportsInterfaceAndConfig(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(baseTestConfig())

	if origin := Origin("radarr.url"); origin != OriginConfig {
		t.Errorf("Origin() with no override = %q, want %q", origin, OriginConfig)
	}

	if err := SetOverride("radarr.url", json.RawMessage(`"http://override.example.com"`)); err != nil {
		t.Fatalf("SetOverride failed: %v", err)
	}

	if origin := Origin("radarr.url"); origin != OriginInterface {
		t.Errorf("Origin() with override = %q, want %q", origin, OriginInterface)
	}
}

func TestSetOverride_RejectsBootstrapKey(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(baseTestConfig())

	err := SetOverride("database.host", json.RawMessage(`"evil-host"`))
	if !errors.Is(err, ErrBootstrapField) {
		t.Errorf("expected ErrBootstrapField, got %v", err)
	}

	var count int64
	database.Get().Model(&models.SettingsOverride{}).Where("key = ?", "database.host").Count(&count)
	if count != 0 {
		t.Error("expected no row stored for a rejected bootstrap key")
	}
}

func TestSetOverride_RejectsUnknownKey(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(baseTestConfig())

	err := SetOverride("does.not.exist", json.RawMessage(`"value"`))
	if !errors.Is(err, ErrUnknownField) {
		t.Errorf("expected ErrUnknownField, got %v", err)
	}
}

func TestClearOverride_RejectsBootstrapAndUnknownKeys(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(baseTestConfig())

	if err := ClearOverride("database.host"); !errors.Is(err, ErrBootstrapField) {
		t.Errorf("expected ErrBootstrapField, got %v", err)
	}
	if err := ClearOverride("does.not.exist"); !errors.Is(err, ErrUnknownField) {
		t.Errorf("expected ErrUnknownField, got %v", err)
	}
}
