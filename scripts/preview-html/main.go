package main

import (
	"log"
	"os"

	"miniflux-digest/internal/templates"
	"miniflux-digest/internal/testutil"
)

func main() {
	data := testutil.NewMockHTMLTemplateData()

	file, err := os.Create("web/preview.html")
	if err != nil {
		log.Fatalf("Failed to create preview file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Failed to close preview file: %v", err)
		}
	}()

	if err := templates.ArchiveTemplate.Execute(file, data); err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	log.Println("Successfully generated web/preview.html")
}