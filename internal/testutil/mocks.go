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
	SendFunc func(smtpConfig config.ConfigSmtp, digestConfig config.ConfigDigest, overviewFile *os.File, groupedEntryFiles []*os.File, data *models.OverviewTemplateData) error
}

func (m *MockEmailService) Send(smtpConfig config.ConfigSmtp, digestConfig config.ConfigDigest, overviewFile *os.File, groupedEntryFiles []*os.File, data *models.OverviewTemplateData) error {
	return m.SendFunc(smtpConfig, digestConfig, overviewFile, groupedEntryFiles, data)
}

type MockDigestService struct {
	services.DigestService
	BuildDigestDataFunc func(entries []*models.Entry, icons map[int64]*models.FeedIcon, view string, minifluxHost string, digestHost string, categories []config.ConfigCategory, digestTitle string) *models.OverviewTemplateData
}

func (m *MockDigestService) BuildDigestData(entries []*models.Entry, icons map[int64]*models.FeedIcon, view string, minifluxHost string, digestHost string, categories []config.ConfigCategory, digestTitle string) *models.OverviewTemplateData {
	if m.BuildDigestDataFunc != nil {
		return m.BuildDigestDataFunc(entries, icons, view, minifluxHost, digestHost, categories, digestTitle)
	}
	return nil
}

// Helper functions for tests
func FindPrimaryGroup(groups []*models.PrimaryGroupDigestData, title string) *models.PrimaryGroupDigestData {
	for _, g := range groups {
		if g.Title == title {
			return g
		}
	}
	return nil
}

func FindSubGroup(groups []*models.EntryGroup, title string) *models.EntryGroup {
	for _, g := range groups {
		if g.Title == title {
			return g
		}
	}
	return nil
}