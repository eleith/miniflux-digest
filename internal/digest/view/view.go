package view

import (
	"context"
	"miniflux-digest/internal/app/services"
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/utils"
	"sort"
	"time"
)

func BuildDigestDataForView(entries []*models.Entry, icons map[int64]*models.FeedIcon, viewType string, minifluxHost string, digestHost string, llmService services.LLMService, categories []config.ConfigCategory, digestTitle string) *models.OverviewTemplateData {
	iconsSlice := make([]*models.FeedIcon, 0, len(icons))
	for _, icon := range icons {
		iconsSlice = append(iconsSlice, icon)
	}

	var resultPrimaryGroups []*models.PrimaryGroupDigestData
	

	switch viewType {
	case "date":
		resultPrimaryGroups = BuildDigestDataByDate(entries)
	case "category", "default":
		resultPrimaryGroups = BuildDigestDataByCategory(entries)
	case "ai":
		resultPrimaryGroups = BuildDigestDataByAI(entries, context.Background(), llmService, categories)
	}

	if viewType == "date" || viewType == "category" {
		for _, pg := range resultPrimaryGroups {
			for _, sg := range pg.SubGroups {
				sort.Slice(sg.Entries, func(i, j int) bool {
					return sg.Entries[i].Date.Before(sg.Entries[j].Date)
				})
			}
		}
	}

	for _, pg := range resultPrimaryGroups {
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
	for _, pg := range resultPrimaryGroups {
		allEntryGroups = append(allEntryGroups, pg.SubGroups...)
	}

	uniqueFeedIDs := make(map[int64]bool)
	for _, entry := range entries {
		uniqueFeedIDs[entry.FeedID] = true
	}

	return &models.OverviewTemplateData{
		DigestTitle:     digestTitle,
		DigestSlug:      utils.Slugify(digestTitle),
		Entries:         entries,
		GeneratedDate:   time.Now(),
		FeedIcons:       iconsSlice,
		PrimaryGroups:   resultPrimaryGroups,
		EntryGroups:     allEntryGroups,
		
		TotalEntries:    len(entries),
		TotalFeeds:      len(uniqueFeedIDs),
		MinifluxHost:    minifluxHost,
		DigestHost:      digestHost,
	}
}


