package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"google.golang.org/genai"
	"miniflux-digest/internal/app/services"
)

const (
	Model               = "gemini-2.5-flash"
	maxRetries          = 3
	perTryTimeout       = 5 * time.Minute
	Temperature float32 = 0.4
)

type modelClient interface {
	GenerateContent(ctx context.Context, model string, parts []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error)
}

type GeminiService struct {
	client    modelClient
	modelName string
	sleep     func(time.Duration)
}

func NewGeminiService(apiKey string) (services.LLMService, error) {
	if apiKey == "" {
		return &GeminiService{modelName: Model, sleep: time.Sleep}, nil
	}

	ctx := context.Background()
	clientConfig := genai.ClientConfig{APIKey: apiKey}
	client, err := genai.NewClient(ctx, &clientConfig)
	if err != nil {
		return nil, err
	}

	return &GeminiService{client: client.Models, modelName: Model, sleep: time.Sleep}, nil
}

func (s *GeminiService) generateContentWithRetry(ctx context.Context, prompt string, schema *genai.Schema) (*genai.GenerateContentResponse, error) {
	var lastErr error

	for i := 1; i <= maxRetries; i++ {
		tryCtx, tryCancel := context.WithTimeout(ctx, perTryTimeout)
		resp, err := s.client.GenerateContent(tryCtx, s.modelName, genai.Text(prompt), &genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
			ResponseSchema:   schema,
			Temperature:      genai.Ptr(Temperature),
		})
		tryCancel()

		if err == nil {
			return resp, nil
		}

		lastErr = err

		log.Printf("LLM attempt %d/%d failed: %v", i, maxRetries, err)

		var apiError genai.APIError
		if errors.As(err, &apiError) && (apiError.Code == 503 || apiError.Code == 500) {
			if i < maxRetries {
				log.Printf("Retrying after server error...")
				s.sleep(time.Second * time.Duration(i))
				continue
			}
		}

		break
	}

	return nil, lastErr
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

func (s *GeminiService) GenerateContentWithResponse(ctx context.Context, prompt string, schema *genai.Schema, response interface{}) error {
	llmResponseBytes, err := s.GenerateContent(ctx, prompt, schema)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(llmResponseBytes, &response); err != nil {
		return fmt.Errorf("failed to parse LLM response: %w", err)
	}
	return nil
}