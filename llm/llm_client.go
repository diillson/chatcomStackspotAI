// llm/llm_client.go

package llm

import "github.com/chatcomStackspotAI/models"

type LLMClient interface {
	SendPrompt(prompt string, history []models.Message) (response string, err error)
}
