package view

import (
	"context"
	"encoding/json"
	"fmt"
	"miniflux-digest/internal/app/services"
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/utils"
	
	"sort"
	"log"

	"google.golang.org/genai"
)

const (
	MaxEntryContentLengthForLLM = 1000
	MaxEntriesForSummarization  = 1000
)

type GroupingResponse struct {
	Groups []struct {
		Title    string  `json:"title"`
		EntryIDs []int64 `json:"entry_ids"`
	} `json:"groups"`
}

var GroupingResponseSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"groups": {
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

const groupingPrompt = `You will be given a list of unread JSON entry objects from an RSS feed reader. Each entry includes its ID, Title, URL, Content, and the Title of the Feed it came from.

Your task is to group these entries based on their content. Follow these instructions:

- Group entries that are about the same specific topic or event.
- Create group titles that are concise and descriptive (2-5 words).
- An entry can only belong to one group.
- Aim for groups with 2 to 20 entries.
- If an entry cannot be grouped, leave it out of the response.

Your goal is to create meaningful groups that help the user quickly understand the main themes in their feed.

Below is the list of entries:
----------------------------
`

type llmEntry struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Content   string `json:"content"`
	FeedTitle string `json:"feed_title"`
}

func GroupAIEntries(ctx context.Context, entries []*models.Entry, llmService services.LLMService) ([]*models.PrimaryGroupDigestData, error) {
	const chunkSize = 1000

	entryMap := make(map[int64]*models.Entry)
	for _, entry := range entries {
		entryMap[entry.ID] = entry
	}

	groupedEntryIDs := make(map[int64]bool)
	var allPrimaryGroupDigestData []*models.PrimaryGroupDigestData
	primaryGroupMap := make(map[string]*models.PrimaryGroupDigestData)
	currentPrimaryGroupID := int64(1)

	for i := 0; i < len(entries); i += chunkSize {
		end := min(i + chunkSize, len(entries))
		chunk := entries[i:end]

		var llmEntries []llmEntry
		for _, entry := range chunk {
			content := entry.Content
			if len(content) > MaxEntryContentLengthForLLM {
				content = content[:MaxEntryContentLengthForLLM]
			}
			llmEntries = append(llmEntries, llmEntry{
				ID:        entry.ID,
				Title:     entry.Title,
				URL:       entry.URL,
				Content:   content,
				FeedTitle: entry.FeedTitle,
			})
		}

		entriesJSON, err := json.MarshalIndent(llmEntries, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal entries to JSON: %w", err)
		}

		prompt := groupingPrompt + string(entriesJSON)

		llmResponse, err := llmService.GenerateContent(ctx, prompt, GroupingResponseSchema)
		if err != nil {
			log.Printf("LLM call for 'primary-grouping' failed on chunk %d-%d: %v", i, end, err)
			return nil, fmt.Errorf("LLM service failed for chunk %d-%d: %w", i, end, err)
		}

		var response GroupingResponse
		if err := json.Unmarshal(llmResponse, &response); err != nil {
			return nil, fmt.Errorf("failed to parse LLM response for chunk %d-%d: %w", i, end, err)
		}

		for _, group := range response.Groups {
			var groupEntries []*models.Entry
			for _, entryID := range group.EntryIDs {
				if _, exists := groupedEntryIDs[entryID]; exists {
					continue
				}
				if entry, ok := entryMap[entryID]; ok {
					groupEntries = append(groupEntries, entry)
					groupedEntryIDs[entryID] = true
				}
			}

			if len(groupEntries) > 0 {
				if existingPGDD, ok := primaryGroupMap[group.Title]; ok {
					existingPGDD.SubGroups[0].Entries = append(existingPGDD.SubGroups[0].Entries, groupEntries...)
					existingPGDD.TotalEntries += len(groupEntries)
					existingPGDD.TotalFeeds = getUniqueFeedIDs(existingPGDD.SubGroups[0].Entries)
				} else {
					newPGDD := &models.PrimaryGroupDigestData{
						ID:           currentPrimaryGroupID,
						Title:        group.Title,
						Slug:         utils.Slugify(group.Title),
						SubGroups:    []*models.EntryGroup{
							{
								Title:   group.Title,
								Entries: groupEntries,
								Slug:    utils.Slugify(group.Title),
							},
						},
						Summary:      "",
						TotalEntries: len(groupEntries),
						TotalFeeds:   getUniqueFeedIDs(groupEntries),
					}
					allPrimaryGroupDigestData = append(allPrimaryGroupDigestData, newPGDD)
					primaryGroupMap[group.Title] = newPGDD
					currentPrimaryGroupID++
				}
			}
		}
	}

	var ungroupedEntries []*models.Entry

	for _, entry := range entries {
		if !groupedEntryIDs[entry.ID] {
			ungroupedEntries = append(ungroupedEntries, entry)
		}
	}

	if len(ungroupedEntries) > 0 {
		allPrimaryGroupDigestData = append(allPrimaryGroupDigestData, &models.PrimaryGroupDigestData{
			ID:           currentPrimaryGroupID,
			Title:        "Uncategorized",
			Slug:         "uncategorized",
			SubGroups:    []*models.EntryGroup{
				{
					Title:   "Uncategorized",
					Entries: ungroupedEntries,
					Slug:    "uncategorized",
				},
			},
			Summary:      "",
			TotalEntries: len(ungroupedEntries),
			TotalFeeds:   getUniqueFeedIDs(ungroupedEntries),
		})
	}

	return allPrimaryGroupDigestData, nil
}

type AIGroupSummaryResponse struct {
	Summary string `json:"summary"`
}

type AISubGroupingResponse struct {
	SubGroups []struct {
		Title    string  `json:"title"`
		EntryIDs []int64 `json:"entry_ids"`
	} `json:"sub_groups"`
}

var SummaryResponseSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"summary": {
			Type: genai.TypeString,
		},
	},
}

var SubGroupingResponseSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
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

const summaryPrompt = `You will be given a list of JSON entry objects from a primary group titled '%s'. Generate a 2-3 sentence summary (less than 150 words) of the main topics. To improve readability, insert two line breaks (\n\n) after every 1-2 sentences.

Below is the list of entries for the primary group:
----------------------------
`

const subGroupingPrompt = `You will be given a list of JSON entry objects from a primary group titled '%s'. Your task is to identify relevant sub-groups within it.

- Create sub-groups for entries about the same specific topic or event.
- Create concise, descriptive titles (2-5 words) for each sub-group.
- An entry can only belong to one sub-group.
- Aim for sub-groups with 2 to 20 entries.
- If an entry cannot be sub-grouped, leave it out of the response.

Your goal is to create meaningful sub-groups that help the user navigate the entries.

Below is the list of entries for the primary group:
----------------------------
`

func ProcessPrimaryGroupWithAI(ctx context.Context, pg *models.PrimaryGroup, llmService services.LLMService) (string, []*models.EntryGroup, error) {
	var llmEntries []llmEntry
	entriesToProcess := pg.Entries
	if len(entriesToProcess) > MaxEntriesForSummarization {
		entriesToProcess = entriesToProcess[:MaxEntriesForSummarization]
	}

	entryMap := make(map[int64]*models.Entry)
	for _, entry := range entriesToProcess {
		entryMap[entry.ID] = entry
		content := entry.Content
		if len(content) > MaxEntryContentLengthForLLM {
			content = content[:MaxEntryContentLengthForLLM]
		}
		llmEntries = append(llmEntries, llmEntry{
			ID:        entry.ID,
			Title:     entry.Title,
			URL:       entry.URL,
			Content:   content,
			FeedTitle: entry.FeedTitle,
		})
	}

	entriesJSON, err := json.MarshalIndent(llmEntries, "", "  ")
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal entries to JSON for AI processing: %w", err)
	}

	// 1. Get Summary
	var summary string
	summaryPromptFmt := fmt.Sprintf(summaryPrompt, pg.Title)
	summaryFullPrompt := summaryPromptFmt + string(entriesJSON)

	summaryResponseBytes, err := llmService.GenerateContent(ctx, summaryFullPrompt, SummaryResponseSchema)
	if err != nil {
		log.Printf("LLM service failed during summarization for primary group '%s': %v", pg.Title, err)
	} else {
		var summaryResponse AIGroupSummaryResponse
		if err := json.Unmarshal(summaryResponseBytes, &summaryResponse); err != nil {
			log.Printf("Failed to parse LLM summarization response for primary group '%s': %v", pg.Title, err)
		} else {
			summary = summaryResponse.Summary
		}
	}

	// 2. Get Sub-groups
	var subGroups []*models.EntryGroup
	subGroupingPromptFmt := fmt.Sprintf(subGroupingPrompt, pg.Title)
	subGroupingFullPrompt := subGroupingPromptFmt + string(entriesJSON)

	subGroupingResponseBytes, err := llmService.GenerateContent(ctx, subGroupingFullPrompt, SubGroupingResponseSchema)
	if err != nil {
		log.Printf("LLM service failed during sub-grouping for primary group '%s': %v. Creating single 'Uncategorized' group.", pg.Title, err)
	} else {
		var subGroupingResponse AISubGroupingResponse
		if err := json.Unmarshal(subGroupingResponseBytes, &subGroupingResponse); err != nil {
			log.Printf("Failed to parse LLM sub-grouping response for primary group '%s': %v. Creating single 'Uncategorized' group.", pg.Title, err)
		} else {
			groupedEntryIDs := make(map[int64]bool)
			for _, llmSubGroup := range subGroupingResponse.SubGroups {
				var currentSubGroupEntries []*models.Entry
				for _, entryID := range llmSubGroup.EntryIDs {
					if entry, ok := entryMap[entryID]; ok {
						currentSubGroupEntries = append(currentSubGroupEntries, entry)
						groupedEntryIDs[entryID] = true
					}
				}
				if len(currentSubGroupEntries) > 0 {
					subGroups = append(subGroups, &models.EntryGroup{
						Title:        llmSubGroup.Title,
						Entries:      currentSubGroupEntries,
						Slug:         utils.Slugify(llmSubGroup.Title),
						TotalEntries: len(currentSubGroupEntries),
						TotalFeeds:   getUniqueFeedIDs(currentSubGroupEntries),
					})
				}
			}
			// Handle ungrouped entries
			var ungroupedEntries []*models.Entry
			for _, entry := range entriesToProcess {
				if !groupedEntryIDs[entry.ID] {
					ungroupedEntries = append(ungroupedEntries, entry)
				}
			}
			if len(ungroupedEntries) > 0 {
				subGroups = append(subGroups, &models.EntryGroup{
					Title:        "Uncategorized",
					Entries:      ungroupedEntries,
					Slug:         "uncategorized",
					TotalEntries: len(ungroupedEntries),
					TotalFeeds:   getUniqueFeedIDs(ungroupedEntries),
				})
			}
		}
	}

	// Fallback if sub-grouping failed or returned no groups
	if len(subGroups) == 0 {
		subGroups = append(subGroups, &models.EntryGroup{
			Title:        "Uncategorized",
			Entries:      entriesToProcess,
			Slug:         "uncategorized",
			TotalEntries: len(entriesToProcess),
			TotalFeeds:   getUniqueFeedIDs(entriesToProcess),
		})
	}

	return summary, subGroups, nil
}

func getUniqueFeedIDs(entries []*models.Entry) int {
	uniqueFeedIDs := make(map[int64]bool)
	for _, entry := range entries {
		uniqueFeedIDs[entry.FeedID] = true
	}
	return len(uniqueFeedIDs)
}

func BuildDigestDataByAI(entries []*models.Entry, ctx context.Context, llmService services.LLMService) []*models.PrimaryGroupDigestData {
	highLevelGroups, err := GroupAIEntries(ctx, entries, llmService)
	if err != nil {
		categoryGroups := BuildDigestDataByCategory(entries)
		if len(categoryGroups) == 1 && len(entries) > 1 {
			return BuildDigestDataByDate(entries)
		}
		return categoryGroups
	}

	var allPrimaryGroups []*models.PrimaryGroupDigestData

	for _, hlg := range highLevelGroups {
		var allEntriesInHlg []*models.Entry
		for _, subGroup := range hlg.SubGroups {
			allEntriesInHlg = append(allEntriesInHlg, subGroup.Entries...)
		}

		pg := &models.PrimaryGroup{
			ID:      hlg.ID,
			Title:   hlg.Title,
			Entries: allEntriesInHlg,
		}

		summary, subGroups, err := ProcessPrimaryGroupWithAI(ctx, pg, llmService)
		if err != nil {
			log.Printf("Failed to process primary group '%s' with AI: %v. Using original group.", pg.Title, err)
			// If processing fails, use the original primary group's entries as a single sub-group
			hlg.SubGroups = []*models.EntryGroup{
				{
					Title:   pg.Title,
					Entries: pg.Entries,
					Slug:    utils.Slugify(pg.Title),
					TotalEntries: len(pg.Entries),
					TotalFeeds:   getUniqueFeedIDs(pg.Entries),
				},
			}
		} else {
			hlg.Summary = summary
			hlg.SubGroups = subGroups
		}

		allPrimaryGroups = append(allPrimaryGroups, hlg)
	}

	sort.Slice(allPrimaryGroups, func(i, j int) bool {
		return allPrimaryGroups[i].Title < allPrimaryGroups[j].Title
	})

	
	return allPrimaryGroups
}
