package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chatcomStackspotAI/models"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)

type ClaudeAIClient struct {
	apiKey string
	model  string
	logger *zap.Logger
	client *http.Client
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

// SendPrompt monta a requisição com o histórico e a envia para a ClaudeAI, retornando a resposta formatada
func (c *ClaudeAIClient) SendPrompt(ctx context.Context, prompt string, history []models.Message) (string, error) {
	messages := c.buildMessages(prompt, history)

	reqBody := map[string]interface{}{
		"model":      c.model,
		"max_tokens": 8192,
		"messages":   messages,
	}
	reqJSON, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", strings.NewReader(string(reqJSON)))
	if err != nil {
		c.logger.Error("Erro ao criar a requisição de prompt", zap.Error(err))
		return "", fmt.Errorf("erro ao criar a requisição: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-api-key", c.apiKey)
	req.Header.Add("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("Erro ao fazer a requisição de prompt", zap.Error(err))
		return "", fmt.Errorf("erro ao fazer a requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error("Erro ao obter resposta da ClaudeAI", zap.Int("status", resp.StatusCode), zap.String("body", string(body)))
		return "", fmt.Errorf("erro ao obter resposta da ClaudeAI: status %d, body %s", resp.StatusCode, string(body))
	}

	return c.parseResponse(resp)
}

// buildMessages monta o histórico de mensagens para incluir na requisição
func (c *ClaudeAIClient) buildMessages(prompt string, history []models.Message) []map[string]string {
	messages := make([]map[string]string, len(history))

	// Processa o histórico, garantindo que role e content estejam bem definidos
	for i, msg := range history {
		role := "user"
		if msg.Role == "assistant" {
			role = "assistant"
		}
		messages[i] = map[string]string{"role": role, "content": msg.Content}
	}

	// Adiciona a mensagem atual do usuário ao final
	messages = append(messages, map[string]string{"role": "user", "content": prompt})

	return messages
}

// parseResponse decodifica e processa a resposta da ClaudeAI
func (c *ClaudeAIClient) parseResponse(resp *http.Response) (string, error) {
	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error("Erro ao decodificar a resposta da ClaudeAI", zap.Error(err))
		return "", fmt.Errorf("erro ao decodificar a resposta: %w", err)
	}

	var responseText string
	for _, content := range result.Content {
		if content.Type == "text" {
			responseText += content.Text
		}
	}

	if responseText == "" {
		c.logger.Error("Nenhum conteúdo de texto encontrado na resposta da ClaudeAI")
		return "", fmt.Errorf("erro ao obter a resposta da ClaudeAI")
	}

	return responseText, nil
}
