package group_by

import (
	"miniflux-digest/internal/models"
)

func GroupByCategory(entries []*models.Entry) []*models.PrimaryGroup {
	groups := make(map[int64]*models.PrimaryGroup)

	for _, entry := range entries {
		var groupID int64
		var groupTitle string

		groupID = entry.GroupID
		groupTitle = entry.GroupTitle
		if groupTitle == "" {
			groupTitle = "Uncategorized"
		}

		if _, ok := groups[groupID]; !ok {
			groups[groupID] = &models.PrimaryGroup{
				ID:    groupID,
				Title: groupTitle,
			}
		}
		groups[groupID].Entries = append(groups[groupID].Entries, entry)
	}

	var result []*models.PrimaryGroup
	for _, pg := range groups {
		result = append(result, pg)
	}

	return result
}
