package processor

import (
	"log"
	"os"
	"testing"
	"time"

	"miniflux-digest/internal/app"
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/models"

	miniflux "miniflux.app/v2/client"
)

type mockMinifluxClientService struct {
	getAllUnreadEntriesFunc func() ([]*models.Entry, error)
	feedIconFunc            func(feedID int64) (*models.FeedIcon, error)
	updateEntriesFunc       func(entryIDs []int64, status string) error
}

func (m *mockMinifluxClientService) FeedIcon(feedID int64) (*models.FeedIcon, error) {
	return m.feedIconFunc(feedID)
}
func (m *mockMinifluxClientService) GetAllUnreadEntries() ([]*models.Entry, error) {
	return m.getAllUnreadEntriesFunc()
}
func (m *mockMinifluxClientService) UpdateEntries(entryIDs []int64, status string) error {
	return m.updateEntriesFunc(entryIDs, status)
}

type mockDigestService struct {
	buildDigestDataFunc func(entries []*models.Entry, icons map[int64]*models.FeedIcon, view string, minifluxHost string, digestHost string, categories []config.ConfigCategory) *models.OverviewTemplateData
}

func (m *mockDigestService) BuildDigestData(entries []*models.Entry, icons map[int64]*models.FeedIcon, view string, minifluxHost string, digestHost string, categories []config.ConfigCategory) *models.OverviewTemplateData {
	return m.buildDigestDataFunc(entries, icons, view, minifluxHost, digestHost, categories)
}

type mockArchiveService struct {
	makeArchiveHTMLFunc func(data *models.OverviewTemplateData, compress bool) (*os.File, []*os.File, error)
	cleanArchiveFunc    func(maxAge time.Duration)
}

func (m *mockArchiveService) MakeArchiveHTML(data *models.OverviewTemplateData, compress bool) (*os.File, []*os.File, error) {
	return m.makeArchiveHTMLFunc(data, compress)
}
func (m *mockArchiveService) CleanArchive(maxAge time.Duration) {
	m.cleanArchiveFunc(maxAge)
}

func TestProcessDigest_Success(t *testing.T) {
	log.SetOutput(os.Stdout)
	defer log.SetOutput(os.Stderr)

	expectedEntries := []*models.Entry{
		{ID: 1, FeedID: 101, Title: "Entry 1"},
		{ID: 2, FeedID: 102, Title: "Entry 2"},
	}
	expectedIcons := map[int64]*models.FeedIcon{
		101: {FeedID: 101, Data: "icon1"},
		102: {FeedID: 102, Data: "icon2"},
	}
	expectedOverviewData := &models.OverviewTemplateData{}

	mockMinifluxClient := &mockMinifluxClientService{
		getAllUnreadEntriesFunc: func() ([]*models.Entry, error) {
			return expectedEntries, nil
		},
		feedIconFunc: func(feedID int64) (*models.FeedIcon, error) {
			return expectedIcons[feedID], nil
		},
		updateEntriesFunc: func(entryIDs []int64, status string) error {
			if status != miniflux.EntryStatusRead {
				t.Errorf("Expected status %s, got %s", miniflux.EntryStatusRead, status)
			}
			return nil
		},
	}
	mockDigestService := &mockDigestService{
		buildDigestDataFunc: func(entries []*models.Entry, icons map[int64]*models.FeedIcon, view string, minifluxHost string, digestHost string, categories []config.ConfigCategory) *models.OverviewTemplateData {
			return expectedOverviewData
		},
	}
	mockArchiveService := &mockArchiveService{
		makeArchiveHTMLFunc: func(data *models.OverviewTemplateData, compress bool) (*os.File, []*os.File, error) {
			overviewFile, _ := os.CreateTemp("", "overview-*.html")
			groupedFile1, _ := os.CreateTemp("", "grouped-*.html")
			return overviewFile, []*os.File{groupedFile1}, nil
		},
	}

	mockApp := app.NewApp(
		app.WithConfig(&config.Config{
			Digests: []config.ConfigDigest{
				{
					MarkAsRead: true,
					Compress:   true,
					View:       "category",
				},
			},
			Miniflux: config.ConfigMiniflux{Host: "http://miniflux.test"},
		}),
		app.WithMinifluxClientService(mockMinifluxClient),
		app.WithDigestService(mockDigestService),
		app.WithArchiveService(mockArchiveService),
	)

	overviewFile, groupedEntryFiles, data, err := ProcessDigest(mockApp, 0)

	if err != nil {
		t.Fatalf("ProcessDigest returned an unexpected error: %v", err)
	}
	if overviewFile == nil {
		t.Error("Overview file should not be nil")
	}
	if len(groupedEntryFiles) == 0 {
		t.Error("Grouped entry files should not be empty")
	}
	if data != expectedOverviewData {
		t.Errorf("Expected overview data %v, got %v", expectedOverviewData, data)
	}

	_ = overviewFile.Close()
	_ = os.Remove(overviewFile.Name())
	for _, f := range groupedEntryFiles {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}
}
