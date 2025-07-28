package llm

import "context"

type LLMService interface {
	GenerateContent(ctx context.Context, prompt string) (string, error)
}
