package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/chatcomStackspotAI/models"
	"io/ioutil"
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
