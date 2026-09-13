package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/glefebvre/stalkeer/internal/settings"
)

// M3USourceResponse is the API-facing shape of one M3U source (origin,
// runtime, or effective), masking its auth password per "Sensitive Source
// Field Masking".
type M3USourceResponse struct {
	Name            string `json:"name"`
	FilePath        string `json:"file_path"`
	Enabled         bool   `json:"enabled"`
	URL             string `json:"url"`
	ArchiveDir      string `json:"archive_dir"`
	RetentionCount  int    `json:"retention_count"`
	MaxFileSizeMB   int64  `json:"max_file_size_mb"`
	TimeoutSeconds  int    `json:"timeout_seconds"`
	RetryAttempts   int    `json:"retry_attempts"`
	AuthUsername    string `json:"auth_username"`
	HasAuthPassword bool   `json:"has_auth_password"`
	IsRuntime       bool   `json:"is_runtime"`
}

// M3USourceRequest is the create/update payload for a runtime M3U source's
// download settings. A nil AuthPassword means "omitted": keep the name's
// current effective password unchanged; an explicit empty string clears
// it. See "Sensitive Source Field Masking".
type M3USourceRequest struct {
	FilePath       string  `json:"file_path" binding:"required"`
	Enabled        bool    `json:"enabled"`
	URL            string  `json:"url"`
	ArchiveDir     string  `json:"archive_dir"`
	RetentionCount int     `json:"retention_count"`
	MaxFileSizeMB  int64   `json:"max_file_size_mb"`
	TimeoutSeconds int     `json:"timeout_seconds"`
	RetryAttempts  int     `json:"retry_attempts"`
	AuthUsername   string  `json:"auth_username"`
	AuthPassword   *string `json:"auth_password"`
}

// M3USourceCreateRequest is the POST /api/v1/m3u/sources request body: the
// same fields as M3USourceRequest, plus the new source's name.
type M3USourceCreateRequest struct {
	Name string `json:"name" binding:"required"`
	M3USourceRequest
}

// listOriginM3USources handles GET /api/v1/m3u/sources/origin.
func (s *Server) listOriginM3USources(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"sources": toM3USourceResponses(settings.OriginSources())})
}

// listEffectiveM3USources handles GET /api/v1/m3u/sources.
func (s *Server) listEffectiveM3USources(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"sources": toM3USourceResponses(settings.EffectiveEntries())})
}

// createM3USource handles POST /api/v1/m3u/sources.
func (s *Server) createM3USource(c *gin.Context) {
	var req M3USourceCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	s.storeM3USource(c, req.Name, req.M3USourceRequest)
}

// updateM3USource handles PUT /api/v1/m3u/sources/:name.
func (s *Server) updateM3USource(c *gin.Context) {
	var req M3USourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	s.storeM3USource(c, c.Param("name"), req)
}

func (s *Server) storeM3USource(c *gin.Context, name string, req M3USourceRequest) {
	input := settings.M3USourceInput{
		FilePath:       req.FilePath,
		Enabled:        req.Enabled,
		URL:            req.URL,
		ArchiveDir:     req.ArchiveDir,
		RetentionCount: req.RetentionCount,
		MaxFileSizeMB:  req.MaxFileSizeMB,
		TimeoutSeconds: req.TimeoutSeconds,
		RetryAttempts:  req.RetryAttempts,
		AuthUsername:   req.AuthUsername,
		AuthPassword:   req.AuthPassword,
	}

	source, err := settings.SetSource(name, input)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "database_error", "failed to store M3U source")
		return
	}

	c.JSON(http.StatusOK, toM3USourceResponse(settings.M3USourceEntry{Source: source, IsRuntime: true}))
}

// deleteM3USource handles DELETE /api/v1/m3u/sources/:name. Deleting a
// name with no stored runtime source is rejected as not found, whether or
// not an origin source of that name exists; the origin source, if any, is
// unaffected. See "Removing a Runtime Source".
func (s *Server) deleteM3USource(c *gin.Context) {
	name := c.Param("name")

	if err := settings.DeleteSource(name); err != nil {
		if errors.Is(err, settings.ErrM3USourceNotFound) {
			respondError(c, http.StatusNotFound, "not_found", fmt.Sprintf("no runtime source stored for %q", name))
			return
		}
		respondError(c, http.StatusInternalServerError, "database_error", "failed to delete M3U source")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "M3U source deleted successfully"})
}

func toM3USourceResponses(entries []settings.M3USourceEntry) []M3USourceResponse {
	out := make([]M3USourceResponse, len(entries))
	for i, e := range entries {
		out[i] = toM3USourceResponse(e)
	}
	return out
}

func toM3USourceResponse(e settings.M3USourceEntry) M3USourceResponse {
	return M3USourceResponse{
		Name:            e.Source.Name,
		FilePath:        e.Source.FilePath,
		Enabled:         e.Source.Download.Enabled,
		URL:             e.Source.Download.URL,
		ArchiveDir:      e.Source.Download.ArchiveDir,
		RetentionCount:  e.Source.Download.RetentionCount,
		MaxFileSizeMB:   e.Source.Download.MaxFileSizeMB,
		TimeoutSeconds:  e.Source.Download.TimeoutSeconds,
		RetryAttempts:   e.Source.Download.RetryAttempts,
		AuthUsername:    e.Source.Download.AuthUsername,
		HasAuthPassword: e.HasAuthPassword(),
		IsRuntime:       e.IsRuntime,
	}
}
