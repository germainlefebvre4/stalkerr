package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/external/tmdb"
)

// Server represents the API server
type Server struct {
	router     *gin.Engine
	httpServer *http.Server
	tmdbClient *tmdb.Client
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

	s := &Server{
		router:     router,
		tmdbClient: tmdbClient,
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
			items.GET("/:id", s.getItem)
			items.PUT("/:id", s.updateItem)
			items.POST("/search", s.searchItems)
			items.POST("/:id/override", s.overrideItem)
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
		v1.GET("/config/paths", s.getConfigPaths)
	}
}
