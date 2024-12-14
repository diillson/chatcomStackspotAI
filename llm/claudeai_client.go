package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/chatcomStackspotAI/models"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"time"
)

type ClaudeAIClient struct {
	apiKey string
	model  string
	logger *zap.Logger
}

func NewClaudeAIClient(apiKey, model string, logger *zap.Logger) *ClaudeAIClient {
	return &ClaudeAIClient{
		apiKey: apiKey,
		model:  model,
		logger: logger,
	}
}

func (c *ClaudeAIClient) GetModelName() string {
	return c.model
}

func (c *ClaudeAIClient) SendPrompt(ctx context.Context, prompt string, history []models.Message) (string, error) {
	url := "https://api.claudeai.com/v1/completions"

	// Construir o array de mensagens
	messages := []map[string]string{}

	// Adicionar o histórico
	for _, msg := range history {
		messages = append(messages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	// Adicionar a nova mensagem do usuário
	messages = append(messages, map[string]string{
		"role":    "user",
		"content": prompt,
	})

	payload := map[string]interface{}{
		"model":    c.model,
		"messages": messages,
	}

	jsonValue, _ := json.Marshal(payload)

	maxAttempts := 3
	backoff := time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonValue))
		if err != nil {
			return "", fmt.Errorf("erro ao criar a requisição: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)

		client := &http.Client{
			Timeout: 30 * time.Second,
		}
		resp, err := client.Do(req)
		if err != nil {
			if isTemporaryError(err) {
				c.logger.Warn("Erro temporário ao chamar ClaudeAI", zap.Int("attempt", attempt), zap.Error(err))
				if attempt < maxAttempts {
					time.Sleep(backoff)
					backoff *= 2 // Backoff exponencial
					continue
				}
			}
			return "", fmt.Errorf("erro ao fazer a requisição para ClaudeAI: %w", err)
		}
		defer resp.Body.Close()

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("erro ao ler a resposta da ClaudeAI: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			errMsg := fmt.Sprintf("Erro na requisição à ClaudeAI: status %d, resposta: %s", resp.StatusCode, string(bodyBytes))
			return "", fmt.Errorf(errMsg)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			return "", fmt.Errorf("erro ao decodificar a resposta da ClaudeAI: %w", err)
		}

		choices, ok := result["choices"].([]interface{})
		if !ok || len(choices) == 0 {
			return "", fmt.Errorf("Nenhuma resposta recebida da ClaudeAI")
		}

		firstChoice := choices[0].(map[string]interface{})
		content := firstChoice["text"].(string)

		return content, nil
	}

	return "", fmt.Errorf("Falha ao obter resposta da ClaudeAI após %d tentativas", maxAttempts)
}
