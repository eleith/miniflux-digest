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
	// Create a mock EntryGroup
	mockEntryGroup := &models.EntryGroup{
		Title:   "Test Group Title",
		Summary: "Test Group Summary",
		Entries: *testutil.NewMockEntries(),
	}

	// Create GroupedEntriesTemplateData
	data := models.GroupedEntriesTemplateData{
		EntryGroup:    mockEntryGroup,
		GeneratedDate: time.Now(),
		FeedIcons:     testutil.NewMockFeedIcons(),
		MinifluxHost:  "http://localhost:8080",
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
		Category: testutil.NewMockCategory(),
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
