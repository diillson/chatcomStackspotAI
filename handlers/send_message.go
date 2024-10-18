// handlers/send_message.go

package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/chatcomStackspotAI/llm"
	"github.com/chatcomStackspotAI/models"
	"github.com/google/uuid"
	"log"
	"net/http"
)

func SendMessageHandler(manager *llm.LLMManager, store *ResponseStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var data struct {
				Provider string           `json:"provider"`
				Prompt   string           `json:"prompt"`
				History  []models.Message `json:"history"`
			}
			err := json.NewDecoder(r.Body).Decode(&data)
			if err != nil {
				log.Printf("Erro ao decodificar o JSON: %v", err)
				http.Error(w, "Dados inválidos", http.StatusBadRequest)
				return
			}

			client, err := manager.GetClient(data.Provider)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// Gerar um ID único para a mensagem
			messageID := uuid.New().String()

			// Armazenar o status inicial como "processing"
			store.SetResponse(messageID, &models.ResponseData{
				Status: "processing",
			})

			// Iniciar o processamento em background
			go func(messageID string, client llm.LLMClient, prompt string, history []models.Message) {
				llmResponse, err := client.SendPrompt(prompt, history)
				if err != nil {
					log.Printf("Erro ao obter a resposta da LLM: %v", err)
					store.SetResponse(messageID, &models.ResponseData{
						Status:  "error",
						Message: fmt.Sprintf("Erro ao obter a resposta: %v", err),
					})
					return
				}

				// Armazenar a resposta com status "completed"
				store.SetResponse(messageID, &models.ResponseData{
					Status:   "completed",
					Response: llmResponse,
				})
			}(messageID, client, data.Prompt, data.History)

			// Retornar o messageID para o cliente
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"message_id": messageID,
			})
		} else {
			http.Error(w, "Método não suportado", http.StatusMethodNotAllowed)
		}
	}
}
