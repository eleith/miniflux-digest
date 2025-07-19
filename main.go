package main

import (
	"log"

	"github.com/go-co-op/gocron/v2"
	miniflux "miniflux.app/v2/client"

	"miniflux-digest/config"
	"miniflux-digest/internal/email"
	"miniflux-digest/internal/archive"
	"miniflux-digest/internal/category"
)

func categoryEntryCheck() {
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

				err = email.Send(cfg, file, data)

				defer func() {
					if err := file.Close(); err != nil {
						log.Printf("Error closing file for category '%s': %v", data.Category.Title, err)
					}

					category.MarkAsRead(client, data.Category)
				}()

				if err != nil {
					log.Fatalf("Error sending email for category '%s': %v", data.Category.Title, err)
				}
			}
		}()
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

	_, err = scheduler.NewJob(gocron.CronJob(cfg.DigestSchedule, true), gocron.NewTask(func() {
		log.Println("Starting digest job...")
		categoryEntryCheck()
		log.Println("Digest job completed.")
	}))

	if err != nil {
		log.Fatalf("Error creating job: %v", err)
	}

	scheduler.Start()

	select {}
}
