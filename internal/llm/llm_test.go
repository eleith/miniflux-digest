package llm

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/genai"
)

// MockLLMService is a mock implementation of the llm.LLMService interface.
type MockLLMService struct{}

func (m *MockLLMService) GenerateContent(ctx context.Context, prompt string, schema *genai.Schema) (string, error) {
	return "", nil
}

// mockModelClient is a mock implementation of the modelClient interface.
type mockModelClient struct {
	GenerateContentFunc func(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error)
}

func (m *mockModelClient) GenerateContent(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	if m.GenerateContentFunc != nil {
		return m.GenerateContentFunc(ctx, model, contents, config)
	}
	return nil, errors.New("GenerateContentFunc not implemented")
}

func TestNewGeminiService(t *testing.T) {
	// Test with an empty API key
	service, err := NewGeminiService("")
	if err != nil {
		t.Fatalf("NewGeminiService with empty API key should not return an error, but got: %v", err)
	}
	if service == nil {
		t.Fatal("Service should not be nil")
	}
	if service.client != nil {
		t.Error("Service client should be nil for empty API key")
	}
}

func TestGeminiService_GenerateContent_Success(t *testing.T) {
	// Create a mock model that returns a successful response
	mockClient := &mockModelClient{
		GenerateContentFunc: func(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
			return &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []*genai.Part{{Text: "test response"}},
						},
					},
				},
			}, nil
		},
	}

	// Create a GeminiService with the mock model
	service := &GeminiService{client: mockClient, modelName: "test-model"}

	// Call GenerateContent and assert that it returns the expected response
	resp, err := service.GenerateContent(context.Background(), "test prompt", nil)
	if err != nil {
		t.Fatalf("GenerateContent should not return an error, but got: %v", err)
	}
	if resp != "test response" {
		t.Errorf("Expected response to be 'test response', but got: %s", resp)
	}
}

func TestGeminiService_GenerateContent_Disabled(t *testing.T) {
	// Create a GeminiService with a nil client
	service := &GeminiService{client: nil}

	// Call GenerateContent and assert that it returns an error
	_, err := service.GenerateContent(context.Background(), "test prompt", nil)
	if err == nil {
		t.Fatal("GenerateContent should return an error when the service is disabled")
	}
	if err.Error() != "LLM service is disabled: no API key provided" {
		t.Errorf("Expected error message to be 'LLM service is disabled: no API key provided', but got: %s", err.Error())
	}
}

func TestGeminiService_GenerateContent_Error(t *testing.T) {
	// Create a mock model that returns an an error
	mockClient := &mockModelClient{
		GenerateContentFunc: func(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
			return nil, errors.New("test error")
		},
	}

	// Create a GeminiService with the mock model
	service := &GeminiService{client: mockClient, modelName: "test-model"}

	// Call GenerateContent and assert that it returns an error
	_, err := service.GenerateContent(context.Background(), "test prompt", nil)
	if err == nil {
		t.Fatal("GenerateContent should return an error")
	}
	if err.Error() != "test error" {
		t.Errorf("Expected error message to be 'test error', but got: %s", err.Error())
	}
}

func TestGeminiService_GenerateContent_EmptyResponse(t *testing.T) {
	// Create a mock model that returns an empty response
	mockClient := &mockModelClient{
		GenerateContentFunc: func(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
			return &genai.GenerateContentResponse{}, nil
		},
	}

	// Create a GeminiService with the mock model
	service := &GeminiService{client: mockClient, modelName: "test-model"}

	// Call GenerateContent and assert that it returns an error
	_, err := service.GenerateContent(context.Background(), "test prompt", nil)
	if err == nil {
		t.Fatal("GenerateContent should return an error for an empty response")
	}
	if err.Error() != "no content returned from LLM" {
		t.Errorf("Expected error message to be 'no content returned from LLM', but got: %s", err.Error())
	}
}
