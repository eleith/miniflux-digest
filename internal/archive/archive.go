package archive

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"miniflux-digest/internal/app"
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/templates"
	"miniflux-digest/internal/utils"
	"os"
	"path/filepath"
	"time"
)

var archiveBaseDir string

func init() {
	executablePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	executableDir := filepath.Dir(executablePath)

	// Assuming the executable is in cmd/miniflux-digest/miniflux-digest
	// We need to go up two levels to reach the project root
	projectRoot := filepath.Join(executableDir, "..", "..")

	archiveBaseDir = filepath.Join(projectRoot, "web", "miniflux-archive")
}

type ArchiveServiceImpl struct{}

var _ app.ArchiveService = (*ArchiveServiceImpl)(nil)

func getHTML(data *models.CategoryData) (string, error) {
	var buf bytes.Buffer
	var htmlOutput string

	err := templates.ArchiveTemplate.Execute(&buf, data)

	if err == nil {
		htmlOutput = buf.String()
	}

	return htmlOutput, err
}

func makeArchiveFile(data *models.CategoryData) (*os.File, error) {
	categorySlug := utils.Slugify(data.Category.Title)
	categoryFolderPath := fmt.Sprintf("%s/%s", archiveBaseDir, categorySlug)
	filename := fmt.Sprintf("%s/%s.html", categoryFolderPath, data.GeneratedDate.Format("2006-01-02"))
	err := os.MkdirAll(categoryFolderPath, os.ModePerm)

	if err == nil {
		file, err := os.Create(filename)
		return file, err
	}

	return nil, err
}

func (s *ArchiveServiceImpl) MakeArchiveHTML(data *models.CategoryData) (*os.File, error) {
	file, err := makeArchiveFile(data)

	if err != nil {
		log.Printf("Error creating HTML file for category '%s': %v", data.Category.Title, err)
		return nil, err
	}

	htmlOutput, err := getHTML(data)

	if err != nil {
		log.Printf("Error generating HTML for category %s: %v", data.Category.Title, err)
		return file, err
	}

	_, err = file.WriteString(htmlOutput)

	if err != nil {
		log.Printf("Error writing HTML to file for category '%s': %v", data.Category.Title, err)
	}

	return file, err
}

func removeOldArchiveFiles(maxAge time.Duration) {
	cutoffTime := time.Now().Add(-maxAge)

	err := filepath.WalkDir(archiveBaseDir, func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if dir.IsDir() {
			return nil
		}

		info, err := dir.Info()
		if err != nil {
			log.Printf("Warning: could not get info for file %s: %v", path, err)
			return nil
		}

		if info.ModTime().Before(cutoffTime) {
			if err := os.Remove(path); err != nil {
				log.Printf("Warning: failed to delete file %s: %v", path, err)
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("Error cleaning archive files: %v", err)
	}
}

func isDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}

	defer func() {
		if err = f.Close(); err != nil {
			log.Printf("Warning: failed to close directory %s: %v", name, err)
		}
	}()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

func removeEmptyCategoryFolders(archivePath string) {
	dirs, err := os.ReadDir(archivePath)
	if err != nil {
		log.Printf("Warning: could not read archive directory %s: %v", archivePath, err)
		return
	}

	for _, dir := range dirs {
		if dir.IsDir() {
			categoryPath := filepath.Join(archivePath, dir.Name())
			empty, err := isDirEmpty(categoryPath)
			if err != nil {
				log.Printf("Warning: could not check if directory %s is empty: %v", categoryPath, err)
				continue
			}
			if empty {
				if err := os.Remove(categoryPath); err != nil {
					log.Printf("Warning: failed to delete empty directory %s: %v", categoryPath, err)
				}
			}
		}
	}
}

func (s *ArchiveServiceImpl) CleanArchive(maxAge time.Duration) {
	removeOldArchiveFiles(maxAge)
	removeEmptyCategoryFolders(archiveBaseDir)
}