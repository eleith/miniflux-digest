package services

import (
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/models"
)

type DigestService interface {
	BuildDigestData(entries []*models.Entry, icons map[int64]*models.FeedIcon, view string, minifluxHost string, digestHost string, categories []config.ConfigCategory, digestTitle string) *models.OverviewTemplateData
}
