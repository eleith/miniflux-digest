package app

import (
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/models"
	"os"
	"time"
	miniflux "miniflux.app/v2/client"
)

type ArchiveService interface {
	MakeArchiveHTML(archivePath string, data *models.CategoryData) (*os.File, error)
	CleanArchive(archivePath string, maxAge time.Duration)
}

type EmailService interface {
		Send(cfg *config.Config, file *os.File, data *models.CategoryData) error
}

type MinifluxClientService interface {
	MarkCategoryAsRead(categoryID int64) error
	Categories() ([]*miniflux.Category, error)
	CategoryEntries(categoryID int64, filter *miniflux.Filter) (*miniflux.Entries, error)
	CategoryFeeds(categoryID int64) ([]*miniflux.Feed, error)
	FeedIcon(feedID int64) (*miniflux.FeedIcon, error)
}

type CategoryService interface {
	StreamData(client MinifluxClientService) <-chan *models.CategoryData
}
