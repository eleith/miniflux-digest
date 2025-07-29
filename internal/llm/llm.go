package llm

import (
	"context"
	"errors"

	"google.golang.org/genai"
)

type GeminiService struct {
	client *genai.Client
	modelName string
}

func NewGeminiService(apiKey, modelName string) (*GeminiService, error) {
	ctx := context.Background()
	clientConfig := genai.ClientConfig{APIKey: apiKey}
	client, err := genai.NewClient(ctx, &clientConfig)
	if err != nil {
		return nil, err
	}

	return &GeminiService{client: client, modelName: modelName}, nil
}

func (s *GeminiService) GenerateContent(ctx context.Context, prompt string, schema *genai.Schema) (string, error) {
	resp, err := s.client.Models.GenerateContent(ctx, s.modelName, genai.Text(prompt), &genai.GenerateContentConfig{
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
