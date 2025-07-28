package app

import (
	"context"
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/models"
	"os"
	"time"
	miniflux "miniflux.app/v2/client"

	"miniflux-digest/internal/digest"
)

type ArchiveService interface {
	MakeArchiveHTML(data *models.HTMLTemplateData, compress bool) (*os.File, error)
	CleanArchive(maxAge time.Duration)
}

type EmailService interface {
		Send(cfg *config.Config, file *os.File, data *models.HTMLTemplateData) error
}

type DigestService interface {
	BuildDigestData(category *miniflux.Category, entries *miniflux.Entries, icons map[int64]*models.FeedIcon, groupBy digest.GroupingType) *models.HTMLTemplateData
}

type MinifluxClientService interface {
	MarkCategoryAsRead(categoryID int64) error
	CategoryEntries(categoryID int64, filter *miniflux.Filter) (*miniflux.Entries, error)
	CategoryFeeds(categoryID int64) ([]*miniflux.Feed, error)
	FeedIcon(feedID int64) (*miniflux.FeedIcon, error)
	StreamAllCategoryData() <-chan *RawCategoryData
}

type LLMService interface {
	GenerateContent(ctx context.Context, prompt string) (string, error)
}
