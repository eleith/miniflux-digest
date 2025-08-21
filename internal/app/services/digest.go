package services

import (
	"miniflux-digest/internal/models"
)

type DigestService interface {
	BuildDigestData(entries []*models.Entry, icons map[int64]*models.FeedIcon, groupBy string, subGroupBy string, sortBy string, minifluxHost string) *models.OverviewTemplateData
}
