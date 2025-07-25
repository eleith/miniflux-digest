package main

import (
	"miniflux-digest/internal/app"
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/archive"
	"miniflux-digest/internal/email"
	"miniflux-digest/internal/testutil"
	"miniflux-digest/internal/models"
	"testing"

	"github.com/go-co-op/gocron/v2"
	miniflux "miniflux.app/v2/client"
)

func TestCategoriesCheckJob(t *testing.T) {
	mockMinifluxClient := &testutil.MockMinifluxClient{
		StreamAllCategoryDataFunc: func() <-chan *models.CategoryData {
			out := make(chan *models.CategoryData)
			go func() {
				defer close(out)
				out <- &models.CategoryData{Category: &miniflux.Category{ID: 1, Title: "Test 1"}, Entries: &miniflux.Entries{}}
				out <- &models.CategoryData{Category: &miniflux.Category{ID: 2, Title: "Test 2"}, Entries: &miniflux.Entries{}}
				out <- &models.CategoryData{Category: &miniflux.Category{ID: 3, Title: "Test 3"}, Entries: &miniflux.Entries{}}
			}()
			return out
		},
	}

	archiveSvc := &archive.ArchiveServiceImpl{}
	emailSvc := &email.EmailServiceImpl{}
	application := app.NewApp(&config.Config{}, mockMinifluxClient, archiveSvc, emailSvc)

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
	cfg := &config.Config{DigestSchedule: "@daily"}
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	clientWrapper := app.NewMinifluxClientWrapper(miniflux.NewClient("http://localhost", "test-token"))
	archiveSvc := &archive.ArchiveServiceImpl{}
	emailSvc := &email.EmailServiceImpl{}
	application := app.NewApp(cfg, clientWrapper, archiveSvc, emailSvc)

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
