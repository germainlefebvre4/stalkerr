package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glefebvre/stalkeer/internal/config"
)

func TestListScheduleWindows_EmptyWhenNoneCreated(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(&config.Config{})
	server := NewServer()

	req, _ := http.NewRequest("GET", "/api/v1/bandwidth-schedule/windows", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body struct {
		Windows []ScheduleWindowResponse `json:"windows"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if len(body.Windows) != 0 {
		t.Errorf("expected 0 windows, got %+v", body.Windows)
	}
}

func TestCreateScheduleWindow_StoresAndReturnsIt(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(&config.Config{})
	server := NewServer()

	body, _ := json.Marshal(ScheduleWindowRequest{
		DaysOfWeek: []string{"monday", "tuesday"},
		StartTime:  "08:00",
		EndTime:    "18:00",
		Action:     "throttle",
	})
	req, _ := http.NewRequest("POST", "/api/v1/bandwidth-schedule/windows", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var resp ScheduleWindowResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.ID == 0 || resp.Action != "throttle" || len(resp.DaysOfWeek) != 2 {
		t.Errorf("unexpected create response: %+v", resp)
	}
}

func TestCreateScheduleWindow_InvalidActionRejected(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(&config.Config{})
	server := NewServer()

	body, _ := json.Marshal(ScheduleWindowRequest{
		DaysOfWeek: []string{"monday"},
		StartTime:  "08:00",
		EndTime:    "18:00",
		Action:     "bogus",
	})
	req, _ := http.NewRequest("POST", "/api/v1/bandwidth-schedule/windows", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for an invalid action, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpdateScheduleWindow_ChangesFields(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(&config.Config{})
	server := NewServer()

	created := createTestScheduleWindow(t, server, ScheduleWindowRequest{
		DaysOfWeek: []string{"monday"}, StartTime: "08:00", EndTime: "18:00", Action: "throttle",
	})

	body, _ := json.Marshal(ScheduleWindowRequest{
		DaysOfWeek: []string{"friday", "saturday"}, StartTime: "22:00", EndTime: "07:00", Action: "stop",
	})
	req, _ := http.NewRequest("PUT", "/api/v1/bandwidth-schedule/windows/"+itoa(created.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp ScheduleWindowResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Action != "stop" || resp.StartTime != "22:00" {
		t.Errorf("unexpected update response: %+v", resp)
	}
}

func TestUpdateScheduleWindow_NotFoundReturns404(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(&config.Config{})
	server := NewServer()

	body, _ := json.Marshal(ScheduleWindowRequest{
		DaysOfWeek: []string{"monday"}, StartTime: "08:00", EndTime: "18:00", Action: "throttle",
	})
	req, _ := http.NewRequest("PUT", "/api/v1/bandwidth-schedule/windows/999", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteScheduleWindow_LeavesOthersIntact(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(&config.Config{})
	server := NewServer()

	a := createTestScheduleWindow(t, server, ScheduleWindowRequest{DaysOfWeek: []string{"monday"}, StartTime: "08:00", EndTime: "18:00", Action: "throttle"})
	b := createTestScheduleWindow(t, server, ScheduleWindowRequest{DaysOfWeek: []string{"friday"}, StartTime: "22:00", EndTime: "07:00", Action: "stop"})

	req, _ := http.NewRequest("DELETE", "/api/v1/bandwidth-schedule/windows/"+itoa(a.ID), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	listReq, _ := http.NewRequest("GET", "/api/v1/bandwidth-schedule/windows", nil)
	listW := httptest.NewRecorder()
	server.router.ServeHTTP(listW, listReq)
	var listBody struct {
		Windows []ScheduleWindowResponse `json:"windows"`
	}
	json.Unmarshal(listW.Body.Bytes(), &listBody)
	if len(listBody.Windows) != 1 || listBody.Windows[0].ID != b.ID {
		t.Errorf("expected only window b to remain, got %+v", listBody.Windows)
	}
}

func TestDeleteScheduleWindow_NotFoundReturns404(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(&config.Config{})
	server := NewServer()

	req, _ := http.NewRequest("DELETE", "/api/v1/bandwidth-schedule/windows/999", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// fakeActivePlaybackChecker lets tests control the Jellyfin signal without a
// real HTTP client.
type fakeActivePlaybackChecker struct{ active bool }

func (f fakeActivePlaybackChecker) ActivePlayback(ctx context.Context) bool { return f.active }

func withFakeActivePlayback(t *testing.T, active bool) {
	t.Helper()
	original := activePlaybackChecker
	activePlaybackChecker = func(cfg *config.Config) interface {
		ActivePlayback(ctx context.Context) bool
	} {
		return fakeActivePlaybackChecker{active: active}
	}
	t.Cleanup(func() { activePlaybackChecker = original })
}

func TestGetEffectivePolicy_NoneWhenNeitherSignalActive(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(&config.Config{})
	server := NewServer()

	req, _ := http.NewRequest("GET", "/api/v1/bandwidth-schedule/effective-policy", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp EffectivePolicyResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Action != "none" || resp.ScheduleAction != "none" || resp.JellyfinAction != "none" {
		t.Errorf("expected all-none response, got %+v", resp)
	}
}

func TestGetEffectivePolicy_ScheduleOnlyContributes(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(&config.Config{})
	server := NewServer()

	createTestScheduleWindow(t, server, ScheduleWindowRequest{
		DaysOfWeek: []string{
			"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday",
		},
		StartTime: "00:00", EndTime: "23:59", Action: "throttle",
	})

	req, _ := http.NewRequest("GET", "/api/v1/bandwidth-schedule/effective-policy", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	var resp EffectivePolicyResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Action != "throttle" || resp.ScheduleAction != "throttle" || resp.JellyfinAction != "none" {
		t.Errorf("expected schedule-only throttle, got %+v", resp)
	}
}

func TestGetEffectivePolicy_JellyfinOnlyContributes(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(&config.Config{Jellyfin: config.JellyfinConfig{
		URL:                  "http://jellyfin.example.com",
		PlaybackCheckEnabled: true,
		PlaybackAction:       "stop",
	}})
	server := NewServer()
	withFakeActivePlayback(t, true)

	req, _ := http.NewRequest("GET", "/api/v1/bandwidth-schedule/effective-policy", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	var resp EffectivePolicyResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Action != "stop" || resp.ScheduleAction != "none" || resp.JellyfinAction != "stop" {
		t.Errorf("expected jellyfin-only stop, got %+v", resp)
	}
}

func TestGetEffectivePolicy_BothContributeMostRestrictiveWins(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(&config.Config{Jellyfin: config.JellyfinConfig{
		URL:                  "http://jellyfin.example.com",
		PlaybackCheckEnabled: true,
		PlaybackAction:       "stop",
	}})
	server := NewServer()
	withFakeActivePlayback(t, true)

	createTestScheduleWindow(t, server, ScheduleWindowRequest{
		DaysOfWeek: []string{
			"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday",
		},
		StartTime: "00:00", EndTime: "23:59", Action: "throttle",
	})

	req, _ := http.NewRequest("GET", "/api/v1/bandwidth-schedule/effective-policy", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	var resp EffectivePolicyResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Action != "stop" || resp.ScheduleAction != "throttle" || resp.JellyfinAction != "stop" {
		t.Errorf("expected stop to win as most restrictive, got %+v", resp)
	}
}

func createTestScheduleWindow(t *testing.T, server *Server, req ScheduleWindowRequest) ScheduleWindowResponse {
	t.Helper()
	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", "/api/v1/bandwidth-schedule/windows", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, httpReq)
	if w.Code != http.StatusCreated {
		t.Fatalf("failed to create schedule window: %d: %s", w.Code, w.Body.String())
	}
	var resp ScheduleWindowResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp
}

func itoa(id uint) string {
	return fmt.Sprintf("%d", id)
}
