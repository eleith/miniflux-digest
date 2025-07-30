package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
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
)

const (
	JitterSeconds         = 30
	ArchiveCleanupDays    = 21
	ArchiveBasePath       = "web/miniflux-archive"
	HealthCheckPort       = ":8080"
)

func registerCategoryDigestJob(application *app.App, scheduler gocron.Scheduler, rawData *app.RawCategoryData) {
	task := func(rawData *app.RawCategoryData) {
		processor.CategoryDigestJob(application, rawData, application.Config.Digest.MarkAsRead)
	}

	jitter := time.Duration(rand.Intn(JitterSeconds)) * time.Second
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
	_, err := scheduler.NewJob(gocron.CronJob(application.Config.Digest.Schedule, true), gocron.NewTask(func() {
		categoriesCheckJob(application, scheduler)
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

func SetupServer(archiveBasePath string) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprintf(w, "OK"); err != nil {
			log.Printf("Error writing healthcheck response: %v", err)
		}
	})

	fs := http.FileServer(http.Dir(archiveBasePath))
	mux.Handle("/archive/", http.StripPrefix("/archive/", fs))

	return mux
}

func requestSanitizerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "..") {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/") && len(r.URL.Path) > 1 {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
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

	registerCategoriesCheckJob(application, scheduler)
	registerArchiveCleanupJob(application, scheduler)

	if application.Config.Digest.RunOnStartup {
		go categoriesCheckJob(application, scheduler)
	}

	go func() {
		mux := SetupServer(ArchiveBasePath)
		log.Printf("Internal web server starting on port %s", HealthCheckPort)

		if err := http.ListenAndServe(HealthCheckPort, requestSanitizerMiddleware(mux)); err != nil {
			log.Fatalf("Internal web server failed to start: %v", err)
		}
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

	archiveSvc := archive.NewArchiveService(ArchiveBasePath)
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