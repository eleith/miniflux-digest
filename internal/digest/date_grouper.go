package digest

import (
	"miniflux-digest/internal/models"
	"sort"
	"time"
)

func GroupByDate(entries []*models.Entry) []*primaryGroup {
	groups := make(map[string]*primaryGroup)

	for _, entry := range entries {
		dateKey := entry.Date.Format("2006-01-02")
		if _, ok := groups[dateKey]; !ok {
			groups[dateKey] = &primaryGroup{
				Title: entry.Date.Format("Jan 2, 2006"),
			}
		}
		groups[dateKey].Entries = append(groups[dateKey].Entries, entry)
	}

	var result []*primaryGroup
	for _, pg := range groups {
		result = append(result, pg)
	}

	sort.Slice(result, func(i, j int) bool {
		iDate, _ := time.Parse("Jan 2, 2006", result[i].Title)
		jDate, _ := time.Parse("Jan 2, 2006", result[j].Title)
		return iDate.Before(jDate)
	})

	return result
}
