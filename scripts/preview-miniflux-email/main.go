package main

import (
	"log"

	"miniflux-digest/internal/app"
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/archive"
	"miniflux-digest/internal/category"
	"miniflux-digest/internal/email"
	"miniflux-digest/internal/processor"
	miniflux "miniflux.app/v2/client"
)

func main() {
	cfg, err := config.Load("./config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	minifluxClient := miniflux.NewClient(cfg.MinifluxHost, cfg.MinifluxApiToken)
	clientWrapper := app.NewMinifluxClientWrapper(minifluxClient)

	archiveSvc := &archive.ArchiveServiceImpl{}
	emailSvc := &email.EmailServiceImpl{}
	categorySvc := &category.CategoryServiceImpl{}

	application := app.NewApp(cfg, clientWrapper, archiveSvc, emailSvc, categorySvc)

	for data := range application.CategoryService.StreamData(application.MinifluxClientService) {
		processor.ProcessCategory(application, data, false)
		break
	}
}
