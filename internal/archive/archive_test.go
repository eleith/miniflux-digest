package archive

import (
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

	// Create a mock EntryGroup (sub-group)
	mockSubGroup := &models.EntryGroup{
		Title:   "Test SubGroup Title",
		Summary: "Test SubGroup Summary",
		Entries: *testutil.NewMockEntries(),
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

	html, err := archiveService.getHTML(templates.ArchiveTemplate, &data, true)
	if err != nil {
		t.Fatalf("getHTML failed: %v", err)
	}
	if len(html) == 0 {
		t.Error("Expected HTML to be non-empty")
	}
}

func TestMakeArchiveFile(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()
	archiveService := NewArchiveService(tempDir, templates.OverviewTemplate, templates.ArchiveTemplate)

	// Create a mock EntryGroup (sub-group)
	mockSubGroup := &models.EntryGroup{
		Title:   "Test SubGroup Title",
		Summary: "Test SubGroup Summary",
		Entries: *testutil.NewMockEntries(),
		Slug:    "test-subgroup-title",
	}

	// Create PrimaryGroupDigestData
	data := models.PrimaryGroupDigestData{
		Title:     "Test Primary Group Title",
		Slug:      "test-primary-group-title",
		SubGroups: []*models.EntryGroup{mockSubGroup},
		Summary:   "Test Primary Group Summary",
	}

	dateFolderPath := filepath.Join(tempDir, time.Now().Format("2006-01-02"))
	// Ensure the date folder exists for the test
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
	// Check if the file was created in the correct hardcoded path
	expectedPath := filepath.Join(dateFolderPath, "digests", data.Slug+".html")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist", expectedPath)
	}
}

func TestCleanArchive(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()
	archiveService := NewArchiveService(tempDir, templates.OverviewTemplate, templates.ArchiveTemplate)

	// Create an old directory with a file
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

	// Create a new directory with a file
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
