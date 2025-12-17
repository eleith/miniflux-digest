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
	"miniflux-digest/internal/manager"
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

func generateMockDigest(cfg *config.Config, digestIndex int) *models.OverviewTemplateData {
	log.Println("generateDigestData: Starting...")

	llmService, err := llm.NewGeminiService(cfg.AI.ApiKey)
	if err != nil {
		log.Fatalf("Failed to create LLM service: %v", err)
	}

	digestManager, err := manager.NewDigestManager(cfg.Digests)
	if err != nil {
		log.Fatalf("Error creating digest manager: %v", err)
	}

	digestSvc := digest.NewDigestService(llmService)
	log.Println("generateDigestData: DigestService initialized.")

	var entries []*models.Entry
	log.Println("generateDigestData: Using mock data...")
	entries = testutil.CreateMockEntries(200)

	var digestEntries []*models.Entry
	digestConfig := cfg.Digests[digestIndex]

	for _, entry := range entries {
		if digestManager.GetOwningDigest(entry) == digestIndex {
			digestEntries = append(digestEntries, entry)
		}
	}

	icons := make(map[int64]*models.FeedIcon)
	for _, entry := range digestEntries {
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
	data := digestSvc.BuildDigestData(
		digestEntries,
		icons,
		digestConfig.View,
		cfg.Miniflux.Host,
		digestConfig.Host,
		digestConfig.Categories,
		digestConfig.Title,
	)

	for i := range data.PrimaryGroups {
		data.PrimaryGroups[i].Summary = fmt.Sprintf("This is a mock summary for group %d to test the layout.", i+1)
	}

	return data
}

func generateMinifluxDigest(cfg *config.Config, digestIndex int, minifluxClientService services.MinifluxClientService) *models.OverviewTemplateData {
	log.Println("generateDigestData: Starting...")

	llmService, err := llm.NewGeminiService(cfg.AI.ApiKey)
	if err != nil {
		log.Fatalf("Failed to create LLM service: %v", err)
	}

	digestManager, err := manager.NewDigestManager(cfg.Digests)
	if err != nil {
		log.Fatalf("Error creating digest manager: %v", err)
	}

	digestSvc := digest.NewDigestService(llmService)
	log.Println("generateDigestData: DigestService initialized.")

	var entries []*models.Entry
	log.Println("generateDigestData: Fetching real Miniflux data...")

	entries, err = minifluxClientService.GetAllUnreadEntries()
	if err != nil {
		log.Fatalf("Failed to fetch entries: %v", err)
	}

	var digestEntries []*models.Entry
	digestConfig := cfg.Digests[digestIndex]

	for _, entry := range entries {
		if digestManager.GetOwningDigest(entry) == digestIndex {
			digestEntries = append(digestEntries, entry)
		}
	}

	icons := make(map[int64]*models.FeedIcon)
	for _, entry := range digestEntries {
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
	return digestSvc.BuildDigestData(
		digestEntries,
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

func generateAndArchiveHTML(cfg *config.Config, digestIndex int, minifluxFlag bool) (*os.File, []*os.File, string, *models.OverviewTemplateData) {
	var data *models.OverviewTemplateData
	digestConfig := cfg.Digests[digestIndex]

	if minifluxFlag {
		minifluxClient := miniflux.NewClient(cfg.Miniflux.Host, cfg.Miniflux.ApiToken)
		clientWrapper := app.NewMinifluxClientWrapper(minifluxClient)
		data = generateMinifluxDigest(cfg, digestIndex, clientWrapper)
	} else {
		data = generateMockDigest(cfg, digestIndex)
	}

	log.Println("main: Digest data generated.")

	archiveSvc := archive.NewArchiveService(webserver.ArchiveBasePath, templates.ArchiveTemplate, templates.OverviewTemplate)
	overviewFile, groupedEntryFiles, err := archiveSvc.MakeArchiveHTML(data, *digestConfig.Compress)
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

	overviewURL := fmt.Sprintf("http://localhost%s/%s", webserver.Port, relativePath)
	if digestConfig.Host != "" {
		overviewURL = fmt.Sprintf("%s/archive/%s", digestConfig.Host, relativePath)
	}
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
	numDigest := len(cfg.Digests)

	if emailFlag {
		for digestIndex := range numDigest {
			overviewFile, groupedEntryFiles, _, data := generateAndArchiveHTML(cfg, digestIndex, minifluxFlag)
			handleEmail(cfg, overviewFile, groupedEntryFiles, data)
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
