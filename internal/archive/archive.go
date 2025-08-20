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

	htmlTemplate "html/template"

	miniflux "miniflux.app/v2/client"
)

type ArchiveServiceImpl struct{
	ArchiveBaseDir string
}

var _ app.ArchiveService = (*ArchiveServiceImpl)(nil)

func NewArchiveService(archiveBaseDir string) *ArchiveServiceImpl {
	return &ArchiveServiceImpl{ArchiveBaseDir: archiveBaseDir}
}

func (s *ArchiveServiceImpl) getHTML(template *htmlTemplate.Template, data *models.HTMLTemplateData, compress bool) ([]byte, error) {
	var buf bytes.Buffer

	err := template.Execute(&buf, data)
	if err != nil {
		return nil, err
	}

	return digest.MinifyHTML(buf.Bytes(), compress)
}

func (s *ArchiveServiceImpl) makeGroupedEntriesArchiveFile(data *models.HTMLTemplateData) (*os.File, error) {
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

func (s *ArchiveServiceImpl) makeOverviewArchiveFile(data *models.HTMLTemplateData) (*os.File, error) {
	dateFolderPath := fmt.Sprintf("%s/%s", s.ArchiveBaseDir, data.GeneratedDate.Format("2006-01-02"))
	filename := fmt.Sprintf("%s/index.html", dateFolderPath)
	err := os.MkdirAll(dateFolderPath, 0755)

	if err == nil {
		file, err := os.Create(filename)
		return file, err
	}

	return nil, err
}

func (s *ArchiveServiceImpl) MakeArchiveHTML(data *models.HTMLTemplateData, compress bool) (*os.File, error) {
	// Generate overview HTML
	overviewFile, err := s.makeOverviewArchiveFile(data)
	if err != nil {
		log.Printf("Error creating overview HTML file: %v", err)
		return nil, err
	}
	overviewHTML, err := s.getHTML(templates.OverviewTemplate, data, compress)
	if err != nil {
		log.Printf("Error generating overview HTML: %v", err)
		return overviewFile, err
	}
	_, err = overviewFile.Write(overviewHTML)
	if err != nil {
		log.Printf("Error writing overview HTML to file: %v", err)
	}

	// Generate grouped entries HTML
	for _, entryGroup := range data.EntryGroups {
		groupData := &models.HTMLTemplateData{
			Category:      data.Category,
			Entries:       (*miniflux.Entries)(&entryGroup.Entries),
			GeneratedDate: data.GeneratedDate,
			FeedIcons:     data.FeedIcons,
			EntryGroups:   []*models.EntryGroup{entryGroup},
			Summary:       entryGroup.Title,
			MinifluxHost:  data.MinifluxHost,
		}
		groupedEntriesFile, err := s.makeGroupedEntriesArchiveFile(groupData)
		if err != nil {
			log.Printf("Error creating grouped entries HTML file: %v", err)
			return nil, err
		}
		groupedEntriesHTML, err := s.getHTML(templates.ArchiveTemplate, groupData, compress)
		if err != nil {
			log.Printf("Error generating grouped entries HTML: %v", err)
			return groupedEntriesFile, err
		}
		_, err = groupedEntriesFile.Write(groupedEntriesHTML)
		if err != nil {
			log.Printf("Error writing grouped entries HTML to file: %v", err)
		}
	}

	return overviewFile, err
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