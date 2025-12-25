package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"miniflux-digest/internal/app"
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/email"
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/webserver"
)

func setupTestArchive(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "miniflux-digest-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to remove temp dir: %v", err)
		}
	})

	categoryDir := filepath.Join(tmpDir, "test-category")
	if err := os.Mkdir(categoryDir, 0755); err != nil {
		t.Fatalf("Failed to create category dir: %v", err)
	}
	filePath := filepath.Join(categoryDir, "test-file.html")
	fileContent := "<html><body><h1>Test File</h1></body></html>"
	if err := os.WriteFile(filePath, []byte(fileContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	categoryWithIndexDir := filepath.Join(tmpDir, "test-category-with-index")
	if err := os.Mkdir(categoryWithIndexDir, 0755); err != nil {
		t.Fatalf("Failed to create category with index dir: %v", err)
	}
	indexPath := filepath.Join(categoryWithIndexDir, "index.html")
	indexContent := "<html><body><h1>Index File</h1></body></html>"
	if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
		t.Fatalf("Failed to write index file: %v", err)
	}

	categoryNoIndexDir := filepath.Join(tmpDir, "test-category-no-index")
	if err := os.Mkdir(categoryNoIndexDir, 0755); err != nil {
		t.Fatalf("Failed to create category no index dir: %v", err)
	}

	return tmpDir
}

func TestHealthCheckHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthcheck", nil)
	rr := httptest.NewRecorder()
	mux := webserver.SetupServer("")
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "OK"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestServeArchiveFile_Success(t *testing.T) {
	archiveBasePath := setupTestArchive(t)
	mux := webserver.SetupServer(archiveBasePath)

	req := httptest.NewRequest("GET", "/archive/test-category/test-file.html", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "<html><body><h1>Test File</h1></body></html>"
	body, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	if string(body) != expected {
		t.Errorf("handler returned unexpected body: got %q want %q",
			string(body), expected)
	}
}

func TestServeArchiveFile_NotFound(t *testing.T) {
	archiveBasePath := setupTestArchive(t)
	mux := webserver.SetupServer(archiveBasePath)

	req := httptest.NewRequest("GET", "/archive/test-category/not-found.html", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}
}

func TestServeArchiveFile_PathTraversal(t *testing.T) {
	archiveBasePath := setupTestArchive(t)
	mux := webserver.SetupServer(archiveBasePath)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse 
		},
	}

	req, err := http.NewRequest("GET", ts.URL+"/archive/../main_test.go", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusMovedPermanently {
		t.Errorf("handler returned wrong status code for path traversal attempt: got %v want %v",
			resp.StatusCode, http.StatusMovedPermanently)
	}

	redirectURL, err := resp.Location()
	if err != nil {
		t.Fatalf("Failed to get redirect location: %v", err)
	}

	req, err = http.NewRequest("GET", redirectURL.String(), nil)
	if err != nil {
		t.Fatalf("Failed to create redirect request: %v", err)
	}

	clientWithRedirect := &http.Client{}
	resp, err = clientWithRedirect.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform redirect request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("handler returned wrong status code after redirect for path traversal attempt: got %v want %v",
			resp.StatusCode, http.StatusNotFound)
	}
}

func TestServeArchiveFile_DirectoryListingDisabled(t *testing.T) {
	archiveBasePath := setupTestArchive(t)
	handler := webserver.SetupServer(archiveBasePath)

	if err := os.Mkdir(filepath.Join(archiveBasePath, "empty-dir"), 0755); err != nil {
		t.Fatalf("Failed to create empty directory: %v", err)
	}

	req := httptest.NewRequest("GET", "/archive/empty-dir/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code for directory listing attempt: got %v want %v",
			status, http.StatusNotFound)
	}
}

func TestServeArchiveFile_NoDirectoryListingAndIndexHtml(t *testing.T) {
	archiveBasePath := setupTestArchive(t)
	mux := webserver.SetupServer(archiveBasePath)

	req1 := httptest.NewRequest("GET", "/archive/test-category-with-index/", nil)
	rr1 := httptest.NewRecorder()
	mux.ServeHTTP(rr1, req1)

	if status := rr1.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code for directory with index.html: got %v want %v",
			status, http.StatusOK)
	}

	expected1 := "<html><body><h1>Index File</h1></body></html>"
	body1, err := io.ReadAll(rr1.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	if string(body1) != expected1 {
		t.Errorf("handler returned unexpected body for directory with index.html: got %q want %q",
			string(body1), expected1)
	}

	req2 := httptest.NewRequest("GET", "/archive/test-category-no-index/", nil)
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, req2)

	if status := rr2.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code for directory without index.html: got %v want %v",
			status, http.StatusNotFound)
	}
}

func TestInitScheduler(t *testing.T) {
	scheduler, err := initScheduler()
	if err != nil {
		t.Fatalf("initScheduler() returned an error: %v", err)
	}
	if scheduler == nil {
		t.Fatal("initScheduler() returned a nil scheduler")
	}
}

func TestInitServices(t *testing.T) {
	cfg := &config.Config{
		Miniflux: config.ConfigMiniflux{
			Host:     "http://localhost",
			ApiToken: "test",
		},
		AI: config.ConfigAI{
			ApiKey: "test",
		},
	}

	app, err := initServices(cfg)
	if err != nil {
		t.Fatalf("initServices returned an unexpected error: %v", err)
	}

	if app.ArchiveService == nil {
		t.Error("ArchiveService should not be nil")
	}

	if app.EmailService == nil {
		t.Error("EmailService should not be nil")
	}

	if app.MinifluxClientService == nil {
		t.Error("MinifluxClientService should not be nil")
	}

	if app.DigestService == nil {
		t.Error("DigestService should not be nil")
	}

	if app.LLMService == nil {
		t.Error("LLMService should not be nil")
	}

	emailServiceImpl, ok := app.EmailService.(*email.EmailServiceImpl)
	if !ok {
		t.Fatal("EmailService is not of type *email.EmailServiceImpl")
	}

	if emailServiceImpl.EmailTemplate == nil {
		t.Error("EmailTemplate should not be nil")
	}
}

type mockMinifluxClient struct {
	getAllUnreadEntriesFunc func() ([]*models.Entry, error)
}

func (m *mockMinifluxClient) FeedIcon(feedID int64) (*models.FeedIcon, error) { return nil, nil }
func (m *mockMinifluxClient) GetAllUnreadEntries() ([]*models.Entry, error) {
	return m.getAllUnreadEntriesFunc()
}
func (m *mockMinifluxClient) UpdateEntries(entryIDs []int64, status string) error { return nil }

type mockEmailService struct {
	sendCalled bool
}

func (m *mockEmailService) Send(smtpConfig config.ConfigSmtp, digestConfig config.ConfigDigest, overviewFile *os.File, groupedEntryFiles []*os.File, data *models.OverviewTemplateData) error {
	m.sendCalled = true
	return nil
}

type mockDigestService struct{}

func (m *mockDigestService) BuildDigestData(entries []*models.Entry, icons map[int64]*models.FeedIcon, view string, minifluxHost string, digestHost string, categories []config.ConfigCategory, digestTitle string) *models.OverviewTemplateData {
	return &models.OverviewTemplateData{Entries: entries}
}

type mockArchiveService struct{}

func (m *mockArchiveService) MakeArchiveHTML(data *models.OverviewTemplateData, compress bool) (*os.File, []*os.File, error) {
	return nil, nil, nil
}
func (m *mockArchiveService) CleanArchive(maxAge time.Duration) {}

func TestDigestJob_EmptyDigest(t *testing.T) {
	tests := []struct {
		name         string
		sendIfEmpty  bool
		expectSend   bool
		smtpHost     string
	}{
		{
			name:        "send_if_empty=false, no email sent",
			sendIfEmpty: false,
			expectSend:  false,
			smtpHost:    "smtp.test.com",
		},
		{
			name:        "send_if_empty=true, email sent",
			sendIfEmpty: true,
			expectSend:  true,
			smtpHost:    "smtp.test.com",
		},
		{
			name:        "send_if_empty=true, but no SMTP configured, no email sent",
			sendIfEmpty: true,
			expectSend:  false,
			smtpHost:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEmail := &mockEmailService{}
			mockMiniflux := &mockMinifluxClient{
				getAllUnreadEntriesFunc: func() ([]*models.Entry, error) {
					return []*models.Entry{}, nil
				},
			}

			cfg := &config.Config{
				Smtp: config.ConfigSmtp{Host: tt.smtpHost},
				Digests: []config.ConfigDigest{
					{
						Title:       "Test Digest",
						SendIfEmpty: &tt.sendIfEmpty,
					},
				},
			}

			mockApp := app.NewApp(
				app.WithConfig(cfg),
				app.WithEmailService(mockEmail),
				app.WithMinifluxClientService(mockMiniflux),
				app.WithDigestService(&mockDigestService{}),
				app.WithArchiveService(&mockArchiveService{}),
			)

			digestJob(mockApp, 0, "test")

			if mockEmail.sendCalled != tt.expectSend {
				t.Errorf("Expected sendCalled to be %v, got %v", tt.expectSend, mockEmail.sendCalled)
			}
		})
	}
}
