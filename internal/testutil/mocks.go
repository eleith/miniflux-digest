package testutil

import (
	"miniflux-digest/internal/app"
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/models"
	"os"
	"time"

	miniflux "miniflux.app/v2/client"
)

type MockMinifluxClient struct {
	app.MinifluxClientService
	MarkAsReadFunc func(categoryID int64) error
	CategoriesFunc func() ([]*miniflux.Category, error)
	CategoryEntriesFunc func(categoryID int64, filter *miniflux.Filter) (*miniflux.Entries, error)
	CategoryFeedsFunc func(categoryID int64) ([]*miniflux.Feed, error)
	FeedIconFunc func(feedID int64) (*miniflux.FeedIcon, error)
	FetchCategoryDataFunc func(categoryID int64) (*models.CategoryData, error)
	StreamAllCategoryDataFunc func() <-chan *models.CategoryData
}

func (m *MockMinifluxClient) MarkCategoryAsRead(categoryID int64) error {
	if m.MarkAsReadFunc != nil {
		return m.MarkAsReadFunc(categoryID)
	}
	return nil
}

func (m *MockMinifluxClient) Categories() ([]*miniflux.Category, error) {
	if m.CategoriesFunc != nil {
		return m.CategoriesFunc()
	}
	return nil, nil
}

func (m *MockMinifluxClient) CategoryEntries(categoryID int64, filter *miniflux.Filter) (*miniflux.Entries, error) {
	if m.CategoryEntriesFunc != nil {
		return m.CategoryEntriesFunc(categoryID, filter)
	}
	return &miniflux.Entries{}, nil
}

func (m *MockMinifluxClient) CategoryFeeds(categoryID int64) ([]*miniflux.Feed, error) {
	if m.CategoryFeedsFunc != nil {
		return m.CategoryFeedsFunc(categoryID)
	}
	return nil, nil
}

func (m *MockMinifluxClient) FeedIcon(feedID int64) (*miniflux.FeedIcon, error) {
	if m.FeedIconFunc != nil {
		return m.FeedIconFunc(feedID)
	}
	return nil, nil
}

func (m *MockMinifluxClient) FetchCategoryData(categoryID int64) (*models.CategoryData, error) {
	if m.FetchCategoryDataFunc != nil {
		return m.FetchCategoryDataFunc(categoryID)
	}
	return nil, nil
}

func (m *MockMinifluxClient) StreamAllCategoryData() <-chan *models.CategoryData {
	if m.StreamAllCategoryDataFunc != nil {
		return m.StreamAllCategoryDataFunc()
	}
	out := make(chan *models.CategoryData)
	close(out)
	return out
}

type MockArchiveService struct {
	app.ArchiveService
	MakeArchiveHTMLFunc func(data *models.CategoryData) (*os.File, error)
}

func (m *MockArchiveService) MakeArchiveHTML(data *models.CategoryData) (*os.File, error) {
	return m.MakeArchiveHTMLFunc(data)
}

func (m *MockArchiveService) CleanArchive(maxAge time.Duration) {}

type MockEmailService struct {
	app.EmailService
	SendFunc func(cfg *config.Config, file *os.File, data *models.CategoryData) error
}

func (m *MockEmailService) Send(cfg *config.Config, file *os.File, data *models.CategoryData) error {
	return m.SendFunc(cfg, file, data)
}
