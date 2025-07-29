package processor

import (
	"log"

	"miniflux-digest/internal/app"
)

func CategoryDigestJob(application *app.App, rawData *app.RawCategoryData, markAsRead bool) {
	data := application.DigestService.BuildDigestData(rawData.Category, rawData.Entries, rawData.Icons, application.Config.Digest.GroupBy, application.Config.Miniflux.Host)

	if len(*data.Entries) > 0 {
		file, err := application.ArchiveService.MakeArchiveHTML(data, application.Config.Digest.Compress)

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
