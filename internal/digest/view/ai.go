package view

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"miniflux-digest/internal/app/services"
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/llm"
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/utils"
	"sort"
	"strings"
	"sync"
	"text/template"

	"google.golang.org/genai"
)

const (
	MaxEntryContentLengthForLLM = 250
	MaxEntriesPerJob  = 150
)

type llmEntry struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Content   string `json:"content"`
	FeedTitle string `json:"feed_title"`
}



func GroupAIEntries(ctx context.Context, entries []*models.Entry, llmService services.LLMService, categories []config.ConfigCategory) (map[string][]*models.Entry, map[int64]bool, map[string][]*models.Entry) {
	const chunkSize = 200

	// Prepare for processing
	entryMap := make(map[int64]*models.Entry)
	for _, entry := range entries {
		entryMap[entry.ID] = entry
	}

	promptTemplate, err := template.New("grouping").Parse(initialGroupingPrompt)
	if err != nil {
		failedChunks := make(map[string][]*models.Entry)
		failedChunks["Prompt template parsing failed"] = entries
		return nil, nil, failedChunks
	}

	workerConfig := llmWorkerConfig{
		promptTemplate: promptTemplate,
		promptData:     categories,
		responseSchema: InitialGroupingResponseSchema,
		prepareEntries: func(entries []*models.Entry) interface{} {
			return prepareLLMEntries(entries)
		},
	}
	worker := createLLMWorker[*InitialGroupingResponse](ctx, llmService, workerConfig, "")

	// Run the parallel processing
	chunkResponses, failedChunks := llm.ProcessInChunks(entries, chunkSize, worker)

	// Process the results
	rawGroups := make(map[string][]*models.Entry)
	groupedEntryIDs := make(map[int64]bool)

	for _, response := range chunkResponses {
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
	}

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

	var response ConsolidationResponse
	if err := llmService.GenerateContentWithResponse(ctx, prompt.String(), ConsolidationResponseSchema, &response); err != nil {
		// Fallback: If LLM fails, use the original, unconsolidated groups
		log.Printf("LLM consolidation failed, falling back to original groups: %v", err)
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



func getSummaryForGroup(ctx context.Context, groupTitle string, entries []*models.Entry, llmService services.LLMService) (string, error) {
	llmResponseWorker := createLLMWorker[SummaryResponse](ctx, llmService, llmWorkerConfig{
		promptFormat:   summaryPrompt,
		responseSchema: SummaryResponseSchema,
		prepareEntries: func(entries []*models.Entry) interface{} {
			return prepareLLMEntries(entries)
		},
	}, groupTitle)

	summaryWorker := func(chunk []*models.Entry) (string, error) {
		resp, err := llmResponseWorker(chunk)
		if err != nil {
			return "", err
		}
		return resp.Summary, nil
	}

	partialSummaries, failedChunks := llm.ProcessInChunks(entries, MaxEntriesPerJob, summaryWorker)

	if len(partialSummaries) == 0 {
		return "", fmt.Errorf("all summary chunks failed for group '%s'", groupTitle)
	}

	// Reduce step: summarize the summaries
	summariesString := strings.Join(partialSummaries, "\n\n---\n\n")
	finalSummaryPromptStr := fmt.Sprintf(summaryOfSummariesPrompt, summariesString)

	var finalSummaryResponse SummaryResponse
	if err := llmService.GenerateContentWithResponse(ctx, finalSummaryPromptStr, SummaryResponseSchema, &finalSummaryResponse); err != nil {
		return "", fmt.Errorf("summary of summaries failed: %w", err)
	}

	finalSummary := finalSummaryResponse.Summary
	var failedItems []*models.Entry
	for _, items := range failedChunks {
		failedItems = append(failedItems, items...)
	}
	if len(failedItems) > 0 {
		finalSummary += fmt.Sprintf("\n\n(Note: %d entries could not be included in the summary due to processing errors.)", len(failedItems))
	}

	return finalSummary, nil
}

func getSubGroupsForGroup(ctx context.Context, groupTitle string, entries []*models.Entry, llmService services.LLMService) ([]*models.EntryGroup, error) {
	entryMap := make(map[int64]*models.Entry)
	for _, entry := range entries {
		entryMap[entry.ID] = entry
	}

	subGroupWorker := createLLMWorker[*SubGroupingResponse](ctx, llmService, llmWorkerConfig{
		promptFormat:   subGroupingPrompt,
		responseSchema: SubGroupingResponseSchema,
		prepareEntries: func(entries []*models.Entry) interface{} {
			return prepareLLMEntries(entries)
		},
	}, groupTitle)

	partialSubGroupResponses, failedChunks := llm.ProcessInChunks(entries, MaxEntriesPerJob, subGroupWorker)

	if len(partialSubGroupResponses) == 0 {
		return nil, fmt.Errorf("all sub-grouping chunks failed for group '%s'", groupTitle)
	}

	// Reduce step: merge sub-groups
	mergedSubGroups := make(map[string][]*models.Entry)
	allGroupedEntryIDs := make(map[int64]bool)

	for _, response := range partialSubGroupResponses {
		for _, llmSubGroup := range response.SubGroups {
			if _, ok := mergedSubGroups[llmSubGroup.Title]; !ok {
				mergedSubGroups[llmSubGroup.Title] = []*models.Entry{}
			}
			for _, entryID := range llmSubGroup.EntryIDs {
				if _, exists := allGroupedEntryIDs[entryID]; exists {
					continue // Avoid duplicating entries if they appear in multiple chunk responses
				}
				if entry, ok := entryMap[entryID]; ok {
					mergedSubGroups[llmSubGroup.Title] = append(mergedSubGroups[llmSubGroup.Title], entry)
					allGroupedEntryIDs[entryID] = true
				}
			}
		}
	}

	var finalSubGroups []*models.EntryGroup
	for title, groupEntries := range mergedSubGroups {
		if len(groupEntries) > 0 {
			finalSubGroups = append(finalSubGroups, &models.EntryGroup{
				Title:        title,
				Entries:      groupEntries,
				Slug:         utils.Slugify(title),
				TotalEntries: len(groupEntries),
				TotalFeeds:   getUniqueFeedIDs(groupEntries),
			})
		}
	}

	// Handle all ungrouped entries from all chunks
	var allUngroupedEntries []*models.Entry
	for _, entry := range entries {
		if !allGroupedEntryIDs[entry.ID] {
			allUngroupedEntries = append(allUngroupedEntries, entry)
		}
	}

	// Collect all failed items from the map into a single slice
	var allFailedItems []*models.Entry
	for _, items := range failedChunks {
		allFailedItems = append(allFailedItems, items...)
	}

	// Add items from failed chunks to a separate group
	if len(allFailedItems) > 0 {
		finalSubGroups = append(finalSubGroups, &models.EntryGroup{
			Title:        "Uncategorized - Processing Failed",
			Entries:      allFailedItems,
			Slug:         "uncategorized-processing-failed",
			TotalEntries: len(allFailedItems),
			TotalFeeds:   getUniqueFeedIDs(allFailedItems),
		})
	}

	if len(allUngroupedEntries) > 0 {
		// Find existing "Uncategorized" group to append to, if it exists
		var uncategorizedGroup *models.EntryGroup
		for _, sg := range finalSubGroups {
			if sg.Title == "Uncategorized" {
				uncategorizedGroup = sg
				break
			}
		}
		if uncategorizedGroup != nil {
			uncategorizedGroup.Entries = append(uncategorizedGroup.Entries, allUngroupedEntries...)
		} else {
			finalSubGroups = append(finalSubGroups, &models.EntryGroup{
				Title:        "Uncategorized",
				Entries:      allUngroupedEntries,
				Slug:         "uncategorized",
				TotalEntries: len(allUngroupedEntries),
				TotalFeeds:   getUniqueFeedIDs(allUngroupedEntries),
			})
		}
	}

	return finalSubGroups, nil
}

func ProcessPrimaryGroupWithAI(ctx context.Context, pg *models.PrimaryGroup, llmService services.LLMService) (string, []*models.EntryGroup, error) {
	entriesToProcess := pg.Entries

	// If the group is large, process in chunks concurrently.
	var finalSummary string
	var finalSubGroups []*models.EntryGroup
	var errSummary, errSubgroup error
	var wg sync.WaitGroup

	wg.Add(2)

	// Goroutine for Summarization
	go func() {
		defer wg.Done()
		finalSummary, errSummary = getSummaryForGroup(ctx, pg.Title, entriesToProcess, llmService)
	}()

	// Goroutine for Sub-grouping
	go func() {
		defer wg.Done()
		finalSubGroups, errSubgroup = getSubGroupsForGroup(ctx, pg.Title, entriesToProcess, llmService)
	}()

	wg.Wait()

	// Combine errors if any. For now, we just log them but don't halt everything.
	if errSummary != nil {
		log.Printf("Error generating summary for group '%s': %v", pg.Title, errSummary)
	}
	if errSubgroup != nil {
		log.Printf("Error generating sub-groups for group '%s': %v", pg.Title, errSubgroup)
		// If sub-grouping completely failed, create a single group as a fallback.
		if finalSubGroups == nil {
			finalSubGroups = []*models.EntryGroup{{ 
				Title:        "Uncategorized",
				Entries:      entriesToProcess,
				Slug:         "uncategorized",
				TotalEntries: len(entriesToProcess),
				TotalFeeds:   getUniqueFeedIDs(entriesToProcess),
			}}
		}
	}

	return finalSummary, finalSubGroups, nil
}

func getUniqueFeedIDs(entries []*models.Entry) int {
	uniqueFeedIDs := make(map[int64]bool)
	for _, entry := range entries {
		uniqueFeedIDs[entry.FeedID] = true
	}
	return len(uniqueFeedIDs)
}



func prepareLLMEntries(entries []*models.Entry) []llmEntry {
	var llmEntries []llmEntry
	for _, entry := range entries {
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
	return llmEntries
}

func BuildDigestDataByAI(entries []*models.Entry, ctx context.Context, llmService services.LLMService, categories []config.ConfigCategory) []*models.PrimaryGroupDigestData {
	rawGroups, groupedEntryIDs, failedChunks, fallbackGroups := initialAIGGrouping(ctx, entries, llmService, categories)
	if fallbackGroups != nil {
		return fallbackGroups
	}

	highLevelGroups := consolidateGroups(ctx, rawGroups, llmService)

	highLevelGroups = handleFailedAndUngroupedEntries(highLevelGroups, entries, groupedEntryIDs, failedChunks)

	allPrimaryGroups := processHighLevelGroupsConcurrently(ctx, highLevelGroups, llmService)

	sortPrimaryGroups(allPrimaryGroups)

	return allPrimaryGroups
}

func initialAIGGrouping(ctx context.Context, entries []*models.Entry, llmService services.LLMService, categories []config.ConfigCategory) (map[string][]*models.Entry, map[int64]bool, map[string][]*models.Entry, []*models.PrimaryGroupDigestData) {
	rawGroups, groupedEntryIDs, failedChunks := GroupAIEntries(ctx, entries, llmService, categories)

	// If all chunks failed and we have no raw groups, fall back to category grouping.
	if len(rawGroups) == 0 && len(failedChunks) > 0 {
		log.Printf("All AI grouping chunks failed, falling back to category grouping.")
		categoryGroups := BuildDigestDataByCategory(entries)
		if len(categoryGroups) == 1 && len(entries) > 1 {
			return nil, nil, nil, BuildDigestDataByDate(entries)
		}
		return nil, nil, nil, categoryGroups
	}
	return rawGroups, groupedEntryIDs, failedChunks, nil
}

func consolidateGroups(ctx context.Context, rawGroups map[string][]*models.Entry, llmService services.LLMService) []*models.PrimaryGroupDigestData {
	highLevelGroups, err := consolidatePrimaryGroups(ctx, rawGroups, llmService)
	if err != nil {
		// This case should ideally not be hit if consolidatePrimaryGroups has a fallback
		log.Printf("AI group consolidation failed unexpectedly: %v", err)
		highLevelGroups = convertRawGroupsToDigestData(rawGroups)
	}
	return highLevelGroups
}

func handleFailedAndUngroupedEntries(highLevelGroups []*models.PrimaryGroupDigestData, allEntries []*models.Entry, groupedEntryIDs map[int64]bool, failedChunks map[string][]*models.Entry) []*models.PrimaryGroupDigestData {
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
	for _, entry := range allEntries {
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
	return highLevelGroups
}

func processHighLevelGroupsConcurrently(ctx context.Context, highLevelGroups []*models.PrimaryGroupDigestData, llmService services.LLMService) []*models.PrimaryGroupDigestData {
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
	return allPrimaryGroups
}

func sortPrimaryGroups(allPrimaryGroups []*models.PrimaryGroupDigestData) {
	sort.Slice(allPrimaryGroups, func(i, j int) bool {
		return allPrimaryGroups[i].Title < allPrimaryGroups[j].Title
	})
}

// llmWorkerConfig holds configuration for creating an LLM worker function.
type llmWorkerConfig struct {
	promptTemplate *template.Template
	promptData     interface{} // Data to pass to template execution
	promptFormat   string // For prompts that use Sprintf
	responseSchema *genai.Schema
	prepareEntries func([]*models.Entry) interface{} // Function to prepare entries for LLM
}

// createLLMWorker creates a worker function for llm.ProcessInChunks.
func createLLMWorker[R any](
	ctx context.Context,
	llmService services.LLMService,
	config llmWorkerConfig,
	groupTitle string, // Optional, for prompts that need it
) llm.WorkerFunc[*models.Entry, R] {
	return func(chunk []*models.Entry) (R, error) {
		var zero R // Initialize a zero value of type R

		// 1. Prepare LLM entries
		var llmEntriesJSON []byte
		if config.prepareEntries != nil {
			prepared := config.prepareEntries(chunk)
			var err error
			llmEntriesJSON, err = json.MarshalIndent(prepared, "", "  ")
			if err != nil {
				return zero, fmt.Errorf("failed to marshal entries to JSON: %w", err)
			}
		}

		// 2. Construct the prompt
		var prompt bytes.Buffer
		if config.promptTemplate != nil {
			if err := config.promptTemplate.Execute(&prompt, config.promptData); err != nil {
				return zero, fmt.Errorf("failed to execute prompt template: %w", err)
			}
		} else if config.promptFormat != "" {
			prompt.WriteString(fmt.Sprintf(config.promptFormat, groupTitle))
		}

		if len(llmEntriesJSON) > 0 {
			prompt.Write(llmEntriesJSON)
		}

		// 3. Call LLM service
		var response R
		if err := llmService.GenerateContentWithResponse(ctx, prompt.String(), config.responseSchema, &response); err != nil {
			return zero, err
		}
		return response, nil
	}
}