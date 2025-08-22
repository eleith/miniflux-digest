package llm

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/genai"
)

func TestGenerateContentWithRetry(t *testing.T) {
	mockClient := &mockModelClient{}
	service := &GeminiService{
		client:    mockClient,
		modelName: "test-model",
	}

	// Setup the mock to fail twice with a 503 error, then succeed.
	attempts := 0
	mockClient.GenerateContentFunc = func(ctx context.Context, model string, parts []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
		attempts++
		if attempts <= 2 {
			return nil, genai.APIError{Code: 503, Message: "Model overloaded"}
		}
		return &genai.GenerateContentResponse{}, nil
	}

	_, err := service.generateContentWithRetry(context.Background(), "test prompt", nil)

	if err != nil {
		t.Fatalf("Expected generateContentWithRetry to succeed, but it failed: %v", err)
	}

	if mockClient.callCount != 3 {
		t.Errorf("Expected GenerateContent to be called 3 times, but it was called %d times", mockClient.callCount)
	}
}

func TestGenerateContentWithRetry_NonRecoverableError(t *testing.T) {
	mockClient := &mockModelClient{}
	service := &GeminiService{
		client:    mockClient,
		modelName: "test-model",
	}

	// Setup the mock to fail with a non-503 error.
	mockClient.GenerateContentFunc = func(ctx context.Context, model string, parts []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
		return nil, errors.New("a non-recoverable error")
	}

	_, err := service.generateContentWithRetry(context.Background(), "test prompt", nil)

	if err == nil {
		t.Fatal("Expected generateContentWithRetry to fail, but it succeeded")
	}

	if mockClient.callCount != 1 {
		t.Errorf("Expected GenerateContent to be called 1 time, but it was called %d times", mockClient.callCount)
	}
}