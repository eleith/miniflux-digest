package main

import (
	"log"
	"os"
	"strconv"

	"miniflux-digest/internal/app"
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/archive"
	"miniflux-digest/internal/email"
	"miniflux-digest/internal/processor"
	miniflux "miniflux.app/v2/client"
)

func main() {
	cfg, err := config.Load("./config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if len(os.Args) < 2 {
		log.Fatalf("Usage: go run main.go <category_id>")
	}

	categoryID, err := strconv.ParseInt(os.Args[1], 10, 64)
	if err != nil {
		log.Fatalf("Invalid category ID: %v", err)
	}

	minifluxClient := miniflux.NewClient(cfg.MinifluxHost, cfg.MinifluxApiToken)
	clientWrapper := app.NewMinifluxClientWrapper(minifluxClient)

	archiveSvc := &archive.ArchiveServiceImpl{}
	emailSvc := &email.EmailServiceImpl{}

	application := app.NewApp(cfg, clientWrapper, archiveSvc, emailSvc)

	data, err := clientWrapper.FetchCategoryData(categoryID)
	if err != nil {
		log.Fatalf("Failed to fetch category data for preview: %v", err)
	}

	processor.CategoryDigestJob(application, data, false)
}