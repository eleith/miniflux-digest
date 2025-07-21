package testutil

import (
	"miniflux-digest/internal/category"
	"time"

	miniflux "miniflux.app/v2/client"
)

// NewMockCategoryData creates a standardized CategoryData object for testing.
func NewMockCategoryData() *category.CategoryData {
	return &category.CategoryData{
		Category: &miniflux.Category{
			ID:    1,
			Title: "Test Category",
		},
		Entries: &miniflux.Entries{
			&miniflux.Entry{
				ID:      1,
				Title:   "Test Entry",
				Content: "Test Content",
			},
		},
		// Use a fixed time for deterministic test results
		GeneratedDate: time.Date(2025, 7, 21, 12, 0, 0, 0, time.UTC),
	}
}
