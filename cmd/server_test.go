package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/api"
	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/glefebvre/stalkeer/internal/shutdown"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupServerMetricsTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.ProcessingLog{},
		&models.JobRun{},
		&models.DownloadInfo{},
	))

	database.SetDB(db)
	return db
}

// dialFails reports whether a TCP connection to addr fails within timeout.
func dialFails(addr string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return true
	}
	conn.Close()
	return false
}

// 6.2: the metrics port is never opened at all when cfg.Metrics.Enabled is
// false (the prometheus-metrics spec's "absent when disabled" requirement).
func TestMaybeStartMetricsServer_DisabledNeverOpensPort(t *testing.T) {
	setupServerMetricsTestDB(t)
	config.SetConfig(&config.Config{Metrics: config.MetricsConfig{Enabled: false, Port: 18199, Path: "/metrics"}})
	cfg := config.Get()

	server := api.NewServer()
	shutdownHandler := shutdown.New(5 * time.Second)
	log := logger.NewWithLevelAndFormat("error", "text")

	errCh := maybeStartMetricsServer(cfg, server, shutdownHandler, log, "127.0.0.1")

	// Give any (unexpected) listener goroutine a moment to bind before probing.
	time.Sleep(50 * time.Millisecond)

	addr := fmt.Sprintf("127.0.0.1:%d", cfg.Metrics.Port)
	require.True(t, dialFails(addr, 200*time.Millisecond), "expected no listener on %s when metrics disabled", addr)

	select {
	case err := <-errCh:
		t.Fatalf("expected no error from a disabled metrics server, got %v", err)
	default:
	}
}

// 6.1: GET :<metrics.port><metrics.path> returns 200 with a Prometheus text
// body when metrics exposition is enabled.
func TestMaybeStartMetricsServer_EnabledServesPrometheusBody(t *testing.T) {
	setupServerMetricsTestDB(t)
	config.SetConfig(&config.Config{Metrics: config.MetricsConfig{Enabled: true, Port: 18200, Path: "/metrics"}})
	cfg := config.Get()

	server := api.NewServer()
	shutdownHandler := shutdown.New(5 * time.Second)
	log := logger.NewWithLevelAndFormat("error", "text")

	errCh := maybeStartMetricsServer(cfg, server, shutdownHandler, log, "127.0.0.1")
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		server.ShutdownMetrics(ctx)
	})

	addr := fmt.Sprintf("127.0.0.1:%d", cfg.Metrics.Port)
	require.Eventually(t, func() bool {
		return !dialFails(addr, 100*time.Millisecond)
	}, 2*time.Second, 20*time.Millisecond, "expected metrics listener to open on %s", addr)

	resp, err := http.Get(fmt.Sprintf("http://%s%s", addr, cfg.Metrics.Path))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "stalkeer_")

	select {
	case err := <-errCh:
		t.Fatalf("expected no error from the metrics server, got %v", err)
	default:
	}
}
