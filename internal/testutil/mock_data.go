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
	entries := make([]*models.Entry, 0, n)
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

	for i := 1; i <= n; i++ {
		var feedTitle string
		var groupTitle string
		switch (i - 1) % 4 {
		case 0:
			feedTitle = "Tech News"
			groupTitle = "Technology"
		case 1:
			feedTitle = "The Daily Bugle - A Very Long Feed Name to Test Overflow"
			groupTitle = "News"
		case 2:
			feedTitle = "Comics"
			groupTitle = "Comics"
		case 3:
			feedTitle = "Uncategorized Feed"
			groupTitle = "Uncategorized" // Or empty string if truly uncategorized
		}

		title := titleTemplates[(i-1)%len(titleTemplates)]
		content := contentTemplates[(i-1)%len(contentTemplates)]

		entry := &models.Entry{
			ID:            int64(i),
			Title:         fmt.Sprintf("Entry %d - %s", i, title),
			URL:           fmt.Sprintf("https://example.com/%d", i),
			Date:          time.Now().Add(time.Duration(-i) * time.Hour),
			Content:       content,
			FeedID:        int64((i-1)%4 + 1), // Assign a dummy FeedID
			FeedTitle:     feedTitle,
			GroupTitle:    groupTitle,
		}
		entries = append(entries, entry)
	}
	return entries
}

func NewMockEntries() []*models.Entry {
	return MockNumEntries(20)
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
		Entries: MockNumEntries(5),
		Slug:    "test-group",
	}
}
