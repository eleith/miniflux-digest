package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"miniflux-digest/internal/app"
	"miniflux-digest/internal/archive"
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/digest"
	"miniflux-digest/internal/email"
	"miniflux-digest/internal/llm"
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/templates"
	"miniflux-digest/internal/testutil"
	"miniflux-digest/internal/webserver"

	miniflux "miniflux.app/v2/client"
)

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
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

func generateDigestData(cfg *config.Config, useMiniflux bool) *models.OverviewTemplateData {
	log.Println("generateDigestData: Starting...")

	llmService, err := llm.NewGeminiService(cfg.AI.ApiKey)
	if err != nil {
		log.Fatalf("Failed to create LLM service: %v", err)
	}

	digestSvc := digest.NewDigestService(llmService)
	log.Println("generateDigestData: DigestService initialized.")

	var entries *miniflux.Entries
	if useMiniflux {
		log.Println("generateDigestData: Fetching real Miniflux data...")
		minifluxClient := miniflux.NewClient(cfg.Miniflux.Host, cfg.Miniflux.ApiToken)
		clientWrapper := app.NewMinifluxClientWrapper(minifluxClient)

		entries, err = clientWrapper.GetAllUnreadEntries()
		if err != nil {
			log.Fatalf("Failed to fetch entries: %v", err)
		}
	} else {
		log.Println("generateDigestData: Using mock data...")
		entries = testutil.MockNumEntries(200)
	}

	// For the preview, we'll just use a mock category and all entries.
	// The full grouping logic will be in the processor.
	log.Println("generateDigestData: Building digest data...")
	return digestSvc.BuildDigestData(
		testutil.NewMockCategory(),
		entries,
		map[int64]*models.FeedIcon{
			1: testutil.NewMockFeedIconRed(),
			2: testutil.NewMockFeedIconYellow(),
			3: testutil.NewMockFeedIconGreen(),
		},
		cfg.Digest.GroupBy,
		cfg.Digest.SubGroupBy,
		cfg.Digest.SortBy,
		cfg.Miniflux.Host,
	)
}



func main() {
	log.Println("main: Starting preview script...")
	emailFlag := flag.Bool("email", false, "Send the generated HTML as an email")
	minifluxFlag := flag.Bool("miniflux", false, "Use live data from Miniflux")
	flag.Parse()

	cfg, err := config.Load("./config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Println("main: Config loaded.")

	data := generateDigestData(cfg, *minifluxFlag)
	log.Println("main: Digest data generated.")

	archiveSvc := archive.NewArchiveService(webserver.ArchiveBasePath, templates.ArchiveTemplate, templates.OverviewTemplate)
	overviewFile, err := archiveSvc.MakeArchiveHTML(data, cfg.Digest.Compress)
	if err != nil {
		log.Fatalf("Failed to generate HTML: %v", err)
	}
	log.Println("main: HTML generated.")

	if *emailFlag {
		log.Println("main: Email flag is true, sending email...")
		emailSvc := &email.EmailServiceImpl{}
		// For preview, we only send the overview file
		if err := emailSvc.Send(cfg, overviewFile, []*os.File{}, data); err != nil {
			log.Fatalf("Failed to send email: %v", err)
		}
		log.Printf("Successfully generated and sent email.")
	} else {
		log.Printf("Successfully generated.")
	}

	// Start web server
	mux := webserver.SetupServer(webserver.ArchiveBasePath)
	server := &http.Server{Addr: webserver.Port, Handler: webserver.RequestSanitizerMiddleware(mux)}

	go func() {
		webserver.StartServer(webserver.Port, webserver.RequestSanitizerMiddleware(mux))
	}()

	// Open browser
	relativePath, err := filepath.Rel(webserver.ArchiveBasePath, overviewFile.Name())
	if err != nil {
		log.Fatalf("Failed to get relative path: %v", err)
	}
	previewURL := fmt.Sprintf("http://localhost%s/archive/%s", webserver.Port, relativePath)
	log.Printf("Preview available at: %s", previewURL)
	log.Println("main: Attempting to open browser...")
	if err := openBrowser(previewURL); err != nil {
		log.Printf("Failed to open browser: %v", err)
	}
	log.Println("main: Browser open attempt finished.")

	log.Println("Press Ctrl+C to stop the preview server.")
	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down preview server...")

	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Preview server gracefully stopped.")
	os.Exit(0)
}
