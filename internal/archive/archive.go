package archive

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"miniflux-digest/internal/digest"
	"miniflux-digest/internal/models"
	"os"
	"path/filepath"
	"time"
)

type HTMLTemplate interface {
	Execute(wr io.Writer, data any) error
}

type ArchiveServiceImpl struct {
	ArchiveBaseDir   string
	ArchiveTemplate  HTMLTemplate
	OverviewTemplate HTMLTemplate
}

func NewArchiveService(archiveBaseDir string, archiveTemplate HTMLTemplate, overviewTemplate HTMLTemplate) *ArchiveServiceImpl {
	return &ArchiveServiceImpl{ArchiveBaseDir: archiveBaseDir, ArchiveTemplate: archiveTemplate, OverviewTemplate: overviewTemplate}
}

func (s *ArchiveServiceImpl) getHTML(template HTMLTemplate, data any, compress bool) ([]byte, error) {
	var buf bytes.Buffer

	err := template.Execute(&buf, data)
	if err != nil {
		return nil, err
	}

	return digest.MinifyHTML(buf.Bytes(), compress)
}

func (s *ArchiveServiceImpl) makeGroupedEntriesArchiveFile(data *models.PrimaryGroupDigestData, dateFolderPath string) (*os.File, error) {
	groupSlug := data.Slug
	groupedFolderPath := filepath.Join(dateFolderPath, "digests")
	if err := os.MkdirAll(groupedFolderPath, 0755); err != nil {
		return nil, err
	}
	filename := fmt.Sprintf("%s/%s.html", groupedFolderPath, groupSlug)
	file, err := os.Create(filename)
	return file, err
}

func (s *ArchiveServiceImpl) makeOverviewArchiveFile(data *models.OverviewTemplateData) (*os.File, string, error) {
	dateFolderPath := fmt.Sprintf("%s/%s", s.ArchiveBaseDir, data.GeneratedDate.Format("2006-01-02"))
	filename := fmt.Sprintf("%s/index.html", dateFolderPath)
	err := os.MkdirAll(dateFolderPath, 0755)

	if err == nil {
		file, err := os.Create(filename)
		return file, dateFolderPath, err
	}

	return nil, "", err
}

func (s *ArchiveServiceImpl) MakeArchiveHTML(data *models.OverviewTemplateData, compress bool) (*os.File, []*os.File, error) {
	overviewFile, dateFolderPath, err := s.makeOverviewArchiveFile(data)
	if err != nil {
		log.Printf("Error creating overview HTML file: %v", err)
		return nil, nil, err
	}
	overviewHTML, err := s.getHTML(s.OverviewTemplate, data, compress)
	if err != nil {
		log.Printf("Error generating overview HTML: %v", err)
		return overviewFile, nil, err
	}
	_, err = overviewFile.Write(overviewHTML)
	if err != nil {
		log.Printf("Error writing overview HTML to file: %v", err)
	}

	_, err = overviewFile.Seek(0, io.SeekStart)
	if err != nil {
		log.Printf("Error rewinding overview HTML file: %v", err)
		return overviewFile, nil, err
	}

	var groupedEntryFiles []*os.File

	feedIconsMap := make(map[int64]*models.FeedIcon)
	for _, icon := range data.FeedIcons {
		feedIconsMap[icon.FeedID] = icon
	}

	for _, primaryGroup := range data.PrimaryGroups {
		totalEntriesInPrimaryGroup := 0
		for _, subGroup := range primaryGroup.SubGroups {
			totalEntriesInPrimaryGroup += len(subGroup.Entries)
		}

		groupIconsMap := make(map[int64]*models.FeedIcon)
		for _, subGroup := range primaryGroup.SubGroups {
			for _, entry := range subGroup.Entries {
				if icon, ok := feedIconsMap[entry.FeedID]; ok {
					groupIconsMap[entry.FeedID] = icon
				}
			}
		}

		var groupIconsSlice []*models.FeedIcon
		for _, icon := range groupIconsMap {
			groupIconsSlice = append(groupIconsSlice, icon)
		}

		groupedPageData := &models.GroupedDigestPageData{
			PrimaryGroup:  primaryGroup,
			FeedIcons:     groupIconsSlice,
			MinifluxHost:  data.MinifluxHost,
			GeneratedDate: data.GeneratedDate,
			TotalEntries:  totalEntriesInPrimaryGroup,
		}
		groupedEntriesFile, err := s.makeGroupedEntriesArchiveFile(primaryGroup, dateFolderPath)
		if err != nil {
			log.Printf("Error creating grouped entries HTML file: %v", err)
			return nil, nil, err
		}
		groupedEntryFiles = append(groupedEntryFiles, groupedEntriesFile)
		groupedEntriesHTML, err := s.getHTML(s.ArchiveTemplate, groupedPageData, compress)
		if err != nil {
			log.Printf("Error generating grouped entries HTML: %v", err)
			return overviewFile, groupedEntryFiles, err
		}
		_, err = groupedEntriesFile.Write(groupedEntriesHTML)
		if err != nil {
			log.Printf("Error writing grouped entries HTML to file: %v", err)
		}
	}

	return overviewFile, groupedEntryFiles, err
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

func (s *ArchiveServiceImpl) removeEmptyDirs() {
	var dirs []string

	err := filepath.WalkDir(s.ArchiveBaseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	})

	if err != nil {
		log.Printf("Error walking directory tree: %v", err)
		return
	}

	for i := len(dirs) - 1; i >= 0; i-- {
		dir := dirs[i]

		if dir == s.ArchiveBaseDir {
			continue
		}

		empty, err := s.isDirEmpty(dir)
		if err != nil {
			log.Printf("Warning: could not check if directory %s is empty: %v", dir, err)
			continue
		}

		if empty {
			if err := os.Remove(dir); err != nil {
				log.Printf("Warning: failed to delete empty directory %s: %v", dir, err)
			}
		}
	}
}

func (s *ArchiveServiceImpl) isDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Warning: failed to close directory %s: %v", name, err)
		}
	}()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

func (s *ArchiveServiceImpl) CleanArchive(maxAge time.Duration) {
	s.removeOldArchiveFiles(maxAge)
	s.removeEmptyDirs()
}
