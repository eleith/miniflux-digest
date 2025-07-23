package app

import (
	"miniflux-digest/config"
	"miniflux-digest/internal/models"
	"os"
)

type ArchiveService interface {
	MakeArchiveHTML(archivePath string, data *models.CategoryData) (*os.File, error)
}

type EmailService interface {
		Send(cfg *config.Config, file *os.File, data *models.CategoryData) error
}

type MinifluxClientService interface {
	MarkCategoryAsRead(categoryID int64) error
}
