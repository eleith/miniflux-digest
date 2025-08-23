package templates

import (
	"bytes"
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/testutil"
	"testing"
	"time"
)

func TestTemplates(t *testing.T) {
	if ArchiveTemplate == nil {
		t.Error("ArchiveTemplate should not be nil")
	}

	if EmailTemplate == nil {
		t.Error("EmailTemplate should not be nil")
	}
}

func TestArchiveTemplateExecution(t *testing.T) {
	mockSubGroup := &models.EntryGroup{
		Title:   "Test SubGroup Title",
		Summary: "Test SubGroup Summary",
		Entries: testutil.NewMockEntries(),
		Slug:    "test-subgroup-title",
	}

	mockPrimaryGroup := &models.PrimaryGroupDigestData{
		Title:     "Test Primary Group Title",
		Slug:      "test-primary-group-title",
		SubGroups: []*models.EntryGroup{mockSubGroup},
		Summary:   "Test Primary Group Summary",
	}

	data := models.GroupedDigestPageData{
		PrimaryGroup: mockPrimaryGroup,
		FeedIcons:    testutil.NewMockFeedIcons(),
		MinifluxHost: "http://localhost:8080",
	}

	var buf bytes.Buffer
	err := ArchiveTemplate.Execute(&buf, data)
	if err != nil {
		t.Errorf("Failed to execute ArchiveTemplate: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("ArchiveTemplate execution resulted in empty output")
	}
}

func TestEmailTemplateExecution(t *testing.T) {
	data := models.OverviewTemplateData{
		Entries: testutil.NewMockEntries(),
		FeedIcons: testutil.NewMockFeedIcons(),
	}
	textData := &EmailTemplateData{
		OverviewTemplateData: data,
		URL:          "https://example.com",
	}
	var buf bytes.Buffer
	err := EmailTemplate.Execute(&buf, textData)
	if err != nil {
		t.Errorf("Failed to execute EmailTemplate: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("EmailTemplate execution resulted in empty output")
	}
}

func TestOverviewTemplateExecution(t *testing.T) {
	data := models.OverviewTemplateData{
		Entries: testutil.NewMockEntries(),
		GeneratedDate: time.Now(),
		FeedIcons: testutil.NewMockFeedIcons(),
		OverviewSummary: "This is a test summary.",
		TotalEntries: 5,
		TotalFeeds: 2,
		MinifluxHost: "http://localhost:8080",
		PrimaryGroups: []*models.PrimaryGroupDigestData{
			{
				Title: "Category A",
				Slug: "category-a",
				Summary: "Summary for Category A",
				SubGroups: []*models.EntryGroup{
					{
						Title: "Feed 1",
						Slug: "feed-1",
						Entries: testutil.NewMockEntries(),
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := OverviewTemplate.Execute(&buf, data)
	if err != nil {
		t.Errorf("Failed to execute OverviewTemplate: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("OverviewTemplate execution resulted in empty output")
	}
}

func TestTextToHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "text with newline",
			input:    "Line 1\nLine 2",
			expected: "Line 1<br>Line 2",
		},
		{
			name:     "text with multiple newlines",
			input:    "Line 1\n\nLine 3",
			expected: "Line 1<br><br>Line 3",
		},
		{
			name:     "text with HTML special characters",
			input:    "<script>alert('hi')</script>",
			expected: "&lt;script&gt;alert(&#39;hi&#39;)&lt;/script&gt;",
		},
		{
			name:     "text with newlines and HTML special characters",
			input:    "Line 1\n<p>Line 2</p>",
			expected: "Line 1<br>&lt;p&gt;Line 2&lt;/p&gt;",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := textToHTML(tt.input)
			if string(result) != tt.expected {
				t.Errorf("textToHTML(%q): got %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
