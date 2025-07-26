package email

import (
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/templates"
	"miniflux-digest/internal/testutil"
	"os"
	"strings"
	"testing"
)

func TestSend(t *testing.T) {
	cfg := &config.Config{
		SmtpHost:        "localhost",
		SmtpPort:        1025,
		SmtpUser:        "test-user",
		SmtpPassword:    "test-password",
		DigestEmailTo:   "to@example.com",
		DigestEmailFrom: "from@example.com",
		DigestHost:      "https://example.com",
	}

	tmpFile, err := os.CreateTemp("", "test.html")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()
	if _, err := tmpFile.WriteString("<html><body><h1>Test</h1></body></html>"); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	file, err := os.Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open temp file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("Failed to close file: %v", err)
		}
	}()

	data := testutil.NewMockHTMLTemplateData()

	// In a real scenario, you would use a mock SMTP server.
	// For this test, we are just checking if the function executes without error.
	// The go-mail library does not make it easy to mock the SMTP client.
	emailService := &EmailServiceImpl{}
	err = emailService.Send(cfg, file, data)
	if err != nil {
		// We expect an error because we are not running a real SMTP server.
		// The important part is that the function attempts to connect.
		if !strings.Contains(err.Error(), "connection refused") {
			t.Errorf("Expected connection refused error, got: %v", err)
		}
	}
}

func TestTextTemplateData(t *testing.T) {
	htmlTemplateData := *testutil.NewMockHTMLTemplateData()
	url := "https://example.com"

	textData := templates.EmailTemplateData{
		HTMLTemplateData: htmlTemplateData,
		URL:          url,
	}

	if textData.URL != url {
		t.Errorf("Expected URL to be %s, got %s", url, textData.URL)
	}

	if textData.Category.Title != "Test Category" {
		t.Errorf("Expected category title to be 'Test Category', got %s", textData.Category.Title)
	}
}
