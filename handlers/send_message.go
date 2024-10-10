// handlers/send_message.go

package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/chatcomStackspotAI/llm"
	"github.com/chatcomStackspotAI/models"
	"log"
	"net/http"
)

func SendMessageHandler(manager *llm.LLMManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var data struct {
				Prompt  string           `json:"prompt"`
				History []models.Message `json:"history"`
			}
			err := json.NewDecoder(r.Body).Decode(&data)
			if err != nil {
				log.Printf("Erro ao decodificar o JSON: %v", err)
				http.Error(w, "Dados inválidos", http.StatusBadRequest)
				return
			}

			client, _, err := manager.GetClient()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			llmResponse, err := client.SendPrompt(data.Prompt, data.History)
			if err != nil {
				log.Printf("Erro ao obter a resposta da LLM: %v", err)
				http.Error(w, fmt.Sprintf("Erro ao obter a resposta: %v", err), http.StatusInternalServerError)
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
}
