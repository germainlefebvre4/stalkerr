package notifier

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestNtfyNotifier_Notify_SuccessfulPost(t *testing.T) {
	var gotPath, gotTitle, gotPriority, gotTags, gotAuth, gotBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotTitle = r.Header.Get("Title")
		gotPriority = r.Header.Get("Priority")
		gotTags = r.Header.Get("Tags")
		gotAuth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := newNtfyNotifier(server.URL, "my-topic", "secret-token")

	err := n.Notify(context.Background(), Event{
		Severity: Critical,
		Title:    "Download failed",
		Message:  "3 items failed to download",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/my-topic" {
		t.Errorf("expected path /my-topic, got %s", gotPath)
	}
	if gotTitle != "Download failed" {
		t.Errorf("expected title %q, got %q", "Download failed", gotTitle)
	}
	if gotPriority != "high" {
		t.Errorf("expected priority 'high' for Critical severity, got %q", gotPriority)
	}
	if gotTags != "rotating_light" {
		t.Errorf("expected tags 'rotating_light' for Critical severity, got %q", gotTags)
	}
	if gotAuth != "Bearer secret-token" {
		t.Errorf("expected Authorization header 'Bearer secret-token', got %q", gotAuth)
	}
	if gotBody != "3 items failed to download" {
		t.Errorf("expected body %q, got %q", "3 items failed to download", gotBody)
	}
}

func TestNtfyNotifier_Notify_WarningSeverity(t *testing.T) {
	var gotPriority, gotTags string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPriority = r.Header.Get("Priority")
		gotTags = r.Header.Get("Tags")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := newNtfyNotifier(server.URL, "my-topic", "")

	err := n.Notify(context.Background(), Event{Severity: Warning, Title: "t", Message: "m"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPriority != "default" {
		t.Errorf("expected priority 'default' for Warning severity, got %q", gotPriority)
	}
	if gotTags != "warning" {
		t.Errorf("expected tags 'warning' for Warning severity, got %q", gotTags)
	}
}

func TestNtfyNotifier_Notify_NoAuthTokenOmitsHeader(t *testing.T) {
	var gotAuth string
	authSeen := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth, authSeen = r.Header.Get("Authorization"), r.Header.Get("Authorization") != ""
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := newNtfyNotifier(server.URL, "my-topic", "")

	if err := n.Notify(context.Background(), Event{Severity: Warning, Title: "t", Message: "m"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if authSeen {
		t.Errorf("expected no Authorization header when auth token is empty, got %q", gotAuth)
	}
}

func TestNtfyNotifier_Notify_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	n := newNtfyNotifier(server.URL, "my-topic", "")

	err := n.Notify(context.Background(), Event{Severity: Warning, Title: "t", Message: "m"})
	if err == nil {
		t.Fatal("expected error for non-200 response, got nil")
	}
}

func TestNtfyNotifier_Notify_RetriesTransientFailure(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&attempts, 1) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := newNtfyNotifier(server.URL, "my-topic", "")
	n.retry.InitialBackoff = 0
	n.retry.MaxBackoff = 0

	err := n.Notify(context.Background(), Event{Severity: Warning, Title: "t", Message: "m"})
	if err != nil {
		t.Fatalf("expected delivery to succeed after one retry, got error: %v", err)
	}
	if got := atomic.LoadInt32(&attempts); got != 2 {
		t.Errorf("expected 2 attempts, got %d", got)
	}
}

func TestNtfyNotifier_Notify_AllAttemptsFail(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	n := newNtfyNotifier(server.URL, "my-topic", "")
	n.retry.InitialBackoff = 0
	n.retry.MaxBackoff = 0

	err := n.Notify(context.Background(), Event{Severity: Warning, Title: "t", Message: "m"})
	if err == nil {
		t.Fatal("expected error when all attempts fail, got nil")
	}
	if got := int(atomic.LoadInt32(&attempts)); got != n.retry.MaxAttempts {
		t.Errorf("expected %d attempts, got %d", n.retry.MaxAttempts, got)
	}
}
