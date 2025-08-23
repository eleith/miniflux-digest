package digest

import (
	"context"
	"encoding/json"
	"log"
	"miniflux-digest/internal/app/services"
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/utils"
	"time"

	"google.golang.org/genai"
)

const (
	LLMTimeout = 10 * time.Minute
)

type LLMResponse struct {
	Overview              string `json:"overview"`
	PrimaryGroupSummaries []struct {
		PrimaryGroupID int64  `json:"primary_group_id"`
		Summary        string `json:"summary"`
	} `json:"primary_group_summaries"`
	SubGroups []struct {
		Title    string  `json:"title"`
		EntryIDs []int64 `json:"entry_ids"`
	} `json:"sub_groups"`
}

type LLMGrouper struct {
	LLMService services.LLMService
}

type llmEntry struct {
	ID             int64  `json:"id"`
	Title          string `json:"title"`
	URL            string `json:"url"`
	Content        string `json:"content"`
	FeedTitle      string `json:"feed_title"`
	PrimaryGroupID int64  `json:"primary_group_id"`
}

const llmPrompt = `You will be given a large list of unread entry JSON Objects from a RSS feed reader. these entries help a user follow updates from blogs, social networks, and news sources.

Your task is to generate the metadata according to the instructions below:

	- an 'overview'
		- this is a 3-5 sentence overview for ALL entries, less than 250 words.
		- the overview is really a brief overview. you are not to summarize every single entry
		- The overview should be concise. don't go beyond 5 sentences. 
  	- The overview should help the user quickly understand your overview of all entries in total
		- the overview should help the user decide whether to skim fast or slow and what to narrow in on. 
		- the overview could highlight critical entries or summarize the main topics or both, you decide

	- a list of 'primary_group_summaries'
		- the 'primary_group_id' will be found associated to one or more entry objects
		- the summary is really a brief overview using all entries associated to the primary group
		- you don't need to summarize every single entry
		- entries are only ever are associated with one primary group id
		- this is a 2-3 sentence overview, less than 150 words.
  	- The overview should be concise.
		- the overview should help the user decide whether to skim all entries in this group fast or slow and what to narrow in on. 
		- the overview could highlight critical entries or summarize the main topics or both whatever helps the user skip, skim or find the most important entries

- 'sub_groups':
  - The sub_groups field should contain a list of subgroup objects
  - Each subgroup should have a 'title' that you come up with
	- a title is generally only 2-3 words, please no more than 70 characters total (including spaces)
  - Each subgroup should have an array of 'entry_ids'
	- an entry id can ONLY be associated to one sub_group
  - all entry_ids in a sub_group must share the same group_id.
  - The purpose of subgroups is to group related entries together to allow for faster skimming. The grouping can be based on topic, relevance, or any other criteria that you think would be helpful to the user.
	- sometimes you'll want to subgroup entries because they are all about the same topic
	- sometimes you'll want to subgroup entries because they are just basically leading to the same article itself
	- sometimes you'll want to subgroup entries because they are just all complementing each other and worth reading together

Your goal is to help the user get through their unread entries as efficiently as possible. The metadata you generate will be used to create a digest that allows the user to quickly understand and navigate their unread entries.

Below is the list of entries:
----------------------------
`

var llmResponseSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"overview": {
			Type: genai.TypeString,
		},
		"primary_group_summaries": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"primary_group_id": {
						Type: genai.TypeInteger,
					},
					"summary": {
						Type: genai.TypeString,
					},
				},
			},
		},
		"sub_groups": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"title": {
						Type: genai.TypeString,
					},
					"entry_ids": {
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

func (g *LLMGrouper) GroupEntries(pgs []*primaryGroup) ([]*models.PrimaryGroupDigestData, *string) {
	var llmEntries []llmEntry
	entryMap := make(map[int64]*models.Entry)
	primaryGroupMap := make(map[int64]*primaryGroup)

	for _, pg := range pgs {
		primaryGroupMap[pg.ID] = pg
		for _, entry := range pg.Entries {
			content := entry.Content
			if len(content) > 1000 {
				content = content[:1000]
			}
			llmEntries = append(llmEntries, llmEntry{
				ID:             entry.ID,
				Title:        entry.Title,
				URL:          entry.URL,
				Content:      content,
				PrimaryGroupID: pg.ID,
			})
			entryMap[entry.ID] = entry
		}
	}

	entriesJSON, err := json.MarshalIndent(llmEntries, "", "  ")
	if err != nil {
		return fallbackToFeedGrouper(pgs)
	}

	prompt := llmPrompt + string(entriesJSON)

	ctx, cancel := context.WithTimeout(context.Background(), LLMTimeout)
	defer cancel()

	llmResponse, err := g.LLMService.GenerateContent(ctx, prompt, llmResponseSchema)
	if err != nil {
		log.Printf("LLM service failed, falling back to feed grouping: %v\n", err)
		return fallbackToFeedGrouper(pgs)
	}

	var response LLMResponse
	if err := json.Unmarshal(llmResponse, &response); err != nil {
		log.Printf("Failed to parse LLM response, falling back to feed grouping: %v\n", err)
		log.Printf("Raw LLM response:\n%s", llmResponse)
		return fallbackToFeedGrouper(pgs)
	}

	finalPrimaryGroupsMap := make(map[int64]*models.PrimaryGroupDigestData)
	for _, pg := range pgs {
		finalPrimaryGroupsMap[pg.ID] = &models.PrimaryGroupDigestData{
			ID:        pg.ID,
			Title:     pg.Title,
			Slug:      utils.Slugify(pg.Title),
			SubGroups: []*models.EntryGroup{},
		}
	}

	for _, summary := range response.PrimaryGroupSummaries {
		if pg, ok := finalPrimaryGroupsMap[summary.PrimaryGroupID]; ok {
			pg.Summary = summary.Summary
		}
	}

	groupedEntryIDs := make(map[int64]bool)
	for _, llmSubGroup := range response.SubGroups {
		subGroupEntriesByPrimaryGroup := make(map[int64][]*models.Entry)

		for _, entryID := range llmSubGroup.EntryIDs {
			if _, exists := groupedEntryIDs[entryID]; exists {
				continue
			}
			if entry, ok := entryMap[entryID]; ok {
				primaryGroupID := entry.GroupID
				subGroupEntriesByPrimaryGroup[primaryGroupID] = append(subGroupEntriesByPrimaryGroup[primaryGroupID], entry)
				groupedEntryIDs[entryID] = true
			}
		}

		for primaryGroupID, entriesInGroup := range subGroupEntriesByPrimaryGroup {
			if pg, ok := finalPrimaryGroupsMap[primaryGroupID]; ok {
				if len(entriesInGroup) > 0 {
					pg.SubGroups = append(pg.SubGroups, &models.EntryGroup{
						Title:   llmSubGroup.Title,
						Entries: entriesInGroup,
						Slug:    utils.Slugify(llmSubGroup.Title),
					})
				}
			}
		}
	}

	for _, pg := range pgs {
		var ungroupedInThisPrimaryGroup []*models.Entry
		for _, entry := range pg.Entries {
			if !groupedEntryIDs[entry.ID] {
				ungroupedInThisPrimaryGroup = append(ungroupedInThisPrimaryGroup, entry)
			}
		}

		if len(ungroupedInThisPrimaryGroup) > 0 {
			finalPrimaryGroupsMap[pg.ID].SubGroups = append(finalPrimaryGroupsMap[pg.ID].SubGroups, &models.EntryGroup{
				Title:   "Uncategorized",
				Entries: ungroupedInThisPrimaryGroup,
				Slug:    "uncategorized",
			})
		}
	}

	var finalPrimaryGroups []*models.PrimaryGroupDigestData
	for _, pg := range finalPrimaryGroupsMap {
		finalPrimaryGroups = append(finalPrimaryGroups, pg)
	}

	return finalPrimaryGroups, &response.Overview
}

func fallbackToFeedGrouper(pgs []*primaryGroup) ([]*models.PrimaryGroupDigestData, *string) {
	return (&FeedGrouper{}).GroupEntries(pgs)
}
