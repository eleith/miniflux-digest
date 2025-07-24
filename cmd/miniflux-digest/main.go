package main

import (
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"
	miniflux "miniflux.app/v2/client"

	"miniflux-digest/internal/app"
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/processor"
	"miniflux-digest/internal/archive"
	"miniflux-digest/internal/email"
	"miniflux-digest/internal/category"
)

func checkAndSendDigests(application *app.App) {
	for data := range application.CategoryService.StreamData(application.MinifluxClientService) {
		processor.ProcessCategory(application, data, true)
	}
}

func registerDigestsJob(application *app.App, scheduler gocron.Scheduler) {
	_, err := scheduler.NewJob(gocron.CronJob(application.Config.DigestSchedule, true), gocron.NewTask(func() {
		checkAndSendDigests(application)
	}))

	if err != nil {
		log.Fatalf("Error creating job: %v", err)
	}
}

func registerArchiveCleanupJob(application *app.App, scheduler gocron.Scheduler) {
	_, err := scheduler.NewJob(gocron.DurationJob(time.Hour*24), gocron.NewTask(func() {
		application.ArchiveService.CleanArchive(application.Config.ArchivePath, time.Hour*24*21)
	}))

	if err != nil {
		log.Fatalf("Error creating job: %v", err)
	}
}

func main() {
	cfg, err := config.Load("./config.yaml")

	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	clientWrapper := app.NewMinifluxClientWrapper(miniflux.NewClient(cfg.MinifluxHost, cfg.MinifluxApiToken))

	archiveSvc := &archive.ArchiveServiceImpl{}
	emailSvc := &email.EmailServiceImpl{}
	categorySvc := &category.CategoryServiceImpl{}

	application := app.NewApp(cfg, clientWrapper, archiveSvc, emailSvc, categorySvc)

	scheduler, err := gocron.NewScheduler()

	defer func() {
		if err := scheduler.Shutdown(); err != nil {
			log.Printf("Error stopping scheduler: %v", err)
		}
	}()

	if err != nil {
		log.Fatalf("Error creating scheduler: %v", err)
	}

	registerDigestsJob(application, scheduler)
	registerArchiveCleanupJob(application, scheduler)

	scheduler.Start()

	select {}
}
