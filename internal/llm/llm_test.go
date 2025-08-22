package llm

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/genai"
)



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
	service, err := NewGeminiService("")
	if err != nil {
		t.Fatalf("NewGeminiService with empty API key should not return an error, but got: %v", err)
	}
	if service == nil {
		t.Fatal("Service should not be nil")
	}
}

func TestGeminiService_GenerateContent_Success(t *testing.T) {
	mockClient := &mockModelClient{
		GenerateContentFunc: func(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
			part := genai.Text("{\"overview_summary\": \"test summary\", \"group_summaries\": []}")
			return &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: part[0],
					},
				},
			}, nil
		},
	}

	service := &GeminiService{client: mockClient, modelName: "test-model"}

	resp, err := service.GenerateContent(context.Background(), "test prompt", nil)
	if err != nil {
		t.Fatalf("GenerateContent should not return an error, but got: %v", err)
	}
	if resp != "{\"overview_summary\": \"test summary\", \"group_summaries\": []}" {
		t.Errorf("Expected response to be '{\"overview_summary\": \"test summary\", \"group_summaries\": []}', but got: %s", resp)
	}
}

func TestGeminiService_GenerateContent_Disabled(t *testing.T) {
	service := &GeminiService{client: nil}

	_, err := service.GenerateContent(context.Background(), "test prompt", nil)
	if err == nil {
		t.Fatal("GenerateContent should return an error when the service is disabled")
	}
	if err.Error() != "LLM service is disabled: no API key provided" {
		t.Errorf("Expected error message to be 'LLM service is disabled: no API key provided', but got: %s", err.Error())
	}
}

func TestGeminiService_GenerateContent_Error(t *testing.T) {
	mockClient := &mockModelClient{
		GenerateContentFunc: func(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
			return nil, errors.New("test error")
		},
	}

	service := &GeminiService{client: mockClient, modelName: "test-model"}

	_, err := service.GenerateContent(context.Background(), "test prompt", nil)
	if err == nil {
		t.Fatal("GenerateContent should return an error")
	}
	if err.Error() != "test error" {
		t.Errorf("Expected error message to be 'test error', but got: %s", err.Error())
	}
}

func TestGeminiService_GenerateContent_EmptyResponse(t *testing.T) {
	mockClient := &mockModelClient{
		GenerateContentFunc: func(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
			return &genai.GenerateContentResponse{}, nil
		},
	}

	service := &GeminiService{client: mockClient, modelName: "test-model"}

	_, err := service.GenerateContent(context.Background(), "test prompt", nil)
	if err == nil {
		t.Fatal("GenerateContent should return an error for an empty response")
	}
	if err.Error() != "no content returned from LLM" {
		t.Errorf("Expected error message to be 'no content returned from LLM', but got: %s", err.Error())
	}
}
