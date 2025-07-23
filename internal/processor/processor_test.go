package processor

import (
	"bytes"
	"errors"
	"io"
	"log"
	"miniflux-digest/config"
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/testutil"
	miniflux "miniflux.app/v2/client"
	"os"
	"testing"
)

// MockMinifluxClient is a mock implementation of app.MinifluxClientService
type MockMinifluxClient struct {
	MarkAsReadFunc func(categoryID int64) error
}

func (m *MockMinifluxClient) MarkCategoryAsRead(categoryID int64) error {
	return m.MarkAsReadFunc(categoryID)
}

// MockArchiveService is a mock implementation of app.ArchiveService
type MockArchiveService struct {
	MakeArchiveHTMLFunc func(archivePath string, data *models.CategoryData) (*os.File, error)
}

func (m *MockArchiveService) MakeArchiveHTML(archivePath string, data *models.CategoryData) (*os.File, error) {
	return m.MakeArchiveHTMLFunc(archivePath, data)
}

// MockEmailService is a mock implementation of app.EmailService
type MockEmailService struct {
	SendFunc func(cfg *config.Config, file *os.File, data *models.CategoryData) error
}

func (m *MockEmailService) Send(cfg *config.Config, file *os.File, data *models.CategoryData) error {
	return m.SendFunc(cfg, file, data)
}

func TestProcessCategory(t *testing.T) {
	// Suppress log output during tests
	log.SetOutput(io.Discard)
	defer log.SetOutput(os.Stderr)

	cfg := &config.Config{}
	archivePath := "/tmp/archive"

	t.Run("no entries", func(t *testing.T) {
		client := &MockMinifluxClient{}
		archiveService := &MockArchiveService{}
		emailService := &MockEmailService{}
		data := &models.CategoryData{Entries: &miniflux.Entries{}}
		ProcessCategory(cfg, client, data, archiveService, emailService, archivePath, true)
		// No panics or errors expected
	})

	t.Run("send email and mark as read", func(t *testing.T) {
		markedAsRead := false
		sentEmail := false

		client := &MockMinifluxClient{
			MarkAsReadFunc: func(categoryID int64) error {
				markedAsRead = true
				return nil
			},
		}

		archiveService := &MockArchiveService{
			MakeArchiveHTMLFunc: func(archivePath string, data *models.CategoryData) (*os.File, error) {
				// Create a dummy file for testing
				return os.CreateTemp("", "test-archive-*.html")
			},
		}

		emailService := &MockEmailService{
			SendFunc: func(cfg *config.Config, file *os.File, data *models.CategoryData) error {
				sentEmail = true
				return nil
			},
		}

		data := testutil.NewMockCategoryData()
		ProcessCategory(cfg, client, data, archiveService, emailService, archivePath, true)

		if !markedAsRead {
			t.Error("Expected category to be marked as read, but it was not")
		}
		if !sentEmail {
			t.Error("Expected email to be sent, but it was not")
		}
	})

	t.Run("do not mark as read", func(t *testing.T) {
		markedAsRead := false
		sentEmail := false

		client := &MockMinifluxClient{
			MarkAsReadFunc: func(categoryID int64) error {
				markedAsRead = true
				return nil
			},
		}

		archiveService := &MockArchiveService{
			MakeArchiveHTMLFunc: func(archivePath string, data *models.CategoryData) (*os.File, error) {
				return os.CreateTemp("", "test-archive-*.html")
			},
		}

		emailService := &MockEmailService{
			SendFunc: func(cfg *config.Config, file *os.File, data *models.CategoryData) error {
				sentEmail = true
				return nil
			},
		}

		data := testutil.NewMockCategoryData()
		ProcessCategory(cfg, client, data, archiveService, emailService, archivePath, false)

		if markedAsRead {
			t.Error("Expected category not to be marked as read, but it was")
		}
		if !sentEmail {
			t.Error("Expected email to be sent, but it was not")
		}
	})

	t.Run("error making archive html", func(t *testing.T) {
		client := &MockMinifluxClient{}
		emailService := &MockEmailService{}
		data := testutil.NewMockCategoryData()

		archiveService := &MockArchiveService{
			MakeArchiveHTMLFunc: func(archivePath string, data *models.CategoryData) (*os.File, error) {
				return nil, errors.New("failed to make archive html")
			},
		}

		// Capture log output
		var buf bytes.Buffer
		log.SetOutput(&buf)

		ProcessCategory(cfg, client, data, archiveService, emailService, archivePath, true)

		if !bytes.Contains(buf.Bytes(), []byte("Error generating File")) {
			t.Error("Expected error log for archive generation, but not found")
		}
	})

	t.Run("error sending email", func(t *testing.T) {
		client := &MockMinifluxClient{
			MarkAsReadFunc: func(categoryID int64) error {
				return nil // No-op for this test case
			},
		}
		data := testutil.NewMockCategoryData()

		archiveService := &MockArchiveService{
			MakeArchiveHTMLFunc: func(archivePath string, data *models.CategoryData) (*os.File, error) {
				return os.CreateTemp("", "test-archive-*.html")
			},
		}

		emailService := &MockEmailService{
			SendFunc: func(cfg *config.Config, file *os.File, data *models.CategoryData) error {
				return errors.New("failed to send email")
			},
		}

		// Capture log output
		var buf bytes.Buffer
		log.SetOutput(&buf)

		ProcessCategory(cfg, client, data, archiveService, emailService, archivePath, true)

		if !bytes.Contains(buf.Bytes(), []byte("Error sending email")) {
			t.Error("Expected error log for email sending, but not found")
		}
	})
}
