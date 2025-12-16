package processor

import (
	"log"
	"os"

	"miniflux-digest/internal/app"
	"miniflux-digest/internal/manager"
	"miniflux-digest/internal/models"

	miniflux "miniflux.app/v2/client"
)

func ProcessDigest(application *app.App, digestIndex int) (*os.File, []*os.File, *models.OverviewTemplateData, error) {
	digestConfig := application.Config.Digests[digestIndex]

	entries, err := application.MinifluxClientService.GetAllUnreadEntries()
	if err != nil {
		log.Printf("Error getting all unread entries: %v", err)
		return nil, nil, nil, err
	}

	digestManager, err := manager.NewDigestManager(application.Config.Digests)
	if err != nil {
		log.Printf("Error creating digest manager: %v", err)
		return nil, nil, nil, err
	}

	var digestEntries []*models.Entry
	for _, entry := range entries {
		if digestManager.GetOwningDigest(entry) == digestIndex {
			digestEntries = append(digestEntries, entry)
		}
	}

	// If no entries belong to this digest, we might still want to generate an empty report
	// or just return early. The original logic continued even if empty?
	// The original logic iterated over empty entries -> empty icons -> BuildDigestData -> MakeArchiveHTML.
	// So we should continue with digestEntries.

	icons := make(map[int64]*models.FeedIcon)
	for _, entry := range digestEntries {
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
		digestEntries,
		icons,
		digestConfig.View,
		application.Config.Miniflux.Host,
		digestConfig.Host,
		digestConfig.Categories,
	)

	overviewFile, groupedEntryFiles, err := application.ArchiveService.MakeArchiveHTML(data, digestConfig.Compress)
	if err != nil {
		log.Printf("Error generating archive HTML: %v", err)
		return nil, nil, nil, err
	}

	if digestConfig.MarkAsRead {
		var entryIDs []int64
		for _, entry := range digestEntries {
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