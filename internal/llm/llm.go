package llm

import (
	"context"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiService struct {
	client *genai.GenerativeModel
}

func NewGeminiService(apiKey, modelName string) (*GeminiService, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	model := client.GenerativeModel(modelName)
	return &GeminiService{client: model}, nil
}

func (s *GeminiService) GenerateContent(ctx context.Context, prompt string) (string, error) {
	resp, err := s.client.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", nil
	}

	return string(resp.Candidates[0].Content.Parts[0].(genai.Text)), nil
}
