package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"miniflux-digest/internal/app"
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/digest"
	"miniflux-digest/internal/email"
	"miniflux-digest/internal/llm"
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/templates"
	"miniflux-digest/internal/testutil"
	miniflux "miniflux.app/v2/client"
)

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "linux":
		cmd = "xdg-open"
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func generateDigestData(cfg *config.Config, minifluxID int64) *models.HTMLTemplateData {
	log.Println("generateDigestData: Starting...")
	var llmService llm.LLMService

	llmService, err := llm.NewGeminiService(cfg.AI.ApiKey)
	if err != nil {
		log.Fatalf("Failed to create LLM service: %v", err)
	}

	digestSvc := digest.NewDigestService(llmService)
	log.Println("generateDigestData: DigestService initialized.")

	if minifluxID != 0 {
		log.Println("generateDigestData: Fetching real Miniflux data...")
		minifluxClient := miniflux.NewClient(cfg.Miniflux.Host, cfg.Miniflux.ApiToken)
		clientWrapper := app.NewMinifluxClientWrapper(minifluxClient)

		rawData, err := clientWrapper.FetchRawCategoryData(minifluxID)
		if err != nil {
			log.Fatalf("Failed to fetch category data: %v", err)
		}
		log.Println("generateDigestData: Building digest data with real Miniflux data...")
		return digestSvc.BuildDigestData(rawData.Category, rawData.Entries, rawData.Icons, cfg.Digest.GroupBy, cfg.Miniflux.Host)
	} else {
		log.Println("generateDigestData: Building digest data with mock data...")
		return digestSvc.BuildDigestData(
			testutil.NewMockCategory(),
			testutil.NewMockEntries(),
			map[int64]*models.FeedIcon{
				1: testutil.NewMockFeedIconRed(),
				2: testutil.NewMockFeedIconYellow(),
				3: testutil.NewMockFeedIconGreen(),
			},
			cfg.Digest.GroupBy,
			cfg.Miniflux.Host,
		)
	}
}

func generateHTML(data *models.HTMLTemplateData, compress bool) ([]byte, error) {
	log.Println("generateHTML: Starting...")
	var buf bytes.Buffer
	if err := templates.ArchiveTemplate.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	html, err := digest.MinifyHTML(buf.Bytes(), compress)
	if err != nil {
		return nil, fmt.Errorf("failed to minify HTML: %w", err)
	}
	log.Println("generateHTML: Finished.")
	return html, nil
}

func writeHTMLToFile(html []byte) (string, error) {
	log.Println("writeHTMLToFile: Starting...")
	tmpDir, err := os.MkdirTemp("", "miniflux-digest-preview")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}
	filePath := filepath.Join(tmpDir, "preview.html")

	if err := os.WriteFile(filePath, html, 0644); err != nil {
		return "", fmt.Errorf("failed to write HTML to file: %w", err)
	}
	log.Println("writeHTMLToFile: Finished.")
	return filePath, nil
}

func main() {
	log.Println("main: Starting preview script...")
	emailFlag := flag.Bool("email", false, "Send the generated HTML as an email")
	minifluxID := flag.Int64("miniflux", 0, "Miniflux category ID to fetch entries")
	flag.Parse()

	cfg, err := config.Load("./config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Println("main: Config loaded.")

	data := generateDigestData(cfg, *minifluxID)
	log.Println("main: Digest data generated.")

	html, err := generateHTML(data, cfg.Digest.Compress)
	if err != nil {
		log.Fatalf("Failed to generate HTML: %v", err)
	}
	log.Println("main: HTML generated.")

	filePath, err := writeHTMLToFile(html)
	if err != nil {
		log.Fatalf("Failed to write HTML to file: %v", err)
	}
	log.Println("main: HTML written to file.")

	if *emailFlag {
		log.Println("main: Email flag is true, sending email...")
		emailSvc := &email.EmailServiceImpl{}
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatalf("Failed to open HTML file for email: %v", err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				log.Printf("Error closing file: %v", err)
			}
		}()
		if err := emailSvc.Send(cfg, file, data); err != nil {
			log.Fatalf("Failed to send email: %v", err)
		}
		log.Printf("Successfully generated %s and sent email.", filePath)
	} else {
		log.Printf("Successfully generated %s.", filePath)
	}

	log.Printf("Preview available at: file://%s", filePath)

	log.Println("main: Attempting to open browser...")
	if err := openBrowser(fmt.Sprintf("file://%s", filePath)); err != nil {
		log.Printf("Failed to open browser: %v", err)
	}
	log.Println("main: Browser open attempt finished.")
}
