package app_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"miniflux-digest/internal/app"
	miniflux "miniflux.app/v2/client"
)



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