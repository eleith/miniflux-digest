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

func generateMockDigest(cfg *config.Config) *models.OverviewTemplateData {
	log.Println("generateDigestData: Starting...")

	llmService, err := llm.NewGeminiService(cfg.AI.ApiKey)
	if err != nil {
		log.Fatalf("Failed to create LLM service: %v", err)
	}

	digestSvc := digest.NewDigestService(llmService)
	log.Println("generateDigestData: DigestService initialized.")

	var entries []*models.Entry
	log.Println("generateDigestData: Using mock data...")
	entries = testutil.CreateMockEntries(200)

	icons := make(map[int64]*models.FeedIcon)
	for _, entry := range entries {
		if _, ok := icons[entry.FeedID]; !ok {
			switch entry.FeedID {
			case 1:
				icons[entry.FeedID] = testutil.NewMockFeedIconRed()
			case 2:
				icons[entry.FeedID] = testutil.NewMockFeedIconYellow()
			case 3:
				icons[entry.FeedID] = testutil.NewMockFeedIconGreen()
			default:
				icons[entry.FeedID] = testutil.NewMockFeedIconGreen()
			}
		}
	}

	log.Println("generateDigestData: Building digest data...")
	digestConfig := cfg.Digests[0]
	data := digestSvc.BuildDigestData(
		entries,
		icons,
		digestConfig.View,
		cfg.Miniflux.Host,
		digestConfig.Host,
		digestConfig.Categories,
		digestConfig.Title,
	)

	if len(data.PrimaryGroups) > 0 {
		data.PrimaryGroups[0].Summary = "This is a mock summary for the first group to test the layout."
	}
	if len(data.PrimaryGroups) > 2 {
		data.PrimaryGroups[2].Summary = "This is another mock summary for a different group, showing that not all groups have summaries."
	}

	return data
}

func generateMinifluxDigest(cfg *config.Config, minifluxClientService services.MinifluxClientService) *models.OverviewTemplateData {
	log.Println("generateDigestData: Starting...")

	llmService, err := llm.NewGeminiService(cfg.AI.ApiKey)
	if err != nil {
		log.Fatalf("Failed to create LLM service: %v", err)
	}

	digestSvc := digest.NewDigestService(llmService)
	log.Println("generateDigestData: DigestService initialized.")

	var entries []*models.Entry
	log.Println("generateDigestData: Fetching real Miniflux data...")

	entries, err = minifluxClientService.GetAllUnreadEntries()
	if err != nil {
		log.Fatalf("Failed to fetch entries: %v", err)
	}

	icons := make(map[int64]*models.FeedIcon)
	for _, entry := range entries {
		if _, ok := icons[entry.FeedID]; !ok {
			icon, err := minifluxClientService.FeedIcon(entry.FeedID)
			if err != nil {
				log.Printf("Warning: failed to fetch icon for feed %d: %v", entry.FeedID, err)
				continue
			}
			icons[entry.FeedID] = icon
		}
	}

	log.Println("generateDigestData: Building digest data...")
	digestConfig := cfg.Digests[0]
	return digestSvc.BuildDigestData(
		entries,
		icons,
		digestConfig.View,
		cfg.Miniflux.Host,
		digestConfig.Host,
		digestConfig.Categories,
		digestConfig.Title,
	)
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

func generateAndArchiveHTML(cfg *config.Config, minifluxFlag bool) (*os.File, []*os.File, string, *models.OverviewTemplateData) {
	var data *models.OverviewTemplateData

	if minifluxFlag {
		minifluxClient := miniflux.NewClient(cfg.Miniflux.Host, cfg.Miniflux.ApiToken)
		clientWrapper := app.NewMinifluxClientWrapper(minifluxClient)
		data = generateMinifluxDigest(cfg, clientWrapper)
	} else {
		data = generateMockDigest(cfg)
	}

	log.Println("main: Digest data generated.")

	digestConfig := cfg.Digests[0]
	archiveSvc := archive.NewArchiveService(webserver.ArchiveBasePath, templates.ArchiveTemplate, templates.OverviewTemplate)
	overviewFile, groupedEntryFiles, err := archiveSvc.MakeArchiveHTML(data, digestConfig.Compress)
	if err != nil {
		log.Fatalf("Failed to generate HTML: %v", err)
	}
	log.Println("main: HTML generated.")

	log.Printf("overviewFile.Name(): %s", overviewFile.Name())
	log.Printf("webserver.ArchiveBasePath: %s", webserver.ArchiveBasePath)

	relativePath, err := filepath.Rel(webserver.ArchiveBasePath, overviewFile.Name())
	if err != nil {
		log.Fatalf("Failed to get relative path: %v", err)
	}
	log.Printf("relativePath: %s", relativePath)
	overviewURL := fmt.Sprintf("http://localhost%s/archive/%s", webserver.Port, relativePath)
	log.Printf("overviewURL: %s", overviewURL)

	return overviewFile, groupedEntryFiles, overviewURL, data
}

func handleEmail(cfg *config.Config, overviewFile *os.File, groupedEntryFiles []*os.File, data *models.OverviewTemplateData) {
	log.Println("main: Email flag is true, sending email...")
	emailSvc := &email.EmailServiceImpl{
		EmailTemplate: templates.EmailTemplate,
	}

	digestConfig := cfg.Digests[0]
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
	log.Printf("Attempting to open %s with the browser...", url)
	if err := openBrowser(url); err != nil {
		log.Printf("Failed to open browser: %v", err)
	}
}

func main() {
	log.Println("main: Starting preview script...")
	emailFlag, minifluxFlag, htmlFlag := setupAndParseFlags()

	cfg := loadConfig()

	if emailFlag {
		overviewFile, groupedEntryFiles, _, data := generateAndArchiveHTML(cfg, minifluxFlag)
		handleEmail(cfg, overviewFile, groupedEntryFiles, data)
	} else if htmlFlag {
		_, _, overviewURL, _ := generateAndArchiveHTML(cfg, minifluxFlag)
		handleOpenUrl(overviewURL)
	} else if minifluxFlag {
		_, _, overviewURL, _ := generateAndArchiveHTML(cfg, minifluxFlag)
		handleOpenUrl(overviewURL)
	} else {
		log.Println("main: Starting web server")
		handleWebServer()
	}
}
