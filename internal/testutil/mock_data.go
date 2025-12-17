package testutil

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"miniflux-digest/internal/models"
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

func NewMockFeedIconRed() *models.FeedIcon {
	icon := loadImageAsBase64("internal/testutil/images/red.png")
	return &models.FeedIcon{
		FeedID: 1,
		Data:   "image/png;base64," + icon,
	}
}

func NewMockFeedIconYellow() *models.FeedIcon {
	icon := loadImageAsBase64("internal/testutil/images/yellow.png")
	return &models.FeedIcon{
		FeedID: 2,
		Data:   "image/png;base64," + icon,
	}
}

func NewMockFeedIconGreen() *models.FeedIcon {
	icon := loadImageAsBase64("internal/testutil/images/green.png")
	return &models.FeedIcon{
		FeedID: 3,
		Data:   "image/png;base64," + icon,
	}
}

func MockNumEntries(n int) []*models.Entry {
	return CreateMockEntries(n)
}

func NewMockEntries() []*models.Entry {
	return CreateMockEntries(20)
}

func NewMockFeedIcons() []*models.FeedIcon {
	return []*models.FeedIcon{
		NewMockFeedIconRed(),
		NewMockFeedIconYellow(),
		NewMockFeedIconGreen(),
	}
}

func NewMockEntryGroup() *models.EntryGroup {
	return &models.EntryGroup{
		Title:   "Test Group",
		Summary: "This is a test group summary.",
		Entries: CreateMockEntries(5),
		Slug:    "test-group",
	}
}

type feedTemplate struct {
	FeedID     int64
	FeedTitle  string
	GroupID    int64
	GroupTitle string
}

func createHardcodedMockEntries() []*models.Entry {
	return []*models.Entry{
		// Category A
		{ID: 1, Title: "Entry 1", Date: time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC), FeedID: 100, FeedTitle: "Feed A", GroupID: 1, GroupTitle: "Category A", SiteURL: "www.example.com", URL: "www.example.com/1"},
		{ID: 3, Title: "Entry 3", Date: time.Date(2024, 1, 2, 11, 0, 0, 0, time.UTC), FeedID: 100, FeedTitle: "Feed A", GroupID: 1, GroupTitle: "Category A", SiteURL: "www.example2.com", URL: "www.example2.com/1"},
		{ID: 5, Title: "Entry 5", Date: time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC), FeedID: 300, FeedTitle: "Feed C", GroupID: 1, GroupTitle: "Category A", SiteURL: "www.example3.com", URL: "www.example3.com/1"},
		{ID: 6, Title: "Entry 6", Date: time.Date(2024, 1, 2, 13, 0, 0, 0, time.UTC), FeedID: 400, FeedTitle: "Feed D", GroupID: 1, GroupTitle: "Category A", SiteURL: "www.example4.com", URL: "www.example4.com/1"},

		// Category B
		{ID: 2, Title: "Entry 2", Date: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), FeedID: 200, FeedTitle: "Feed B", GroupID: 2, GroupTitle: "Category B", SiteURL: "www.example5.com", URL: "www.example.com/1"},
		{ID: 4, Title: "Entry 4", Content: "Content of entry 4", Date: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), FeedID: 200, FeedTitle: "Feed B", GroupID: 2, GroupTitle: "Category B", SiteURL: "www.example5.com", URL: "www.example5.com/2"},
		{ID: 7, Title: "Entry 7", Date: time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC), FeedID: 500, FeedTitle: "Feed E", GroupID: 2, GroupTitle: "Category B", SiteURL: "www.example5.com", URL: "www.example5.com/3"},

		// Category C
		{ID: 8, Title: "Entry 8", Date: time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC), FeedID: 600, FeedTitle: "Feed F", GroupID: 3, GroupTitle: "Category C", SiteURL: "www.example6.com", URL: "www.example7.com/1"},
		{ID: 9, Title: "Entry 9", Date: time.Date(2024, 1, 3, 11, 0, 0, 0, time.UTC), FeedID: 700, FeedTitle: "Feed G", GroupID: 3, GroupTitle: "Category C", SiteURL: "www.example7.com", URL: "www.example7.com/1"},
	}
}

func CreateMockEntries(n int) []*models.Entry {
	if n == 0 {
		return createHardcodedMockEntries()
	}

	titleTemplates := []string{
		"Short title",
		"This is a much longer title to test how the UI handles overflow",
		"Another title",
	}
	contentTemplates := []string{
		"This is short content.",
		"",
		"This is a long paragraph of content to test how the UI handles overflow. It should wrap to multiple lines and give a good sense of how the layout will look with a more substantial amount of content.",
		"<h1>HTML Content</h1><p>This is a paragraph with <strong>strong</strong> text and a <a href=\"https://example.com\">link</a>.</p><ul><li>This is a list item</li><li>This is another list item</li></ul>",
		"<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.</p><p>Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.</p><p>Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo.</p><p>Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt.</p><p>Neque porro quisquam est, qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit, sed quia non numquam eius modi tempora incidunt ut labore et dolore magnam aliquam quaerat voluptatem.</p><pre>This is a single, very long line of text inside a pre block that should cause the container to overflow horizontally if the CSS is not correctly handling it. This_is_a_very_long_word_that_will_not_break_and_should_force_a_horizontal_scrollbar. a_b_c_d_e_f_g_h_i_j_k_l_m_n_o_p_q_r_s_t_u_v_w_x_y_z_a_b_c_d_e_f_g_h_i_j_k_l_m_n_o_p_q_r_s_t_u_v_w_x_y_z_a_b_c_d_e_f_g_h_i_j_k_l_m_n_o_p_q_r_s_t_u_v_w_x_y_z</pre>",
	}

	templates := []feedTemplate{
		{FeedID: 1, FeedTitle: "Feed A", GroupID: 1, GroupTitle: "Category A"},
		{FeedID: 2, FeedTitle: "Feed B", GroupID: 1, GroupTitle: "Category A"},
		{FeedID: 3, FeedTitle: "Feed C", GroupID: 2, GroupTitle: "Category B"},
		{FeedID: 4, FeedTitle: "Feed D", GroupID: 2, GroupTitle: "Category B"},
		{FeedID: 5, FeedTitle: "Feed E", GroupID: 2, GroupTitle: "Category B"},
		{FeedID: 6, FeedTitle: "Feed F", GroupID: 3, GroupTitle: "Category C"},
		{FeedID: 7, FeedTitle: "Feed G", GroupID: 3, GroupTitle: "Category C"},
	}

	entries := make([]*models.Entry, 0, n)
	for i := 1; i <= n; i++ {
		template := templates[(i-1)%len(templates)]
		title := titleTemplates[(i-1)%len(titleTemplates)]
		content := contentTemplates[(i-1)%len(contentTemplates)]
		var commentsURL string
		if i%2 == 0 {
			commentsURL = fmt.Sprintf("https://example.com/comments/%d", i)
		}

		siteURL := fmt.Sprintf("https://www.example%d.com", template.FeedID)
		entryURL := fmt.Sprintf("%s/item/%d", siteURL, i)

		entry := &models.Entry{
			ID:          int64(i),
			Title:       fmt.Sprintf("Entry %d - %s", i, title),
			URL:         entryURL,
			CommentsURL: commentsURL,
			Date:        time.Now().Add(time.Duration(-i) * time.Hour),
			Content:     content,
			FeedID:      template.FeedID,
			FeedTitle:   template.FeedTitle,
			GroupID:     template.GroupID,
			GroupTitle:  template.GroupTitle,
			SiteURL:     siteURL,
		}
		entries = append(entries, entry)
	}
	return entries
}
