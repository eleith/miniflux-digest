package processor

import (
	"log"
	"os"

	"miniflux-digest/internal/app"
	"miniflux-digest/internal/models"

	miniflux "miniflux.app/v2/client"
)

func ProcessAndSendDigest(application *app.App) (*os.File, []*os.File, *models.OverviewTemplateData, error) {
	log.Println("Starting to process digest...")

	entries, err := application.MinifluxClientService.GetAllUnreadEntries()
	if err != nil {
		log.Printf("Error getting all unread entries: %v", err)
		return nil, nil, nil, err
	}

	log.Printf("Found %d unread entries", len(*entries))

	// Fetch feed icons
	icons := make(map[int64]*models.FeedIcon)
	for _, entry := range *entries {
		if _, ok := icons[entry.FeedID]; !ok {
			icon, err := application.MinifluxClientService.FeedIcon(entry.FeedID)
			if err != nil {
				log.Printf("Warning: failed to fetch icon for feed %d: %v", entry.FeedID, err)
				continue
			}
			icons[entry.FeedID] = &models.FeedIcon{
				FeedID: icon.ID,
				Data:   icon.Data,
			}
		}
	}

	// Generate digest data
	data := application.DigestService.BuildDigestData(
		&miniflux.Category{Title: "All"},
		entries,
		icons,
		application.Config.Digest.GroupBy,
		application.Config.Digest.SubGroupBy,
		application.Config.Digest.SortBy,
		application.Config.Miniflux.Host,
	)

	// Generate HTML files
	overviewFile, err := application.ArchiveService.MakeArchiveHTML(data, application.Config.Digest.Compress)
	if err != nil {
		log.Printf("Error generating archive HTML: %v", err)
		return nil, nil, nil, err
	}

	// TODO: Mark as read

	return overviewFile, []*os.File{}, data, nil
}

func RunOnStartup(application *app.App) {
	if application.Config.Digest.RunOnStartup {
		_, _, _, _ = ProcessAndSendDigest(application)
	}
}
