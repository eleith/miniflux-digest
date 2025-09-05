package digest

import (
	"miniflux-digest/internal/app/services"
	"miniflux-digest/internal/digest/view"
	"miniflux-digest/internal/models"
)

func NewDigestService(llmService services.LLMService) services.DigestService {
	return &digestServiceImpl{llmService: llmService}
}

type digestServiceImpl struct {
	llmService services.LLMService
}

func (s *digestServiceImpl) BuildDigestData(entries []*models.Entry, icons map[int64]*models.FeedIcon, viewType string, minifluxHost string) *models.OverviewTemplateData {
	return view.BuildDigestDataForView(entries, icons, viewType, minifluxHost, s.llmService)
}