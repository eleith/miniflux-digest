package digest

import (
	"miniflux-digest/internal/app/services"
	"miniflux-digest/internal/models"
	"sort"
	"time"
)



func GroupEntries(entries []*models.Entry, groupBy string) []*primaryGroup {
	groups := make(map[int64]*primaryGroup)

	for _, entry := range entries {
		var groupID int64
		var groupTitle string

		if groupBy == "feed" {
			groupID = entry.FeedID
			groupTitle = entry.FeedTitle
		} else { // category
			groupID = entry.GroupID
			groupTitle = entry.GroupTitle
			if groupTitle == "" {
				groupTitle = "Uncategorized"
			}
		}

		if _, ok := groups[groupID]; !ok {
			groups[groupID] = &primaryGroup{
				ID:    groupID,
				Title: groupTitle,
			}
		}
		groups[groupID].Entries = append(groups[groupID].Entries, entry)
	}

	var result []*primaryGroup
	for _, pg := range groups {
		result = append(result, pg)
	}

	return result
}

func NewDigestService(llmService services.LLMService) services.DigestService {
	return &digestServiceImpl{llmService: llmService}
}

type digestServiceImpl struct {
	llmService services.LLMService
}

func (s *digestServiceImpl) BuildDigestData(entries []*models.Entry, icons map[int64]*models.FeedIcon, view string, minifluxHost string) *models.OverviewTemplateData {
	iconsSlice := make([]*models.FeedIcon, 0, len(icons))
	for _, icon := range icons {
		iconsSlice = append(iconsSlice, icon)
	}

	var allPrimaryGroups []*models.PrimaryGroupDigestData
	var overallDigestSummary *string

	switch view {
	case "date":
		primaryGroups := GroupByDate(entries)
		allPrimaryGroups, overallDigestSummary = SubGroupByFeed(primaryGroups)
	case "category":
		primaryGroups := GroupEntries(entries, "category")
		allPrimaryGroups, overallDigestSummary = SubGroupByFeed(primaryGroups)
		sort.Slice(allPrimaryGroups, func(i, j int) bool {
			return allPrimaryGroups[i].Title < allPrimaryGroups[j].Title
		})
	case "ai":
		primaryGroups := GroupEntries(entries, "category")
		allPrimaryGroups, overallDigestSummary = SubGroupByAI(primaryGroups, s.llmService)
		sort.Slice(allPrimaryGroups, func(i, j int) bool {
			return allPrimaryGroups[i].Title < allPrimaryGroups[j].Title
		})
	default: // category
		primaryGroups := GroupEntries(entries, "category")
		allPrimaryGroups, overallDigestSummary = SubGroupByFeed(primaryGroups)
		sort.Slice(allPrimaryGroups, func(i, j int) bool {
			return allPrimaryGroups[i].Title < allPrimaryGroups[j].Title
		})
	}

	// Sort entries by date for date and category views
	if view == "date" || view == "category" {
		for _, pg := range allPrimaryGroups {
			for _, sg := range pg.SubGroups {
				sort.Slice(sg.Entries, func(i, j int) bool {
					return sg.Entries[i].Date.Before(sg.Entries[j].Date)
				})
			}
		}
	}

	for _, primaryGroup := range allPrimaryGroups {
		primaryGroupTotalEntries := 0
		primaryGroupUniqueFeedIDs := make(map[int64]bool)
		for _, subGroup := range primaryGroup.SubGroups {
			primaryGroupTotalEntries += len(subGroup.Entries)
			for _, entry := range subGroup.Entries {
				primaryGroupUniqueFeedIDs[entry.FeedID] = true
			}
		}
		primaryGroup.TotalEntries = primaryGroupTotalEntries
		primaryGroup.TotalFeeds = len(primaryGroupUniqueFeedIDs)
	}

	var allEntryGroups []*models.EntryGroup
	for _, primaryGroup := range allPrimaryGroups {
		allEntryGroups = append(allEntryGroups, primaryGroup.SubGroups...)
	}

	uniqueFeedIDs := make(map[int64]bool)
	for _, entry := range entries {
		uniqueFeedIDs[entry.FeedID] = true
	}

	return &models.OverviewTemplateData{
		Entries:         entries,
		GeneratedDate:   time.Now(),
		FeedIcons:       iconsSlice,
		PrimaryGroups:   allPrimaryGroups,
		EntryGroups:     allEntryGroups,
		OverviewSummary: safeDerefString(overallDigestSummary),
		TotalEntries:    len(entries),
		TotalFeeds:      len(uniqueFeedIDs),
		MinifluxHost:    minifluxHost,
	}
}

func safeDerefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
