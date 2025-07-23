package main

import (
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"
	miniflux "miniflux.app/v2/client"

	"miniflux-digest/config"
	"miniflux-digest/internal/archive"
	"miniflux-digest/internal/category"
	"miniflux-digest/internal/processor"
)



func checkAndSendDigests(cfg *config.Config, archivePath string) {

	client := miniflux.NewClient(cfg.MinifluxHost, cfg.MinifluxApiToken)

	for data := range category.StreamData(client) {
		processor.ProcessCategory(cfg, client, data, archivePath, true)
	}
}

func registerDigestsJob(cfg *config.Config, scheduler gocron.Scheduler, archivePath string) {
	_, err := scheduler.NewJob(gocron.CronJob(cfg.DigestSchedule, true), gocron.NewTask(func() {
		checkAndSendDigests(cfg, archivePath)
	}))

	if err != nil {
		log.Fatalf("Error creating job: %v", err)
	}
}

func registerArchiveCleanupJob(scheduler gocron.Scheduler, archivePath string) {
	_, err := scheduler.NewJob(gocron.DurationJob(time.Hour*24), gocron.NewTask(func() {
		archive.CleanArchive(archivePath, time.Hour*24*21)
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

	archivePath := "./web/miniflux-archive"

	scheduler, err := gocron.NewScheduler()

	defer func() {
		if err := scheduler.Shutdown(); err != nil {
			log.Printf("Error stopping scheduler: %v", err)
		}
	}()

	if err != nil {
		log.Fatalf("Error creating scheduler: %v", err)
	}

	registerDigestsJob(cfg, scheduler, archivePath)
	registerArchiveCleanupJob(scheduler, archivePath)

	scheduler.Start()

	select {}
}
