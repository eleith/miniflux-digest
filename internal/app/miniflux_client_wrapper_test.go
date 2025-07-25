package app_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"miniflux-digest/internal/app"
	miniflux "miniflux.app/v2/client"
)

func TestMinifluxClientWrapper_StreamAllCategoryData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/categories":
			if _, err := fmt.Fprintln(w, `[{"id": 1, "title": "Test Category 1"}, {"id": 2, "title": "Test Category 2"}]`); err != nil {
				panic(err)
			}
		case "/v1/categories/1/entries":
			if _, err := fmt.Fprintln(w, `{"total": 1, "entries": [{"id": 101, "title": "Entry 1A"}]}`); err != nil {
				panic(err)
			}
		case "/v1/categories/2/entries":
			if _, err := fmt.Fprintln(w, `{"total": 1, "entries": [{"id": 102, "title": "Entry 2A"}]}`); err != nil {
				panic(err)
			}
		case "/v1/categories/1/feeds", "/v1/categories/2/feeds":
			if _, err := fmt.Fprintln(w, `[{"id": 1, "title": "Test Feed"}]`); err != nil {
				panic(err)
			}
		case "/v1/feeds/1/icon":
			if _, err := fmt.Fprintln(w, `{"data": "icon-data", "mime_type": "image/png"}`); err != nil {
				panic(err)
			}
		default:
			t.Fatalf("Unexpected request path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := miniflux.NewClient(server.URL, "test-token")
	wrapper := app.NewMinifluxClientWrapper(client)

	dataChannel := wrapper.StreamAllCategoryData()

	var receivedCount int
	for data := range dataChannel {
		receivedCount++
		if data.Category == nil {
			t.Errorf("Expected category data, but got nil category")
		}
		if data.Entries == nil {
			t.Errorf("Expected entries data, but got nil entries")
		}
	}

	expectedCount := 2
	if receivedCount != expectedCount {
		t.Errorf("Expected to receive data for %d categories, but got %d", expectedCount, receivedCount)
	}
}

func TestMinifluxClientWrapper_MarkCategoryAsRead(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := miniflux.NewClient(server.URL, "test-token")
	wrapper := app.NewMinifluxClientWrapper(client)

	if err := wrapper.MarkCategoryAsRead(1); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestMinifluxClientWrapper_CategoryEntries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprintln(w, `{"total": 1, "entries": [{"id": 101, "title": "Entry 1A"}]}`); err != nil {
			panic(err)
		}
	}))
	defer server.Close()

	client := miniflux.NewClient(server.URL, "test-token")
	wrapper := app.NewMinifluxClientWrapper(client)

	entries, err := wrapper.CategoryEntries(1, nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(*entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(*entries))
	}
}

func TestMinifluxClientWrapper_CategoryFeeds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprintln(w, `[{"id": 1, "title": "Test Feed"}]`); err != nil {
			panic(err)
		}
	}))
	defer server.Close()

	client := miniflux.NewClient(server.URL, "test-token")
	wrapper := app.NewMinifluxClientWrapper(client)

	feeds, err := wrapper.CategoryFeeds(1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(feeds) != 1 {
		t.Errorf("Expected 1 feed, got %d", len(feeds))
	}
}

func TestMinifluxClientWrapper_FeedIcon(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprintln(w, `{"data": "icon-data", "mime_type": "image/png"}`); err != nil {
			panic(err)
		}
	}))
	defer server.Close()

	client := miniflux.NewClient(server.URL, "test-token")
	wrapper := app.NewMinifluxClientWrapper(client)

	icon, err := wrapper.FeedIcon(1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if icon.Data != "icon-data" {
		t.Errorf("Expected icon data 'icon-data', got '%s'", icon.Data)
	}
}