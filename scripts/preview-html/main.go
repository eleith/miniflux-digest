package main

import (
	"bytes"
	"log"
	"os"

	"miniflux-digest/internal/config"
	"miniflux-digest/internal/digest"
	"miniflux-digest/internal/templates"
	"miniflux-digest/internal/testutil"
)

func main() {
	cfg, err := config.Load("./config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	data := testutil.NewMockHTMLTemplateData(cfg.DigestGroupBy)

	file, err := os.Create("web/preview.html")
	if err != nil {
		log.Fatalf("Failed to create preview file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Failed to close preview file: %v", err)
		}
	}()

	var buf bytes.Buffer
	if err := templates.ArchiveTemplate.Execute(&buf, data); err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	html, err := digest.MinifyHTML(buf.Bytes(), cfg.DigestCompress)
	if err != nil {
		log.Fatalf("Failed to minify HTML: %v", err)
	}

	if _, err := file.Write(html); err != nil {
		log.Fatalf("Failed to write to preview file: %v", err)
	}

	log.Println("Successfully generated web/preview.html")
}
