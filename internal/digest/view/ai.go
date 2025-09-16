package view

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"miniflux-digest/internal/app/services"
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/utils"
	"sort"
	"strings"
	"sync"
	"text/template"
)

const (
	MaxEntryContentLengthForLLM = 1000
	MaxEntriesForSummarization  = 200
)

type llmEntry struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Content   string `json:"content"`
	FeedTitle string `json:"feed_title"`
}

func GroupAIEntries(ctx context.Context, entries []*models.Entry, llmService services.LLMService) (map[string][]*models.Entry, map[int64]bool, map[string][]*models.Entry) {
	const chunkSize = 200
	var wg sync.WaitGroup
	var mu sync.Mutex

	rawGroups := make(map[string][]*models.Entry)
	groupedEntryIDs := make(map[int64]bool)
	failedChunks := make(map[string][]*models.Entry)
	entryMap := make(map[int64]*models.Entry)
	for _, entry := range entries {
		entryMap[entry.ID] = entry
	}

	promptTemplate, err := template.New("grouping").Parse(initialGroupingPrompt)
	if err != nil {
		// This is a startup failure, not a runtime one, so we can't easily recover.
		// We'll return it as a failed chunk.
		failedChunks["Prompt template parsing failed"] = entries
		return nil, nil, failedChunks
	}

	for i := 0; i < len(entries); i += chunkSize {
		end := min(i+chunkSize, len(entries))
		chunk := entries[i:end]
		wg.Add(1)

		go func(chunk []*models.Entry, i, end int) {
			defer wg.Done()

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
				log.Printf("Failed to marshal entries to JSON for chunk %d-%d: %v", i, end, err)
				mu.Lock()
				failedChunks["JSON Marshal Error"] = append(failedChunks["JSON Marshal Error"], chunk...)
				mu.Unlock()
				return
			}

			var prompt bytes.Buffer
			if err := promptTemplate.Execute(&prompt, nil); err != nil {
				log.Printf("Failed to execute prompt template for chunk %d-%d: %v", i, end, err)
				mu.Lock()
				failedChunks["Prompt Execution Error"] = append(failedChunks["Prompt Execution Error"], chunk...)
				mu.Unlock()
				return
			}
			prompt.Write(entriesJSON)

			llmResponse, err := llmService.GenerateContent(ctx, prompt.String(), InitialGroupingResponseSchema)
			if err != nil {
				log.Printf("LLM service failed for chunk %d-%d: %v", i, end, err)
				errorKey := "LLM Error"
				if strings.Contains(err.Error(), "DEADLINE_EXCEEDED") {
					errorKey = "Processing Timeout"
				}
				mu.Lock()
				failedChunks[errorKey] = append(failedChunks[errorKey], chunk...)
				mu.Unlock()
				return
			}

			var response InitialGroupingResponse
			if err := json.Unmarshal(llmResponse, &response); err != nil {
				log.Printf("Failed to parse LLM response for chunk %d-%d: %v", i, end, err)
				mu.Lock()
				failedChunks["LLM Response Parse Error"] = append(failedChunks["LLM Response Parse Error"], chunk...)
				mu.Unlock()
				return
			}

			mu.Lock()
			defer mu.Unlock()
			for _, group := range response.Groups {
				if _, ok := rawGroups[group.Title]; !ok {
					rawGroups[group.Title] = []*models.Entry{}
				}
				for _, entryID := range group.EntryIDs {
					if _, exists := groupedEntryIDs[entryID]; exists {
						continue
					}
					if entry, ok := entryMap[entryID]; ok {
						rawGroups[group.Title] = append(rawGroups[group.Title], entry)
						groupedEntryIDs[entryID] = true
					}
				}
			}
		}(chunk, i, end)
	}

	wg.Wait()

	return rawGroups, groupedEntryIDs, failedChunks
}

func consolidatePrimaryGroups(ctx context.Context, rawGroups map[string][]*models.Entry, llmService services.LLMService) ([]*models.PrimaryGroupDigestData, error) {
	var initialTitles []string
	for title := range rawGroups {
		initialTitles = append(initialTitles, title)
	}

	if len(initialTitles) <= 10 { // No need to consolidate if we already have a small number of groups
		return convertRawGroupsToDigestData(rawGroups), nil
	}

	tmpl, err := template.New("consolidation").Parse(consolidationPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse consolidation prompt template: %w", err)
	}

	var prompt bytes.Buffer
	if err := tmpl.Execute(&prompt, initialTitles); err != nil {
		return nil, fmt.Errorf("failed to execute consolidation prompt template: %w", err)
	}

	llmResponse, err := llmService.GenerateContent(ctx, prompt.String(), ConsolidationResponseSchema)
	if err != nil {
		// Fallback: If LLM fails, use the original, unconsolidated groups
		log.Printf("LLM consolidation failed, falling back to original groups: %v", err)
		return convertRawGroupsToDigestData(rawGroups), nil
	}

	var response ConsolidationResponse
	if err := json.Unmarshal(llmResponse, &response); err != nil {
		log.Printf("Failed to parse LLM consolidation response, falling back to original groups: %v", err)
		return convertRawGroupsToDigestData(rawGroups), nil
	}

	// Merge groups based on consolidation response
	consolidatedGroups := make(map[string][]*models.Entry)
	mappedOldTitles := make(map[string]bool)
	for _, cg := range response.ConsolidatedGroups {
		if _, ok := consolidatedGroups[cg.NewTitle]; !ok {
			consolidatedGroups[cg.NewTitle] = []*models.Entry{}
		}
		for _, oldTitle := range cg.OldTitles {
			if entries, ok := rawGroups[oldTitle]; ok {
				consolidatedGroups[cg.NewTitle] = append(consolidatedGroups[cg.NewTitle], entries...)
				mappedOldTitles[oldTitle] = true
			}
		}
	}

	// Add any groups that the LLM may have missed
	for oldTitle, entries := range rawGroups {
		if !mappedOldTitles[oldTitle] {
			consolidatedGroups[oldTitle] = entries
		}
	}

	return convertRawGroupsToDigestData(consolidatedGroups), nil
}

func convertRawGroupsToDigestData(rawGroups map[string][]*models.Entry) []*models.PrimaryGroupDigestData {
	var digestData []*models.PrimaryGroupDigestData
	currentPrimaryGroupID := int64(1)
	for title, entries := range rawGroups {
		if len(entries) == 0 {
			continue
		}
		newPGDD := &models.PrimaryGroupDigestData{
			ID:    currentPrimaryGroupID,
			Title: title,
			Slug:  utils.Slugify(title),
			SubGroups: []*models.EntryGroup{
				{
					Title:   title, // Initial sub-group is the same as the primary group
					Entries: entries,
					Slug:    utils.Slugify(title),
				},
			},
			Summary:      "",
			TotalEntries: len(entries),
			TotalFeeds:   getUniqueFeedIDs(entries),
		}
		digestData = append(digestData, newPGDD)
		currentPrimaryGroupID++
	}
	return digestData
}

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
		var summaryResponse SummaryResponse
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
		var subGroupingResponse SubGroupingResponse
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
	rawGroups, groupedEntryIDs, failedChunks := GroupAIEntries(ctx, entries, llmService)

	// If all chunks failed and we have no raw groups, fall back to category grouping.
	if len(rawGroups) == 0 && len(failedChunks) > 0 {
		log.Printf("All AI grouping chunks failed, falling back to category grouping.")
		categoryGroups := BuildDigestDataByCategory(entries)
		if len(categoryGroups) == 1 && len(entries) > 1 {
			return BuildDigestDataByDate(entries)
		}
		return categoryGroups
	}

	highLevelGroups, err := consolidatePrimaryGroups(ctx, rawGroups, llmService)
	if err != nil {
		// This case should ideally not be hit if consolidatePrimaryGroups has a fallback
		log.Printf("AI group consolidation failed unexpectedly: %v", err)
		highLevelGroups = convertRawGroupsToDigestData(rawGroups)
	}

	failedEntryIDs := make(map[int64]bool)
	for reason, chunkEntries := range failedChunks {
		title := fmt.Sprintf("Uncategorized - %s", reason)
		highLevelGroups = append(highLevelGroups, &models.PrimaryGroupDigestData{
			ID:    int64(len(highLevelGroups) + 1),
			Title: title,
			Slug:  utils.Slugify(title),
			SubGroups: []*models.EntryGroup{
				{
					Title:        title,
					Entries:      chunkEntries,
					Slug:         utils.Slugify(title),
					TotalEntries: len(chunkEntries),
					TotalFeeds:   getUniqueFeedIDs(chunkEntries),
				},
			},
			Summary:      "These entries could not be processed by the AI during the initial grouping.",
			TotalEntries: len(chunkEntries),
			TotalFeeds:   getUniqueFeedIDs(chunkEntries),
		})
		for _, entry := range chunkEntries {
			failedEntryIDs[entry.ID] = true
		}
	}

	// Handle entries that were not grouped by the initial LLM call
	var otherUngroupedEntries []*models.Entry
	for _, entry := range entries {
		if !groupedEntryIDs[entry.ID] && !failedEntryIDs[entry.ID] {
			otherUngroupedEntries = append(otherUngroupedEntries, entry)
		}
	}

	if len(otherUngroupedEntries) > 0 {
		highLevelGroups = append(highLevelGroups, &models.PrimaryGroupDigestData{
			ID:    int64(len(highLevelGroups) + 1),
			Title: "Uncategorized",
			Slug:  "uncategorized",
			SubGroups: []*models.EntryGroup{
				{
					Title:        "Uncategorized",
					Entries:      otherUngroupedEntries,
					Slug:         "uncategorized",
					TotalEntries: len(otherUngroupedEntries),
					TotalFeeds:   getUniqueFeedIDs(otherUngroupedEntries),
				},
			},
			Summary:      "",
			TotalEntries: len(otherUngroupedEntries),
			TotalFeeds:   getUniqueFeedIDs(otherUngroupedEntries),
		})
	}

	var allPrimaryGroups []*models.PrimaryGroupDigestData
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, hlg := range highLevelGroups {
		wg.Add(1)
		go func(hlg *models.PrimaryGroupDigestData) {
			defer wg.Done()

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
				hlg.SubGroups = []*models.EntryGroup{
					{
						Title:        pg.Title,
						Entries:      pg.Entries,
						Slug:         utils.Slugify(pg.Title),
						TotalEntries: len(pg.Entries),
						TotalFeeds:   getUniqueFeedIDs(pg.Entries),
					},
				}
			} else {
				hlg.Summary = summary
				hlg.SubGroups = subGroups
			}

			mu.Lock()
			allPrimaryGroups = append(allPrimaryGroups, hlg)
			mu.Unlock()
		}(hlg)
	}

	wg.Wait()

	sort.Slice(allPrimaryGroups, func(i, j int) bool {
		return allPrimaryGroups[i].Title < allPrimaryGroups[j].Title
	})

	return allPrimaryGroups
}
