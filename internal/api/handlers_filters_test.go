package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/models"
)

// A duplicate name across two *different* attributes still violates the
// unique name constraint (attribute-based replacement only applies within
// the same attribute) and must still surface as filter_create_failed.
func TestCreateFilter_DuplicateNameAcrossAttributesReturnsFilterCreateFailed(t *testing.T) {
	db := setupTestDB(t)

	db.Create(&models.FilterConfig{
		Name:      "Existing Filter",
		Attribute: "tvg_name",
		IsRuntime: true,
	})

	server := NewServer()

	body, _ := json.Marshal(CreateFilterRequest{
		Name:      "Existing Filter",
		Attribute: "group_title",
	})
	req, _ := http.NewRequest("POST", "/api/v1/filters", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("Expected status 500, got %d. Body: %s", w.Code, w.Body.String())
	}
	var errResp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}
	if errResp.Error != "filter_create_failed" {
		t.Errorf("Expected error code 'filter_create_failed', got %q", errResp.Error)
	}
}

// filter-override-policy: creating a runtime filter for an attribute that
// already has an active override replaces it rather than erroring or
// leaving both active.
func TestCreateFilter_ReplacesExistingOverrideOnSameAttribute(t *testing.T) {
	db := setupTestDB(t)

	db.Create(&models.FilterConfig{
		Name:      "Old Override",
		Attribute: "group_title",
		IsRuntime: true,
	})

	server := NewServer()

	body, _ := json.Marshal(CreateFilterRequest{
		Name:      "New Override",
		Attribute: "group_title",
	})
	req, _ := http.NewRequest("POST", "/api/v1/filters", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var remaining []models.FilterConfig
	if err := db.Where("attribute = ?", "group_title").Find(&remaining).Error; err != nil {
		t.Fatalf("failed to query filters: %v", err)
	}
	if len(remaining) != 1 {
		t.Fatalf("expected exactly 1 runtime filter for group_title, got %d", len(remaining))
	}
	if remaining[0].Name != "New Override" {
		t.Errorf("expected the new override to replace the old one, got name %q", remaining[0].Name)
	}
}

// filter-override-policy: updating an existing filter's attribute onto one
// that already has a different active override replaces that override.
func TestUpdateFilter_ReplacesExistingOverrideOnTargetAttribute(t *testing.T) {
	db := setupTestDB(t)

	toMove := models.FilterConfig{Name: "Moving Filter", Attribute: "tvg_name", IsRuntime: true}
	db.Create(&toMove)
	db.Create(&models.FilterConfig{Name: "Target Attribute Override", Attribute: "group_title", IsRuntime: true})

	server := NewServer()

	newAttribute := "group_title"
	body, _ := json.Marshal(UpdateFilterRequest{Attribute: &newAttribute})
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/filters/%d", toMove.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var remaining []models.FilterConfig
	if err := db.Where("attribute = ?", "group_title").Find(&remaining).Error; err != nil {
		t.Fatalf("failed to query filters: %v", err)
	}
	if len(remaining) != 1 {
		t.Fatalf("expected exactly 1 runtime filter for group_title, got %d", len(remaining))
	}
	if remaining[0].Name != "Moving Filter" {
		t.Errorf("expected the moved filter to replace the target attribute's override, got name %q", remaining[0].Name)
	}
}

func TestListOriginFilters_ReturnsConfigYmlPatterns(t *testing.T) {
	setupTestDB(t)
	original := config.Get()
	t.Cleanup(func() { config.SetConfig(original) })

	cfg := &config.Config{}
	cfg.Filter.GroupTitle.IncludePatterns = []string{"FRENCH", "TRUEFRENCH"}
	cfg.Filter.GroupTitle.ExcludePatterns = []string{"VOSTFR"}
	// TvgName left with no configured patterns.
	config.SetConfig(cfg)

	server := NewServer()

	req, _ := http.NewRequest("GET", "/api/v1/filters/origin", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Origin []FilterOriginEntry `json:"origin"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(resp.Origin) != 2 {
		t.Fatalf("expected 2 origin entries, got %d", len(resp.Origin))
	}

	byAttribute := map[string]FilterOriginEntry{}
	for _, entry := range resp.Origin {
		byAttribute[entry.Attribute] = entry
	}

	groupTitle, ok := byAttribute["group_title"]
	if !ok {
		t.Fatal("expected a group_title entry")
	}
	if len(groupTitle.IncludePatterns) != 2 || groupTitle.IncludePatterns[0] != "FRENCH" {
		t.Errorf("unexpected group_title include patterns: %v", groupTitle.IncludePatterns)
	}
	if len(groupTitle.ExcludePatterns) != 1 || groupTitle.ExcludePatterns[0] != "VOSTFR" {
		t.Errorf("unexpected group_title exclude patterns: %v", groupTitle.ExcludePatterns)
	}

	tvgName, ok := byAttribute["tvg_name"]
	if !ok {
		t.Fatal("expected a tvg_name entry")
	}
	if tvgName.IncludePatterns == nil || len(tvgName.IncludePatterns) != 0 {
		t.Errorf("expected an empty (non-nil) include patterns list for tvg_name, got %v", tvgName.IncludePatterns)
	}
	if tvgName.ExcludePatterns == nil || len(tvgName.ExcludePatterns) != 0 {
		t.Errorf("expected an empty (non-nil) exclude patterns list for tvg_name, got %v", tvgName.ExcludePatterns)
	}
}
