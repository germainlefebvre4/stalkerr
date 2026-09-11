package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/retry"
)

type testItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func testClient(baseURL string) *Client {
	return New(baseURL, "test-key", 5*time.Second, retry.Config{MaxAttempts: 1}, nil)
}

func TestGetDecodesSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "test-key" {
			t.Errorf("expected X-Api-Key header to be set")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(testItem{ID: 1, Name: "widget"})
	}))
	defer server.Close()

	item, err := Get[testItem](context.Background(), testClient(server.URL), "/items/1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != 1 || item.Name != "widget" {
		t.Errorf("unexpected item: %+v", item)
	}
}

func TestGetPageDecodesSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			TotalRecords int        `json:"totalRecords"`
			Records      []testItem `json:"records"`
		}{
			TotalRecords: 2,
			Records:      []testItem{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}},
		})
	}))
	defer server.Close()

	records, total, err := GetPage[testItem](context.Background(), testClient(server.URL), "/items")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(records) != 2 || records[0].ID != 1 || records[1].ID != 2 {
		t.Errorf("unexpected records: %+v", records)
	}
}

func TestGetNonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	_, err := Get[testItem](context.Background(), testClient(server.URL), "/items/1")
	if err == nil {
		t.Fatal("expected an error")
	}
	want := "unexpected status code 500: boom"
	if err.Error() != want {
		t.Errorf("expected error %q, got %q", want, err.Error())
	}
}

func TestPutNonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("nope"))
	}))
	defer server.Close()

	err := Put(context.Background(), testClient(server.URL), "/items/1", testItem{ID: 1})
	if err == nil {
		t.Fatal("expected an error")
	}
	want := "unexpected status code 400: nope"
	if err.Error() != want {
		t.Errorf("expected error %q, got %q", want, err.Error())
	}
}

func TestPutSuccessAcceptsOKAndAccepted(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusAccepted} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
		}))

		err := Put(context.Background(), testClient(server.URL), "/items/1", testItem{ID: 1})
		server.Close()

		if err != nil {
			t.Errorf("status %d: unexpected error: %v", status, err)
		}
	}
}

func TestGetDecodeFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	_, err := Get[testItem](context.Background(), testClient(server.URL), "/items/1")
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.HasPrefix(err.Error(), "failed to decode response: ") {
		t.Errorf("expected decode error prefix, got %q", err.Error())
	}
}

func TestNewRequestMarshalFailure(t *testing.T) {
	c := testClient("http://example.invalid")

	_, err := c.NewRequest(context.Background(), "PUT", "/items/1", func() {})
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.HasPrefix(err.Error(), "failed to marshal request body: ") {
		t.Errorf("expected marshal error prefix, got %q", err.Error())
	}
}
