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
			fmt.Fprintln(w, `[{"id": 1, "title": "Test Category 1"}, {"id": 2, "title": "Test Category 2"}]`)
		case "/v1/categories/1/entries":
			fmt.Fprintln(w, `{"total": 1, "entries": [{"id": 101, "title": "Entry 1A"}]}`)
		case "/v1/categories/2/entries":
			fmt.Fprintln(w, `{"total": 1, "entries": [{"id": 102, "title": "Entry 2A"}]}`)
		case "/v1/categories/1/feeds", "/v1/categories/2/feeds":
			fmt.Fprintln(w, `[{"id": 1, "title": "Test Feed"}]`)
		case "/v1/feeds/1/icon":
			fmt.Fprintln(w, `{"data": "icon-data", "mime_type": "image/png"}`)
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