package digest

import (
	
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/utils"
	"sort"
)

func SubGroupByFeed(pgs []*primaryGroup) ([]*models.PrimaryGroupDigestData, *string) {
	var allPrimaryGroups []*models.PrimaryGroupDigestData

	for _, pg := range pgs {
		entryGroupsMap := make(map[int64]*models.EntryGroup)
		for _, entry := range pg.Entries {
			if _, ok := entryGroupsMap[entry.FeedID]; !ok {
				entryGroupsMap[entry.FeedID] = &models.EntryGroup{
					Title:   entry.FeedTitle,
					Entries: []*models.Entry{},
					Slug:    utils.Slugify(entry.FeedTitle),
				}
			}
			entryGroupsMap[entry.FeedID].Entries = append(entryGroupsMap[entry.FeedID].Entries, entry)
		}

		entryGroups := make([]*models.EntryGroup, 0, len(entryGroupsMap))
		for _, group := range entryGroupsMap {
			uniqueFeedIDs := make(map[int64]bool)
			for _, entry := range group.Entries {
				uniqueFeedIDs[entry.FeedID] = true
			}
			group.TotalEntries = len(group.Entries)
			group.TotalFeeds = len(uniqueFeedIDs)
			entryGroups = append(entryGroups, group)
		}

		sort.Slice(entryGroups, func(i, j int) bool {
			return entryGroups[i].Title < entryGroups[j].Title
		})

		for _, group := range entryGroups {
			sort.Slice(group.Entries, func(i, j int) bool {
				return group.Entries[i].Date.Before(group.Entries[j].Date)
			})
		}

		

		allPrimaryGroups = append(allPrimaryGroups, &models.PrimaryGroupDigestData{
			ID:        pg.ID,
			Title:     pg.Title,
			Slug:      utils.Slugify(pg.Title),
			SubGroups: entryGroups,
		})
	}

	return allPrimaryGroups, nil
}
