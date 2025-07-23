package processor

import (
	"log"

	"miniflux-digest/config"
	"miniflux-digest/internal/app"
	"miniflux-digest/internal/models"
)

func ProcessCategory(cfg *config.Config, client app.MinifluxClientService, data *models.CategoryData, archiveService app.ArchiveService, emailService app.EmailService, archivePath string, markAsRead bool) {
	if len(*data.Entries) > 0 {
		file, err := archiveService.MakeArchiveHTML(archivePath, data)

		if err != nil {
			log.Printf("Error generating File for category %s: %v", data.Category.Title, err)
			return
		}

		defer func() {
			if err := file.Close(); err != nil {
				log.Printf("Error closing file for category '%s': %v", data.Category.Title, err)
			}
		}()

		err = emailService.Send(cfg, file, data)

		if err != nil {
			log.Printf("Error sending email for category '%s': %v", data.Category.Title, err)
		}

		if markAsRead {
			if err := client.MarkCategoryAsRead(data.Category.ID); err != nil {
				log.Printf("Error marking category as read for category '%s': %v", data.Category.Title, err)
			}
		}
	}
}
