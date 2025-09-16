package app

import (
	"context"
	"os"
	"testing"
	"time"

	"google.golang.org/genai"

	"miniflux-digest/internal/config"
	"miniflux-digest/internal/models"
)

type mockArchiveService struct{}

func (m *mockArchiveService) MakeArchiveHTML(data *models.OverviewTemplateData, compress bool) (*os.File, []*os.File, error) {
	return nil, nil, nil
}
func (m *mockArchiveService) CleanArchive(maxAge time.Duration) {}

type mockEmailService struct{}

func (m *mockEmailService) Send(cfg *config.Config, overviewFile *os.File, groupedEntryFiles []*os.File, data *models.OverviewTemplateData) error {
	return nil
}

type mockMinifluxClientService struct{}

func (m *mockMinifluxClientService) FeedIcon(feedID int64) (*models.FeedIcon, error) {
	return nil, nil
}
func (m *mockMinifluxClientService) GetAllUnreadEntries() ([]*models.Entry, error) {
	return nil, nil
}
func (m *mockMinifluxClientService) UpdateEntries(entryIDs []int64, status string) error {
	return nil
}

type mockDigestService struct{}

func (m *mockDigestService) BuildDigestData(entries []*models.Entry, icons map[int64]*models.FeedIcon, view string, minifluxHost string, digestHost string) *models.OverviewTemplateData {
	return nil
}

type mockLLMService struct{}

func (m *mockLLMService) GenerateContent(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
	return nil, nil
}

func (m *mockLLMService) GenerateContentWithResponse(ctx context.Context, prompt string, schema *genai.Schema, response interface{}) error {
	return nil
}

func TestNewApp(t *testing.T) {
	mockArchiveSvc := &mockArchiveService{}
	mockEmailSvc := &mockEmailService{}
	mockMinifluxClientSvc := &mockMinifluxClientService{}
	mockDigestSvc := &mockDigestService{}
	mockLLMSvc := &mockLLMService{}
	cfg := &config.Config{}

	app := NewApp(
		WithConfig(cfg),
		WithArchiveService(mockArchiveSvc),
		WithEmailService(mockEmailSvc),
		WithMinifluxClientService(mockMinifluxClientSvc),
		WithDigestService(mockDigestSvc),
		WithLLMService(mockLLMSvc),
	)

	if app == nil {
		t.Error("NewApp() should not return nil")
		return
	}

	if app.Config != cfg {
		t.Errorf("NewApp() did not set the config correctly. Got %v, want %v", app.Config, cfg)
	}
	if app.ArchiveService != mockArchiveSvc {
		t.Errorf("NewApp() did not set the archive service correctly. Got %v, want %v", app.ArchiveService, mockArchiveSvc)
	}
	if app.EmailService != mockEmailSvc {
		t.Errorf("NewApp() did not set the email service correctly. Got %v, want %v", app.EmailService, mockEmailSvc)
	}
	if app.MinifluxClientService != mockMinifluxClientSvc {
		t.Errorf("NewApp() did not set the miniflux client service correctly. Got %v, want %v", app.MinifluxClientService, mockMinifluxClientSvc)
	}
	if app.DigestService != mockDigestSvc {
		t.Errorf("NewApp() did not set the digest service correctly. Got %v, want %v", app.DigestService, mockDigestSvc)
	}
	if app.LLMService != mockLLMSvc {
		t.Errorf("NewApp() did not set the LLM service correctly. Got %v, want %v", app.LLMService, mockLLMSvc)
	}
}

func TestWithConfig(t *testing.T) {
	cfg := &config.Config{}
	app := NewApp(WithConfig(cfg))
	if app.Config != cfg {
		t.Errorf("WithConfig() did not set the config correctly. Got %v, want %v", app.Config, cfg)
	}
}

func TestWithArchiveService(t *testing.T) {
	mockSvc := &mockArchiveService{}
	app := NewApp(WithArchiveService(mockSvc))
	if app.ArchiveService != mockSvc {
		t.Errorf("WithArchiveService() did not set the archive service correctly. Got %v, want %v", app.ArchiveService, mockSvc)
	}
}

func TestWithEmailService(t *testing.T) {
	mockSvc := &mockEmailService{}
	app := NewApp(WithEmailService(mockSvc))
	if app.EmailService != mockSvc {
		t.Errorf("WithEmailService() did not set the email service correctly. Got %v, want %v", app.EmailService, mockSvc)
	}
}

func TestWithMinifluxClientService(t *testing.T) {
	mockSvc := &mockMinifluxClientService{}
	app := NewApp(WithMinifluxClientService(mockSvc))
	if app.MinifluxClientService != mockSvc {
		t.Errorf("WithMinifluxClientService() did not set the miniflux client service correctly. Got %v, want %v", app.MinifluxClientService, mockSvc)
	}
}

func TestWithDigestService(t *testing.T) {
	mockSvc := &mockDigestService{}
	app := NewApp(WithDigestService(mockSvc))
	if app.DigestService != mockSvc {
		t.Errorf("WithDigestService() did not set the digest service correctly. Got %v, want %v", app.DigestService, mockSvc)
	}
}

func TestWithLLMService(t *testing.T) {
	mockSvc := &mockLLMService{}
	app := NewApp(WithLLMService(mockSvc))
	if app.LLMService != mockSvc {
		t.Errorf("WithLLMService() did not set the LLM service correctly. Got %v, want %v", app.LLMService, mockSvc)
	}
}
