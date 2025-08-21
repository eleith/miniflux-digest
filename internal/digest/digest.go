package digest

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"miniflux-digest/internal/app/services"
	"miniflux-digest/internal/llm"
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/utils"
	"sort"
	"strings"
	"time"

	"google.golang.org/genai"
	miniflux "miniflux.app/v2/client"
)

const (
	LLMTimeout        = 5 * time.Minute
	DayGroupLayout    = "2006-01-02"
	DayGroupTitleLayout = "Jan 2, 2006"
)

func GroupEntries(entries *miniflux.Entries, groupBy string) map[string][]*miniflux.Entry {
	groups := make(map[string][]*miniflux.Entry)

	switch groupBy {
	case "feed":
		for _, entry := range *entries {
			groups[entry.Feed.Title] = append(groups[entry.Feed.Title], entry)
		}
	default:
		for _, entry := range *entries {
			var groupName string
			if entry.Feed.Category != nil {
				groupName = entry.Feed.Category.Title
			} else {
				groupName = "Uncategorized"
			}
			groups[groupName] = append(groups[groupName], entry)
		}
	}

	return groups
}

func NewSubGrouper(subGroupBy string, llmService services.LLMService) SubGrouper {
	switch subGroupBy {
	case "ai":
		return &LLMGrouper{LLMService: llmService}
	case "feed":
		return &FeedGrouper{}
	default:
		return &DayGrouper{}
	}
}

func NewDigestService(llmService services.LLMService) services.DigestService {
	return &digestServiceImpl{llmService: llmService}
}

type digestServiceImpl struct {
	llmService services.LLMService
}



func (s *digestServiceImpl) BuildDigestData(category *miniflux.Category, entries *miniflux.Entries, icons map[int64]*models.FeedIcon, groupBy string, subGroupBy string, sortBy string, minifluxHost string) *models.OverviewTemplateData {
	// Convert map to slice
	iconsSlice := make([]*models.FeedIcon, 0, len(icons))
	for _, icon := range icons {
		iconsSlice = append(iconsSlice, icon)
	}

	// Group entries by primary grouping (category or feed)
	primaryGroups := GroupEntries(entries, groupBy)

	var allPrimaryGroups []*models.PrimaryGroupDigestData
	var overallSummary string

	// Process each primary group
	for primaryGroupName, primaryGroupEntries := range primaryGroups {
		// Create a miniflux.Entries wrapper for the primary group entries
		// This is needed because SubGrouper.GroupEntries expects *miniflux.Entries
		primaryGroupEntriesWrapper := (*miniflux.Entries)(&primaryGroupEntries)

		// Apply sub-grouping and sorting
		grouper := NewSubGrouper(subGroupBy, s.llmService)
		subEntryGroups, subSummary := grouper.GroupEntries(primaryGroupEntriesWrapper)

		// Add primary group title to sub-groups if not already present
		if subGroupBy != "category" && subGroupBy != "feed" {
			for _, seg := range subEntryGroups {
				seg.Title = fmt.Sprintf("%s - %s", primaryGroupName, seg.Title)
			}
		}

		allPrimaryGroups = append(allPrimaryGroups, &models.PrimaryGroupDigestData{
			Title:     primaryGroupName,
			Slug:      utils.Slugify(primaryGroupName),
			SubGroups: subEntryGroups,
			Summary:   subSummary,
		})
		overallSummary += subSummary + "\n"
	}

	// Sort the top-level groups (e.g., by primary group name)
	sort.Slice(allPrimaryGroups, func(i, j int) bool {
		return allPrimaryGroups[i].Title < allPrimaryGroups[j].Title
	})

	// Flatten the sub-groups into a single list for the email template
	var allEntryGroups []*models.EntryGroup
	for _, primaryGroup := range allPrimaryGroups {
		allEntryGroups = append(allEntryGroups, primaryGroup.SubGroups...)
	}

	return &models.OverviewTemplateData{
		Category:        category,
		Entries:         entries,
		GeneratedDate:   time.Now(),
		FeedIcons:       iconsSlice,
		PrimaryGroups:   allPrimaryGroups,
		EntryGroups:     allEntryGroups,
		OverviewSummary: overallSummary,
		MinifluxHost:    minifluxHost,
	}
}

type SubGrouper interface {
	GroupEntries(entries *miniflux.Entries) ([]*models.EntryGroup, string)
}

type DayGrouper struct{}

func (g *DayGrouper) GroupEntries(entries *miniflux.Entries) ([]*models.EntryGroup, string) {
	entryGroupsMap := make(map[string]*models.EntryGroup)
	for _, entry := range *entries {
		dateKey := entry.Date.Format(DayGroupLayout)
		if _, ok := entryGroupsMap[dateKey]; !ok {
			entryGroupsMap[dateKey] = &models.EntryGroup{
				Title:   entry.Date.Format(DayGroupTitleLayout),
				Entries: []*miniflux.Entry{},
				Slug:    utils.Slugify(entry.Date.Format(DayGroupTitleLayout)),
			}
		}
		entryGroupsMap[dateKey].Entries = append(entryGroupsMap[dateKey].Entries, entry)
	}

	// Convert map to sorted slice of EntryGroups
	sortedEntryGroups := make([]*models.EntryGroup, 0, len(entryGroupsMap))
	for _, group := range entryGroupsMap {
		sortedEntryGroups = append(sortedEntryGroups, group)
	}

	// Sort groups by date (older to newer)
	sort.Slice(sortedEntryGroups, func(i, j int) bool {
		// Dates are stored as strings, so we need to parse them back to time.Time
		iDate, _ := time.Parse("Jan 2, 2006", sortedEntryGroups[i].Title)
		jDate, _ := time.Parse("Jan 2, 2006", sortedEntryGroups[j].Title)
		return iDate.Before(jDate)
	})

	// Sort entries within each group by date (older to newer)
	for _, group := range sortedEntryGroups {
		sort.Slice(group.Entries, func(i, j int) bool {
			return group.Entries[i].Date.Before(group.Entries[j].Date)
		})
	}

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
				Slug:    utils.Slugify(entry.Feed.Title),
			}
		}
		entryGroupsMap[entry.FeedID].Entries = append(entryGroupsMap[entry.FeedID].Entries, entry)
	}

	// Convert map to slice of EntryGroups
	entryGroups := make([]*models.EntryGroup, 0, len(entryGroupsMap))
	for _, group := range entryGroupsMap {
		entryGroups = append(entryGroups, group)
	}

	// Sort groups by feed title (alphabetically)
	sort.Slice(entryGroups, func(i, j int) bool {
		return entryGroups[i].Title < entryGroups[j].Title
	})

	// Sort entries within each group by date (older to newer)
	for _, group := range entryGroups {
		sort.Slice(group.Entries, func(i, j int) bool {
			return group.Entries[i].Date.Before(group.Entries[j].Date)
		})
	}

	return entryGroups, fmt.Sprintf("You have %d entries from %d feeds", len(*entries), len(entryGroups))
}

type LLMGrouper struct {
	LLMService services.LLMService
}

type llmEntry struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Content   string `json:"content"`
	FeedTitle string `json:"feed_title"`
}

const llmPrompt = `You are an expert content curator with a talent for identifying the most important and interesting information from a large volume of content. Your primary goal is to save the user time by providing a high-level, insightful overview of their content feeds.

The content can come from a variety of sources, including news sites, blogs, forums like Reddit, and social media like Bluesky. Your tone should be that of a helpful assistant, not a news editor.

Given the following list of entries, your task is to perform two distinct functions:

1.  **Create an insightful 'summary':**
    *   This must be a single, concise paragraph.
    *   Your task is to identify and highlight the most significant themes, trends, or critical events from the provided entries.
    *   **Do not** simply list the topics of every article. Instead, synthesize a compelling narrative. For example, you might point out a recurring theme across several articles or highlight a single entry if it represents a major, must-read development. Your summary should be opinionated and selective, giving the user a clear sense of what matters most.

2.  **Generate intelligent 'groups' for all entries:**
    *   Your goal is to cluster the entries into a set of meaningful, thematic groups that will help the user quickly navigate the content.
    *   The number of groups should be driven by the content itself. **Do not create a group for a single entry unless it represents a major, unique event.** The ideal number of groups is one that best helps a user skim the content. A group can contain many entries if they are all highly related.
    *   Group titles should be short, descriptive, and useful for skimming (e.g., "AI Industry News," "Project Updates," "Global Politics").
    *   Within each group, you must rank the 'entries' by importance, with the most significant or actionable item appearing first.
    *   The goal of grouping is to help with reading, so if a few entries are about some similar topic, those are worth grouping because the user can just read one and skim the rest.

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

	ctx, cancel := context.WithTimeout(context.Background(), LLMTimeout)
	defer cancel()

	llmResponse, err := g.LLMService.GenerateContent(ctx, prompt, llmResponseSchema)

	if err != nil {
		log.Printf("LLM service failed, falling back to day grouping: %v\n", err)
		return (&DayGrouper{}).GroupEntries(entries)
	}

	var response llm.LLMResponse
	if err := json.Unmarshal([]byte(llmResponse), &response); err != nil {
		log.Printf("Failed to parse LLM response, falling back to day grouping: %v\n", err)
		return (&DayGrouper{}).GroupEntries(entries)
	}

	entryMap := make(map[int64]*miniflux.Entry)
	for _, entry := range *entries {
		entryMap[entry.ID] = entry
	}


	var entryGroups []*models.EntryGroup
	groupedEntryIDs := make(map[int64]bool)

	for _, groupData := range response.GroupSummaries {
		var groupEntries []*miniflux.Entry
		for _, entryID := range groupData.EntryIDs {
			if _, exists := groupedEntryIDs[int64(entryID)]; !exists {
				if entry, ok := entryMap[int64(entryID)]; ok {
					groupEntries = append(groupEntries, entry)
					groupedEntryIDs[int64(entryID)] = true
				}
			}
		}
		
		entryGroups = append(entryGroups, &models.EntryGroup{
			Title:   groupData.Title,
			Entries: groupEntries,
			Slug:    utils.Slugify(groupData.Title),
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
				Slug:    utils.Slugify("Uncategorized"),
			})
		}
	}

	feedIDs := make(map[int64]bool)
	for _, entry := range *entries {
		feedIDs[entry.FeedID] = true
	}
	
	statsSummary := fmt.Sprintf("You have %d entries from %d feeds.", len(*entries), len(feedIDs))

	return entryGroups, fmt.Sprintf("%s\n\n%s", response.OverviewSummary, statsSummary)
}
