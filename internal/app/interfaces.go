package app

import (
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/models"
	"os"
	"time"

	miniflux "miniflux.app/v2/client"
)

type ArchiveService interface {
	MakeArchiveHTML(data *models.HTMLTemplateData, compress bool) (*os.File, error)
	CleanArchive(maxAge time.Duration)
}

type EmailService interface {
	Send(cfg *config.Config, overviewFile *os.File, groupedEntryFiles []*os.File, data *models.HTMLTemplateData) error
}



type MinifluxClientService interface {
	MarkCategoryAsRead(categoryID int64) error
	CategoryEntries(categoryID int64, filter *miniflux.Filter) (*miniflux.Entries, error)
	CategoryFeeds(categoryID int64) ([]*miniflux.Feed, error)
	FeedIcon(feedID int64) (*miniflux.FeedIcon, error)
	StreamAllCategoryData() <-chan *RawCategoryData
	GetAllUnreadEntries() (*miniflux.Entries, error)
}

type RawCategoryData struct {
	Category *miniflux.Category
	Entries  *miniflux.Entries
	Feeds    []*miniflux.Feed
	Icons    map[int64]*models.FeedIcon
	Err      error
}