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

func TestCategoryDigestJob(t *testing.T) {
	log.SetOutput(io.Discard)
	defer log.SetOutput(os.Stderr)

	t.Run("no entries", func(t *testing.T) {
		mockApp := &app.App{
			Config:                &config.Config{},
			MinifluxClientService: &testutil.MockMinifluxClient{},
			DigestService: &testutil.MockDigestService{
				BuildDigestDataFunc: func(category *miniflux.Category, entries *miniflux.Entries, icons map[int64]*models.FeedIcon) *models.HTMLTemplateData {
					return &models.HTMLTemplateData{Entries: &miniflux.Entries{}}
				},
			},
			ArchiveService: &testutil.MockArchiveService{},
			EmailService:   &testutil.MockEmailService{},
		}
		data := &app.RawCategoryData{Entries: &miniflux.Entries{}}
		CategoryDigestJob(mockApp, data, true)
	})

	t.Run("error making archive html", func(t *testing.T) {
		mockApp := &app.App{
			Config:                &config.Config{},
			MinifluxClientService: &testutil.MockMinifluxClient{},
			DigestService: &testutil.MockDigestService{
				BuildDigestDataFunc: func(category *miniflux.Category, entries *miniflux.Entries, icons map[int64]*models.FeedIcon) *models.HTMLTemplateData {
					return &models.HTMLTemplateData{Entries: &miniflux.Entries{{ID: 1}}, Category: &miniflux.Category{Title: "title"}}
				},
			},
			EmailService: &testutil.MockEmailService{},
			ArchiveService: &testutil.MockArchiveService{
				MakeArchiveHTMLFunc: func(data *models.HTMLTemplateData, minify bool) (*os.File, error) {
					return nil, errors.New("failed to make archive html")
				},
			},
		}
		data := &app.RawCategoryData{Entries: &miniflux.Entries{{ID: 1}}}

		var buf bytes.Buffer
		log.SetOutput(&buf)

		CategoryDigestJob(mockApp, data, true)

		if !bytes.Contains(buf.Bytes(), []byte("Error generating File")) {
			t.Error("Expected error log for archive generation, but not found")
		}
	})

	t.Run("error sending email", func(t *testing.T) {
		mockApp := &app.App{
			Config:                &config.Config{},
			MinifluxClientService: &testutil.MockMinifluxClient{},
			DigestService: &testutil.MockDigestService{
				BuildDigestDataFunc: func(category *miniflux.Category, entries *miniflux.Entries, icons map[int64]*models.FeedIcon) *models.HTMLTemplateData {
					return &models.HTMLTemplateData{Entries: &miniflux.Entries{{ID: 1}}, Category: &miniflux.Category{Title: "title"}, FeedIcons: []*models.FeedIcon{}}
				},
			},
			ArchiveService: &testutil.MockArchiveService{
				MakeArchiveHTMLFunc: func(data *models.HTMLTemplateData, minify bool) (*os.File, error) {
					return os.CreateTemp("", "test-archive-*.html")
				},
			},
			EmailService: &testutil.MockEmailService{
				SendFunc: func(cfg *config.Config, file *os.File, data *models.HTMLTemplateData) error {
					return errors.New("failed to send email")
				},
			},
		}
		data := &app.RawCategoryData{Entries: &miniflux.Entries{{ID: 1}}}

		var buf bytes.Buffer
		log.SetOutput(&buf)

		CategoryDigestJob(mockApp, data, true)

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
			Config:                &config.Config{},
			MinifluxClientService: mockMinifluxClient,
			DigestService: &testutil.MockDigestService{
				BuildDigestDataFunc: func(category *miniflux.Category, entries *miniflux.Entries, icons map[int64]*models.FeedIcon) *models.HTMLTemplateData {
					return &models.HTMLTemplateData{Entries: &miniflux.Entries{{ID: 1}}, Category: &miniflux.Category{Title: "title"}, FeedIcons: []*models.FeedIcon{}}
				},
			},
			ArchiveService: &testutil.MockArchiveService{
				MakeArchiveHTMLFunc: func(data *models.HTMLTemplateData, minify bool) (*os.File, error) {
					return os.CreateTemp("", "test-archive-*.html")
				},
			},
			EmailService: &testutil.MockEmailService{
				SendFunc: func(cfg *config.Config, file *os.File, data *models.HTMLTemplateData) error {
					return nil
				},
			},
		}
		data := &app.RawCategoryData{Entries: &miniflux.Entries{{ID: 1}}}

		CategoryDigestJob(mockApp, data, true)

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
			Config:                &config.Config{},
			MinifluxClientService: mockMinifluxClient,
			DigestService: &testutil.MockDigestService{
				BuildDigestDataFunc: func(category *miniflux.Category, entries *miniflux.Entries, icons map[int64]*models.FeedIcon) *models.HTMLTemplateData {
					return &models.HTMLTemplateData{Entries: &miniflux.Entries{{ID: 1}}, Category: &miniflux.Category{Title: "title"}, FeedIcons: []*models.FeedIcon{}}
				},
			},
			ArchiveService: &testutil.MockArchiveService{
				MakeArchiveHTMLFunc: func(data *models.HTMLTemplateData, minify bool) (*os.File, error) {
					return os.CreateTemp("", "test-archive-*.html")
				},
			},
			EmailService: &testutil.MockEmailService{
				SendFunc: func(cfg *config.Config, file *os.File, data *models.HTMLTemplateData) error {
					return nil
				},
			},
		}
		data := &app.RawCategoryData{Entries: &miniflux.Entries{{ID: 1}}}

		CategoryDigestJob(mockApp, data, true)

		if !bytes.Contains(buf.Bytes(), []byte("Error marking category as read")) {
			t.Error("Expected error log for marking as read, but not found")
		}
	})
}
