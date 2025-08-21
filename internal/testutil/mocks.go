package testutil

import (
	"miniflux-digest/internal/app/services"
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/models"
	"os"
	"time"

	miniflux "miniflux.app/v2/client"
)

type MockMinifluxClient struct {
	services.MinifluxClientService
	
	
	
	
	FeedIconFunc func(feedID int64) (*miniflux.FeedIcon, error)
	
	GetAllUnreadEntriesFunc func() (*miniflux.Entries, error)
	UpdateEntriesFunc func(entryIDs []int64, status string) error
}









func (m *MockMinifluxClient) FeedIcon(feedID int64) (*miniflux.FeedIcon, error) {
	if m.FeedIconFunc != nil {
		return m.FeedIconFunc(feedID)
	}
	return nil, nil
}





func (m *MockMinifluxClient) GetAllUnreadEntries() (*miniflux.Entries, error) {
	if m.GetAllUnreadEntriesFunc != nil {
		return m.GetAllUnreadEntriesFunc()
	}
	return &miniflux.Entries{}, nil
}

func (m *MockMinifluxClient) UpdateEntries(entryIDs []int64, status string) error {
	if m.UpdateEntriesFunc != nil {
		return m.UpdateEntriesFunc(entryIDs, status)
	}
	return nil
}

type MockArchiveService struct {
	services.ArchiveService
	MakeArchiveHTMLFunc func(data *models.OverviewTemplateData, minify bool) (*os.File, error)
}

func (m *MockArchiveService) MakeArchiveHTML(data *models.OverviewTemplateData, minify bool) (*os.File, error) {
	return m.MakeArchiveHTMLFunc(data, minify)
}

func (m *MockArchiveService) CleanArchive(maxAge time.Duration) {}

type MockEmailService struct {
	services.EmailService
	SendFunc func(cfg *config.Config, overviewFile *os.File, groupedEntryFiles []*os.File, data *models.OverviewTemplateData) error
}

func (m *MockEmailService) Send(cfg *config.Config, overviewFile *os.File, groupedEntryFiles []*os.File, data *models.OverviewTemplateData) error {
	return m.SendFunc(cfg, overviewFile, groupedEntryFiles, data)
}

type MockDigestService struct {
	services.DigestService
	BuildDigestDataFunc func(category *miniflux.Category, entries *miniflux.Entries, icons map[int64]*models.FeedIcon, subGroupBy string, sortBy string, minifluxHost string) *models.OverviewTemplateData
}

func (m *MockDigestService) BuildDigestData(category *miniflux.Category, entries *miniflux.Entries, icons map[int64]*models.FeedIcon, subGroupBy string, sortBy string, minifluxHost string) *models.OverviewTemplateData {
	if m.BuildDigestDataFunc != nil {
		return m.BuildDigestDataFunc(category, entries, icons, subGroupBy, sortBy, minifluxHost)
	}
	return nil
}
