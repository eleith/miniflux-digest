package testutil

import (
	"time"

	"miniflux-digest/internal/models"
	miniflux "miniflux.app/v2/client"
)

func NewMockCategory() *miniflux.Category {
	return &miniflux.Category{
		ID:    1,
		Title: "Test Category",
	}
}

func NewMockFeed() *miniflux.Feed {
	return &miniflux.Feed{
		ID:    1,
		Title: "Test Feed",
	}
}

func NewMockEntry() *miniflux.Entry {
	return &miniflux.Entry{
		ID:      1,
		Title:   "Test Entry",
		Content: "<p>Test Content</p>",
		Feed:    NewMockFeed(),
	}
}

func NewMockFeedIcon() *models.FeedIcon {
	return &models.FeedIcon{
		FeedID: 1,
		Data:   "image/png;base64,abcd1234=",
	}
}

func NewMockCategoryData() *models.CategoryData {
	return &models.CategoryData{
		Category:      NewMockCategory(),
		Entries:       &miniflux.Entries{NewMockEntry()},
		GeneratedDate: time.Date(2025, 7, 21, 12, 0, 0, 0, time.UTC),
		FeedIcons:     []*models.FeedIcon{NewMockFeedIcon()},
	}
}
