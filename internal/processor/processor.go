package processor

import (
	"log"

	"miniflux-digest/config"
	"miniflux-digest/internal/archive"
	"miniflux-digest/internal/category"
	"miniflux-digest/internal/email"
	miniflux "miniflux.app/v2/client"
)

func ProcessCategory(cfg *config.Config, client *miniflux.Client, data *category.CategoryData, archivePath string, markAsRead bool) {
	if len(*data.Entries) > 0 {
		file, err := archive.MakeArchiveHTML(archivePath, data)

		if err != nil {
			log.Printf("Error generating File for category %s: %v", data.Category.Title, err)
			return
		}

		defer func() {
			if err := file.Close(); err != nil {
				log.Printf("Error closing file for category '%s': %v", data.Category.Title, err)
			}
		}()

		err = email.Send(cfg, file, data)

		if err != nil {
			log.Printf("Error sending email for category '%s': %v", data.Category.Title, err)
		}

		if markAsRead {
			category.MarkAsRead(client, data.Category)
		}
	}
}
