package digest

import (
	
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/utils"
	"sort"
	"time"
)

const (
	DayGroupLayout      = "2006-01-02"
	DayGroupTitleLayout = "Jan 2, 2006"
)

func SubGroupByDay(pgs []*primaryGroup) ([]*models.PrimaryGroupDigestData, *string) {
	var allPrimaryGroups []*models.PrimaryGroupDigestData

	for _, pg := range pgs {
		entryGroupsMap := make(map[string]*models.EntryGroup)
		for _, entry := range pg.Entries {
			dateKey := entry.Date.Format(DayGroupLayout)
			if _, ok := entryGroupsMap[dateKey]; !ok {
				entryGroupsMap[dateKey] = &models.EntryGroup{
					Title:   entry.Date.Format(DayGroupTitleLayout),
					Entries: []*models.Entry{},
					Slug:    utils.Slugify(entry.Date.Format(DayGroupTitleLayout)),
				}
			}
			entryGroupsMap[dateKey].Entries = append(entryGroupsMap[dateKey].Entries, entry)
		}

		sortedEntryGroups := make([]*models.EntryGroup, 0, len(entryGroupsMap))
		for _, group := range entryGroupsMap {
			uniqueFeedIDs := make(map[int64]bool)
			for _, entry := range group.Entries {
				uniqueFeedIDs[entry.FeedID] = true
			}
			group.TotalEntries = len(group.Entries)
			group.TotalFeeds = len(uniqueFeedIDs)
			sortedEntryGroups = append(sortedEntryGroups, group)
		}

		sort.Slice(sortedEntryGroups, func(i, j int) bool {
			iDate, _ := time.Parse("Jan 2, 2006", sortedEntryGroups[i].Title)
			jDate, _ := time.Parse("Jan 2, 2006", sortedEntryGroups[j].Title)
			return iDate.Before(jDate)
		})

		for _, group := range sortedEntryGroups {
			sort.Slice(group.Entries, func(i, j int) bool {
				return group.Entries[i].Date.Before(group.Entries[j].Date)
			})
		}

		

		allPrimaryGroups = append(allPrimaryGroups, &models.PrimaryGroupDigestData{
			ID:        pg.ID,
			Title:     pg.Title,
			Slug:      utils.Slugify(pg.Title),
			SubGroups: sortedEntryGroups,
		})
	}

	return allPrimaryGroups, nil
}
