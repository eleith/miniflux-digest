package main

import (
	"miniflux-digest/internal/app"
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/archive"
	"miniflux-digest/internal/email"
	"miniflux-digest/internal/digest"
	"miniflux-digest/internal/testutil"
	
	"testing"

	"github.com/go-co-op/gocron/v2"
	miniflux "miniflux.app/v2/client"
)

func TestCategoriesCheckJob(t *testing.T) {
	mockMinifluxClient := &testutil.MockMinifluxClient{
		StreamAllCategoryDataFunc: func() <-chan *app.RawCategoryData {
			out := make(chan *app.RawCategoryData)
			go func() {
				defer close(out)
				out <- &app.RawCategoryData{Category: &miniflux.Category{ID: 1, Title: "Test 1"}, Entries: &miniflux.Entries{}}
				out <- &app.RawCategoryData{Category: &miniflux.Category{ID: 2, Title: "Test 2"}, Entries: &miniflux.Entries{}}
				out <- &app.RawCategoryData{Category: &miniflux.Category{ID: 3, Title: "Test 3"}, Entries: &miniflux.Entries{}}
			}()
			return out
		},
	}

	archiveSvc := &archive.ArchiveServiceImpl{}
	emailSvc := &email.EmailServiceImpl{}
	digestSvc := digest.NewDigestService()
	application := app.NewApp(
		app.WithConfig(&config.Config{}),
		app.WithMinifluxClientService(mockMinifluxClient),
		app.WithArchiveService(archiveSvc),
		app.WithEmailService(emailSvc),
		app.WithDigestService(digestSvc),
	)

	scheduler, err := gocron.NewScheduler()
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	categoriesCheckJob(application, scheduler)

	jobs := scheduler.Jobs()
	if len(jobs) != 3 {
		t.Errorf("Expected 3 jobs to be scheduled, got %d", len(jobs))
	}
}

func TestJobRegistration(t *testing.T) {
	cfg := &config.Config{
		Digest: config.ConfigDigest{
			Schedule: "@daily",
		},
	}
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	clientWrapper := app.NewMinifluxClientWrapper(miniflux.NewClient("http://localhost", "test-token"))
	archiveSvc := &archive.ArchiveServiceImpl{}
	emailSvc := &email.EmailServiceImpl{}
	digestSvc := digest.NewDigestService()
	application := app.NewApp(
		app.WithConfig(cfg),
		app.WithMinifluxClientService(clientWrapper),
		app.WithArchiveService(archiveSvc),
		app.WithEmailService(emailSvc),
		app.WithDigestService(digestSvc),
	)

	registerCategoriesCheckJob(application, scheduler)
	registerArchiveCleanupJob(application, scheduler)

	jobs := scheduler.Jobs()
	if len(jobs) != 2 {
		t.Errorf("Expected 2 jobs to be registered, got %d", len(jobs))
	}

	if err := scheduler.Shutdown(); err != nil {
		t.Fatalf("Failed to shutdown scheduler: %v", err)
	}
}
