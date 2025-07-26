package testutil

import (
	"encoding/base64"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"miniflux-digest/internal/digest"
	"miniflux-digest/internal/models"
	miniflux "miniflux.app/v2/client"
)

func loadImageAsBase64(path string) string {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	imagePath := filepath.Join(basepath, "images", filepath.Base(path))

	file, err := os.ReadFile(imagePath)
	if err != nil {
		log.Fatalf("Failed to read image file: %v", err)
	}

	return base64.StdEncoding.EncodeToString(file)
}

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
	redSquare := loadImageAsBase64("internal/testutil/images/red.png")
	return &models.FeedIcon{
		FeedID: 1,
		Data:   "image/png;base64," + redSquare,
	}
}

func NewMockHTMLTemplateData() *models.HTMLTemplateData {
	redSquare := loadImageAsBase64("internal/testutil/images/red.png")
	yellowSquare := loadImageAsBase64("internal/testutil/images/yellow.png")
	greenSquare := loadImageAsBase64("internal/testutil/images/green.png")

	return digest.NewDigestService().BuildDigestData(
		NewMockCategory(),
		&miniflux.Entries{
			{
				ID:          1,
				UserID:      1,
				FeedID:      1,
				Status:      miniflux.EntryStatusUnread,
				Title:       "A Short and Sweet Title",
				URL:         "https://example.com/1",
				Date:        time.Now().Add(-1 * time.Hour),
				Content:     "This is a short and sweet entry.",
				Author:      "Test Author 1",
				Feed:        NewMockFeed(),
			},
			{
				ID:          2,
				FeedID:      2,
				Title:       "A Longer Entry with a Paragraph of Text",
				URL:         "https://example.com/2",
				Date:        time.Now().Add(-3 * time.Hour),
				Content:     "This is a longer entry that contains a full paragraph of text. It is meant to simulate a more realistic entry that a user might encounter in their feed. It has enough text to wrap to multiple lines and give a good sense of how the layout will look with a more substantial amount of content.",
				Author:      "Test Author 2",
				Feed:        NewMockFeed(),
			},
			{
				ID:          3,
				FeedID:      3,
				Title:       "An Entry with HTML Content",
				URL:         "https://example.com/3",
				Date:        time.Now().Add(-4 * time.Hour),
				Content:     "<h1>This is a heading</h1><p>This is a paragraph with <strong>strong</strong> text and a <a href=\"https://example.com\">link</a>.</p><ul><li>This is a list item</li><li>This is another list item</li></ul>",
				Feed:        NewMockFeed(),
			},
			{
				ID:          4,
				UserID:      1,
				FeedID:      1,
				Status:      miniflux.EntryStatusUnread,
				Title:       "Another Entry - Day 2",
				URL:         "https://example.com/4",
				Date:        time.Now().AddDate(0, 0, -1), // One day earlier
				Content:     "This entry is from a different day.",
				Author:      "Test Author 4",
				Feed:        NewMockFeed(),
			},
			{
				ID:          5,
				UserID:      1,
				FeedID:      2,
				Status:      miniflux.EntryStatusUnread,
				Title:       "Fifth Entry - Day 2",
				URL:         "https://example.com/5",
				Date:        time.Now().AddDate(0, 0, -1).Add(-2 * time.Hour), // One day earlier, different time
				Content:     "This is the fifth entry, also from day 2.",
				Author:      "Test Author 5",
				Feed:        NewMockFeed(),
			},
			{
				ID:          6,
				UserID:      1,
				FeedID:      3,
				Status:      miniflux.EntryStatusUnread,
				Title:       "Sixth Entry - Day 2",
				URL:         "https://example.com/6",
				Date:        time.Now().AddDate(0, 0, -1).Add(-5 * time.Hour), // One day earlier, different time
				Content:     "This is the sixth entry, also from day 2.",
				Author:      "Test Author 6",
				Feed:        NewMockFeed(),
			},
		},
		map[int64]*models.FeedIcon{
			1: {FeedID: 1, Data: "image/png;base64," + redSquare},
			2: {FeedID: 2, Data: "image/png;base64," + yellowSquare},
			3: {FeedID: 3, Data: "image/png;base64," + greenSquare},
		},
	)
}