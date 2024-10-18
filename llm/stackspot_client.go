// llm/stackspot_client.go

package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/chatcomStackspotAI/models"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type StackSpotClient struct {
	tokenManager *TokenManager
	slug         string
}

func NewStackSpotClient(tokenManager *TokenManager, slug string) *StackSpotClient {
	return &StackSpotClient{
		tokenManager: tokenManager,
		slug:         slug,
	}
}

func (c *StackSpotClient) GetModelName() string {
	return "StackSpotAI"
}

// Função para formatar o histórico da conversa
func formatConversationHistory(history []models.Message) string {
	var conversationBuilder strings.Builder
	for _, msg := range history {
		role := "Usuário"
		if msg.Role == "assistant" {
			role = "Assistente"
		}
		conversationBuilder.WriteString(fmt.Sprintf("%s: %s\n", role, msg.Content))
	}
	return conversationBuilder.String()
}

func (c *StackSpotClient) SendPrompt(prompt string, history []models.Message) (string, error) {
	token, err := c.tokenManager.GetAccessToken()
	if err != nil {
		log.Printf("Erro ao obter o token: %v", err)
		return "", fmt.Errorf("Erro ao obter o token: %v", err)
	}

	// Formatar o histórico da conversa
	conversationHistory := formatConversationHistory(history)

	// Concatenar o histórico com o prompt atual
	fullPrompt := fmt.Sprintf("%sUsuário: %s", conversationHistory, prompt)

	// Enviar o prompt completo e obter o responseID
	responseID, err := c.sendRequestToLLMWithRetry(fullPrompt, token)
	if err != nil {
		log.Printf("Erro ao enviar a requisição para a LLM: %v", err)
		return "", fmt.Errorf("Erro ao enviar a requisição: %v", err)
	}

	var llmResponse string
	maxAttempts := 50
	for i := 0; i < maxAttempts; i++ {
		time.Sleep(2 * time.Second)

		llmResponse, err = c.getLLMResponseWithRetry(responseID, token)
		if err == nil {
			return llmResponse, nil
		}

		if strings.Contains(err.Error(), "resposta ainda não está pronta") {
			log.Printf("Resposta ainda não está pronta, tentativa %d/%d", i+1, maxAttempts)
			continue
		}

		if strings.Contains(err.Error(), "a execução da LLM falhou") {
			log.Printf("Falha na execução da LLM: %v", err)
			return "", fmt.Errorf("A LLM não pôde processar a solicitação.")
		}

		log.Printf("Erro ao obter a resposta da LLM: %v", err)
		return "", fmt.Errorf("Erro ao obter a resposta: %v", err)
	}

	log.Printf("Timeout ao obter a resposta da LLM: %v", err)
	return "", fmt.Errorf("Timeout ao obter a resposta da LLM")
}

// Implementação das funções auxiliares com retry

func (c *StackSpotClient) sendRequestToLLMWithRetry(prompt, accessToken string) (string, error) {
	maxAttempts := 3
	backoff := time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		responseID, err := c.sendRequestToLLM(prompt, accessToken)
		if err != nil {
			if isTemporaryError(err) {
				log.Printf("Erro temporário ao enviar requisição para StackSpotAI (tentativa %d/%d): %v", attempt, maxAttempts, err)
				if attempt < maxAttempts {
					time.Sleep(backoff)
					backoff *= 2 // Backoff exponencial
					continue
				}
			}
			return "", err
		}
		return responseID, nil
	}

	return "", fmt.Errorf("Falha ao enviar requisição para StackSpotAI após %d tentativas", maxAttempts)
}

func (c *StackSpotClient) getLLMResponseWithRetry(responseID, accessToken string) (string, error) {
	maxAttempts := 3
	backoff := time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		llmResponse, err := c.getLLMResponse(responseID, accessToken)
		if err != nil {
			if isTemporaryError(err) {
				log.Printf("Erro temporário ao obter resposta da StackSpotAI (tentativa %d/%d): %v", attempt, maxAttempts, err)
				if attempt < maxAttempts {
					time.Sleep(backoff)
					backoff *= 2 // Backoff exponencial
					continue
				}
			}
			return "", err
		}
		return llmResponse, nil
	}

	return "", fmt.Errorf("Falha ao obter resposta da StackSpotAI após %d tentativas", maxAttempts)
}

func (c *StackSpotClient) sendRequestToLLM(prompt, accessToken string) (string, error) {
	conversationID := uuid.New().String()

	url := fmt.Sprintf("https://genai-code-buddy-api.stackspot.com/v1/quick-commands/create-execution/%s?conversation_id=%s", c.slug, conversationID)
	log.Printf("Fazendo POST para URL: %s", url)

	requestBody := map[string]string{
		"input_data": prompt,
	}
	jsonValue, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Erro na requisição à LLM: status %d, resposta: %s", resp.StatusCode, string(bodyBytes))
		return "", fmt.Errorf("erro na requisição à LLM: status %d, resposta: %s", resp.StatusCode, string(bodyBytes))
	}

	var responseID string
	if err := json.Unmarshal(bodyBytes, &responseID); err != nil {
		log.Printf("Erro ao deserializar o responseID: %v", err)
		return "", err
	}

	log.Printf("Response ID recebido: %s", responseID)
	return responseID, nil
}

func (c *StackSpotClient) getLLMResponse(responseID, accessToken string) (string, error) {
	url := fmt.Sprintf("https://genai-code-buddy-api.stackspot.com/v1/quick-commands/callback/%s", responseID)
	log.Printf("Fazendo GET para URL: %s", url)
	log.Printf("Usando Token de Acesso: %s...", accessToken[:10])

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Erro ao criar a requisição GET: %v", err)
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Erro na requisição GET para a LLM: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Erro ao ler o corpo da resposta da LLM: %v", err)
		return "", err
	}

	log.Printf("Status Code: %d", resp.StatusCode)
	log.Printf("Headers da Resposta: %v", resp.Header)
	log.Printf("Corpo da Resposta: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("erro na requisição de callback: status %d, resposta: %s", resp.StatusCode, string(bodyBytes))
	}

	var callbackResponse CallbackResponse
	if err := json.Unmarshal(bodyBytes, &callbackResponse); err != nil {
		log.Printf("Erro ao deserializar a resposta JSON: %v", err)
		return "", err
	}

	switch callbackResponse.Progress.Status {
	case "COMPLETED":
		if len(callbackResponse.Steps) > 0 {
			lastStepIndex := len(callbackResponse.Steps) - 1
			lastStep := callbackResponse.Steps[lastStepIndex]
			llmAnswer := lastStep.StepResult.Answer
			return llmAnswer, nil
		} else {
			return "", fmt.Errorf("nenhuma resposta disponível")
		}
	case "FAILURE":
		log.Printf("A execução falhou com status: %s", callbackResponse.Progress.Status)
		return "", fmt.Errorf("a execução da LLM falhou")
	default:
		log.Printf("Status da execução: %s", callbackResponse.Progress.Status)
		return "", fmt.Errorf("resposta ainda não está pronta")
	}
}

// Estruturas para decodificar a resposta da LLM

type CallbackResponse struct {
	ExecutionID      string   `json:"execution_id"`
	QuickCommandSlug string   `json:"quick_command_slug"`
	ConversationID   string   `json:"conversation_id"`
	Progress         Progress `json:"progress"`
	Steps            []Step   `json:"steps"`
	Result           string   `json:"result"`
}

type Progress struct {
	Start               string  `json:"start"`
	End                 string  `json:"end"`
	Duration            int     `json:"duration"`
	ExecutionPercentage float64 `json:"execution_percentage"`
	Status              string  `json:"status"`
}

type Step struct {
	StepName       string     `json:"step_name"`
	ExecutionOrder int        `json:"execution_order"`
	Type           string     `json:"type"`
	StepResult     StepResult `json:"step_result"`
}

type Source struct {
	Type          string  `json:"type,omitempty"`
	Name          string  `json:"name,omitempty"`
	Slug          string  `json:"slug,omitempty"`
	DocumentType  string  `json:"document_type,omitempty"`
	DocumentScore float64 `json:"document_score,omitempty"`
	DocumentID    string  `json:"document_id,omitempty"`
}

type StepResult struct {
	Answer  string   `json:"answer"`
	Sources []Source `json:"sources"`
}

// Implementação do TokenManager

type TokenManager struct {
	clientID     string
	clientSecret string
	accessToken  string
	expiresAt    time.Time
	mu           sync.RWMutex
}

func NewTokenManager(clientID, clientSecret string) *TokenManager {
	return &TokenManager{
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

func (tm *TokenManager) GetAccessToken() (string, error) {
	tm.mu.RLock()
	token := tm.accessToken
	expiration := tm.expiresAt
	tm.mu.RUnlock()

	if time.Until(expiration) > 60*time.Second && token != "" {
		return token, nil
	}

	return tm.refreshToken()
}

func (tm *TokenManager) refreshToken() (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	log.Println("Renovando o access token...")

	tokenURL := "https://idm.stackspot.com/zup/oidc/oauth/token"
	data := strings.NewReader(fmt.Sprintf(
		"grant_type=client_credentials&client_id=%s&client_secret=%s",
		tm.clientID, tm.clientSecret))

	req, err := http.NewRequest("POST", tokenURL, data)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
		errMsg := fmt.Sprintf("falha ao obter o token: %s", string(bodyBytes))
		log.Printf(errMsg)
		return "", fmt.Errorf(errMsg)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	accessToken, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("não foi possível obter o access_token")
	}

	expiresIn, ok := result["expires_in"].(float64)
	if !ok {
		return "", fmt.Errorf("não foi possível obter expires_in")
	}

	tm.accessToken = accessToken
	tm.expiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
	log.Printf("Token renovado com sucesso. Expira em: %s", tm.expiresAt)

	return tm.accessToken, nil
}
