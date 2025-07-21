package archive

import (
	"miniflux-digest/internal/testutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMakeArchiveHTML(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "miniflux-digest-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	}()

	// Set the archivePath to the temporary directory
	archivePath = tempDir

	// Get mock data from the test utility
	mockData := testutil.NewMockCategoryData()

	// Call the function to be tested
	file, err := MakeArchiveHTML(mockData)
	if err != nil {
		t.Fatalf("MakeArchiveHTML failed: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("Failed to close file: %v", err)
		}
	}()

	// Check if the file was created
	info, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}

	if info.Name() == "" {
		t.Error("File name is empty")
	}

	// Check if the file has content
	if info.Size() == 0 {
		t.Error("File is empty")
	}
}

func TestCleanArchive(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "miniflux-digest-test-clean")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	}()

	archivePath = tempDir
	maxAge := 7 * 24 * time.Hour // 7 days

	// Create dummy files and directories
	categoryDir := filepath.Join(tempDir, "test-category")
	emptyDir := filepath.Join(tempDir, "empty-category")
	if err := os.Mkdir(categoryDir, 0755); err != nil {
		t.Fatalf("Failed to create category dir: %v", err)
	}
	if err := os.Mkdir(emptyDir, 0755); err != nil {
		t.Fatalf("Failed to create empty dir: %v", err)
	}

	// File that should be kept
	newFile, err := os.Create(filepath.Join(categoryDir, "new.html"))
	if err != nil {
		t.Fatalf("Failed to create new file: %v", err)
	}
	if err := newFile.Close(); err != nil {
		t.Fatalf("Failed to close new file: %v", err)
	}

	// File that should be deleted
	oldFile, err := os.Create(filepath.Join(categoryDir, "old.html"))
	if err != nil {
		t.Fatalf("Failed to create old file: %v", err)
	}
	if err := oldFile.Close(); err != nil {
		t.Fatalf("Failed to close old file: %v", err)
	}
	twoWeeksAgo := time.Now().Add(-2 * maxAge)
	if err := os.Chtimes(oldFile.Name(), twoWeeksAgo, twoWeeksAgo); err != nil {
		t.Fatalf("Failed to change old file mod time: %v", err)
	}

	CleanArchive(maxAge)

	// Check that old file is deleted
	if _, err := os.Stat(oldFile.Name()); !os.IsNotExist(err) {
		t.Errorf("Expected old file to be deleted, but it still exists")
	}

	// Check that new file still exists
	if _, err := os.Stat(newFile.Name()); os.IsNotExist(err) {
		t.Errorf("Expected new file to exist, but it was deleted")
	}

	// Check that empty directory is deleted
	if _, err := os.Stat(emptyDir); !os.IsNotExist(err) {
		t.Errorf("Expected empty directory to be deleted, but it still exists")
	}

	// Check that non-empty directory still exists
	if _, err := os.Stat(categoryDir); os.IsNotExist(err) {
		t.Errorf("Expected non-empty directory to exist, but it was deleted")
	}
}