package main

import (
	"fmt"
	"miniflux-digest/config"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-co-op/gocron/v2"
)

func TestCheckAndSendDigestsWithDryRun(t *testing.T) {
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
			if _, err := fmt.Fprintln(w, `{"data": "icon-data", "mime_type": "image/png"}`); err != nil {
				panic(err)
			}
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		MinifluxHost: server.URL,
		MinifluxApiToken: "test-token",
		DigestDryRun: true,
	}
	checkAndSendDigests(cfg, t.TempDir())
}

func TestJobRegistration(t *testing.T) {
	cfg := &config.Config{DigestSchedule: "@daily"}
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	registerDigestsJob(cfg, scheduler, t.TempDir())
	registerArchiveCleanupJob(scheduler, t.TempDir())

	jobs := scheduler.Jobs()
	if len(jobs) != 2 {
		t.Errorf("Expected 2 jobs to be registered, got %d", len(jobs))
	}

	if err := scheduler.Shutdown(); err != nil {
		t.Fatalf("Failed to shutdown scheduler: %v", err)
	}
}
