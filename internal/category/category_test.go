package category

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"miniflux-digest/internal/models"
	miniflux "miniflux.app/v2/client"
)

func mockHandler(response string, statusCode int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		if statusCode != http.StatusNoContent {
			if _, err := fmt.Fprintln(w, response); err != nil {
				panic(err)
			}
		}
	}
}

func TestMarkAsRead(t *testing.T) {
	server := httptest.NewServer(mockHandler(`{}`, http.StatusNoContent))
	defer server.Close()

	client := miniflux.NewClient(server.URL, "testUser", "testPassword")
	category := &miniflux.Category{ID: 1, Title: "Test"}

	MarkAsRead(client, category)
}

func TestFetchData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/categories/1/entries":
			if _, err := fmt.Fprintln(w, `{"entries": [{"id": 1, "title": "Test Entry"}], "total": 1}`); err != nil {
				panic(err)
			}
		case "/v1/categories/1/feeds":
			if _, err := fmt.Fprintln(w, `[{"id": 1, "title": "Test Feed"}]`); err != nil {
				panic(err)
			}
		case "/v1/feeds/1/icon":
			if _, err := fmt.Fprintln(w, `{"data": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAAHElEQVQ4T2Nksr/85wADGGYw4oBMAHkAAAD//wMA/wEAP2D3e/gAAAAASUVORK5CYII=", "mime_type": "image/png"}`); err != nil {
				panic(err)
			}
		}
	}))
	defer server.Close()

	client := miniflux.NewClient(server.URL, "testUser", "testPassword")
	category := &miniflux.Category{ID: 1, Title: "Test"}

	data, err := fetchData(client, category)
	if err != nil {
		t.Fatalf("fetchData failed: %v", err)
	}

	if data.Category.ID != 1 {
		t.Errorf("Expected category ID 1, got %d", data.Category.ID)
	}

	if len(*data.Entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(*data.Entries))
	}

	if (*data.Entries)[0].Title != "Test Entry" {
		t.Errorf("Expected entry title 'Test Entry', got %s", (*data.Entries)[0].Title)
	}

	if len(data.FeedIcons) != 1 {
		t.Errorf("Expected 1 feed icon, got %d", len(data.FeedIcons))
	}

	if data.FeedIcons[0].Data != "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAAHElEQVQ4T2Nksr/85wADGGYw4oBMAHkAAAD//wMA/wEAP2D3e/gAAAAASUVORK5CYII=" {
		t.Errorf("Expected feed icon data 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAAHElEQVQ4T2Nksr/85wADGGYw4oBMAHkAAAD//wMA/wEAP2D3e/gAAAAASUVORK5CYII=', got %s", data.FeedIcons[0].Data)
	}
}

func TestStreamData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/categories":
			if _, err := fmt.Fprintln(w, `[{"id": 1, "title": "Test Category"}]`); err != nil {
				panic(err)
			}
		case "/v1/categories/1/entries":
			if _, err := fmt.Fprintln(w, `{"entries": [{"id": 1, "title": "Test Entry"}], "total": 1}`); err != nil {
				panic(err)
			}
		case "/v1/categories/1/feeds":
			if _, err := fmt.Fprintln(w, `[{"id": 1, "title": "Test Feed"}]`); err != nil {
				panic(err)
			}
		case "/v1/feeds/1/icon":
			if _, err := fmt.Fprintln(w, `{"data": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAAHElEQVQ4T2Nksr/85wADGGYw4oBMAHkAAAD//wMA/wEAP2D3e/gAAAAASUVORK5CYII=", "mime_type": "image/png"}`); err != nil {
				panic(err)
			}
		}
	}))
	defer server.Close()

	client := miniflux.NewClient(server.URL, "testUser", "testPassword")
	ch := StreamData(client)

	var results []*models.CategoryData
	for data := range ch {
		results = append(results, data)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].Category.Title != "Test Category" {
		t.Errorf("Expected category title 'Test Category', got %s", results[0].Category.Title)
	}
}

func TestFeedIcon(t *testing.T) {
	icon := &models.FeedIcon{
		FeedID: 1,
		Data:   "test",
	}

	if icon.FeedID != 1 {
		t.Errorf("Expected FeedID to be 1, got %d", icon.FeedID)
	}

	if icon.Data != "test" {
		t.Errorf("Expected Data to be 'test', got %s", icon.Data)
	}
}

func TestCategoryData(t *testing.T) {
	now := time.Now()
	category := &miniflux.Category{ID: 1, Title: "Test"}
	entries := &miniflux.Entries{}
	icons := []*models.FeedIcon{}

	data := &models.CategoryData{
		Category:      category,
		Entries:       entries,
		GeneratedDate: now,
		FeedIcons:     icons,
	}

	if !reflect.DeepEqual(data.Category, category) {
		t.Errorf("Expected Category to be %v, got %v", category, data.Category)
	}

	if !reflect.DeepEqual(data.Entries, entries) {
		t.Errorf("Expected Entries to be %v, got %v", entries, data.Entries)
	}

	if !reflect.DeepEqual(data.GeneratedDate, now) {
		t.Errorf("Expected GeneratedDate to be %v, got %v", now, data.GeneratedDate)
	}

	if !reflect.DeepEqual(data.FeedIcons, icons) {
		t.Errorf("Expected FeedIcons to be %v, got %v", icons, data.FeedIcons)
	}
}
