package main

import (
	"log"

	"miniflux-digest/config"
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

	client := miniflux.NewClient(cfg.MinifluxHost, cfg.MinifluxApiToken)

	for data := range category.StreamData(client) {
		processor.ProcessCategory(cfg, client, data, &archive.ArchiveServiceImpl{}, &email.EmailServiceImpl{}, "./web/miniflux-archive", false)
		break
	}
}
