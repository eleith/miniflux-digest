package digest

import (
	"miniflux-digest/internal/models"
	"sort"
	miniflux "miniflux.app/v2/client"
	"time"
)

type DigestService struct{}

func NewDigestService() *DigestService {
	return &DigestService{}
}

func (s *DigestService) BuildDigestData(category *miniflux.Category, entries *miniflux.Entries, icons map[int64]*models.FeedIcon) *models.HTMLTemplateData {
	// Convert map to slice
	iconsSlice := make([]*models.FeedIcon, 0, len(icons))
	for _, icon := range icons {
		iconsSlice = append(iconsSlice, icon)
	}

	// Group entries by day
	entryGroups := make(map[string]*models.EntryGroup)
	for _, entry := range *entries {
		dateKey := entry.Date.Format("2006-01-02")
		if _, ok := entryGroups[dateKey]; !ok {
			entryGroups[dateKey] = &models.EntryGroup{
				Date:  entry.Date,
				Entries: []*miniflux.Entry{},
			}
		}
		entryGroups[dateKey].Entries = append(entryGroups[dateKey].Entries, entry)
	}

	// Convert map to sorted slice of EntryGroups
	sortedEntryGroups := make([]*models.EntryGroup, 0, len(entryGroups))
	for _, group := range entryGroups {
		sortedEntryGroups = append(sortedEntryGroups, group)
	}

	// Sort groups by date (most recent first)
	sort.Slice(sortedEntryGroups, func(i, j int) bool {
		return sortedEntryGroups[i].Date.After(sortedEntryGroups[j].Date)
	})

	return &models.HTMLTemplateData{
		Category:      category,
		Entries:       entries,
		GeneratedDate: time.Now(),
		FeedIcons:     iconsSlice,
		EntryGroups:   sortedEntryGroups,
	}
}
