package testutil

import (
	"miniflux-digest/internal/app"
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/models"
	"os"
	"time"
)

type MockMinifluxClient struct {
	app.MinifluxClientService
	MarkAsReadFunc func(categoryID int64) error
}

func (m *MockMinifluxClient) MarkCategoryAsRead(categoryID int64) error {
	return m.MarkAsReadFunc(categoryID)
}

type MockArchiveService struct {
	app.ArchiveService
	MakeArchiveHTMLFunc func(archivePath string, data *models.CategoryData) (*os.File, error)
}

func (m *MockArchiveService) MakeArchiveHTML(archivePath string, data *models.CategoryData) (*os.File, error) {
	return m.MakeArchiveHTMLFunc(archivePath, data)
}

func (m *MockArchiveService) CleanArchive(archivePath string, maxAge time.Duration) {}

type MockEmailService struct {
	app.EmailService
	SendFunc func(cfg *config.Config, file *os.File, data *models.CategoryData) error
}

func (m *MockEmailService) Send(cfg *config.Config, file *os.File, data *models.CategoryData) error {
	return m.SendFunc(cfg, file, data)
}
