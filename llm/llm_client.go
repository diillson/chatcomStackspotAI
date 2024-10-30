package llm

import (
	"context"
	"github.com/chatcomStackspotAI/models"
)

type LLMClient interface {
	SendPrompt(ctx context.Context, prompt string, history []models.Message) (response string, err error)
	GetModelName() string
}
