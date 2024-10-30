package handlers

import (
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
)

func GetResponseHandler(store *ResponseStore, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Método não suportado", http.StatusMethodNotAllowed)
			return
		}

		// Capturar message_id e session_id da URL da requisição
		messageID := r.URL.Query().Get("message_id")
		sessionID := r.URL.Query().Get("session_id")

		// Verificar se ambos foram fornecidos
		if messageID == "" || sessionID == "" {
			http.Error(w, "message_id ou session_id não fornecido", http.StatusBadRequest)
			return
		}

		// Obter a resposta da store
		data, exists := store.GetResponse(sessionID, messageID)
		if !exists {
			http.Error(w, "message_id não encontrado", http.StatusNotFound)
			return
		}

		// Enviar a resposta como JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
}
