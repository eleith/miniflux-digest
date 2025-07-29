package llm

import (
	"context"

	"google.golang.org/genai"
)

type LLMService interface {
	GenerateContent(ctx context.Context, prompt string, schema *genai.Schema) (string, error)
}