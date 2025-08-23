package llm

import (
	"context"
	"errors"
	"log"
	"time"

	"google.golang.org/genai"
	"miniflux-digest/internal/app/services"
)

const (
	Model               = "gemini-2.5-pro"
	maxRetries          = 3
	Temperature float32 = 0.4
)

type modelClient interface {
	GenerateContent(ctx context.Context, model string, parts []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error)
}

type GeminiService struct {
	client    modelClient
	modelName string
}

func NewGeminiService(apiKey string) (services.LLMService, error) {
	if apiKey == "" {
		return &GeminiService{modelName: Model}, nil
	}

	ctx := context.Background()
	clientConfig := genai.ClientConfig{APIKey: apiKey}
	client, err := genai.NewClient(ctx, &clientConfig)
	if err != nil {
		return nil, err
	}

	return &GeminiService{client: client.Models, modelName: Model}, nil
}

func (s *GeminiService) generateContentWithRetry(ctx context.Context, prompt string, schema *genai.Schema) (*genai.GenerateContentResponse, error) {
	var resp *genai.GenerateContentResponse
	var err error

	for i := range maxRetries {
		resp, err = s.client.GenerateContent(ctx, s.modelName, genai.Text(prompt), &genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
			ResponseSchema:   schema,
			Temperature:      genai.Ptr(Temperature),
		})

		if err == nil {
			return resp, nil
		}

		log.Printf("LLM call failed: %v", err)

		var apiError genai.APIError
		if errors.As(err, &apiError) {
			log.Printf("LLM API error: Code=%d, Status=%s, Message=%s", apiError.Code, apiError.Status, apiError.Message)
			if apiError.Code == 503 || apiError.Code == 500 {
				log.Printf("Retrying LLM call (%d/%d)...", i+1, maxRetries)
				time.Sleep(time.Second * time.Duration(i+1))
				continue
			}
		}

		return nil, err
	}

	return nil, err
}

func (s *GeminiService) GenerateContent(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
	if s.client == nil {
		return nil, errors.New("LLM service is disabled: no API key provided")
	}

	resp, err := s.generateContentWithRetry(ctx, prompt, schema)
	if err != nil {
		return nil, err
	}

	candidates := resp.Candidates

	if len(candidates) == 0 || len(candidates[0].Content.Parts) == 0 {
		log.Println("LLM: No content returned from LLM (empty candidates or parts).")
		return nil, errors.New("no content returned from LLM")
	}

	candidate := candidates[0]
	part := candidate.Content.Parts[0]
	jsonData := []byte(part.Text)

	if (candidate.FinishReason != genai.FinishReasonStop) {
		log.Printf("LLM: Unexpected finish reason: %s", candidate.FinishReason)
	}

	return jsonData, nil
}
