package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// TokenManager gerencia o access token, cuidando da sua obtenção e renovação.
type TokenManager struct {
	clientID     string
	clientSecret string
	accessToken  string
	expiresAt    time.Time
	mu           sync.RWMutex
}

// NewTokenManager cria uma nova instância de TokenManager.
func NewTokenManager(clientID, clientSecret string) *TokenManager {
	return &TokenManager{
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

// GetAccessToken retorna o token atual. Se estiver expirado ou prestes a expirar, renova-o.
func (tm *TokenManager) GetAccessToken() (string, error) {
	tm.mu.RLock()
	token := tm.accessToken
	expiration := tm.expiresAt
	tm.mu.RUnlock()

	// Se o token estiver prestes a expirar (em menos de 60 segundos), renova-o.
	if time.Until(expiration) > 60*time.Second && token != "" {
		return token, nil
	}

	// Token expirado ou prestes a expirar, renovar.
	return tm.refreshToken()
}

// refreshToken obtém um novo token e atualiza o cache.
func (tm *TokenManager) refreshToken() (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	log.Println("Renovando o access token...")

	tokenURL := "https://idm.stackspot.com/zup/oidc/oauth/token"
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", tm.clientID)
	data.Set("client_secret", tm.clientSecret)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
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

	// Atualizar o token e o tempo de expiração
	tm.accessToken = accessToken
	tm.expiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
	log.Printf("Token renovado com sucesso. Expira em: %s", tm.expiresAt)

	return tm.accessToken, nil
}

// Função para enviar a solicitação para a LLM
func sendRequestToLLM(prompt, accessToken string) (string, error) {
	conversationID := uuid.New().String() // Gerando um conversation_id único
	slug := "SUA SLUG AQUI"

	url := fmt.Sprintf("https://genai-code-buddy-api.stackspot.com/v1/quick-commands/create-execution/%s?conversation_id=%s", slug, conversationID)
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

	client := &http.Client{}
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

	// Decodificar a resposta JSON para obter o responseID sem aspas
	var responseID string
	if err := json.Unmarshal(bodyBytes, &responseID); err != nil {
		log.Printf("Erro ao deserializar o responseID: %v", err)
		return "", err
	}

	log.Printf("Response ID recebido: %s", responseID)
	return responseID, nil
}

// Função para obter a resposta da LLM
func getLLMResponse(responseID, accessToken string) (string, error) {
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

	client := &http.Client{}
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

	if callbackResponse.Progress.Status != "COMPLETED" {
		log.Printf("Status da execução: %s", callbackResponse.Progress.Status)
		return "", fmt.Errorf("resposta ainda não está pronta")
	}

	if len(callbackResponse.Steps) > 0 {
		llmAnswer := callbackResponse.Steps[0].StepResult.Answer
		return llmAnswer, nil
	} else {
		return "", fmt.Errorf("nenhuma resposta disponível")
	}
}

// Função para lidar com o envio da mensagem
func sendMessageHandler(w http.ResponseWriter, r *http.Request, tm *TokenManager) {
	if r.Method == "POST" {
		var data struct {
			Prompt string `json:"prompt"`
		}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			log.Printf("Erro ao decodificar o JSON: %v", err)
			http.Error(w, "Dados inválidos", http.StatusBadRequest)
			return
		}

		token, err := tm.GetAccessToken()
		if err != nil {
			log.Printf("Erro ao obter o token: %v", err)
			http.Error(w, fmt.Sprintf("Erro ao obter o token: %v", err), http.StatusInternalServerError)
			return
		}
		log.Printf("Token obtido com sucesso")

		log.Printf("Enviando prompt para a LLM")
		responseID, err := sendRequestToLLM(data.Prompt, token)
		if err != nil {
			log.Printf("Erro ao enviar a requisição para a LLM: %v", err)
			http.Error(w, fmt.Sprintf("Erro ao enviar a requisição: %v", err), http.StatusInternalServerError)
			return
		}
		log.Printf("Response ID recebido: %s", responseID)

		var llmResponse string
		maxAttempts := 15
		for i := 0; i < maxAttempts; i++ {
			time.Sleep(2 * time.Second) // Esperar antes de tentar novamente

			llmResponse, err = getLLMResponse(responseID, token)
			if err == nil {
				// Sucesso
				break
			}

			if strings.Contains(err.Error(), "resposta ainda não está pronta") {
				log.Printf("Resposta ainda não está pronta, tentativa %d/%d", i+1, maxAttempts)
				continue
			}

			// Outro erro ocorreu
			log.Printf("Erro ao obter a resposta da LLM: %v", err)
			http.Error(w, fmt.Sprintf("Erro ao obter a resposta: %v", err), http.StatusInternalServerError)
			return
		}

		if err != nil {
			log.Printf("Timeout ao obter a resposta da LLM: %v", err)
			http.Error(w, "Timeout ao obter a resposta da LLM", http.StatusGatewayTimeout)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"response": llmResponse,
		})
	} else {
		http.Error(w, "Método não suportado", http.StatusMethodNotAllowed)
	}
}

// Função para lidar com o index
func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(filepath.Join("templates", "index.html"))
	if err != nil {
		http.Error(w, "Erro ao carregar o template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
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

// Função principal
func main() {
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		log.Fatal("CLIENT_ID e CLIENT_SECRET devem estar definidos nas variáveis de ambiente")
	}

	// Instanciando o gerenciador de tokens
	tokenManager := NewTokenManager(clientID, clientSecret)

	// Handlers
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		sendMessageHandler(w, r, tokenManager)
	})
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Servidor iniciado na porta 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
