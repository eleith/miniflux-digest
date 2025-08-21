package services

import (
	miniflux "miniflux.app/v2/client"
)

type MinifluxClientService interface {
	
	FeedIcon(feedID int64) (*miniflux.FeedIcon, error)
	
	GetAllUnreadEntries() (*miniflux.Entries, error)
	UpdateEntries(entryIDs []int64, status string) error
}
