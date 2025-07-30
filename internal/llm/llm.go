package llm

import (
	"context"
	"errors"

	"google.golang.org/genai"
)

const GeminiModel = "gemini-1.5-flash"

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

func (s *GeminiService) GenerateContent(ctx context.Context, prompt string, schema *genai.Schema) (string, error) {
	if s.client == nil {
		return "", errors.New("LLM service is disabled: no API key provided")
	}

	resp, err := s.client.GenerateContent(ctx, s.modelName, genai.Text(prompt), &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema:   schema,
	})
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
