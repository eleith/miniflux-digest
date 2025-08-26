package processor

import (
	"log"
	"os"

	"miniflux-digest/internal/app"
	"miniflux-digest/internal/models"

	miniflux "miniflux.app/v2/client"
)

func ProcessDigest(application *app.App) (*os.File, []*os.File, *models.OverviewTemplateData, error) {
	log.Println("Starting to process digest...")

	entries, err := application.MinifluxClientService.GetAllUnreadEntries()
	if err != nil {
		log.Printf("Error getting all unread entries: %v", err)
		return nil, nil, nil, err
	}

	log.Printf("Found %d unread entries", len(entries))

	icons := make(map[int64]*models.FeedIcon)
	for _, entry := range entries {
		if _, ok := icons[entry.FeedID]; !ok {
			icon, err := application.MinifluxClientService.FeedIcon(entry.FeedID)
			if err != nil {
				log.Printf("Warning: failed to fetch icon for feed %d: %v", entry.FeedID, err)
				continue
			}
			icons[entry.FeedID] = &models.FeedIcon{
				FeedID: icon.FeedID,
				Data:   icon.Data,
			}
		}
	}

	data := application.DigestService.BuildDigestData(
		entries,
		icons,
		application.Config.Digest.GroupBy,
		application.Config.Digest.SubGroupBy,
		application.Config.Digest.SortBy,
		application.Config.Miniflux.Host,
	)

	overviewFile, groupedEntryFiles, err := application.ArchiveService.MakeArchiveHTML(data, application.Config.Digest.Compress)
	if err != nil {
		log.Printf("Error generating archive HTML: %v", err)
		return nil, nil, nil, err
	}

	if application.Config.Digest.MarkAsRead {
		var entryIDs []int64
		for _, entry := range entries {
			entryIDs = append(entryIDs, entry.ID)
		}

		if len(entryIDs) > 0 {
			err := application.MinifluxClientService.UpdateEntries(entryIDs, miniflux.EntryStatusRead)
			if err != nil {
				log.Printf("Error marking entries as read: %v", err)
			}
		}
	}

	return overviewFile, groupedEntryFiles, data, nil
}
