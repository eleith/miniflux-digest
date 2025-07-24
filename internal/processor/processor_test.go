package processor

import (
	"bytes"
	"errors"
	"io"
	"log"
	"miniflux-digest/internal/app"
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/testutil"
	"os"
	"testing"

	miniflux "miniflux.app/v2/client"
)

func TestProcessCategory(t *testing.T) {
	log.SetOutput(io.Discard)
	defer log.SetOutput(os.Stderr)

	t.Run("no entries", func(t *testing.T) {
		mockApp := &app.App{
			Config: &config.Config{},
			MinifluxClientService: &testutil.MockMinifluxClient{
				MarkAsReadFunc: func(categoryID int64) error { return nil },
			},
			ArchiveService: &testutil.MockArchiveService{},
			EmailService: &testutil.MockEmailService{},
		}
		data := &models.CategoryData{Entries: &miniflux.Entries{}}
		ProcessCategory(mockApp, data, true)
	})

	t.Run("error making archive html", func(t *testing.T) {
		mockApp := &app.App{
			Config: &config.Config{ArchivePath: "/tmp/archive"},
			MinifluxClientService: &testutil.MockMinifluxClient{
				MarkAsReadFunc: func(categoryID int64) error { return nil },
			},
			EmailService: &testutil.MockEmailService{},
			ArchiveService: &testutil.MockArchiveService{
				MakeArchiveHTMLFunc: func(archivePath string, data *models.CategoryData) (*os.File, error) {
					return nil, errors.New("failed to make archive html")
				},
			},
		}
		data := testutil.NewMockCategoryData()

		var buf bytes.Buffer
		log.SetOutput(&buf)

		ProcessCategory(mockApp, data, true)

		if !bytes.Contains(buf.Bytes(), []byte("Error generating File")) {
			t.Error("Expected error log for archive generation, but not found")
		}
	})

	t.Run("error sending email", func(t *testing.T) {
		mockApp := &app.App{
			Config: &config.Config{ArchivePath: "/tmp/archive"},
			MinifluxClientService: &testutil.MockMinifluxClient{
				MarkAsReadFunc: func(categoryID int64) error {
					return nil // No-op for this test case
				},
			},
			ArchiveService: &testutil.MockArchiveService{
				MakeArchiveHTMLFunc: func(archivePath string, data *models.CategoryData) (*os.File, error) {
					return os.CreateTemp("", "test-archive-*.html")
				},
			},
			EmailService: &testutil.MockEmailService{
				SendFunc: func(cfg *config.Config, file *os.File, data *models.CategoryData) error {
					return errors.New("failed to send email")
				},
			},
		}
		data := testutil.NewMockCategoryData()

		var buf bytes.Buffer
		log.SetOutput(&buf)

		ProcessCategory(mockApp, data, true)

		if !bytes.Contains(buf.Bytes(), []byte("Error sending email")) {
			t.Error("Expected error log for email sending, but not found")
		}
	})

	t.Run("mark as read is called", func(t *testing.T) {
		markAsReadCalled := false
		mockMinifluxClient := &testutil.MockMinifluxClient{
			MarkAsReadFunc: func(categoryID int64) error {
				markAsReadCalled = true
				return nil
			},
		}
		mockApp := &app.App{
			Config:                &config.Config{ArchivePath: "/tmp/archive"},
			MinifluxClientService: mockMinifluxClient,
			ArchiveService: &testutil.MockArchiveService{
				MakeArchiveHTMLFunc: func(archivePath string, data *models.CategoryData) (*os.File, error) {
					return os.CreateTemp("", "test-archive-*.html")
				},
			},
			EmailService: &testutil.MockEmailService{
				SendFunc: func(cfg *config.Config, file *os.File, data *models.CategoryData) error {
					return nil
				},
			},
		}
		data := testutil.NewMockCategoryData()

		ProcessCategory(mockApp, data, true)

		if !markAsReadCalled {
			t.Error("Expected MarkCategoryAsRead to be called, but it was not")
		}
	})

	t.Run("error marking as read", func(t *testing.T) {
		var buf bytes.Buffer
		log.SetOutput(&buf)

		mockMinifluxClient := &testutil.MockMinifluxClient{
			MarkAsReadFunc: func(categoryID int64) error {
				return errors.New("failed to mark as read")
			},
		}
		mockApp := &app.App{
			Config:                &config.Config{ArchivePath: "/tmp/archive"},
			MinifluxClientService: mockMinifluxClient,
			ArchiveService: &testutil.MockArchiveService{
				MakeArchiveHTMLFunc: func(archivePath string, data *models.CategoryData) (*os.File, error) {
					return os.CreateTemp("", "test-archive-*.html")
				},
			},
			EmailService: &testutil.MockEmailService{
				SendFunc: func(cfg *config.Config, file *os.File, data *models.CategoryData) error {
					return nil
				},
			},
		}
		data := testutil.NewMockCategoryData()

		ProcessCategory(mockApp, data, true)

		if !bytes.Contains(buf.Bytes(), []byte("Error marking category as read")) {
			t.Error("Expected error log for marking as read, but not found")
		}
	})
}