package services

import (
	"miniflux-digest/internal/models"

	miniflux "miniflux.app/v2/client"
)

type DigestService interface {
	BuildDigestData(category *miniflux.Category, entries *miniflux.Entries, icons map[int64]*models.FeedIcon, groupBy string, subGroupBy string, sortBy string, minifluxHost string) *models.OverviewTemplateData
}
