package testutil

import (
	"encoding/base64"
	"log"
	"miniflux-digest/internal/category"
	"os"
	"path/filepath"
	"runtime"
	"time"

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

func NewMockCategoryData() *category.CategoryData {
	redSquare := loadImageAsBase64("internal/testutil/images/red.png")
	yellowSquare := loadImageAsBase64("internal/testutil/images/yellow.png")
	greenSquare := loadImageAsBase64("internal/testutil/images/green.png")

	return &category.CategoryData{
		Category: &miniflux.Category{
			ID:	1,
			Title: "Test Category",
		},
		Entries: &miniflux.Entries{
			{
				ID:	1,
				UserID:	1,
				FeedID:	1,
				Status:	miniflux.EntryStatusUnread,
				Hash:	"test-hash-1",
				Title:	"A Short and Sweet Title",
				URL:	"https://example.com/1",
				CommentsURL: "https://example.com/1/comments",
				Date:	time.Now().Add(-1 * time.Hour),
				CreatedAt:	time.Now().Add(-2 * time.Hour),
				Content:	"This is a short and sweet entry.",
				Author:	"Test Author 1",
				ShareCode:	"test-share-code-1",
				Starred:	true,
				ReadingTime: 1,
				Enclosures:	[]*miniflux.Enclosure{{URL: "https://example.com/image.jpg", MimeType: "image/jpeg"}},
				Feed:	&miniflux.Feed{ID: 1, Title: "Test Feed 1"},
			},
			{
				ID:	2,
				FeedID: 2,
				Title:	"A Longer Entry with a Paragraph of Text",
				URL:	"https://example.com/2",
				Date:	time.Now().Add(-3 * time.Hour),
				Content:	"This is a longer entry that contains a full paragraph of text. It is meant to simulate a more realistic entry that a user might encounter in their feed. It has enough text to wrap to multiple lines and give a good sense of how the layout will look with a more substantial amount of content.",
				Author:	"Test Author 2",
				Feed:	&miniflux.Feed{ID: 2, Title: "Test Feed 2"},
			},
			{
				ID:	3,
				FeedID: 3,
				Title:	"An Entry with HTML Content",
				URL:	"https://example.com/3",
				Date:	time.Now().Add(-4 * time.Hour),
				Content:	"<h1>This is a heading</h1><p>This is a paragraph with <strong>strong</strong> text and a <a href=\"https://example.com\">link</a>.</p><ul><li>This is a list item</li><li>This is another list item</li></ul>",
				Feed:	&miniflux.Feed{ID: 3, Title: "Test Feed 3"},
			},
		},
		GeneratedDate: time.Date(2025, 7, 21, 12, 0, 0, 0, time.UTC),
		FeedIcons: []*category.FeedIcon{
			{FeedID: 1, Data: "image/png;base64," + redSquare},
			{FeedID: 2, Data: "image/png;base64," + yellowSquare},
			{FeedID: 3, Data: "image/png;base64," + greenSquare},
		},
	}
}