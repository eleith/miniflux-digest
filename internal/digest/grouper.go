package digest

import (
	"miniflux-digest/internal/app/services"
	"miniflux-digest/internal/models"
)

const (
	DayGroupLayout      = "2006-01-02"
	DayGroupTitleLayout = "Jan 2, 2006"
)

type SubGrouper interface {
	GroupEntries(pgs []*primaryGroup) ([]*models.PrimaryGroupDigestData, *string)
}

func NewSubGrouper(subGroupBy string, llmService services.LLMService) SubGrouper {
	switch subGroupBy {
	case "ai":
		return &LLMGrouper{LLMService: llmService}
	case "feed":
		return &FeedGrouper{}
	default: // day
		return &DayGrouper{}
	}
}
