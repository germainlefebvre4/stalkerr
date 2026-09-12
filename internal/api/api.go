package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/glefebvre/stalkeer/internal/circuitbreaker"
	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/downloader"
	"github.com/glefebvre/stalkeer/internal/external/httpclient"
	"github.com/glefebvre/stalkeer/internal/external/tmdb"
	"github.com/glefebvre/stalkeer/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server represents the API server
type Server struct {
	router          *gin.Engine
	httpServer      *http.Server
	metricsServer   *http.Server
	tmdbClient      *tmdb.Client
	radarrBreaker   *circuitbreaker.CircuitBreaker
	sonarrBreaker   *circuitbreaker.CircuitBreaker
	downloader      *downloader.Downloader
	metricsRegistry *prometheus.Registry
}

// newServiceBreaker builds a circuit breaker for a Radarr/Sonarr client with
// the same defaults as the existing TMDB breaker (see tmdb.NewClient), using
// httpclient.IsSuccessful so a bad request or bad credentials never trips
// the circuit (see internal/external/httpclient.IsSuccessful).
func newServiceBreaker() *circuitbreaker.CircuitBreaker {
	return circuitbreaker.New(circuitbreaker.Config{
		MaxFailures:         5,
		Timeout:             60 * time.Second,
		MaxHalfOpenRequests: 1,
		IsSuccessful:        httpclient.IsSuccessful,
	})
}

// NewServer creates a new API server instance
func NewServer() *Server {
	router := gin.Default()

	// Configure CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"} // TODO: Configure from config file
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	router.Use(cors.New(corsConfig))

	// Add request ID middleware
	router.Use(requestIDMiddleware())

	// Add error handling middleware
	router.Use(errorHandlerMiddleware())

	var tmdbClient *tmdb.Client
	cfg := config.Get()
	if cfg.TMDB.Enabled && cfg.TMDB.APIKey != "" {
		tmdbClient = tmdb.NewClient(tmdb.Config{
			APIKey:            cfg.TMDB.APIKey,
			Language:          cfg.TMDB.Language,
			RequestsPerSecond: cfg.TMDB.RequestsPerSecond,
		})
	}

	metricsRegistry := prometheus.NewRegistry()
	metricsRegistry.MustRegister(metrics.New(database.Get(), tmdbClient))

	s := &Server{
		router:        router,
		tmdbClient:    tmdbClient,
		radarrBreaker: newServiceBreaker(),
		sonarrBreaker: newServiceBreaker(),
		downloader: downloader.New(
			time.Duration(cfg.Downloads.Timeout)*time.Second,
			cfg.Downloads.RetryAttempts,
			cfg.Downloads.MinFileSizeMB,
		),
		metricsRegistry: metricsRegistry,
	}

	s.setupRoutes()

	return s
}

// ServeHTTP allows the Server to handle HTTP requests directly (useful for testing)
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}

// Run starts the API server on the specified port
func (s *Server) Run(port int) error {
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// RunMetrics starts the Prometheus metrics exposition listener on the given
// port and path. This is a separate HTTP server from the main API router (see
// design.md's "second, independently-toggleable HTTP listener" decision) so
// that leaving metrics disabled means this port is never opened at all - not
// merely unauthenticated. Callers should only invoke this when metrics
// exposition is enabled.
func (s *Server) RunMetrics(port int, path string) error {
	mux := http.NewServeMux()
	mux.Handle(path, promhttp.HandlerFor(s.metricsRegistry, promhttp.HandlerOpts{}))

	s.metricsServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s.metricsServer.ListenAndServe()
}

// ShutdownMetrics gracefully shuts down the metrics listener, if it was started.
func (s *Server) ShutdownMetrics(ctx context.Context) error {
	if s.metricsServer != nil {
		return s.metricsServer.Shutdown(ctx)
	}
	return nil
}

func (s *Server) setupRoutes() {
	// Health check endpoint
	s.router.GET("/health", s.healthCheck)

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// TMDB proxy endpoint
		v1.GET("/tmdb/search", s.searchTMDBProxy)

		// Items endpoints
		items := v1.Group("/items")
		{
			items.GET("", s.listItems)
			items.GET("/grouped", s.listItemGroups)
			items.GET("/:id", s.getItem)
			items.PUT("/:id", s.updateItem)
			items.POST("/search", s.searchItems)
			items.POST("/:id/override", s.overrideItem)
			items.POST("/:id/force-download", s.forceDownloadItem)
		}

		// Movies endpoints
		movies := v1.Group("/movies")
		{
			movies.GET("", s.listMovies)
			movies.GET("/:id", s.getMovie)
			movies.POST("/:id/move", s.moveMovieFolder)
			movies.POST("/:id/reset", s.resetMovie)
		}

		// TV shows endpoints
		tvshows := v1.Group("/tvshows")
		{
			tvshows.GET("", s.listTVShows)
			tvshows.GET("/:id", s.getTVShow)
			tvshows.POST("/:id/move", s.moveTVShowFolder)
			tvshows.POST("/:id/reset", s.resetTVShow)
		}

		// Radarr/Sonarr monitoring endpoints
		v1.GET("/radarr/movies", s.listRadarrMonitoredMovies)
		v1.GET("/radarr/movies/:id/matches", s.getRadarrMovieMatches)
		v1.GET("/sonarr/series", s.listSonarrMonitoredSeries)
		v1.GET("/sonarr/series/:id/episodes", s.getSonarrSeriesEpisodes)
		v1.GET("/radarr-sonarr/stats", s.listRadarrSonarrStats)

		// Filter endpoints
		filters := v1.Group("/filters")
		{
			filters.GET("", s.listFilters)
			filters.POST("", s.createFilter)
			filters.PATCH("/:id", s.updateFilter)
			filters.DELETE("/:id", s.deleteFilter)
			filters.DELETE("/runtime", s.clearRuntimeFilters)
		}

		// Dry-run endpoint
		v1.POST("/dryrun", s.executeDryRun)

		// Statistics endpoint
		v1.GET("/stats", s.getStats)

		// Background logs and downloads tracking endpoints
		v1.GET("/processing-logs", s.listProcessingLogs)
		v1.GET("/downloads", s.listDownloadsEnriched)
		v1.GET("/downloads/simple", s.listDownloads)
		v1.POST("/downloads/:id/rename", s.renameDownload)
		v1.POST("/downloads/:id/resync-path", s.resyncDownloadPath)
		v1.POST("/downloads/:id/cancel", s.cancelDownload)
		v1.GET("/config/paths", s.getConfigPaths)

		// System status endpoint
		v1.GET("/system/status", s.getSystemStatus)
	}
}
