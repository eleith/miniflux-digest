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
	data := models.HTMLTemplateData{
		Category: testutil.NewMockCategory(),
		Entries: testutil.NewMockEntries(),
		FeedIcons: testutil.NewMockFeedIcons(),
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
	data := models.HTMLTemplateData{
		Category: testutil.NewMockCategory(),
		Entries: testutil.NewMockEntries(),
		FeedIcons: testutil.NewMockFeedIcons(),
	}
	textData := &EmailTemplateData{
		HTMLTemplateData: data,
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
