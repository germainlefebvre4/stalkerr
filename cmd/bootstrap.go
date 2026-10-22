package main

import (
	"fmt"
	"os"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/settings"
)

// connectDatabase is database.Initialize, swappable in tests.
var connectDatabase = database.Initialize

// loadEffectiveConfig is the one place a CLI command resolves its
// configuration. It connects the database before calling settings.Effective()
// so stored overrides are applied; calling Effective() first silently skips
// them. config.Load() must already have run.
//
// With requireDB, a connection failure is returned and the caller must not
// proceed. Without it, the failure is logged and the file/env/default
// configuration is returned.
func loadEffectiveConfig(requireDB bool) (*config.Config, error) {
	boot := config.Get()
	logger.InitializeLoggersWithFormat(boot.GetAppLogLevel(), boot.GetDatabaseLogLevel(), boot.Logging.Format)

	if err := connectDatabase(); err != nil {
		if requireDB {
			return nil, fmt.Errorf("initializing database: %w", err)
		}
		logger.AppLogger().WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Warn("settings overrides not applied: database unreachable, using file/env configuration")
	}

	cfg := settings.Effective()
	logger.InitializeLoggersWithFormat(cfg.GetAppLogLevel(), cfg.GetDatabaseLogLevel(), cfg.Logging.Format)
	return cfg, nil
}

// mustLoadEffectiveConfig is loadEffectiveConfig for commands that exit on
// failure. Callers keep their own `defer database.Close()`.
func mustLoadEffectiveConfig(requireDB bool) *config.Config {
	cfg, err := loadEffectiveConfig(requireDB)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	return cfg
}
