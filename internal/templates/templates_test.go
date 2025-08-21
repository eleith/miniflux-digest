package templates

import (
	"bytes"
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/testutil"
	"testing"
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
	// Create a mock EntryGroup (sub-group)
	mockSubGroup := &models.EntryGroup{
		Title:   "Test SubGroup Title",
		Summary: "Test SubGroup Summary",
		Entries: testutil.NewMockEntries(),
		Slug:    "test-subgroup-title",
	}

	// Create PrimaryGroupDigestData
	mockPrimaryGroup := &models.PrimaryGroupDigestData{
		Title:     "Test Primary Group Title",
		Slug:      "test-primary-group-title",
		SubGroups: []*models.EntryGroup{mockSubGroup},
		Summary:   "Test Primary Group Summary",
	}

	// Create GroupedDigestPageData
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
