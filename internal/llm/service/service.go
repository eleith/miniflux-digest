package service

import (
	"context"

	"google.golang.org/genai"
)

type LLMResponse struct {
	OverviewSummary string `json:"overview_summary"`
	GroupSummaries  []struct {
		Title   string `json:"title"`
		Summary string `json:"summary"`
		EntryIDs []int64 `json:"entry_ids"`
	} `json:"group_summaries"`
}

type LLMService interface {
	GenerateContent(ctx context.Context, prompt string, schema *genai.Schema) (string, error)
}
