package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glefebvre/stalkeer/internal/models"
)

func TestCreateFilter_DuplicateNameReturnsFilterCreateFailed(t *testing.T) {
	db := setupTestDB(t)

	db.Create(&models.FilterConfig{
		Name:      "Existing Filter",
		Attribute: "group_title",
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
