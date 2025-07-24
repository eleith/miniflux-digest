package app

import (
	miniflux "miniflux.app/v2/client"
)

type MinifluxClientWrapper struct {
	client *miniflux.Client
}

func NewMinifluxClientWrapper(client *miniflux.Client) *MinifluxClientWrapper {
	return &MinifluxClientWrapper{client: client}
}

func (m *MinifluxClientWrapper) MarkCategoryAsRead(categoryID int64) error {
	return m.client.MarkCategoryAsRead(categoryID)
}

func (m *MinifluxClientWrapper) Categories() ([]*miniflux.Category, error) {
	return m.client.Categories()
}

func (m *MinifluxClientWrapper) CategoryEntries(categoryID int64, filter *miniflux.Filter) (*miniflux.Entries, error) {
	entries, err := m.client.CategoryEntries(categoryID, filter)
	if err != nil {
		return nil, err
	}
	return &entries.Entries, nil
}

func (m *MinifluxClientWrapper) CategoryFeeds(categoryID int64) ([]*miniflux.Feed, error) {
	return m.client.CategoryFeeds(categoryID)
}

func (m *MinifluxClientWrapper) FeedIcon(feedID int64) (*miniflux.FeedIcon, error) {
	return m.client.FeedIcon(feedID)
}
