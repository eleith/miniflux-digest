package testutil

import (
	"encoding/base64"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

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

func NewMockFeedRed() *miniflux.Feed {
	return &miniflux.Feed{
		ID:    1,
		Title: "Tech News",
	}
}

func NewMockFeedYellow() *miniflux.Feed {
	return &miniflux.Feed{
		ID:    2,
		Title: "The Daily Bugle - A Very Long Feed Name to Test Overflow",
	}
}

func NewMockFeedGreen() *miniflux.Feed {
	return &miniflux.Feed{
		ID:    3,
		Title: "Comics",
	}
}

func NewMockEntry() *miniflux.Entry {
	return &miniflux.Entry{
		ID:      1,
		Title:   "Test Entry",
		Content: "<p>Test Content</p>",
		Feed:    NewMockFeedRed(),
	}
}

func NewMockFeedIconRed() *models.FeedIcon {
	feed := NewMockFeedRed()
	icon := loadImageAsBase64("internal/testutil/images/red.png")
	return &models.FeedIcon{
		FeedID: feed.ID,
		Data:   "image/png;base64," + icon,
	}
}

func NewMockFeedIconYellow() *models.FeedIcon {
	feed := NewMockFeedYellow()
	icon := loadImageAsBase64("internal/testutil/images/yellow.png")
	return &models.FeedIcon{
		FeedID: feed.ID,
		Data:   "image/png;base64," + icon,
	}
}

func NewMockFeedIconGreen() *models.FeedIcon {
	feed := NewMockFeedGreen()
	icon := loadImageAsBase64("internal/testutil/images/green.png")
	return &models.FeedIcon{
		FeedID: feed.ID,
		Data:   "image/png;base64," + icon,
	}
}

func NewMockEntry1() *miniflux.Entry {
	return &miniflux.Entry{
		ID:          1,
		UserID:      1,
		FeedID:      1,
		Status:      miniflux.EntryStatusUnread,
		Title:       "A Short and Sweet Title",
		URL:         "https://example.com/1",
		Date:        time.Now().Add(-1 * time.Hour),
		Content:     "This is a short and sweet entry.",
		Author:      "Test Author 1",
		Feed:        NewMockFeedRed(),
	}
}

func NewMockEntry2() *miniflux.Entry {
	return &miniflux.Entry{
		ID:          2,
		FeedID:      2,
		Title:       "A Longer Entry with a Paragraph of Text",
		URL:         "https://example.com/2",
		Date:        time.Now().Add(-3 * time.Hour),
		Content:     "This is a longer entry that contains a full paragraph of text. It is meant to simulate a more realistic entry that a user might encounter in their feed. It has enough text to wrap to multiple lines and give a good sense of how the layout will look with a more substantial amount of content.",
		Author:      "Test Author 2",
		Feed:        NewMockFeedYellow(),
	}
}

func NewMockEntry3() *miniflux.Entry {
	return &miniflux.Entry{
		ID:          3,
		FeedID:      3,
		Title:       "An Entry with HTML Content",
		URL:         "https://example.com/3",
		Date:        time.Now().Add(-4 * time.Hour),
		Content:     "<h1>This is a heading</h1><p>This is a paragraph with <strong>strong</strong> text and a <a href=\"https://example.com\">link</a>.</p><ul><li>This is a list item</li><li>This is another list item</li></ul>",
		Feed:        NewMockFeedGreen(),
	}
}

func NewMockEntry4() *miniflux.Entry {
	return &miniflux.Entry{
		ID:          4,
		UserID:      1,
		FeedID:      1,
		Status:      miniflux.EntryStatusUnread,
		Title:       "Another Entry - Day 2",
		URL:         "https://example.com/4",
		Date:        time.Now().AddDate(0, 0, -1), // One day earlier
		Content:     "This entry is from a different day.",
		Author:      "Test Author 4",
		Feed:        NewMockFeedRed(),
	}
}

func NewMockEntry5() *miniflux.Entry {
	return &miniflux.Entry{
		ID:          5,
		UserID:      1,
		FeedID:      2,
		Status:      miniflux.EntryStatusUnread,
		Title:       "Fifth Entry - Day 2",
		URL:         "https://example.com/5",
		Date:        time.Now().AddDate(0, 0, -1).Add(-2 * time.Hour), // One day earlier, different time
		Content:     "This is the fifth entry, also from day 2.",
		Author:      "Test Author 5",
		Feed:        NewMockFeedYellow(),
	}
}

func NewMockEntry6() *miniflux.Entry {
	return &miniflux.Entry{
		ID:          6,
		UserID:      1,
		FeedID:      3,
		Status:      miniflux.EntryStatusUnread,
		Title:       "Sixth Entry - Day 2",
		URL:         "https://example.com/6",
		Date:        time.Now().AddDate(0, 0, -1).Add(-5 * time.Hour), // One day earlier, different time
		Content:     "This is the sixth entry, also from day 2.",
		Author:      "Test Author 6",
		Feed:        NewMockFeedGreen(),
	}
}

func NewMockEntry7() *miniflux.Entry {
	return &miniflux.Entry{
		ID:      7,
		FeedID:  1,
		Title:   "Short and Sweet",
		URL:     "https://example.com/7",
		Date:    time.Now().Add(-2 * time.Hour),
		Content: "Just a little something.",
		Author:  "Test Author 7",
		Feed:    NewMockFeedRed(),
	}
}

func NewMockEntry8() *miniflux.Entry {
	return &miniflux.Entry{
		ID:          8,
		FeedID:      2,
		Title:       "This is a very long title to test how the UI handles overflow and wrapping of text content in the entry header",
		URL:         "https://example.com/8",
		CommentsURL: "https://example.com/8/comments",
		Date:        time.Now().Add(-5 * time.Hour),
		Content:     "This entry has a particularly long title to stress test the layout. It also has comments enabled. The content itself is also quite long, providing a good example of a substantial post that might require scrolling within its own container, depending on the UI design. We want to see how the navigation bar at the bottom behaves with this much content.",
		Author:      "Test Author 8",
		Feed:        NewMockFeedYellow(),
	}
}

func NewMockEntry9() *miniflux.Entry {
	return &miniflux.Entry{
		ID:      9,
		FeedID:  3,
		Title:   "HTML Content Test",
		URL:     "https://example.com/9",
		Date:    time.Now().Add(-6 * time.Hour),
		Content: "<h2>HTML Test</h2><p>This entry includes <code>HTML</code> tags to verify rendering.<ul><li>Item 1</li><li>Item 2</li></ul></p>",
		Author:  "Test Author 9",
		Feed:    NewMockFeedGreen(),
	}
}

func NewMockEntry10() *miniflux.Entry {
	return &miniflux.Entry{
		ID:          10,
		FeedID:      1,
		Title:       "Empty Content Entry",
		URL:         "https://example.com/10",
		CommentsURL: "https://example.com/10/comments",
		Date:        time.Now().Add(-7 * time.Hour),
		Content:     "",
		Author:      "Test Author 10",
		Feed:        NewMockFeedRed(),
	}
}

func NewMockEntry11() *miniflux.Entry {
	return &miniflux.Entry{
		ID:      11,
		FeedID:  2,
		Title:   "Another Very Long Title That Just Keeps Going And Going To See What Happens",
		URL:     "https://example.com/11",
		Date:    time.Now().Add(-8 * time.Hour),
		Content: "Short content, long title.",
		Author:  "Test Author 11",
		Feed:    NewMockFeedYellow(),
	}
}

func NewMockEntry12() *miniflux.Entry {
	return &miniflux.Entry{
		ID:          12,
		FeedID:      3,
		Title:       "Short with Comments",
		URL:         "https://example.com/12",
		CommentsURL: "https://example.com/12/comments",
		Date:        time.Now().Add(-9 * time.Hour),
		Content:     "A brief entry that has comments.",
		Author:      "Test Author 12",
		Feed:        NewMockFeedGreen(),
	}
}

func NewMockEntry13() *miniflux.Entry {
	return &miniflux.Entry{
		ID:      13,
		FeedID:  1,
		Title:   "Plain Text, Long Content",
		URL:     "https://example.com/13",
		Date:    time.Now().Add(-10 * time.Hour),
		Content: "This is a long entry with only plain text content. No HTML tags are included. This is to test the wrapping and scrolling of plain text. It should be long enough to require scrolling on most screens. We need to ensure that the spacing of the bottom navigation bar is correct for this type of content.",
		Author:  "Test Author 13",
		Feed:    NewMockFeedRed(),
	}
}

func NewMockEntry14() *miniflux.Entry {
	return &miniflux.Entry{
		ID:          14,
		FeedID:      2,
		Title:       "HTML and Comments",
		URL:         "https://example.com/14",
		CommentsURL: "https://example.com/14/comments",
		Date:        time.Now().Add(-11 * time.Hour),
		Content:     "<h1>Heading</h1><p>This entry has both HTML content and comments. It's a common combination.</p>",
		Author:      "Test Author 14",
		Feed:        NewMockFeedYellow(),
	}
}

func NewMockEntry15() *miniflux.Entry {
	return &miniflux.Entry{
		ID:      15,
		FeedID:  3,
		Title:   "Short, Empty, No Comments",
		URL:     "https://example.com/15",
		Date:    time.Now().Add(-12 * time.Hour),
		Content: "",
		Author:  "Test Author 15",
		Feed:    NewMockFeedGreen(),
	}
}

func NewMockEntry16() *miniflux.Entry {
	return &miniflux.Entry{
		ID:      16,
		FeedID:  1,
		Title:   "A Very Long Title for an Entry with Short Content",
		URL:     "https://example.com/16",
		Date:    time.Now().Add(-13 * time.Hour),
		Content: "The title is long, the content is not.",
		Author:  "Test Author 16",
		Feed:    NewMockFeedRed(),
	}
}

func NewMockEntry17() *miniflux.Entry {
	return &miniflux.Entry{
		ID:          17,
		FeedID:      2,
		Title:       "Comments and Long Content",
		URL:         "https://example.com/17",
		CommentsURL: "https://example.com/17/comments",
		Date:        time.Now().Add(-14 * time.Hour),
		Content:     "This entry has a lot of content to read through, and it also has comments. This is a good test case for scrolling and making sure the bottom navigation bar is not obscured by the content. The content is intentionally verbose to simulate a real-world article or blog post.",
		Author:      "Test Author 17",
		Feed:        NewMockFeedYellow(),
	}
}

func NewMockEntry18() *miniflux.Entry {
	return &miniflux.Entry{
		ID:      18,
		FeedID:  3,
		Title:   "HTML, No Comments",
		URL:     "https://example.com/18",
		Date:    time.Now().Add(-15 * time.Hour),
		Content: "<b>Bold text</b> and <i>italic text</i> but no comments.",
		Author:  "Test Author 18",
		Feed:    NewMockFeedGreen(),
	}
}

func NewMockEntry19() *miniflux.Entry {
	return &miniflux.Entry{
		ID:      19,
		FeedID:  1,
		Title:   "Empty, No Comments",
		URL:     "https://example.com/19",
		Date:    time.Now().Add(-16 * time.Hour),
		Content: "",
		Author:  "Test Author 19",
		Feed:    NewMockFeedRed(),
	}
}

func NewMockEntry20() *miniflux.Entry {
	return &miniflux.Entry{
		ID:          20,
		FeedID:      2,
		Title:       "The Final Entry: A Very Long Title for a Very Long Entry with Comments",
		URL:         "https://example.com/20",
		CommentsURL: "https://example.com/20/comments",
		Date:        time.Now().Add(-17 * time.Hour),
		Content:     "This is the final mock entry. It has a very long title, a lot of content, and comments. It's the ultimate test case for the UI. We want to make sure that everything looks good and functions correctly with this entry. The content is long enough to require scrolling, and the title is long enough to test wrapping. The comments URL is also present. This entry should help identify any remaining layout issues.",
		Author:      "Test Author 20",
		Feed:        NewMockFeedYellow(),
	}
}

func NewMockEntries() *miniflux.Entries {
	return &miniflux.Entries{
		NewMockEntry1(),
		NewMockEntry2(),
		NewMockEntry3(),
		NewMockEntry4(),
		NewMockEntry5(),
		NewMockEntry6(),
		NewMockEntry7(),
		NewMockEntry8(),
		NewMockEntry9(),
		NewMockEntry10(),
		NewMockEntry11(),
		NewMockEntry12(),
		NewMockEntry13(),
		NewMockEntry14(),
		NewMockEntry15(),
		NewMockEntry16(),
		NewMockEntry17(),
		NewMockEntry18(),
		NewMockEntry19(),
		NewMockEntry20(),
	}
}

func NewMockFeedIcons() []*models.FeedIcon {
	return []*models.FeedIcon{
		NewMockFeedIconRed(),
		NewMockFeedIconYellow(),
		NewMockFeedIconGreen(),
	}
}