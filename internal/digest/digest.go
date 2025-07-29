package digest

import (
	"context"
	"encoding/json"
	"fmt"
	"miniflux-digest/internal/llm"
	"miniflux-digest/internal/models"
	"sort"
	"time"
	"strings"

	"google.golang.org/genai"
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
		Summary:			summary,
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

type LLMResponse struct {
	Summary string `json:"summary"`
	Groups  []struct {
		Title   string `json:"title"`
		Entries []int  `json:"entries"`
	} `json:"groups"`
}

type llmEntry struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Content   string `json:"content"`
	FeedTitle string `json:"feed_title"`
}

const llmPrompt = `You are a personal news assistant that helps organize feeds from various news websites, status updates and link aggregators. Given a list of feed entries, your task is to:
1. provide a concise one paragraph concise 'summary' of the most important entries. it should be quick to read and informative.
2. provide a small number of 'groups' to organize each entry into. the groups could be a topic, theme or keyword relevant to the collection of entries. groups and entries have a 1 to 1 mapping.
3. a 'group.title' is the name of the group and 'group.entries' are a list of entry ids from the entry objects provided below.
4. please do rank the list of entry ids in the group by order of importance, relevance or interest.

Return the response as a JSON object according to the desired responseSchema.

Below are the entries and other relevant metadata for this task:
-----------------

`

var llmResponseSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"summary": {
			Type: genai.TypeString,
		},
		"groups": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"title": {
						Type: genai.TypeString,
					},
					"entries": {
						Type: genai.TypeArray,
						Items: &genai.Schema{
							Type: genai.TypeInteger,
						},
					},
				},
			},
		},
	},
}

func (g *LLMGrouper) GroupEntries(entries *miniflux.Entries) ([]*models.EntryGroup, string) {
	llmEntries := make([]llmEntry, len(*entries))
	for i, entry := range *entries {
		llmEntries[i] = llmEntry{
			ID:        entry.ID,
			Title:     entry.Title,
			URL:       entry.URL,
			Content:   entry.Content,
			FeedTitle: entry.Feed.Title,
		}
	}

	entriesJSON, err := json.MarshalIndent(llmEntries, "", "  ")
	if err != nil {
		return (&DayGrouper{}).GroupEntries(entries)
	}

	prompt := llmPrompt + string(entriesJSON)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	llmResponse, err := g.LLMService.GenerateContent(ctx, prompt, llmResponseSchema)

	if err != nil {
		fmt.Printf("LLM service failed, falling back to day grouping: %v\n", err)
		return (&DayGrouper{}).GroupEntries(entries)
	}

	var response LLMResponse
	if err := json.Unmarshal([]byte(llmResponse), &response); err != nil {
		fmt.Printf("Failed to parse LLM response, falling back to day grouping: %v\n", err)
		return (&DayGrouper{}).GroupEntries(entries)
	}

	entryMap := make(map[int64]*miniflux.Entry)
	for _, entry := range *entries {
		entryMap[entry.ID] = entry
	}

	var entryGroups []*models.EntryGroup
	groupedEntryIDs := make(map[int64]bool)

	for _, groupData := range response.Groups {
		var groupEntries []*miniflux.Entry
		for _, entryID := range groupData.Entries {
			if entry, ok := entryMap[int64(entryID)]; ok {
				groupEntries = append(groupEntries, entry)
				groupedEntryIDs[int64(entryID)] = true
			}
		}
		entryGroups = append(entryGroups, &models.EntryGroup{
			Title:   groupData.Title,
			Entries: groupEntries,
		})
	}

	var ungroupedEntries []*miniflux.Entry
	for _, entry := range *entries {
		if !groupedEntryIDs[entry.ID] {
			ungroupedEntries = append(ungroupedEntries, entry)
		}
	}

	if len(ungroupedEntries) > 0 {
		foundUncategorized := false
		for _, group := range entryGroups {
			if strings.EqualFold(group.Title, "Uncategorized") {
				group.Entries = append(group.Entries, ungroupedEntries...)
				foundUncategorized = true
				break
			}
		}
		if !foundUncategorized {
			entryGroups = append(entryGroups, &models.EntryGroup{
				Title:   "Uncategorized",
				Entries: ungroupedEntries,
			})
		}
	}

	return entryGroups, response.Summary
}
