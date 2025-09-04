package digest

import "miniflux-digest/internal/models"

type primaryGroup struct {
	ID      int64
	Title   string
	Entries []*models.Entry
}
