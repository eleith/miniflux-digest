package services

import (
	"miniflux-digest/internal/models"
)

type DigestService interface {
	BuildDigestData(entries []*models.Entry, icons map[int64]*models.FeedIcon, view string, minifluxHost string) *models.OverviewTemplateData
}
