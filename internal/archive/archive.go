package archive

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"miniflux-digest/internal/app"
	"miniflux-digest/internal/digest"
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/templates"
	"miniflux-digest/internal/utils"
	"os"
	"path/filepath"
	"time"
)

type ArchiveServiceImpl struct{
	ArchiveBaseDir string
}

var _ app.ArchiveService = (*ArchiveServiceImpl)(nil)

func NewArchiveService(archiveBaseDir string) *ArchiveServiceImpl {
	return &ArchiveServiceImpl{ArchiveBaseDir: archiveBaseDir}
}

func (s *ArchiveServiceImpl) getHTML(data *models.HTMLTemplateData, compress bool) ([]byte, error) {
	var buf bytes.Buffer

	err := templates.ArchiveTemplate.Execute(&buf, data)
	if err != nil {
		return nil, err
	}

	return digest.MinifyHTML(buf.Bytes(), compress)
}

func (s *ArchiveServiceImpl) makeArchiveFile(data *models.HTMLTemplateData) (*os.File, error) {
	categorySlug := utils.Slugify(data.Category.Title)
	categoryFolderPath := fmt.Sprintf("%s/%s", s.ArchiveBaseDir, categorySlug)
	filename := fmt.Sprintf("%s/%s.html", categoryFolderPath, data.GeneratedDate.Format("2006-01-02"))
	err := os.MkdirAll(categoryFolderPath, 0755)

	if err == nil {
		file, err := os.Create(filename)
		return file, err
	}

	return nil, err
}

func (s *ArchiveServiceImpl) MakeArchiveHTML(data *models.HTMLTemplateData, compress bool) (*os.File, error) {
	file, err := s.makeArchiveFile(data)

	if err != nil {
		log.Printf("Error creating HTML file for category '%s': %v", data.Category.Title, err)
		return nil, err
	}

	htmlOutput, err := s.getHTML(data, compress)

	if err != nil {
		log.Printf("Error generating HTML for category %s: %v", data.Category.Title, err)
		return file, err
	}

	_, err = file.Write(htmlOutput)

	if err != nil {
		log.Printf("Error writing HTML to file for category '%s': %v", data.Category.Title, err)
	}

	return file, err
}

func (s *ArchiveServiceImpl) removeOldArchiveFiles(maxAge time.Duration) {
	cutoffTime := time.Now().Add(-maxAge)

	err := filepath.WalkDir(s.ArchiveBaseDir, func(path string, dir fs.DirEntry, err error) error {
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

func (s *ArchiveServiceImpl) isDirEmpty(name string) (bool, error) {
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

func (s *ArchiveServiceImpl) removeEmptyCategoryFolders() {
	dirs, err := os.ReadDir(s.ArchiveBaseDir)
	if err != nil {
		log.Printf("Warning: could not read archive directory %s: %v", s.ArchiveBaseDir, err)
		return
	}

	for _, dir := range dirs {
		if dir.IsDir() {
			categoryPath := filepath.Join(s.ArchiveBaseDir, dir.Name())
			empty, err := s.isDirEmpty(categoryPath)
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
	s.removeOldArchiveFiles(maxAge)
	s.removeEmptyCategoryFolders()
}