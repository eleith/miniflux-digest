package category

import (
	"log"
	"time"

	"miniflux-digest/internal/app"
	"miniflux-digest/internal/models"
	miniflux "miniflux.app/v2/client"
)

type CategoryServiceImpl struct{}

var _ app.CategoryService = (*CategoryServiceImpl)(nil)

func (s *CategoryServiceImpl) StreamData(client app.MinifluxClientService) <-chan *models.CategoryData {
	out := make(chan *models.CategoryData)

	go func() {
		defer close(out)

		categories, err := client.Categories()

		if err != nil {
			log.Fatalf("Streamer failed to fetch categories: %v", err)
			return
		}

		for _, category := range categories {
			data, err := fetchData(client, category)
			if err != nil {
				log.Printf("Streamer failed to fetch data for category %q: %v", category.Title, err)
				continue
			}

			out <- &data
		}
	}()

	return out
}

func fetchData(client app.MinifluxClientService, category *miniflux.Category) (models.CategoryData, error) {
	entriesResult, err := client.CategoryEntries(category.ID, &miniflux.Filter{Status: miniflux.EntryStatusUnread})

	if err != nil {
		return models.CategoryData{}, err
	}

	entries := entriesResult
	feeds, err := client.CategoryFeeds(category.ID)
	feedIcons := []*models.FeedIcon{}

	if err != nil {
		return models.CategoryData{}, err
	}

	for _, feed := range feeds {
		feedIcon, err := client.FeedIcon(feed.ID)

		if err != nil {
			continue
		}

		feedIconForTemplate := &models.FeedIcon{
			FeedID: feed.ID,
			Data:   feedIcon.Data,
		}
		feedIcons = append(feedIcons, feedIconForTemplate)
	}

	return models.CategoryData{
		Category:      category,
		Entries:       entries,
		GeneratedDate: time.Now(),
		FeedIcons:     feedIcons,
	}, nil
}
