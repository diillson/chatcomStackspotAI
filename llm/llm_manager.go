// llm/llm_manager.go

package llm

import (
	"fmt"
	"log"
	"os"
	"sync"
)

type LLMManager struct {
	clients     map[string]LLMClient
	currentProv string
	mu          sync.RWMutex
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

	// Definir o provedor atual (pode ser padrão ou lido de uma configuração)
	manager.currentProv = os.Getenv("LLM_PROVIDER")
	if manager.currentProv == "" {
		manager.currentProv = "STACKSPOT" // ou outro padrão
	}

	return manager, nil
}

func (m *LLMManager) GetClient() (LLMClient, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, exists := m.clients[m.currentProv]
	if !exists {
		return nil, "", fmt.Errorf("Provedor LLM '%s' não suportado ou não configurado", m.currentProv)
	}
	return client, m.currentProv, nil
}

// Metodo para atualizar o provedor atual em tempo de execução
func (m *LLMManager) SetCurrentProvider(provider string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.clients[provider]; !exists {
		log.Printf("Provedor LLM '%s' não suportado ou não configurado", provider)
		return fmt.Errorf("Provedor LLM '%s' não suportado ou não configurado", provider)
	}
	m.currentProv = provider
	log.Printf("Provedor LLM atualizado para '%s'", provider)
	return nil
}
