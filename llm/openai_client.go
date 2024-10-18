// llm/openai_client.go

package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/chatcomStackspotAI/models"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type OpenAIClient struct {
	apiKey string
	model  string
}

func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	return &OpenAIClient{
		apiKey: apiKey,
		model:  model,
	}
}

func (c *OpenAIClient) GetModelName() string {
	return c.model
}

func (c *OpenAIClient) SendPrompt(prompt string, history []models.Message) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	// Construir o array de mensagens
	messages := []map[string]string{}

	// Adicionar o histórico existente
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
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)

		client := &http.Client{
			Timeout: 30 * time.Second,
		}
		resp, err := client.Do(req)
		if err != nil {
			// Verificar se o erro é temporário ou timeout
			if isTemporaryError(err) {
				log.Printf("Erro temporário ao chamar OpenAI (tentativa %d/%d): %v", attempt, maxAttempts, err)
				if attempt < maxAttempts {
					time.Sleep(backoff)
					backoff *= 2 // Backoff exponencial
					continue
				}
			}
			return "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			errMsg := fmt.Sprintf("Erro na requisição à OpenAI: status %d, resposta: %s", resp.StatusCode, string(bodyBytes))
			return "", fmt.Errorf(errMsg)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return "", err
		}

		choices, ok := result["choices"].([]interface{})
		if !ok || len(choices) == 0 {
			return "", fmt.Errorf("Nenhuma resposta recebida da OpenAI")
		}

		firstChoice := choices[0].(map[string]interface{})
		message := firstChoice["message"].(map[string]interface{})
		content := message["content"].(string)

		return content, nil
	}

	return "", fmt.Errorf("Falha ao obter resposta da OpenAI após %d tentativas", maxAttempts)
}
