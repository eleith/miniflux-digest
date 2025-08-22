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
	LLMTimeout = 5 * time.Minute
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

const llmPrompt = `You are an expert content curator. Your task is to help fill out a digest built from a collection of feeds that a person follows. these feeds follow updates from blogs, social networks, and news sources. The goal of the digest is to help the person quickly skim through a large number of entries (sometimes more than 1000) and identify the ones that are worth bookmarking for later or to just skim a bunch of overlapping ones to get a gist of what's happened.

You will be given a list of entries, each with a title, content, feed association, and primary group association. Your task is to provide some essential ideas that will be used by a digest to help the user read through their entries in an hour or less. the response will look like the following example:

{
  "primary_group_summaries": [
    {
      "primary_group_id": 123,
      "summary": "an overview of all the entries"
    }
  ],
  "sub_groups": [
    {
      "title": "A short, descriptive title for the subgroup.",
      "entry_ids": [1, 2, 3]
    }
  ]
}

Here are the constraints and guidelines for your response:

- Primary Group Summaries:
  - The primary_group_summaries field should contain a list of summaries for the entries belonging to one primary group.
	- entries only ever are associated with one primary group
  - Each summary should be associated with a primary_group_id.
  - The summary should be a single paragraph, no more than 3-4 sentences long.
  - The summary should help the user quickly understand your overview of the entries in the primary group
	- the summary should help the user decide whether to skim fast or slow and what to narrow in on. 
	- the summary could highlight critical entries or summarize the main topics or both whatever is best for the entries.

- Subgroups:
  - The sub_groups field should contain a list of subgroups.
  - Each subgroup should have a title that you come up with that is short and descriptive.
  - Each subgroup should have a list of entry_ids.
  - Crucially, all entry_ids within a single subgroup must belong to the same primary group. You can determine the primary group of an entry from the primary_group_id field in the input.
  - The purpose of subgroups is to group related entries together to allow for faster skimming. The grouping can be based on topic, relevance, or any other criteria that you think would be helpful to the user.
	- sometimes you'll want to subgroup entries because they are all about the same topic
	- sometimes you'll want to subgroup entries because they are just basically leading to the same article itself
	- sometimes you'll want to subgroup entries because they are just all complementing each other and worth reading together

Your goal is to help the user get through their unread entries as efficiently as possible. The overall summaries and subgroups names and subgroup associations you create are the primary tools to achieve this.

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
			llmEntries = append(llmEntries, llmEntry{
				ID:             entry.ID,
				Title:        entry.Title,
				URL:          entry.URL,
				Content:      entry.Content,
				FeedTitle:    entry.FeedTitle,
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
	if err := json.Unmarshal([]byte(llmResponse), &response); err != nil {
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
		var groupEntries []*models.Entry
		var primaryGroupID int64 = -1

		for _, entryID := range llmSubGroup.EntryIDs {
			if _, exists := groupedEntryIDs[entryID]; !exists {
				if entry, ok := entryMap[entryID]; ok {
					if primaryGroupID == -1 {
						primaryGroupID = entry.GroupID
					} else if primaryGroupID != entry.GroupID {
						log.Printf("LLM returned subgroup with mixed primary group IDs. Falling back to feed grouper.")
						return fallbackToFeedGrouper(pgs)
					}
					groupEntries = append(groupEntries, entry)
					groupedEntryIDs[entryID] = true
				}
			}
		}

		if pg, ok := finalPrimaryGroupsMap[primaryGroupID]; ok {
			pg.SubGroups = append(pg.SubGroups, &models.EntryGroup{
				Title:   llmSubGroup.Title,
				Entries: groupEntries,
				Slug:    utils.Slugify(llmSubGroup.Title),
			})
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
