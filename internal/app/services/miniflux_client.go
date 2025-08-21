package services

import "miniflux-digest/internal/models"

type MinifluxClientService interface {
	FeedIcon(feedID int64) (*models.FeedIcon, error)
	GetAllUnreadEntries() ([]*models.Entry, error)
	UpdateEntries(entryIDs []int64, status string) error
}
