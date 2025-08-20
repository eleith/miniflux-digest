package processor

import (
	"log"
	"os"

	"miniflux-digest/internal/app"
	"miniflux-digest/internal/digest"
	"miniflux-digest/internal/models"

	miniflux "miniflux.app/v2/client"
)

func ProcessAndSendDigest(application *app.App) (*os.File, []*os.File, *models.HTMLTemplateData, error) {
	log.Println("Starting to process digest...")

	entries, err := application.MinifluxClientService.GetAllUnreadEntries()
	if err != nil {
		log.Printf("Error getting all unread entries: %v", err)
		return nil, nil, nil, err
	}

	log.Printf("Found %d unread entries", len(*entries))

	// Generate digest data
	data := application.DigestService.BuildDigestData(
		&miniflux.Category{Title: "All"},
		entries,
		map[int64]*models.FeedIcon{},
		digest.SubGroupingType(application.Config.Digest.SubGroupBy),
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
