// llm/llm_manager.go

package llm

import (
	"fmt"
	"os"
)

type LLMManager struct {
	clients map[string]LLMClient
}

func NewLLMManager() (*LLMManager, error) {
	manager := &LLMManager{
		clients: make(map[string]LLMClient),
	}

	// Inicializar os clientes para cada provedor
	// StackSpotAI
	if os.Getenv("CLIENT_ID") != "" && os.Getenv("CLIENT_SECRET") != "" && os.Getenv("SLUG_NAME") != "" {
		clientID := os.Getenv("CLIENT_ID")
		clientSecret := os.Getenv("CLIENT_SECRET")
		slug := os.Getenv("SLUG_NAME")
		tokenManager := NewTokenManager(clientID, clientSecret)
		manager.clients["STACKSPOT"] = NewStackSpotClient(tokenManager, slug)
	}

	// OpenAI
	if os.Getenv("OPENAI_API_KEY") != "" {
		apiKey := os.Getenv("OPENAI_API_KEY")
		model := os.Getenv("OPENAI_MODEL")
		if model == "" {
			model = "gpt-4o-mini"
		}
		manager.clients["OPENAI"] = NewOpenAIClient(apiKey, model)
	}

	return manager, nil
}

func (m *LLMManager) GetClient(provider string) (LLMClient, error) {
	client, exists := m.clients[provider]
	if !exists {
		return nil, fmt.Errorf("Provedor LLM '%s' não suportado ou não configurado", provider)
	}
	return client, nil
}
