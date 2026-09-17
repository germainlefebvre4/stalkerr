package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/external/jellyfin"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/glefebvre/stalkeer/internal/settings"
)

// ScheduleWindowResponse is the API-facing shape of one bandwidth schedule
// window.
type ScheduleWindowResponse struct {
	ID         uint     `json:"id"`
	DaysOfWeek []string `json:"days_of_week"`
	StartTime  string   `json:"start_time"`
	EndTime    string   `json:"end_time"`
	Action     string   `json:"action"`
}

// ScheduleWindowRequest is the create/update payload for a bandwidth
// schedule window.
type ScheduleWindowRequest struct {
	DaysOfWeek []string `json:"days_of_week" binding:"required"`
	StartTime  string   `json:"start_time" binding:"required"`
	EndTime    string   `json:"end_time" binding:"required"`
	Action     string   `json:"action" binding:"required"`
}

// EffectivePolicyResponse reports the currently effective download policy
// and which signal(s) are contributing to it. See the
// adaptive-download-throttling spec.
type EffectivePolicyResponse struct {
	Action         string `json:"action"`
	ScheduleAction string `json:"schedule_action"`
	JellyfinAction string `json:"jellyfin_action"`
}

// listScheduleWindows handles GET /api/v1/bandwidth-schedule/windows.
func (s *Server) listScheduleWindows(c *gin.Context) {
	windows, err := settings.ListScheduleWindows()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "database_error", "failed to fetch schedule windows")
		return
	}
	c.JSON(http.StatusOK, gin.H{"windows": toScheduleWindowResponses(windows)})
}

// createScheduleWindow handles POST /api/v1/bandwidth-schedule/windows.
func (s *Server) createScheduleWindow(c *gin.Context) {
	var req ScheduleWindowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	window, err := settings.CreateScheduleWindow(toScheduleWindowInput(req))
	if err != nil {
		respondScheduleWindowError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toScheduleWindowResponse(window))
}

// updateScheduleWindow handles PUT /api/v1/bandwidth-schedule/windows/:id.
func (s *Server) updateScheduleWindow(c *gin.Context) {
	id, err := parseScheduleWindowID(c)
	if err != nil {
		return
	}

	var req ScheduleWindowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	window, err := settings.UpdateScheduleWindow(id, toScheduleWindowInput(req))
	if err != nil {
		respondScheduleWindowError(c, err)
		return
	}
	c.JSON(http.StatusOK, toScheduleWindowResponse(window))
}

// deleteScheduleWindow handles DELETE /api/v1/bandwidth-schedule/windows/:id.
// Deleting one window leaves every other window unchanged. See "Managing
// Schedule Windows Independently".
func (s *Server) deleteScheduleWindow(c *gin.Context) {
	id, err := parseScheduleWindowID(c)
	if err != nil {
		return
	}

	if err := settings.DeleteScheduleWindow(id); err != nil {
		respondScheduleWindowError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "schedule window deleted successfully"})
}

// getEffectivePolicy handles GET /api/v1/bandwidth-schedule/effective-policy:
// the currently effective download policy and which signal(s) contribute to
// it. The API server is long-lived (unlike the download/resume-downloads
// commands), so this resolves fresh on every request rather than sharing a
// per-run PolicyEngine: the weekly schedule is evaluated against the
// current instant, and Jellyfin (when configured) is checked live, already
// fail-open on any error. See adaptive-download-throttling's "Effective
// Policy Combination".
func (s *Server) getEffectivePolicy(c *gin.Context) {
	cfg := settings.Effective()

	windows, err := settings.ListScheduleWindows()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "database_error", "failed to fetch schedule windows")
		return
	}
	scheduleAction := settings.ActiveScheduleAction(windows, time.Now())

	jellyfinAction := settings.ActionNone
	if cfg.Jellyfin.PlaybackCheckEnabled && cfg.Jellyfin.URL != "" {
		if activePlaybackChecker(cfg).ActivePlayback(c.Request.Context()) {
			jellyfinAction = cfg.Jellyfin.PlaybackAction
		}
	}

	c.JSON(http.StatusOK, EffectivePolicyResponse{
		Action:         settings.MostRestrictive(scheduleAction, jellyfinAction),
		ScheduleAction: scheduleAction,
		JellyfinAction: jellyfinAction,
	})
}

// activePlaybackChecker is a seam for tests: production always builds a
// real Jellyfin client, but tests substitute this to avoid a real HTTP
// call.
var activePlaybackChecker = func(cfg *config.Config) interface {
	ActivePlayback(ctx context.Context) bool
} {
	return jellyfin.New(jellyfin.Config{
		BaseURL: cfg.Jellyfin.URL,
		APIKey:  cfg.Jellyfin.APIKey,
		Logger:  logger.AppLogger(),
	})
}

func parseScheduleWindowID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", "invalid schedule window id")
		return 0, err
	}
	return uint(id), nil
}

func respondScheduleWindowError(c *gin.Context, err error) {
	var invalid *settings.InvalidScheduleWindowError
	switch {
	case errors.As(err, &invalid):
		respondError(c, http.StatusBadRequest, "invalid_schedule_window", invalid.Error())
	case errors.Is(err, settings.ErrScheduleWindowNotFound):
		respondError(c, http.StatusNotFound, "not_found", "schedule window not found")
	default:
		respondError(c, http.StatusInternalServerError, "database_error", fmt.Sprintf("failed to store schedule window: %v", err))
	}
}

func toScheduleWindowResponses(windows []models.BandwidthScheduleWindow) []ScheduleWindowResponse {
	out := make([]ScheduleWindowResponse, len(windows))
	for i, w := range windows {
		out[i] = toScheduleWindowResponse(w)
	}
	return out
}

func toScheduleWindowResponse(w models.BandwidthScheduleWindow) ScheduleWindowResponse {
	return ScheduleWindowResponse{
		ID:         w.ID,
		DaysOfWeek: []string(w.DaysOfWeek),
		StartTime:  w.StartTime,
		EndTime:    w.EndTime,
		Action:     w.Action,
	}
}

func toScheduleWindowInput(req ScheduleWindowRequest) settings.ScheduleWindowInput {
	return settings.ScheduleWindowInput{
		DaysOfWeek: req.DaysOfWeek,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Action:     req.Action,
	}
}
