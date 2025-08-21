package services

import (
	"miniflux-digest/internal/models"
	miniflux "miniflux.app/v2/client"
)

type MinifluxClientService interface {
	MarkCategoryAsRead(categoryID int64) error
	CategoryEntries(categoryID int64, filter *miniflux.Filter) (*miniflux.Entries, error)
	CategoryFeeds(categoryID int64) ([]*miniflux.Feed, error)
	FeedIcon(feedID int64) (*miniflux.FeedIcon, error)
	StreamAllCategoryData() <-chan *models.RawCategoryData
	GetAllUnreadEntries() (*miniflux.Entries, error)
	UpdateEntries(entryIDs []int64, status string) error
}
