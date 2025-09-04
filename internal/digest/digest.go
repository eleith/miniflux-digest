package digest

import (
	"miniflux-digest/internal/app/services"
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/digest/group_by"
	"sort"
	"time"
)



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
		primaryGroups := group_by.GroupByDate(entries)
		allPrimaryGroups, overallDigestSummary = group_by.SubGroupByFeed(primaryGroups)
	case "category":
		primaryGroups := group_by.GroupByCategory(entries)
		allPrimaryGroups, overallDigestSummary = group_by.SubGroupByFeed(primaryGroups)
		sort.Slice(allPrimaryGroups, func(i, j int) bool {
			return allPrimaryGroups[i].Title < allPrimaryGroups[j].Title
		})
	case "ai":
		primaryGroups := group_by.GroupByCategory(entries)
		allPrimaryGroups, overallDigestSummary = group_by.SubGroupByAI(primaryGroups, s.llmService)
		sort.Slice(allPrimaryGroups, func(i, j int) bool {
			return allPrimaryGroups[i].Title < allPrimaryGroups[j].Title
		})
	default:
		primaryGroups := group_by.GroupByCategory(entries)
		allPrimaryGroups, overallDigestSummary = group_by.SubGroupByFeed(primaryGroups)
		sort.Slice(allPrimaryGroups, func(i, j int) bool {
			return allPrimaryGroups[i].Title < allPrimaryGroups[j].Title
		})
	}

	if view == "date" || view == "category" {
		for _, pg := range allPrimaryGroups {
			for _, sg := range pg.SubGroups {
				sort.Slice(sg.Entries, func(i, j int) bool {
					return sg.Entries[i].Date.Before(sg.Entries[j].Date)
				})
			}
		}
	}

	for _, pg := range allPrimaryGroups {
		pgTotalEntries := 0
		pgUniqueFeedIDs := make(map[int64]bool)
		for _, subGroup := range pg.SubGroups {
			pgTotalEntries += len(subGroup.Entries)
			for _, entry := range subGroup.Entries {
				pgUniqueFeedIDs[entry.FeedID] = true
			}
		}
		pg.TotalEntries = pgTotalEntries
		pg.TotalFeeds = len(pgUniqueFeedIDs)
	}

	var allEntryGroups []*models.EntryGroup
	for _, pg := range allPrimaryGroups {
		allEntryGroups = append(allEntryGroups, pg.SubGroups...)
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
