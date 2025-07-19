package category

import (
	"log"
	"time"

	miniflux "miniflux.app/v2/client"
)

type FeedIcon struct {
	FeedID int64
	Data   string
}

type CategoryData struct {
	Category      *miniflux.Category
	Entries       *miniflux.Entries
	GeneratedDate time.Time
	FeedIcons     []*FeedIcon
}

func MarkAsRead(client *miniflux.Client, category *miniflux.Category) {
	err := client.MarkCategoryAsRead(category.ID)

	if err != nil {
		log.Printf("Failed to mark category %q as read: %v", category.Title, err)
	}
}

func fetchData(client *miniflux.Client, category *miniflux.Category) (CategoryData, error) {
	entriesResult, err := client.CategoryEntries(category.ID, &miniflux.Filter{Status: miniflux.EntryStatusUnread})

	if err != nil {
		return CategoryData{}, err
	}

	entries := entriesResult.Entries
	feeds, err := client.CategoryFeeds(category.ID)
	feedIcons := []*FeedIcon{}

	if err != nil {
		return CategoryData{}, err
	}

	for _, feed := range feeds {
		feedIcon, err := client.FeedIcon(feed.ID)

		if err != nil {
			continue
		}

		feedIconForTemplate := &FeedIcon{
			FeedID: feed.ID,
			Data:   feedIcon.Data,
		}
		feedIcons = append(feedIcons, feedIconForTemplate)
	}

	return CategoryData{
		Category:      category,
		Entries:       &entries,
		GeneratedDate: time.Now(),
		FeedIcons:     feedIcons,
	}, nil
}

func StreamData(client *miniflux.Client) <-chan *CategoryData {
	out := make(chan *CategoryData)

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
