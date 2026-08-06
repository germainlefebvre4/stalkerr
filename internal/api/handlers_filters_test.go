package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glefebvre/stalkeer/internal/models"
)

func TestCreateFilter_DuplicateNameReturnsFilterCreateFailed(t *testing.T) {
	db := setupTestDB(t)

	// Different attribute than the request below, so the new one-override-per-attribute
	// replacement logic doesn't delete this row before the name-uniqueness check fires.
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

func TestCreateFilter_ReplacesExistingActiveOverrideOnSameAttribute(t *testing.T) {
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
	if err := db.Where("attribute = ? AND is_runtime = ?", "group_title", true).Find(&remaining).Error; err != nil {
		t.Fatalf("failed to query remaining filters: %v", err)
	}
	if len(remaining) != 1 {
		t.Fatalf("Expected exactly 1 active override on group_title, got %d", len(remaining))
	}
	if remaining[0].Name != "New Override" {
		t.Errorf("Expected the surviving override to be 'New Override', got %q", remaining[0].Name)
	}
}

func TestCreateFilter_InvalidPatternRejectedAndNotPersisted(t *testing.T) {
	db := setupTestDB(t)

	server := NewServer()

	invalidPattern := "*"
	body, _ := json.Marshal(CreateFilterRequest{
		Name:            "Bad Filter",
		Attribute:       "tvg_name",
		IncludePatterns: &invalidPattern,
	})
	req, _ := http.NewRequest("POST", "/api/v1/filters", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d. Body: %s", w.Code, w.Body.String())
	}
	var errResp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}
	if errResp.Error != "invalid_pattern" {
		t.Errorf("Expected error code 'invalid_pattern', got %q", errResp.Error)
	}

	var count int64
	db.Model(&models.FilterConfig{}).Where("name = ?", "Bad Filter").Count(&count)
	if count != 0 {
		t.Errorf("Expected the invalid filter to not be persisted, found %d rows", count)
	}
}

func TestUpdateFilter_InvalidPatternRejectedAndLeavesFilterUnchanged(t *testing.T) {
	db := setupTestDB(t)

	validPattern := "FRENCH"
	existing := models.FilterConfig{
		Name:            "Good Filter",
		Attribute:       "group_title",
		IncludePatterns: &validPattern,
		IsRuntime:       true,
	}
	db.Create(&existing)

	server := NewServer()

	invalidPattern := "("
	body, _ := json.Marshal(UpdateFilterRequest{
		IncludePatterns: &invalidPattern,
	})
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/filters/%d", existing.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d. Body: %s", w.Code, w.Body.String())
	}
	var errResp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}
	if errResp.Error != "invalid_pattern" {
		t.Errorf("Expected error code 'invalid_pattern', got %q", errResp.Error)
	}

	var reloaded models.FilterConfig
	if err := db.First(&reloaded, existing.ID).Error; err != nil {
		t.Fatalf("failed to reload filter: %v", err)
	}
	if reloaded.IncludePatterns == nil || *reloaded.IncludePatterns != validPattern {
		t.Errorf("Expected filter to remain unchanged with include_patterns %q, got %v", validPattern, reloaded.IncludePatterns)
	}
}

func TestUpdateFilter_AttributeChangeReplacesExistingOverride(t *testing.T) {
	db := setupTestDB(t)

	db.Create(&models.FilterConfig{
		Name:      "Group Title Override",
		Attribute: "group_title",
		IsRuntime: true,
	})
	tvgFilter := models.FilterConfig{
		Name:      "Tvg Name Override",
		Attribute: "tvg_name",
		IsRuntime: true,
	}
	db.Create(&tvgFilter)

	server := NewServer()

	newAttribute := "group_title"
	body, _ := json.Marshal(UpdateFilterRequest{
		Attribute: &newAttribute,
	})
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/filters/%d", tvgFilter.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var remaining []models.FilterConfig
	if err := db.Where("attribute = ? AND is_runtime = ?", "group_title", true).Find(&remaining).Error; err != nil {
		t.Fatalf("failed to query remaining filters: %v", err)
	}
	if len(remaining) != 1 {
		t.Fatalf("Expected exactly 1 active override on group_title, got %d", len(remaining))
	}
	if remaining[0].Name != "Tvg Name Override" {
		t.Errorf("Expected the surviving override to be 'Tvg Name Override', got %q", remaining[0].Name)
	}
}
