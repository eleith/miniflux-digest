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
	JitterSeconds         = 30
	ArchiveCleanupDays    = 21
)

func digestJob(application *app.App) {
	overviewFile, groupedEntryFiles, data, err := processor.ProcessAndSendDigest(application)
	if err != nil {
		log.Printf("Error processing digest: %v", err)
		return
	}

	if err := application.EmailService.Send(application.Config, overviewFile, groupedEntryFiles, data); err != nil {
		log.Printf("Error sending digest email: %v", err)
	}
}

func registerDigestJob(application *app.App, scheduler gocron.Scheduler) {
	_, err := scheduler.NewJob(gocron.CronJob(application.Config.Digest.Schedule, true), gocron.NewTask(func() {
		digestJob(application)
	}))

	if err != nil {
		log.Fatalf("Error creating job: %v", err)
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

	defer func() {
		if err := scheduler.Shutdown(); err != nil {
			log.Printf("Error stopping scheduler: %v", err)
		}
	}()

	registerDigestJob(application, scheduler)
	registerArchiveCleanupJob(application, scheduler)

	if application.Config.Digest.RunOnStartup {
		go digestJob(application)
	}

	go func() {
		mux := webserver.SetupServer(webserver.ArchiveBasePath)
		webserver.StartServer(webserver.Port, webserver.RequestSanitizerMiddleware(mux))
	}()

	scheduler.Start()

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
	emailSvc := &email.EmailServiceImpl{}
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
