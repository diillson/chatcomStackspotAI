package main

import (
	"fmt"
	"github.com/chatcomStackspotAI/handlers"
	"github.com/chatcomStackspotAI/llm"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func indexHandler(manager *llm.LLMManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		client, llmProvider, err := manager.GetClient()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		modelName := client.GetModelName()

		tmpl, err := template.ParseFiles(filepath.Join("templates", "index.html"))
		if err != nil {
			http.Error(w, "Erro ao carregar o template", http.StatusInternalServerError)
			return
		}
		data := map[string]interface{}{
			"LLMProvider": llmProvider,
			"ModelName":   modelName,
		}
		tmpl.Execute(w, data)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	manager, err := llm.NewLLMManager()
	if err != nil {
		log.Fatalf("Erro ao inicializar o LLMManager: %v", err)
	}

	http.HandleFunc("/", indexHandler(manager))
	http.HandleFunc("/send", handlers.SendMessageHandler(manager))
	http.HandleFunc("/change-provider", handlers.ChangeProviderHandler(manager))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Servidor iniciado na porta " + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Erro ao carregar o llm: %v", err)
	}
}
