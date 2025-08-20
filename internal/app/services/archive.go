package services

import (
	"miniflux-digest/internal/models"
	"os"
	"time"
)

type ArchiveService interface {
	MakeArchiveHTML(data *models.HTMLTemplateData, compress bool) (*os.File, error)
	CleanArchive(maxAge time.Duration)
}
