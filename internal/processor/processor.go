package processor

import (
	"log"

	"miniflux-digest/internal/app"
	"miniflux-digest/internal/digest"
)

func ProcessAndSendDigest(application *app.App) {
	log.Println("Starting to process digest...")

	entries, err := application.MinifluxClientService.GetAllUnreadEntries()
	if err != nil {
		log.Printf("Error getting all unread entries: %v", err)
		return
	}

	log.Printf("Found %d unread entries", len(*entries))

	groups := digest.GroupEntries(entries, application.Config.Digest.GroupBy)
	log.Printf("Grouped entries into %d groups", len(groups))

	// TODO: Call LLM
	// TODO: Generate Overview Digest
	// TODO: Generate Group Digests
	// TODO: Send Email
	// TODO: Mark as read
}

func RunOnStartup(application *app.App) {
	if application.Config.Digest.RunOnStartup {
		ProcessAndSendDigest(application)
	}
}
