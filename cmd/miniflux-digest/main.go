package main

import (
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"
	miniflux "miniflux.app/v2/client"

	"miniflux-digest/internal/app"
	"miniflux-digest/internal/archive"
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/digest"
	"miniflux-digest/internal/email"
	"miniflux-digest/internal/llm"
	"miniflux-digest/internal/processor"
	"miniflux-digest/internal/templates"
	"miniflux-digest/internal/webserver"
)

const (
	JitterSeconds      = 30
	ArchiveCleanupDays = 21
)

func digestJob(application *app.App, digestIndex int, source string) {
	digestConfig := application.Config.Digests[digestIndex]
	log.Printf("Starting digest job for '%s' from source: %s", digestConfig.Title, source)
	overviewFile, groupedEntryFiles, data, err := processor.ProcessDigest(application, digestIndex)
	if err != nil {
		log.Printf("Error processing digest '%s': %v", digestConfig.Title, err)
		return
	}

	if len(data.Entries) == 0 && !*digestConfig.SendIfEmpty {
		log.Printf("Skipping empty digest '%s' (send_if_empty=false)", digestConfig.Title)
		return
	}

	emailSent := application.Config.Smtp.Host != ""
	if emailSent {
		if err := application.EmailService.Send(application.Config.Smtp, digestConfig, overviewFile, groupedEntryFiles, data); err != nil {
			log.Printf("Error sending digest email for '%s': %v", digestConfig.Title, err)
		}
	}

	folder := "(none)"
	if overviewFile != nil {
		folder = overviewFile.Name()
	}

	log.Printf("Digest '%s' produced at %s: entries=%d, folder=%s, email_sent=%t, source=%s",
		digestConfig.Title,
		data.GeneratedDate.Format(time.RFC3339),
		len(data.Entries),
		folder,
		emailSent,
		source,
	)
}

func registerDigestJobs(application *app.App, scheduler gocron.Scheduler) {
	for i, d := range application.Config.Digests {
		digestConfig := d // Capture loop variable
		digestIndex := i
		_, err := scheduler.NewJob(
			gocron.CronJob(digestConfig.Schedule, true),
			gocron.NewTask(func() {
				digestJob(application, digestIndex, "scheduler")
			}),
		)

		if err != nil {
			log.Fatalf("Error creating job for digest '%s': %v", digestConfig.Title, err)
		}

		if digestConfig.RunOnStartup {
			go digestJob(application, digestIndex, "startup")
		}
	}
}

func registerArchiveCleanupJob(application *app.App, scheduler gocron.Scheduler) {
	_, err := scheduler.NewJob(gocron.DurationJob(time.Hour*24), gocron.NewTask(func() {
		application.ArchiveService.CleanArchive(time.Hour * 24 * ArchiveCleanupDays)
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

	application, err := initServices(cfg)
	if err != nil {
		log.Fatalf("Error initializing services: %v", err)
	}

	scheduler, err := initScheduler()
	if err != nil {
		log.Fatalf("Error creating scheduler: %v", err)
	}

	scheduler.Start()

	registerDigestJobs(application, scheduler)
	registerArchiveCleanupJob(application, scheduler)

	go webserver.ListenAndServe(webserver.ArchiveBasePath, webserver.Port)

	select {}
}

func initServices(cfg *config.Config) (*app.App, error) {
	minifluxClient := miniflux.NewClient(cfg.Miniflux.Host, cfg.Miniflux.ApiToken)
	clientWrapper := app.NewMinifluxClientWrapper(minifluxClient)

	llmService, err := llm.NewGeminiService(cfg.AI.ApiKey)

	if err != nil {
		return nil, err
	}

	archiveSvc := archive.NewArchiveService(webserver.ArchiveBasePath, templates.ArchiveTemplate, templates.OverviewTemplate)
	emailSvc := &email.EmailServiceImpl{EmailTemplate: templates.EmailTemplate}
	digestService := digest.NewDigestService(llmService)

	application := app.NewApp(
		app.WithConfig(cfg),
		app.WithArchiveService(archiveSvc),
		app.WithEmailService(emailSvc),
		app.WithMinifluxClientService(clientWrapper),
		app.WithDigestService(digestService),
		app.WithLLMService(llmService),
	)

	return application, nil
}

func initScheduler() (gocron.Scheduler, error) {
	return gocron.NewScheduler()
}