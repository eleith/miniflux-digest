package processor

import (
	"log"

	"miniflux-digest/internal/app"
	"miniflux-digest/internal/models"
)

func ProcessCategory(application *app.App, data *models.CategoryData, markAsRead bool) {
	if len(*data.Entries) > 0 {
		file, err := application.ArchiveService.MakeArchiveHTML(application.Config.ArchivePath, data)

		if err != nil {
			log.Printf("Error generating File for category %s: %v", data.Category.Title, err)
			return
		}

		defer func() {
			if err := file.Close(); err != nil {
				log.Printf("Error closing file for category '%s': %v", data.Category.Title, err)
			}
		}()

		err = application.EmailService.Send(application.Config, file, data)

		if err != nil {
			log.Printf("Error sending email for category '%s': %v", data.Category.Title, err)
		}

		if markAsRead {
			if err := application.MinifluxClientService.MarkCategoryAsRead(data.Category.ID); err != nil {
				log.Printf("Error marking category as read for category '%s': %v", data.Category.Title, err)
			}
		}
	}
}
