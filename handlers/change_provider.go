package handlers

import (
	"encoding/json"
	"github.com/chatcomStackspotAI/llm"
	"log"
	"net/http"
)

func ChangeProviderHandler(manager *llm.LLMManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var data struct {
				Provider string `json:"provider"`
			}
			err := json.NewDecoder(r.Body).Decode(&data)
			if err != nil {
				log.Printf("Erro ao decodificar o JSON: %v", err)
				http.Error(w, "Dados inválidos", http.StatusBadRequest)
				return
			}

			log.Printf("Tentando mudar o provedor para: %s", data.Provider)

			err = manager.SetCurrentProvider(data.Provider)
			if err != nil {
				log.Printf("Erro ao mudar o provedor: %v", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			log.Printf("Provedor alterado com sucesso para: %s", data.Provider)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Provedor atualizado com sucesso"))
		} else {
			http.Error(w, "Método não suportado", http.StatusMethodNotAllowed)
		}
	}
}
