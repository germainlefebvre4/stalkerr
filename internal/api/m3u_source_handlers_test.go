package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glefebvre/stalkeer/internal/config"
)

func TestListOriginM3USources_ReturnsConfigYamlSources(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(&config.Config{M3U: config.M3UConfig{Sources: []config.M3USourceConfig{
		{Name: "a", FilePath: "/tmp/a.m3u"},
	}}})
	server := NewServer()

	req, _ := http.NewRequest("GET", "/api/v1/m3u/sources/origin", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body struct {
		Sources []M3USourceResponse `json:"sources"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if len(body.Sources) != 1 || body.Sources[0].Name != "a" || body.Sources[0].IsRuntime {
		t.Errorf("unexpected origin sources response: %+v", body.Sources)
	}
}

func TestCreateM3USource_RuntimeOnly(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(&config.Config{})
	server := NewServer()

	body, _ := json.Marshal(M3USourceCreateRequest{
		Name: "c",
		M3USourceRequest: M3USourceRequest{
			FilePath: "/tmp/c.m3u",
			URL:      "http://c.example.com",
		},
	})
	req, _ := http.NewRequest("POST", "/api/v1/m3u/sources", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp M3USourceResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Name != "c" || !resp.IsRuntime || resp.URL != "http://c.example.com" {
		t.Errorf("unexpected create response: %+v", resp)
	}
}

func TestUpdateM3USource_OverridesOrigin(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(&config.Config{M3U: config.M3UConfig{Sources: []config.M3USourceConfig{
		{Name: "b", FilePath: "/tmp/b-origin.m3u"},
	}}})
	server := NewServer()

	body, _ := json.Marshal(M3USourceRequest{FilePath: "/tmp/b-override.m3u", URL: "http://override.example.com"})
	req, _ := http.NewRequest("PUT", "/api/v1/m3u/sources/b", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	listReq, _ := http.NewRequest("GET", "/api/v1/m3u/sources", nil)
	listW := httptest.NewRecorder()
	server.router.ServeHTTP(listW, listReq)

	var listBody struct {
		Sources []M3USourceResponse `json:"sources"`
	}
	json.Unmarshal(listW.Body.Bytes(), &listBody)
	if len(listBody.Sources) != 1 || !listBody.Sources[0].IsRuntime || listBody.Sources[0].FilePath != "/tmp/b-override.m3u" {
		t.Errorf("unexpected effective sources after override: %+v", listBody.Sources)
	}
}

func TestDeleteM3USource_RuntimeOnlyRemoved(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(&config.Config{})
	server := NewServer()

	createBody, _ := json.Marshal(M3USourceCreateRequest{Name: "c", M3USourceRequest: M3USourceRequest{FilePath: "/tmp/c.m3u"}})
	createReq, _ := http.NewRequest("POST", "/api/v1/m3u/sources", bytes.NewBuffer(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	server.router.ServeHTTP(httptest.NewRecorder(), createReq)

	delReq, _ := http.NewRequest("DELETE", "/api/v1/m3u/sources/c", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, delReq)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteM3USource_NoRuntimeSourceReturns404(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(&config.Config{M3U: config.M3UConfig{Sources: []config.M3USourceConfig{
		{Name: "a", FilePath: "/tmp/a.m3u"},
	}}})
	server := NewServer()

	req, _ := http.NewRequest("DELETE", "/api/v1/m3u/sources/a", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 deleting an origin-only source, got %d: %s", w.Code, w.Body.String())
	}
}

func TestM3USourceResponse_MasksAuthPassword(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(&config.Config{})
	server := NewServer()

	pw := "s3cret"
	body, _ := json.Marshal(M3USourceCreateRequest{
		Name: "c",
		M3USourceRequest: M3USourceRequest{
			FilePath:     "/tmp/c.m3u",
			AuthPassword: &pw,
		},
	})
	req, _ := http.NewRequest("POST", "/api/v1/m3u/sources", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	raw := w.Body.String()
	if bytes.Contains([]byte(raw), []byte("s3cret")) {
		t.Fatalf("response must never contain the raw password: %s", raw)
	}

	var resp M3USourceResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.HasAuthPassword {
		t.Error("expected has_auth_password=true")
	}
}
