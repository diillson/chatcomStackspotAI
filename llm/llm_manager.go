package llm

import (
	"fmt"
	"os"
)

type LLMManager struct {
	clients map[string]func(string) (LLMClient, error)
}

func NewLLMManager() (*LLMManager, error) {
	manager := &LLMManager{
		clients: make(map[string]func(string) (LLMClient, error)),
	}

	// Configurar a fábrica para OpenAI
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY não está definido")
	}

	manager.clients["OPENAI"] = func(model string) (LLMClient, error) {
		if model == "" {
			model = "gpt-3.5-turbo" // Modelo padrão
		}
		return NewOpenAIClient(apiKey, model), nil
	}

	// Configurar a fábrica para StackSpot
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	slug := os.Getenv("SLUG_NAME")
	if clientID == "" || clientSecret == "" || slug == "" {
		return nil, fmt.Errorf("As credenciais do StackSpot não estão definidas")
	}
	tokenManager := NewTokenManager(clientID, clientSecret)

	manager.clients["STACKSPOT"] = func(model string) (LLMClient, error) {
		// StackSpotClient não usa o parâmetro model, mas mantemos a assinatura consistente
		return NewStackSpotClient(tokenManager, slug), nil
	}

	return manager, nil
}

func (m *LLMManager) GetClient(provider string, model string) (LLMClient, error) {
	factoryFunc, ok := m.clients[provider]
	if !ok {
		return nil, fmt.Errorf("Provedor LLM '%s' não suportado", provider)
	}
	client, err := factoryFunc(model)
	if err != nil {
		return nil, err
	}
	return client, nil
}
