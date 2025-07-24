package app

import (
	"fmt"
	"log"
	"time"

	"miniflux-digest/internal/models"
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

func (m *MinifluxClientWrapper) categories() ([]*miniflux.Category, error) {
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

func (m *MinifluxClientWrapper) FetchCategoryData(categoryID int64) (*models.CategoryData, error) {
	categories, err := m.categories()
	if err != nil {
		return nil, err
	}

	var category *miniflux.Category
	for _, c := range categories {
		if c.ID == categoryID {
			category = c
			break
		}
	}

	if category == nil {
		return nil, fmt.Errorf("category with ID %d not found", categoryID)
	}

	entriesResult, err := m.client.CategoryEntries(category.ID, &miniflux.Filter{Status: miniflux.EntryStatusUnread})
	if err != nil {
		return nil, err
	}

	feeds, err := m.client.CategoryFeeds(category.ID)
	feedIcons := []*models.FeedIcon{}

	if err != nil {
		return nil, err
	}

	for _, feed := range feeds {
		feedIcon, err := m.client.FeedIcon(feed.ID)
		if err != nil {
			log.Printf("Warning: failed to fetch icon for feed %d: %v", feed.ID, err)
			continue
		}

		feedIcons = append(feedIcons, &models.FeedIcon{
			FeedID: feed.ID,
			Data:   feedIcon.Data,
		})
	}

	return &models.CategoryData{
		Category:      category,
		Entries:       &entriesResult.Entries,
		GeneratedDate: time.Now(),
		FeedIcons:     feedIcons,
	}, nil
}

func (m *MinifluxClientWrapper) StreamAllCategoryData() <-chan *models.CategoryData {
	out := make(chan *models.CategoryData)

	go func() {
		defer close(out)

		categories, err := m.categories()

		if err != nil {
			log.Printf("Streamer failed to fetch categories: %v", err)
			return
		}

		for _, category := range categories {
			data, err := m.FetchCategoryData(category.ID)
			if err != nil {
				log.Printf("Streamer failed to fetch data for category %q: %v", category.Title, err)
				continue
			}

			out <- data
		}
	}()

	return out
}
