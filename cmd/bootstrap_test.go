package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/external/radarr"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/glefebvre/stalkeer/internal/retry"
	"github.com/glefebvre/stalkeer/internal/settings"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newOverridesDB returns an in-memory database with the settings tables,
// without registering it as the global database.
func newOverridesDB(t *testing.T, migrate bool) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	if migrate {
		require.NoError(t, db.AutoMigrate(&models.SettingsOverride{}, &models.M3USourceConfig{}))
	}
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	return db
}

func bootstrapTestConfig() *config.Config {
	return &config.Config{
		Radarr: config.RadarrConfig{URL: "http://file-radarr.example.com", APIKey: "file-key", Enabled: true},
	}
}

// stubConnect swaps connectDatabase for the duration of the test. A nil db
// simulates an unreachable database.
func stubConnect(t *testing.T, db *gorm.DB) {
	t.Helper()
	orig := connectDatabase
	t.Cleanup(func() {
		connectDatabase = orig
		database.SetDB(nil)
	})
	database.SetDB(nil)
	connectDatabase = func() error {
		if db == nil {
			return errors.New("connection refused")
		}
		database.SetDB(db)
		return nil
	}
}

// captureStdout runs fn and returns what it wrote to os.Stdout, which is
// where the freshly initialised loggers write.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	require.NoError(t, err)
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()

	require.NoError(t, w.Close())
	out, err := io.ReadAll(r)
	require.NoError(t, err)
	return string(out)
}

func TestLoadEffectiveConfig(t *testing.T) {
	tests := []struct {
		name          string
		requireDB     bool
		reachable     bool
		wantErr       bool
		wantAPIKey    string
		wantWarning   bool
		storeOverride bool
	}{
		{name: "required, reachable applies override", requireDB: true, reachable: true, wantAPIKey: "ui-key", storeOverride: true},
		{name: "best effort, reachable applies override", requireDB: false, reachable: true, wantAPIKey: "ui-key", storeOverride: true},
		{name: "required, unreachable errors", requireDB: true, reachable: false, wantErr: true},
		{name: "best effort, unreachable falls back with warning", requireDB: false, reachable: false, wantAPIKey: "file-key", wantWarning: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.SetConfig(bootstrapTestConfig())

			var db *gorm.DB
			if tt.reachable {
				db = newOverridesDB(t, true)
				database.SetDB(db)
				if tt.storeOverride {
					require.NoError(t, settings.SetOverride("radarr.api_key", json.RawMessage(`"ui-key"`)))
				}
			}
			stubConnect(t, db)

			var cfg *config.Config
			var err error
			out := captureStdout(t, func() { cfg, err = loadEffectiveConfig(tt.requireDB) })

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, cfg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantAPIKey, cfg.Radarr.APIKey)
			if tt.wantWarning {
				assert.Contains(t, out, "settings overrides not applied")
			} else {
				assert.NotContains(t, out, "settings overrides not applied")
			}
		})
	}
}

func TestLoadEffectiveConfig_MissingOverridesTableDoesNotAbort(t *testing.T) {
	config.SetConfig(bootstrapTestConfig())
	stubConnect(t, newOverridesDB(t, false))

	var cfg *config.Config
	var err error
	captureStdout(t, func() { cfg, err = loadEffectiveConfig(true) })

	require.NoError(t, err)
	assert.Equal(t, "file-key", cfg.Radarr.APIKey)
}

func TestLoadEffectiveConfig_UsesOverriddenM3USources(t *testing.T) {
	config.SetConfig(bootstrapTestConfig())
	db := newOverridesDB(t, true)
	database.SetDB(db)
	_, err := settings.SetSource("ui-source", settings.M3USourceInput{FilePath: "/data/ui.m3u"})
	require.NoError(t, err)
	stubConnect(t, db)

	var cfg *config.Config
	captureStdout(t, func() { cfg, err = loadEffectiveConfig(false) })

	require.NoError(t, err)
	var names []string
	for _, s := range cfg.M3U.Sources {
		names = append(names, s.Name)
	}
	assert.Contains(t, names, "ui-source")
	assert.Equal(t, names, namesOf(settings.EffectiveSources()))
}

func namesOf(sources []config.M3USourceConfig) []string {
	var names []string
	for _, s := range sources {
		names = append(names, s.Name)
	}
	return names
}

// Reproduces the reported 401: the file/env key is stale, the key stored from
// the UI is the valid one, and the download command builds its Radarr client
// from the effective config.
func TestLoadEffectiveConfig_RadarrClientSendsStoredAPIKey(t *testing.T) {
	var gotKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-Api-Key")
		if gotKey != "ui-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	bc := bootstrapTestConfig()
	bc.Radarr.URL = srv.URL
	config.SetConfig(bc)

	db := newOverridesDB(t, true)
	database.SetDB(db)
	require.NoError(t, settings.SetOverride("radarr.api_key", json.RawMessage(`"ui-key"`)))
	stubConnect(t, db)

	var cfg *config.Config
	var err error
	captureStdout(t, func() { cfg, err = loadEffectiveConfig(true) })
	require.NoError(t, err)

	client := radarr.New(radarr.Config{
		BaseURL:     cfg.Radarr.URL,
		APIKey:      cfg.Radarr.APIKey,
		Timeout:     5 * time.Second,
		RetryConfig: retry.Config{MaxAttempts: 1},
	})
	require.NoError(t, client.SystemStatus(context.Background()))
	assert.Equal(t, "ui-key", gotKey)
	assert.False(t, strings.Contains(gotKey, "file-key"))
}
