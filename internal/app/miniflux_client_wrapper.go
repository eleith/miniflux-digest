package app

import (
	"miniflux-digest/internal/models"
	miniflux "miniflux.app/v2/client"
)

type MinifluxClientWrapper struct {
	client *miniflux.Client
}

func NewMinifluxClientWrapper(client *miniflux.Client) *MinifluxClientWrapper {
	return &MinifluxClientWrapper{client: client}
}

func (m *MinifluxClientWrapper) FeedIcon(feedID int64) (*models.FeedIcon, error) {
	icon, err := m.client.FeedIcon(feedID)

	if err != nil {
		return nil, err
	}

	return &models.FeedIcon{
		FeedID: feedID,
		Data:   icon.Data,
	}, nil
}

func (m *MinifluxClientWrapper) GetAllUnreadEntries() ([]*models.Entry, error) {
	minifluxEntries, err := m.client.Entries(&miniflux.Filter{Status: miniflux.EntryStatusUnread})
	if err != nil {
		return nil, err
	}

	var entries []*models.Entry
	for _, entry := range minifluxEntries.Entries {
		var categoryID int64
		var categoryTitle string
		if entry.Feed != nil && entry.Feed.Category != nil {
			categoryID = entry.Feed.Category.ID
			categoryTitle = entry.Feed.Category.Title
		}
		var feedTitle string
		var siteURL string
		if entry.Feed != nil {
			feedTitle = entry.Feed.Title
			siteURL = entry.Feed.SiteURL
		}

		entries = append(entries, &models.Entry{
			ID:            entry.ID,
			Title:         entry.Title,
			URL:           entry.URL,
			SiteURL:       siteURL,
			Content:       entry.Content,
			FeedID:        entry.FeedID,
			FeedTitle:     feedTitle,
			GroupID:       categoryID,
			GroupTitle:    categoryTitle,
			CommentsURL:   entry.CommentsURL,
			Date:          entry.Date,
		})
	}
	return entries, nil
}

func (m *MinifluxClientWrapper) UpdateEntries(entryIDs []int64, status string) error {
	return m.client.UpdateEntries(entryIDs, status)
}
