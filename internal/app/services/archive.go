package services

import (
	"miniflux-digest/internal/models"
	"os"
	"time"
)

type ArchiveService interface {
	MakeArchiveHTML(data *models.OverviewTemplateData, compress bool) (*os.File, []*os.File, error)
	CleanArchive(maxAge time.Duration)
}
