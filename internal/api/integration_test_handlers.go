package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/glefebvre/stalkeer/internal/external/jellyfin"
	"github.com/glefebvre/stalkeer/internal/external/radarr"
	"github.com/glefebvre/stalkeer/internal/external/sonarr"
	"github.com/glefebvre/stalkeer/internal/retry"
)

// integrationTestRequest is the request body for POST
// /api/v1/settings/integrations/test: a caller-supplied (not necessarily
// saved) service/url/api_key combination to check live.
type integrationTestRequest struct {
	Service string `json:"service"`
	URL     string `json:"url"`
	APIKey  string `json:"api_key"`
}

// testIntegrationConnectivity handles POST /api/v1/settings/integrations/test,
// performing a live, on-demand reachability-and-authentication check against
// the caller-supplied url/api_key for radarr, sonarr, or jellyfin - never the
// currently saved configuration for that service, and never reading or
// writing any stored setting. Unlike the aggregated /system/status checks,
// this bypasses the shared circuit breaker entirely (Breaker: nil), so a
// manual test always attempts a live call and never counts toward that
// breaker's state.
func (s *Server) testIntegrationConnectivity(c *gin.Context) {
	var req integrationTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), systemStatusCheckTimeout)
	defer cancel()

	var result ServiceStatus
	switch req.Service {
	case "radarr":
		result = checkRadarrConnectivity(ctx, req.URL, req.APIKey)
	case "sonarr":
		result = checkSonarrConnectivity(ctx, req.URL, req.APIKey)
	case "jellyfin":
		result = checkJellyfinConnectivity(ctx, req.URL, req.APIKey)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported service"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// checkRadarrConnectivity builds a throwaway, unprotected Radarr client from
// caller-supplied values (Breaker: nil, single attempt) and classifies the
// SystemStatus result the same way the aggregated status check does.
func checkRadarrConnectivity(ctx context.Context, url, apiKey string) ServiceStatus {
	client := radarr.New(radarr.Config{
		BaseURL:     url,
		APIKey:      apiKey,
		Timeout:     systemStatusCheckTimeout,
		RetryConfig: retry.Config{MaxAttempts: 1},
		Breaker:     nil,
	})

	if err := client.SystemStatus(ctx); err != nil {
		return ServiceStatus{Status: statusKO, Reason: classifyReachabilityError(err)}
	}
	return ServiceStatus{Status: statusOK}
}

// checkSonarrConnectivity mirrors checkRadarrConnectivity for Sonarr.
func checkSonarrConnectivity(ctx context.Context, url, apiKey string) ServiceStatus {
	client := sonarr.New(sonarr.Config{
		BaseURL:     url,
		APIKey:      apiKey,
		Timeout:     systemStatusCheckTimeout,
		RetryConfig: retry.Config{MaxAttempts: 1},
		Breaker:     nil,
	})

	if err := client.SystemStatus(ctx); err != nil {
		return ServiceStatus{Status: statusKO, Reason: classifyReachabilityError(err)}
	}
	return ServiceStatus{Status: statusOK}
}

// checkJellyfinConnectivity mirrors checkRadarrConnectivity for Jellyfin. The
// Jellyfin client has no circuit breaker at all, so there is nothing to
// bypass here.
func checkJellyfinConnectivity(ctx context.Context, url, apiKey string) ServiceStatus {
	client := jellyfin.New(jellyfin.Config{
		BaseURL:     url,
		APIKey:      apiKey,
		Timeout:     systemStatusCheckTimeout,
		RetryConfig: retry.Config{MaxAttempts: 1},
	})

	if err := client.SystemStatus(ctx); err != nil {
		return ServiceStatus{Status: statusKO, Reason: classifyReachabilityError(err)}
	}
	return ServiceStatus{Status: statusOK}
}
