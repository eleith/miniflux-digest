package archive

import (
	"io"
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/templates"
	"miniflux-digest/internal/testutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetHTML(t *testing.T) {
	archiveService := NewArchiveService(t.TempDir(), templates.OverviewTemplate, templates.ArchiveTemplate)

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

	html, err := archiveService.getHTML(templates.ArchiveTemplate, &data, true)
	if err != nil {
		t.Fatalf("getHTML failed: %v", err)
	}
	if len(html) == 0 {
		t.Error("Expected HTML to be non-empty")
	}
}

func TestMakeArchiveFile(t *testing.T) {
	tempDir := t.TempDir()
	archiveService := NewArchiveService(tempDir, templates.OverviewTemplate, templates.ArchiveTemplate)

	mockSubGroup := &models.EntryGroup{
		Title:   "Test SubGroup Title",
		Summary: "Test SubGroup Summary",
		Entries: testutil.NewMockEntries(),
		Slug:    "test-subgroup-title",
	}

	data := models.PrimaryGroupDigestData{
		Title:     "Test Primary Group Title",
		Slug:      "test-primary-group-title",
		SubGroups: []*models.EntryGroup{mockSubGroup},
		Summary:   "Test Primary Group Summary",
	}

	dateFolderPath := filepath.Join(tempDir, time.Now().Format("2006-01-02"))
	if err := os.MkdirAll(dateFolderPath, 0755); err != nil {
		t.Fatalf("Failed to create date folder for test: %v", err)
	}

	file, err := archiveService.makeGroupedEntriesArchiveFile(&data, dateFolderPath)
	if err != nil {
		t.Fatalf("makeArchiveFile failed: %v", err)
	}
	if file == nil {
		t.Fatal("Expected file to be non-nil")
	}
	expectedPath := filepath.Join(dateFolderPath, "digests", data.Slug+".html")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist", expectedPath)
	}
}

func TestCleanArchive(t *testing.T) {
	tempDir := t.TempDir()
	archiveService := NewArchiveService(tempDir, templates.OverviewTemplate, templates.ArchiveTemplate)

	oldDirName := time.Now().Add(-48 * time.Hour).Format("2006-01-02")
	oldDirPath := filepath.Join(tempDir, oldDirName)
	if err := os.MkdirAll(oldDirPath, 0755); err != nil {
		t.Fatalf("Failed to create old test directory: %v", err)
	}
	oldFilePath := filepath.Join(oldDirPath, "test.html")
	if err := os.WriteFile(oldFilePath, []byte("old"), 0644); err != nil {
		t.Fatalf("Failed to create old file: %v", err)
	}
	twoDaysAgo := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(oldFilePath, twoDaysAgo, twoDaysAgo); err != nil {
		t.Fatalf("Failed to change file modification time: %v", err)
	}

	newDirName := time.Now().Format("2006-01-02")
	newDirPath := filepath.Join(tempDir, newDirName)
	if err := os.MkdirAll(newDirPath, 0755); err != nil {
		t.Fatalf("Failed to create new test directory: %v", err)
	}
	newFilePath := filepath.Join(newDirPath, "test.html")
	if err := os.WriteFile(newFilePath, []byte("new"), 0644); err != nil {
		t.Fatalf("Failed to create new file: %v", err)
	}

	archiveService.CleanArchive(24 * time.Hour)

	if _, err := os.Stat(oldDirPath); !os.IsNotExist(err) {
		t.Error("Expected old directory to be deleted")
	}

	if _, err := os.Stat(newDirPath); os.IsNotExist(err) {
		t.Error("Expected new directory to be kept")
	}
}

type mockHTMLTemplate struct {
	executeFunc func(wr io.Writer, data any) error
}

func (m *mockHTMLTemplate) Execute(wr io.Writer, data any) error {
	return m.executeFunc(wr, data)
}

func TestMakeOverviewArchiveFile(t *testing.T) {
	tempDir := t.TempDir()
	archiveService := NewArchiveService(tempDir, &mockHTMLTemplate{}, &mockHTMLTemplate{})

	data := &models.OverviewTemplateData{
		GeneratedDate: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
	}

	file, dateFolderPath, err := archiveService.makeOverviewArchiveFile(data)
	if err != nil {
		t.Fatalf("makeOverviewArchiveFile failed: %v", err)
	}
	if file == nil {
		t.Fatal("Expected file to be non-nil")
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("Error closing file: %v", err)
		}
	}()

	expectedPath := filepath.Join(tempDir, "2024-01-01", "index.html")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist", expectedPath)
	}
	if dateFolderPath != filepath.Join(tempDir, "2024-01-01") {
		t.Errorf("Expected date folder path %s, got %s", filepath.Join(tempDir, "2024-01-01"), dateFolderPath)
	}
}

func TestMakeArchiveHTML_Success(t *testing.T) {
	tempDir := t.TempDir()

	mockOverviewTemplate := &mockHTMLTemplate{
		executeFunc: func(wr io.Writer, data any) error {
			_, err := wr.Write([]byte("overview html"))
			return err
		},
	}
	mockArchiveTemplate := &mockHTMLTemplate{
		executeFunc: func(wr io.Writer, data any) error {
			d, ok := data.(*models.GroupedDigestPageData)
			if !ok {
				t.Fatal("unexpected data type for archive template")
			}
			if len(d.FeedIcons) == 0 {
				t.Error("expected feed icons to be present in archive template data")
			}
			_, err := wr.Write([]byte("grouped html"))
			return err
		},
	}
	archiveService := NewArchiveService(tempDir, mockArchiveTemplate, mockOverviewTemplate)

	data := &models.OverviewTemplateData{
		GeneratedDate: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
		PrimaryGroups: []*models.PrimaryGroupDigestData{
			{
				Slug: "group-1",
				SubGroups: []*models.EntryGroup{{Entries: []*models.Entry{{ID: 1, FeedID: 1}}}},
			},
			{
				Slug: "group-2",
				SubGroups: []*models.EntryGroup{{Entries: []*models.Entry{{ID: 2, FeedID: 2}}}},
			},
		},
		FeedIcons: []*models.FeedIcon{
			{FeedID: 1, Data: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII="},
			{FeedID: 2, Data: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII="},
		},
	}

	overviewFile, groupedEntryFiles, err := archiveService.MakeArchiveHTML(data, false)
	if err != nil {
		t.Fatalf("MakeArchiveHTML failed: %v", err)
	}
	if overviewFile == nil {
		t.Fatal("Expected overviewFile to be non-nil")
	}
	if len(groupedEntryFiles) != 2 {
		t.Errorf("Expected 2 groupedEntryFiles, got %d", len(groupedEntryFiles))
	}

	overviewContent, _ := os.ReadFile(overviewFile.Name())
	if string(overviewContent) != "overview html" {
		t.Errorf("Expected overview content 'overview html', got %s", string(overviewContent))
	}

	for _, f := range groupedEntryFiles {
		content, _ := os.ReadFile(f.Name())
		if string(content) != "grouped html" {
			t.Errorf("Expected grouped content 'grouped html', got %s", string(content))
		}
	}

	_ = overviewFile.Close()
	_ = os.RemoveAll(filepath.Dir(overviewFile.Name())) 
}
