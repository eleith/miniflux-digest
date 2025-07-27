package archive

import (
	"miniflux-digest/internal/digest"
	"miniflux-digest/internal/testutil"
	"miniflux-digest/internal/utils"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetHTML(t *testing.T) {
	data := testutil.NewMockHTMLTemplateData(digest.GroupingTypeDay)
	html, err := getHTML(data, true)
	if err != nil {
		t.Fatalf("getHTML failed: %v", err)
	}
	if len(html) == 0 {
		t.Error("Expected HTML to be non-empty")
	}
}

func TestMakeArchiveFile(t *testing.T) {
	// Save original archiveBaseDir and restore it after the test
	oldArchiveBaseDir := archiveBaseDir
	archiveBaseDir = t.TempDir()
	defer func() {
		archiveBaseDir = oldArchiveBaseDir
		if err := os.RemoveAll(archiveBaseDir); err != nil {
			t.Errorf("Failed to clean up original archive directory: %v", err)
		}
	}()

	data := testutil.NewMockHTMLTemplateData(digest.GroupingTypeDay)
	file, err := makeArchiveFile(data)
	if err != nil {
		t.Fatalf("makeArchiveFile failed: %v", err)
	}
	if file == nil {
		t.Fatal("Expected file to be non-nil")
	}
	// Check if the file was created in the correct hardcoded path
	expectedPath := filepath.Join(archiveBaseDir, utils.Slugify(data.Category.Title), data.GeneratedDate.Format("2006-01-02")+".html")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist", expectedPath)
	}
}

func TestMakeArchiveHTML(t *testing.T) {
	// Save original archiveBaseDir and restore it after the test
	oldArchiveBaseDir := archiveBaseDir
	archiveBaseDir = t.TempDir()
	defer func() {
		archiveBaseDir = oldArchiveBaseDir
		if err := os.RemoveAll(archiveBaseDir); err != nil {
			t.Errorf("Failed to clean up original archive directory: %v", err)
		}
	}()

	data := testutil.NewMockHTMLTemplateData(digest.GroupingTypeDay)
	archiveService := &ArchiveServiceImpl{}
	file, err := archiveService.MakeArchiveHTML(data, true)
	if err != nil {
		t.Fatalf("MakeArchiveHTML failed: %v", err)
	}
	if file == nil {
		t.Fatal("Expected file to be non-nil")
	}
	info, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Expected file to have content")
	}
}

func TestCleanArchive(t *testing.T) {
	// Save original archiveBaseDir and restore it after the test
	oldArchiveBaseDir := archiveBaseDir
	archiveBaseDir = t.TempDir()
	defer func() {
		archiveBaseDir = oldArchiveBaseDir
		if err := os.RemoveAll(archiveBaseDir); err != nil {
			t.Errorf("Failed to clean up original archive directory: %v", err)
		}
	}()

	categorySlug := "test-category"
	categoryPath := filepath.Join(archiveBaseDir, categorySlug)
	if err := os.MkdirAll(categoryPath, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	oldFilePath := filepath.Join(categoryPath, "old.html")
	if err := os.WriteFile(oldFilePath, []byte("old"), 0644); err != nil {
		t.Fatalf("Failed to create old file: %v", err)
	}
	twoDaysAgo := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(oldFilePath, twoDaysAgo, twoDaysAgo); err != nil {
		t.Fatalf("Failed to change file modification time: %v", err)
	}

	newFilePath := filepath.Join(categoryPath, "new.html")
	if err := os.WriteFile(newFilePath, []byte("new"), 0644); err != nil {
		t.Fatalf("Failed to create new file: %v", err)
	}

	archiveService := &ArchiveServiceImpl{}
	archiveService.CleanArchive(24*time.Hour)

	if _, err := os.Stat(oldFilePath); !os.IsNotExist(err) {
		t.Error("Expected old file to be deleted")
	}

	if _, err := os.Stat(newFilePath); os.IsNotExist(err) {
		t.Error("Expected new file to be kept")
	}

	if err := os.Remove(newFilePath); err != nil {
		t.Fatalf("Failed to remove new file: %v", err)
	}

	// Test removeEmptyCategoryFolders separately
	emptyCategoryPath := filepath.Join(archiveBaseDir, "empty-category")
	if err := os.MkdirAll(emptyCategoryPath, 0755); err != nil {
		t.Fatalf("Failed to create empty test directory: %v", err)
	}
	removeEmptyCategoryFolders(archiveBaseDir)

	if _, err := os.Stat(emptyCategoryPath); !os.IsNotExist(err) {
		t.Error("Expected empty category directory to be deleted")
	}
}

func TestIsDirEmpty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-empty-dir")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	}()

	empty, err := isDirEmpty(tmpDir)
	if err != nil {
		t.Fatalf("isDirEmpty failed for empty dir: %v", err)
	}
	if !empty {
		t.Error("Expected directory to be empty")
	}

	filePath := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	empty, err = isDirEmpty(tmpDir)
	if err != nil {
		t.Fatalf("isDirEmpty failed for non-empty dir: %v", err)
	}
	if empty {
		t.Error("Expected directory to not be empty")
	}

	_, err = isDirEmpty("non-existent-dir")
	if !os.IsNotExist(err) {
		t.Errorf("Expected IsNotExist error for non-existent dir, got: %v", err)
	}
}
