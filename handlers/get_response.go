// handlers/get_response.go

package handlers

import (
	"encoding/json"
	"net/http"
)

func GetResponseHandler(store *ResponseStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			messageID := r.URL.Query().Get("message_id")
			if messageID == "" {
				http.Error(w, "message_id não fornecido", http.StatusBadRequest)
				return
			}

			sessionID := r.URL.Query().Get("session_id")
			if sessionID == "" {
				http.Error(w, "session_id não fornecido", http.StatusBadRequest)
				return
			}

			data, exists := store.GetResponse(sessionID, messageID)
			if !exists {
				http.Error(w, "message_id não encontrado", http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(data)
		} else {
			http.Error(w, "Método não suportado", http.StatusMethodNotAllowed)
		}
	}
}
