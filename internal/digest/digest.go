package digest

import (
	"fmt"
	"miniflux-digest/internal/llm"
	"miniflux-digest/internal/models"
	"sort"
	"time"

	miniflux "miniflux.app/v2/client"
)

type GroupingType string

const (
	GroupingTypeDay  GroupingType = "day"
	GroupingTypeFeed GroupingType = "feed"
)

func (gt GroupingType) String() string {
	return string(gt)
}

type DigestService struct{
	LLMService llm.LLMService
}

func NewDigestService(llmService llm.LLMService) *DigestService {
	return &DigestService{LLMService: llmService}
}

func NewGrouper(groupBy GroupingType, llmService llm.LLMService) Grouper {
	switch groupBy {
	case "ai":
		return &LLMGrouper{LLMService: llmService}
	case GroupingTypeFeed:
		return &FeedGrouper{}
	default:
		return &DayGrouper{}
	}
}

func (s *DigestService) BuildDigestData(category *miniflux.Category, entries *miniflux.Entries, icons map[int64]*models.FeedIcon, groupBy GroupingType) *models.HTMLTemplateData {
	// Convert map to slice
	iconsSlice := make([]*models.FeedIcon, 0, len(icons))
	for _, icon := range icons {
		iconsSlice = append(iconsSlice, icon)
	}

	// Group entries
	grouper := NewGrouper(groupBy, s.LLMService)
	entryGroups, summary := grouper.GroupEntries(entries)

	return &models.HTMLTemplateData{
		Category:      category,
		Entries:       entries,
		GeneratedDate: time.Now(),
		FeedIcons:     iconsSlice,
		EntryGroups:   entryGroups,
		Summary:			 summary,
	}
}

type Grouper interface {
	GroupEntries(entries *miniflux.Entries) ([]*models.EntryGroup, string)
}

type DayGrouper struct{}

func (g *DayGrouper) GroupEntries(entries *miniflux.Entries) ([]*models.EntryGroup, string) {
	entryGroupsMap := make(map[string]*models.EntryGroup)
	for _, entry := range *entries {
		dateKey := entry.Date.Format("2006-01-02")
		if _, ok := entryGroupsMap[dateKey]; !ok {
			entryGroupsMap[dateKey] = &models.EntryGroup{
				Title:   entry.Date.Format("Jan 2, 2006"),
				Entries: []*miniflux.Entry{},
			}
		}
		entryGroupsMap[dateKey].Entries = append(entryGroupsMap[dateKey].Entries, entry)
	}

	// Convert map to sorted slice of EntryGroups
	sortedEntryGroups := make([]*models.EntryGroup, 0, len(entryGroupsMap))
	for _, group := range entryGroupsMap {
		// Sort entries within each group by date (older to newer)
		sort.Slice(group.Entries, func(i, j int) bool {
			return group.Entries[i].Date.Before(group.Entries[j].Date)
		})
		sortedEntryGroups = append(sortedEntryGroups, group)
	}

	// Sort groups by date (older to newer)
	sort.Slice(sortedEntryGroups, func(i, j int) bool {
		// Dates are stored as strings, so we need to parse them back to time.Time
		iDate, _ := time.Parse("Jan 2, 2006", sortedEntryGroups[i].Title)
		jDate, _ := time.Parse("Jan 2, 2006", sortedEntryGroups[j].Title)
		return iDate.Before(jDate)
	})

	return sortedEntryGroups, fmt.Sprintf("You have %d entries from %d different days", len(*entries), len(sortedEntryGroups))
}

type FeedGrouper struct{}

func (g *FeedGrouper) GroupEntries(entries *miniflux.Entries) ([]*models.EntryGroup, string) {
	entryGroupsMap := make(map[int64]*models.EntryGroup)
	for _, entry := range *entries {
		if _, ok := entryGroupsMap[entry.FeedID]; !ok {
			entryGroupsMap[entry.FeedID] = &models.EntryGroup{
				Title:   entry.Feed.Title,
				Entries: []*miniflux.Entry{},
			}
		}
		entryGroupsMap[entry.FeedID].Entries = append(entryGroupsMap[entry.FeedID].Entries, entry)
	}

	// Convert map to slice of EntryGroups
	entryGroups := make([]*models.EntryGroup, 0, len(entryGroupsMap))
	for _, group := range entryGroupsMap {
		// Sort entries within each group by date (older to newer)
		sort.Slice(group.Entries, func(i, j int) bool {
			return group.Entries[i].Date.Before(group.Entries[j].Date)
		})
		entryGroups = append(entryGroups, group)
	}

	// Sort groups by feed title (alphabetically)
	sort.Slice(entryGroups, func(i, j int) bool {
		return entryGroups[i].Title < entryGroups[j].Title
	})

	return entryGroups, fmt.Sprintf("You have %d entries from %d feeds", len(*entries), len(entryGroups))
}

type LLMGrouper struct {
	LLMService llm.LLMService
}

func (g *LLMGrouper) GroupEntries(entries *miniflux.Entries) ([]*models.EntryGroup, string) {
	// This is a simple prompt that will be optimized in a later branch.
	prompt := "Please provide a summary of the following entries, and then group them by topic. For each group, provide a title and the IDs of the entries in that group.\n\n"
	for _, entry := range *entries {
		prompt += fmt.Sprintf("- %s\n", entry.Title)
	}

	// In a real implementation, we would parse the LLM's response and create EntryGroups and a summary.
	// For now, we'll just return the entries grouped by day as a fallback.
	return (&DayGrouper{}).GroupEntries(entries)
}
