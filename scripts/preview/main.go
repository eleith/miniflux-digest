package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"miniflux-digest/internal/app"
	"miniflux-digest/internal/app/services"
	"miniflux-digest/internal/archive"
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/digest"
	"miniflux-digest/internal/email"
	"miniflux-digest/internal/llm"
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/processor"
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

// PreviewMinifluxClient handles data fetching for the preview.
// It serves either mock data or wraps the real client to ensure safety (preventing writes).
type PreviewMinifluxClient struct {
	RealClient services.MinifluxClientService // If nil, use mock data
	MockData   []*models.Entry
}

func (c *PreviewMinifluxClient) GetAllUnreadEntries() ([]*models.Entry, error) {
	if c.RealClient != nil {
		return c.RealClient.GetAllUnreadEntries()
	}
	return c.MockData, nil
}

func (c *PreviewMinifluxClient) FeedIcon(feedID int64) (*models.FeedIcon, error) {
	if c.RealClient != nil {
		return c.RealClient.FeedIcon(feedID)
	}
	// Mock icons based on feed ID logic from original script
	switch feedID {
	case 1:
		return testutil.NewMockFeedIconRed(), nil
	case 2:
		return testutil.NewMockFeedIconYellow(), nil
	case 3:
		return testutil.NewMockFeedIconGreen(), nil
	default:
		return testutil.NewMockFeedIconGreen(), nil
	}
}

func (c *PreviewMinifluxClient) UpdateEntries(entryIDs []int64, status string) error {
	log.Println("Preview: Skipping UpdateEntries (safety mechanism active)")
	return nil
}

// PreviewDigestService wraps the real service to inject mock summaries when needed.
type PreviewDigestService struct {
	RealService services.DigestService
	InjectMocks bool
}

func (s *PreviewDigestService) BuildDigestData(entries []*models.Entry, icons map[int64]*models.FeedIcon, view string, minifluxHost string, digestHost string, categories []config.ConfigCategory, digestTitle string) *models.OverviewTemplateData {
	data := s.RealService.BuildDigestData(entries, icons, view, minifluxHost, digestHost, categories, digestTitle)
	if s.InjectMocks && data != nil {
		for i := range data.PrimaryGroups {
			data.PrimaryGroups[i].Summary = fmt.Sprintf("This is a mock summary for group %d to test the layout.", i+1)
		}
	}
	return data
}

func setupAndParseFlags() (emailFlag, minifluxFlag, serveOnlyFlag bool) {
	email := flag.Bool("email", false, "send mock data derived digest as an email")
	miniflux := flag.Bool("miniflux", false, "use live data from miniflux to generate digest")
	html := flag.Bool("html", false, "use mock data to generate digest")
	flag.Parse()
	return *email, *miniflux, *html
}

func loadConfig() *config.Config {
	cfg, err := config.Load("./config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Println("main: Config loaded.")
	return cfg
}

func generateAndArchiveHTML(cfg *config.Config, digestIndex int, minifluxFlag bool) (*os.File, []*os.File, string, *models.OverviewTemplateData) {
	log.Println("main: Preparing application state...")

	llmService, err := llm.NewGeminiService(cfg.AI.ApiKey)
	if err != nil {
		log.Fatalf("Failed to create LLM service: %v", err)
	}

	// Setup Miniflux Client (Mock or Real)
	var clientService services.MinifluxClientService
	var mockData []*models.Entry
	if minifluxFlag {
		realClient := miniflux.NewClient(cfg.Miniflux.Host, cfg.Miniflux.ApiToken)
		clientService = &PreviewMinifluxClient{
			RealClient: app.NewMinifluxClientWrapper(realClient),
		}
		log.Println("main: Using Real Miniflux Client (Safe Mode)")
	} else {
		log.Println("main: Using Mock Data")
		mockData = testutil.CreateMockEntries(200)
		clientService = &PreviewMinifluxClient{
			MockData: mockData,
		}
	}

	// Setup Digest Service (Wrapped for Mock Summaries if not using real data)
	realDigestService := digest.NewDigestService(llmService)
	digestService := &PreviewDigestService{
		RealService: realDigestService,
		InjectMocks: !minifluxFlag, // Inject mocks only if NOT using real Miniflux data
	}

	archiveSvc := archive.NewArchiveService(webserver.ArchiveBasePath, templates.ArchiveTemplate, templates.OverviewTemplate)
	emailSvc := &email.EmailServiceImpl{EmailTemplate: templates.EmailTemplate}

	// Create App instance
	application := app.NewApp(
		app.WithConfig(cfg),
		app.WithArchiveService(archiveSvc),
		app.WithEmailService(emailSvc),
		app.WithMinifluxClientService(clientService),
		app.WithDigestService(digestService),
		app.WithLLMService(llmService),
	)

	// Ensure safety for preview: Force MarkAsRead to false in config copy being used
	cfg.Digests[digestIndex].MarkAsRead = false
	// Ensure we process it even if empty to show something?
	// processor.ProcessDigest returns nil if empty.
	// We might want to see "No entries" page? But ProcessDigest doesn't generate one.
	// For preview, let's stick to ProcessDigest behavior.

	log.Println("main: Running ProcessDigest...")
	overviewFile, groupedEntryFiles, data, err := processor.ProcessDigest(application, digestIndex)
	if err != nil {
		log.Fatalf("Failed to process digest: %v", err)
	}

	if overviewFile == nil {
		log.Printf("Digest %d is empty. No HTML generated.", digestIndex)
		return nil, nil, "", data
	}

	log.Println("main: HTML generated.")
	log.Printf("overviewFile.Name(): %s", overviewFile.Name())
	log.Printf("webserver.ArchiveBasePath: %s", webserver.ArchiveBasePath)

	relativePath, err := filepath.Rel(webserver.ArchiveBasePath, overviewFile.Name())
	if err != nil {
		log.Fatalf("Failed to get relative path: %v", err)
	}
	log.Printf("relativePath: %s", relativePath)

	digestConfig := cfg.Digests[digestIndex]
	overviewURL := fmt.Sprintf("http://localhost%s/%s", webserver.Port, relativePath)
	if digestConfig.Host != "" {
		overviewURL = fmt.Sprintf("%s/archive/%s", digestConfig.Host, relativePath)
	}
	log.Printf("overviewURL: %s", overviewURL)

	return overviewFile, groupedEntryFiles, overviewURL, data
}

func handleEmail(cfg *config.Config, digestConfig config.ConfigDigest, overviewFile *os.File, groupedEntryFiles []*os.File, data *models.OverviewTemplateData) {
	if overviewFile == nil {
		log.Println("main: Skipping email (no digest generated).")
		return
	}
	log.Println("main: Email flag is true, sending email...")
	emailSvc := &email.EmailServiceImpl{
		EmailTemplate: templates.EmailTemplate,
	}

	if err := emailSvc.Send(cfg.Smtp, digestConfig, overviewFile, groupedEntryFiles, data); err != nil {
		log.Fatalf("Failed to send email: %v", err)
	}
	log.Printf("Successfully generated and sent email.")
}

func handleWebServer() {
	log.Printf("Successfully generated.")

	go func() {
		webserver.ListenAndServe(webserver.ArchiveBasePath, webserver.Port)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down web server...")
}

func handleOpenUrl(url string) {
	if url == "" {
		return
	}
	log.Printf("Attempting to open %s with the browser...", url)
	if err := openBrowser(url); err != nil {
		log.Printf("Failed to open browser: %v", err)
	}
}

func main() {
	log.Println("main: Starting preview script...")
	emailFlag, minifluxFlag, htmlFlag := setupAndParseFlags()

	cfg := loadConfig()
	numDigest := len(cfg.Digests)

	if emailFlag {
		for digestIndex := range numDigest {
			overviewFile, groupedEntryFiles, _, data := generateAndArchiveHTML(cfg, digestIndex, minifluxFlag)
			handleEmail(cfg, cfg.Digests[digestIndex], overviewFile, groupedEntryFiles, data)
		}
	} else if htmlFlag {
		for digestIndex := range numDigest {
			_, _, overviewURL, _ := generateAndArchiveHTML(cfg, digestIndex, minifluxFlag)
			handleOpenUrl(overviewURL)
		}
	} else if minifluxFlag {
		for digestIndex := range numDigest {
			_, _, overviewURL, _ := generateAndArchiveHTML(cfg, digestIndex, minifluxFlag)
			handleOpenUrl(overviewURL)
		}
	} else {
		log.Println("main: Starting web server")
		handleWebServer()
	}
}
