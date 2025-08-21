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



func (m *MinifluxClientWrapper) FeedIcon(feedID int64) (*miniflux.FeedIcon, error) {
	return m.client.FeedIcon(feedID)
}





		

func (m *MinifluxClientWrapper) GetAllUnreadEntries() (*miniflux.Entries, error) {
	entries, err := m.client.Entries(&miniflux.Filter{Status: miniflux.EntryStatusUnread})
	if err != nil {
		return nil, err
	}
	return &entries.Entries, nil
}

func (m *MinifluxClientWrapper) UpdateEntries(entryIDs []int64, status string) error {
	return m.client.UpdateEntries(entryIDs, status)
}
