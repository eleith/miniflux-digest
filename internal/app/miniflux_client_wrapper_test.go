package app

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	miniflux "miniflux.app/v2/client"
)

func TestMinifluxClientWrapper_MarkCategoryAsRead(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request path is correct
		if r.URL.Path != "/v1/categories/1/mark-all-as-read" {
			t.Errorf("Expected to request '/v1/categories/1/mark-all-as-read', got %s", r.URL.Path)
		}
		// Check if the request method is correct
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Create a new miniflux client that points to the test server
	client := miniflux.NewClient(server.URL, "testUser", "testPassword")
	// Create our wrapper
	wrapper := NewMinifluxClientWrapper(client)

	// Call the function we want to test
	err := wrapper.MarkCategoryAsRead(1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestMinifluxClientWrapper_CategoryEntries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/categories/1/entries" {
			if _, err := fmt.Fprintln(w, `{"total": 1, "entries": [{"id": 123, "title": "Test Entry"}]}`); err != nil {
				panic(err)
			}
		}
	}))
	defer server.Close()

	client := miniflux.NewClient(server.URL, "testUser", "testPassword")
	wrapper := NewMinifluxClientWrapper(client)

	entries, err := wrapper.CategoryEntries(1, &miniflux.Filter{Status: miniflux.EntryStatusUnread})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if entries == nil {
		t.Fatal("Expected entries to be non-nil")
	}

	if len(*entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(*entries))
	}

	if (*entries)[0].ID != 123 {
		t.Errorf("Expected entry ID 123, got %d", (*entries)[0].ID)
	}
}
