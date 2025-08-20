package llm

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

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

const (
	GeminiModel = "gemini-1.5-flash"
	maxRetries  = 3
)

type modelClient interface {
	GenerateContent(ctx context.Context, model string, parts []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error)
}

type GeminiService struct {
	client    modelClient
	modelName string
}

func NewGeminiService(apiKey string) (*GeminiService, error) {
	if apiKey == "" {
		return &GeminiService{modelName: GeminiModel}, nil
	}

	ctx := context.Background()
	clientConfig := genai.ClientConfig{APIKey: apiKey}
	client, err := genai.NewClient(ctx, &clientConfig)
	if err != nil {
		return nil, err
	}

	return &GeminiService{client: client.Models, modelName: GeminiModel}, nil
}

func (s *GeminiService) generateContentWithRetry(ctx context.Context, prompt string, schema *genai.Schema) (*genai.GenerateContentResponse, error) {
	var resp *genai.GenerateContentResponse
	var err error

	for i := 0; i < maxRetries; i++ {
		resp, err = s.client.GenerateContent(ctx, s.modelName, genai.Text(prompt), &genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
			ResponseSchema:   schema,
		})
		if err == nil {
			return resp, nil
		}

		var apiError *genai.APIError
		if errors.As(err, &apiError) {
			log.Printf("LLM API error: Code=%d, Status=%s, Message=%s", apiError.Code, apiError.Status, apiError.Message)
			if apiError.Code == 503 && strings.Contains(apiError.Message, "overloaded") {
				log.Printf("Retrying LLM call (%d/%d)...", i+1, maxRetries-1)
				time.Sleep(time.Second * time.Duration(i+1)) // Wait 1, 2 seconds
				continue
			}
		}

		return nil, err
	}

	return nil, err
}

func (s *GeminiService) GenerateContent(ctx context.Context, prompt string, schema *genai.Schema) (string, error) {
	if s.client == nil {
		return "", errors.New("LLM service is disabled: no API key provided")
	}

	resp, err := s.generateContentWithRetry(ctx, prompt, schema)
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("no content returned from LLM")
	}

	// The API returns JSON as a Text part when a schema is provided
	textPart := resp.Text()
	return textPart, nil
}
