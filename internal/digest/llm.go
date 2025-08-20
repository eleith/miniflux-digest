package digest

import (
	"context"
	"encoding/json"
	"fmt"

	"miniflux-digest/internal/llm"

	"google.golang.org/genai"
	miniflux "miniflux.app/v2/client"
)

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

type DigestLLMService interface {
	GenerateDigestContent(ctx context.Context, entries *miniflux.Entries) (*llm.LLMResponse, error)
}

type LLMService interface {
	GenerateContent(ctx context.Context, prompt string, schema *genai.Schema) (string, error)
}

type digestLLMServiceImpl struct {
	llmService llm.LLMService
}

func NewDigestLLMService(llmService llm.LLMService) DigestLLMService {
	return &digestLLMServiceImpl{llmService: llmService}
}

func (s *digestLLMServiceImpl) GenerateDigestContent(ctx context.Context, entries *miniflux.Entries) (*llm.LLMResponse, error) {
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
		return nil, fmt.Errorf("failed to marshal entries for LLM: %w", err)
	}

	prompt := llmPrompt + string(entriesJSON)

	llmResponse, err := s.llmService.GenerateContent(ctx, prompt, llmResponseSchema)
	if err != nil {
		return nil, err
	}

	var response llm.LLMResponse
	if err := json.Unmarshal([]byte(llmResponse), &response); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return &response, nil
}
