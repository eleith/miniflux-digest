package services

import (
	"context"

	"google.golang.org/genai"
)

// TODO: To make timeouts customizable per-call, consider adding an optional timeout
// parameter to the GenerateContent method, e.g.:
// GenerateContent(ctx context.Context, prompt string, schema *genai.Schema, timeout ...time.Duration) ([]byte, error)

type LLMService interface {
	GenerateContent(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error)
	GenerateContentWithResponse(ctx context.Context, prompt string, schema *genai.Schema, response interface{}) error
}
