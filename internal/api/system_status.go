package api

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glefebvre/stalkeer/internal/circuitbreaker"
	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/downloader"
	"github.com/glefebvre/stalkeer/internal/external/radarr"
	"github.com/glefebvre/stalkeer/internal/external/sonarr"
	"github.com/glefebvre/stalkeer/internal/external/tmdb"
	"github.com/glefebvre/stalkeer/internal/m3udownloader"
	"github.com/glefebvre/stalkeer/internal/retry"
)

// systemStatusCheckTimeout bounds each individual dependency check performed
// by the aggregation endpoint. It is intentionally shorter than
// existenceCheckTimeout: a diagnostic check must fail fast, not wait as long
// as a real data fetch would. A var (not a const) so tests can shrink it to
// exercise the timeout path without waiting out the real 5s bound.
var systemStatusCheckTimeout = 5 * time.Second

// Three-state status values reported for the database/Radarr/Sonarr/TMDB
// sections.
const (
	statusOK            = "ok"
	statusKO            = "ko"
	statusNotConfigured = "not_configured"
)

// KO reason codes: a small fixed set the frontend translates, rather than
// raw error strings.
const (
	reasonUnreachable  = "unreachable"
	reasonUnauthorized = "unauthorized"
	reasonTimeout      = "timeout"
	reasonUnavailable  = "unavailable"
	reasonCircuitOpen  = "circuit_open"
)

// buildVersion, buildCommit, and buildDate hold the running binary's build
// metadata, set once at startup via SetBuildInfo (main can't be imported
// here, since it imports this package).
var (
	buildVersion string
	buildCommit  string
	buildDate    string
)

// SetBuildInfo records the running binary's version, commit, and build date
// so getSystemStatus can report them. Call once at startup before serving
// requests.
func SetBuildInfo(version, commit, date string) {
	buildVersion = version
	buildCommit = commit
	buildDate = date
}

// ServiceStatus is the three-state reachability result for one dependency
// (database, Radarr, Sonarr, or TMDB).
type ServiceStatus struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

// DiskUsageResponse is one deduplicated disk usage entry, covering every
// configured storage path that resolved to the same mounted volume.
type DiskUsageResponse struct {
	Paths       []string `json:"paths"`
	Available   uint64   `json:"available,omitempty"`
	Free        uint64   `json:"free,omitempty"`
	Total       uint64   `json:"total,omitempty"`
	UsedPct     float64  `json:"used_pct,omitempty"`
	Unavailable bool     `json:"unavailable,omitempty"`
	Reason      string   `json:"reason,omitempty"`
}

// SystemStatusResponse is the full aggregated response for GET
// /api/v1/system/status.
type SystemStatusResponse struct {
	Database ServiceStatus       `json:"database"`
	Radarr   ServiceStatus       `json:"radarr"`
	Sonarr   ServiceStatus       `json:"sonarr"`
	TMDB     ServiceStatus       `json:"tmdb"`
	Disk     []DiskUsageResponse `json:"disk"`
	Version  string              `json:"version"`
	Commit   string              `json:"commit"`
	Date     string              `json:"date"`
}

// statusCoder is implemented by each external client's status-check error
// type (radarr.StatusError, sonarr.StatusError, tmdb.StatusError) so
// classifyReachabilityError can distinguish a 401 from other non-2xx
// responses without importing any of those packages' concrete error types.
type statusCoder interface {
	StatusCode() int
}

// getSystemStatus handles GET /api/v1/system/status, aggregating database,
// Radarr, Sonarr, TMDB reachability and deduplicated disk usage into a single
// on-demand response. Every section is computed fresh on each call and each
// check is independent: one dependency's failure never affects another
// section's reported status, and the endpoint always returns HTTP 200.
func (s *Server) getSystemStatus(c *gin.Context) {
	cfg := config.Get()
	ctx := c.Request.Context()

	var wg sync.WaitGroup
	var dbStatus, radarrStatus, sonarrStatus, tmdbStatus ServiceStatus
	var diskEntries []downloader.DiskUsageEntry

	wg.Add(5)
	go func() { defer wg.Done(); dbStatus = checkDatabaseStatus() }()
	go func() { defer wg.Done(); radarrStatus = checkRadarrStatus(ctx, cfg, s.radarrBreaker) }()
	go func() { defer wg.Done(); sonarrStatus = checkSonarrStatus(ctx, cfg, s.sonarrBreaker) }()
	go func() { defer wg.Done(); tmdbStatus = checkTMDBStatus(ctx, s.tmdbClient) }()
	go func() { defer wg.Done(); diskEntries = downloader.GroupDiskUsage(configuredStoragePaths(cfg)) }()
	wg.Wait()

	c.JSON(http.StatusOK, SystemStatusResponse{
		Database: dbStatus,
		Radarr:   radarrStatus,
		Sonarr:   sonarrStatus,
		TMDB:     tmdbStatus,
		Disk:     toDiskUsageResponses(diskEntries),
		Version:  buildVersion,
		Commit:   buildCommit,
		Date:     buildDate,
	})
}

// configuredStoragePaths lists the app's configured storage paths for disk
// usage reporting. The temp dir is omitted when unset (falls back to the OS
// default) rather than reported as its own entry. Every configured M3U
// source contributes its own "archive:<name>" entry, using its effective
// (per-source subdirectory) archive path.
func configuredStoragePaths(cfg *config.Config) []downloader.NamedPath {
	var paths []downloader.NamedPath
	if cfg.Downloads.MoviesPath != "" {
		paths = append(paths, downloader.NamedPath{Label: "movies", Path: cfg.Downloads.MoviesPath})
	}
	if cfg.Downloads.TVShowsPath != "" {
		paths = append(paths, downloader.NamedPath{Label: "tvshows", Path: cfg.Downloads.TVShowsPath})
	}
	if cfg.Downloads.TempDir != "" {
		paths = append(paths, downloader.NamedPath{Label: "temp", Path: cfg.Downloads.TempDir})
	}
	for _, source := range cfg.M3U.Sources {
		if source.Download.ArchiveDir == "" {
			continue
		}
		_, archiveDir := m3udownloader.SourcePaths(source.FilePath, source.Download.ArchiveDir, source.Name)
		paths = append(paths, downloader.NamedPath{Label: "archive:" + source.Name, Path: archiveDir})
	}
	return paths
}

func toDiskUsageResponses(entries []downloader.DiskUsageEntry) []DiskUsageResponse {
	responses := make([]DiskUsageResponse, len(entries))
	for i, e := range entries {
		responses[i] = DiskUsageResponse{
			Paths:       e.Labels,
			Unavailable: e.Unavailable,
			Reason:      e.Reason,
		}
		if e.Space != nil {
			responses[i].Available = e.Space.Available
			responses[i].Free = e.Space.Free
			responses[i].Total = e.Space.Total
			responses[i].UsedPct = e.Space.UsedPct
		}
	}
	return responses
}

func checkDatabaseStatus() ServiceStatus {
	if err := database.HealthCheck(); err != nil {
		return ServiceStatus{Status: statusKO, Reason: reasonUnreachable}
	}
	return ServiceStatus{Status: statusOK}
}

func checkRadarrStatus(ctx context.Context, cfg *config.Config, breaker *circuitbreaker.CircuitBreaker) ServiceStatus {
	if cfg.Radarr.URL == "" || cfg.Radarr.APIKey == "" {
		return ServiceStatus{Status: statusNotConfigured}
	}

	checkCtx, cancel := context.WithTimeout(ctx, systemStatusCheckTimeout)
	defer cancel()

	client := radarr.New(radarr.Config{
		BaseURL:     cfg.Radarr.URL,
		APIKey:      cfg.Radarr.APIKey,
		Timeout:     systemStatusCheckTimeout,
		RetryConfig: retry.Config{MaxAttempts: 1},
		Breaker:     breaker,
	})

	if err := client.SystemStatus(checkCtx); err != nil {
		return ServiceStatus{Status: statusKO, Reason: classifyReachabilityError(err)}
	}
	return ServiceStatus{Status: statusOK}
}

func checkSonarrStatus(ctx context.Context, cfg *config.Config, breaker *circuitbreaker.CircuitBreaker) ServiceStatus {
	if cfg.Sonarr.URL == "" || cfg.Sonarr.APIKey == "" {
		return ServiceStatus{Status: statusNotConfigured}
	}

	checkCtx, cancel := context.WithTimeout(ctx, systemStatusCheckTimeout)
	defer cancel()

	client := sonarr.New(sonarr.Config{
		BaseURL:     cfg.Sonarr.URL,
		APIKey:      cfg.Sonarr.APIKey,
		Timeout:     systemStatusCheckTimeout,
		RetryConfig: retry.Config{MaxAttempts: 1},
		Breaker:     breaker,
	})

	if err := client.SystemStatus(checkCtx); err != nil {
		return ServiceStatus{Status: statusKO, Reason: classifyReachabilityError(err)}
	}
	return ServiceStatus{Status: statusOK}
}

// checkTMDBStatus takes the concrete *tmdb.Client (rather than a
// SystemStatus-shaped interface) so a disabled integration's nil
// Server.tmdbClient stays a nil interface after being passed in - through an
// interface parameter, a nil *tmdb.Client would instead produce a non-nil
// interface value wrapping a nil pointer, defeating this nil check.
func checkTMDBStatus(ctx context.Context, client *tmdb.Client) ServiceStatus {
	if client == nil {
		return ServiceStatus{Status: statusNotConfigured}
	}

	checkCtx, cancel := context.WithTimeout(ctx, systemStatusCheckTimeout)
	defer cancel()

	if err := client.SystemStatus(checkCtx); err != nil {
		return ServiceStatus{Status: statusKO, Reason: classifyReachabilityError(err)}
	}
	return ServiceStatus{Status: statusOK}
}

// classifyReachabilityError maps a reachability check's error into one of the
// small fixed set of KO reason codes, so the frontend can translate a stable
// code instead of parsing raw error text.
func classifyReachabilityError(err error) string {
	if err == nil {
		return ""
	}

	// Checked first: an open (or half-open, request-limited) circuit means
	// no live call was even attempted, which is a distinct condition from a
	// timeout or connection failure observed on an actual attempt.
	if errors.Is(err, circuitbreaker.ErrOpenState) || errors.Is(err, circuitbreaker.ErrTooManyRequests) {
		return reasonCircuitOpen
	}

	// A ctx deadline exceeded surfaces as *url.Error wrapping
	// context.DeadlineExceeded; net.Error.Timeout() catches that (and any
	// other transport-level timeout) before falling through to the
	// connection-failure checks below.
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return reasonTimeout
	}

	var sc statusCoder
	if errors.As(err, &sc) {
		if sc.StatusCode() == http.StatusUnauthorized {
			return reasonUnauthorized
		}
		return reasonUnavailable
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return reasonUnreachable
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return reasonUnreachable
	}

	return reasonUnavailable
}
