package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/go-co-op/gocron/v2"
	miniflux "miniflux.app/v2/client"

	"miniflux-digest/internal/app"
	"miniflux-digest/internal/archive"
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/digest"
	"miniflux-digest/internal/email"
	
	"miniflux-digest/internal/processor"
)

func registerCategoryDigestJob(application *app.App, scheduler gocron.Scheduler, rawData *app.RawCategoryData) {
	task := func(rawData *app.RawCategoryData) {
		processor.CategoryDigestJob(application, rawData, true)
	}

	jitter := time.Duration(rand.Intn(30)) * time.Second
	startTime := time.Now().Add(1*time.Minute + jitter)

	_, err := scheduler.NewJob(
		gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(startTime)),
		gocron.NewTask(task, rawData),
	)
	if err != nil {
		log.Printf("Error creating one-time job for category %d: %v", rawData.Category.ID, err)
	}
}

func categoriesCheckJob(application *app.App, scheduler gocron.Scheduler) {
	for rawData := range application.MinifluxClientService.StreamAllCategoryData() {
		registerCategoryDigestJob(application, scheduler, rawData)
	}
}

func registerCategoriesCheckJob(application *app.App, scheduler gocron.Scheduler) {
	_, err := scheduler.NewJob(gocron.CronJob(application.Config.DigestSchedule, true), gocron.NewTask(func() {
		categoriesCheckJob(application, scheduler)
	}))

	if err != nil {
		log.Fatalf("Error creating job: %v", err)
	}
}

func registerArchiveCleanupJob(application *app.App, scheduler gocron.Scheduler) {
	_, err := scheduler.NewJob(gocron.DurationJob(time.Hour*24), gocron.NewTask(func() {
		application.ArchiveService.CleanArchive(time.Hour*24*21)
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
	digestSvc := digest.NewDigestService()

	application := app.NewApp(cfg, clientWrapper, archiveSvc, emailSvc, digestSvc)

	scheduler, err := gocron.NewScheduler()

	defer func() {
		if err := scheduler.Shutdown(); err != nil {
			log.Printf("Error stopping scheduler: %v", err)
		}
	}()

	if err != nil {
		log.Fatalf("Error creating scheduler: %v", err)
	}

	registerCategoriesCheckJob(application, scheduler)
	registerArchiveCleanupJob(application, scheduler)

	scheduler.Start()

	select {}
}