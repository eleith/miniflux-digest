package main

import (
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"
	miniflux "miniflux.app/v2/client"

	"miniflux-digest/config"
	"miniflux-digest/internal/archive"
	"miniflux-digest/internal/category"
	"miniflux-digest/internal/email"
)

func checkAndSendDigests() {
	cfg, err := config.Load()

	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	client := miniflux.NewClient(cfg.MinifluxHost, cfg.MinifluxApiToken)

	for data := range category.StreamData(client) {
		func() {
			if len(*data.Entries) > 0 {
				file, err := archive.MakeArchiveHTML(data)

				if err != nil {
					log.Printf("Error generating File for category %s: %v", data.Category.Title, err)
					return
				}

				defer func() {
					if err := file.Close(); err != nil {
						log.Printf("Error closing file for category '%s': %v", data.Category.Title, err)
					}
				}()

				err = email.Send(cfg, file, data)

				if err != nil {
					log.Printf("Error sending email for category '%s': %v", data.Category.Title, err)
				}

				category.MarkAsRead(client, data.Category)
			}
		}()
	}
}

func registerDigestsJob(cfg *config.Config, scheduler gocron.Scheduler) {
	_, err := scheduler.NewJob(gocron.CronJob(cfg.DigestSchedule, true), gocron.NewTask(func() {
		checkAndSendDigests()
	}))

	if err != nil {
		log.Fatalf("Error creating job: %v", err)
	}
}

func registerArchiveCleanupJob(scheduler gocron.Scheduler) {
	_, err := scheduler.NewJob(gocron.DurationJob(time.Hour*24), gocron.NewTask(func() {
		archive.CleanArchive(time.Hour * 24 * 21)
	}))

	if err != nil {
		log.Fatalf("Error creating job: %v", err)
	}
}

func main() {
	cfg, err := config.Load()

	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	scheduler, err := gocron.NewScheduler()

	defer func() {
		if err := scheduler.Shutdown(); err != nil {
			log.Printf("Error stopping scheduler: %v", err)
		}
	}()

	if err != nil {
		log.Fatalf("Error creating scheduler: %v", err)
	}

	registerDigestsJob(cfg, scheduler)
	registerArchiveCleanupJob(scheduler)

	scheduler.Start()

	select {}
}
