package services

import (
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/models"
	"os"
)

type EmailService interface {
	Send(cfg *config.Config, overviewFile *os.File, groupedEntryFiles []*os.File, data *models.OverviewTemplateData) error
}
