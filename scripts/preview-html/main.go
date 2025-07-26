package main

import (
	"bytes"
	"flag"
	"log"
	"os"

	"miniflux-digest/internal/digest"
	"miniflux-digest/internal/templates"
	"miniflux-digest/internal/testutil"
)

func main() {
	minify := flag.Bool("minify", true, "Minify the HTML output")
	flag.Parse()

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

	var buf bytes.Buffer
	if err := templates.ArchiveTemplate.Execute(&buf, data); err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	html, err := digest.MinifyHTML(buf.Bytes(), *minify)
	if err != nil {
		log.Fatalf("Failed to minify HTML: %v", err)
	}

	if _, err := file.Write(html); err != nil {
		log.Fatalf("Failed to write to preview file: %v", err)
	}

	log.Println("Successfully generated web/preview.html")
}