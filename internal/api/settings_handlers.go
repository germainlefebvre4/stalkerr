package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/settings"
)

// SettingsFieldResponse represents one overridable (Tier-1) field's
// effective value and metadata. For a sensitive field, Value is omitted and
// IsSet reports only whether a value is currently resolved - see the
// app-settings spec's "Sensitive Field Masking" requirement.
type SettingsFieldResponse struct {
	Key             string      `json:"key"`
	Value           interface{} `json:"value,omitempty"`
	IsSet           *bool       `json:"is_set,omitempty"`
	Sensitive       bool        `json:"sensitive"`
	Origin          string      `json:"origin"`
	RestartRequired bool        `json:"restart_required"`
}

// BootstrapFieldResponse represents one read-only Tier-0 field. Origin is
// always "config": bootstrap fields are never overridden. See "Bootstrap
// Configuration Remains Read-Only".
type BootstrapFieldResponse struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value,omitempty"`
	IsSet     *bool       `json:"is_set,omitempty"`
	Sensitive bool        `json:"sensitive"`
	Origin    string      `json:"origin"`
}

// SetSettingsFieldRequest is the PUT /api/v1/settings/:key request body.
type SetSettingsFieldRequest struct {
	Value json.RawMessage `json:"value" binding:"required"`
}

// listSettings handles GET /api/v1/settings: every overridable field's
// effective value, origin, sensitivity, and restart-required flag.
func (s *Server) listSettings(c *gin.Context) {
	eff := settings.Effective()
	registry := settings.Registry()
	fields := make([]SettingsFieldResponse, 0, len(registry))
	for _, f := range registry {
		fields = append(fields, toSettingsFieldResponse(f, eff))
	}
	c.JSON(http.StatusOK, gin.H{"settings": fields})
}

// getBootstrapSettings handles GET /api/v1/settings/origin: the read-only
// Tier-0 bootstrap configuration, with its origin always "config".
func (s *Server) getBootstrapSettings(c *gin.Context) {
	cfg := config.Get()
	entries := []BootstrapFieldResponse{
		toBootstrapFieldResponse("database.host", cfg.Database.Host, false),
		toBootstrapFieldResponse("database.port", cfg.Database.Port, false),
		toBootstrapFieldResponse("database.user", cfg.Database.User, false),
		toBootstrapFieldResponse("database.password", cfg.Database.Password, true),
		toBootstrapFieldResponse("database.dbname", cfg.Database.DBName, false),
		toBootstrapFieldResponse("database.sslmode", cfg.Database.SSLMode, false),
		toBootstrapFieldResponse("api.port", cfg.API.Port, false),
		toBootstrapFieldResponse("metrics.port", cfg.Metrics.Port, false),
		toBootstrapFieldResponse("metrics.enabled", cfg.Metrics.Enabled, false),
		toBootstrapFieldResponse("metrics.path", cfg.Metrics.Path, false),
	}
	c.JSON(http.StatusOK, gin.H{"bootstrap": entries})
}

// setSettingsField handles PUT /api/v1/settings/:key: sets or replaces the
// stored override for one field.
func (s *Server) setSettingsField(c *gin.Context) {
	key := c.Param("key")

	var req SetSettingsFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	if err := settings.SetOverride(key, req.Value); err != nil {
		respondSettingsError(c, key, err)
		return
	}

	if key == "logging.app.level" || key == "logging.database.level" || key == "logging.format" {
		applyLiveLoggingConfig()
	}

	respondSettingsField(c, key)
}

// clearSettingsField handles DELETE /api/v1/settings/:key: removes the
// stored override, reverting the field to its file/env/default value.
func (s *Server) clearSettingsField(c *gin.Context) {
	key := c.Param("key")

	if err := settings.ClearOverride(key); err != nil {
		respondSettingsError(c, key, err)
		return
	}

	if key == "logging.app.level" || key == "logging.database.level" || key == "logging.format" {
		applyLiveLoggingConfig()
	}

	respondSettingsField(c, key)
}

// applyLiveLoggingConfig re-invokes the logger initializer with the current
// effective logging configuration, so a logging.* override applies
// immediately rather than requiring a restart (see design.md's "Restart-
// required subset").
func applyLiveLoggingConfig() {
	eff := settings.Effective()
	logger.InitializeLoggersWithFormat(eff.GetAppLogLevel(), eff.GetDatabaseLogLevel(), eff.Logging.Format)
}

func respondSettingsField(c *gin.Context, key string) {
	field, ok := settings.FieldByKey(key)
	if !ok {
		// Unreachable in practice: SetOverride/ClearOverride already
		// validated key against the registry above.
		respondError(c, http.StatusInternalServerError, "internal_error", "field lookup failed after write")
		return
	}
	c.JSON(http.StatusOK, toSettingsFieldResponse(field, settings.Effective()))
}

func respondSettingsError(c *gin.Context, key string, err error) {
	switch {
	case errors.Is(err, settings.ErrBootstrapField):
		respondError(c, http.StatusBadRequest, "bootstrap_field_readonly", fmt.Sprintf("%s is bootstrap configuration and cannot be overridden", key))
	case errors.Is(err, settings.ErrUnknownField):
		respondError(c, http.StatusNotFound, "unknown_settings_field", fmt.Sprintf("unknown settings field: %s", key))
	default:
		respondError(c, http.StatusInternalServerError, "database_error", "failed to update setting")
	}
}

func toSettingsFieldResponse(f settings.FieldSpec, cfg *config.Config) SettingsFieldResponse {
	resp := SettingsFieldResponse{
		Key:             f.Key,
		Sensitive:       f.Sensitive,
		Origin:          settings.Origin(f.Key),
		RestartRequired: f.RestartRequired,
	}
	value := f.Get(cfg)
	if f.Sensitive {
		isSet := !isZeroValue(value)
		resp.IsSet = &isSet
	} else {
		resp.Value = value
	}
	return resp
}

func toBootstrapFieldResponse(key string, value interface{}, sensitive bool) BootstrapFieldResponse {
	resp := BootstrapFieldResponse{
		Key:       key,
		Sensitive: sensitive,
		Origin:    settings.OriginConfig,
	}
	if sensitive {
		isSet := !isZeroValue(value)
		resp.IsSet = &isSet
	} else {
		resp.Value = value
	}
	return resp
}

func isZeroValue(v interface{}) bool {
	switch t := v.(type) {
	case string:
		return t == ""
	case int:
		return t == 0
	case int64:
		return t == 0
	case float64:
		return t == 0
	default:
		return v == nil
	}
}
