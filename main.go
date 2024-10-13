// main.go
package main

import (
	"fmt"
	"github.com/chatcomStackspotAI/handlers"
	"github.com/chatcomStackspotAI/llm"
	"github.com/chatcomStackspotAI/middlewares"
	"github.com/joho/godotenv"
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
	// Carrega variáveis de ambiente do arquivo .env, se existir
	err := godotenv.Load()
	if err != nil {
		log.Println("Nenhum arquivo .env encontrado")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	manager, err := llm.NewLLMManager()
	if err != nil {
		log.Fatalf("Erro ao inicializar o LLMManager: %v", err)
	}

	// Cria um novo mux para aplicar o middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler(manager))
	mux.HandleFunc("/send", handlers.SendMessageHandler(manager))
	mux.HandleFunc("/change-provider", handlers.ChangeProviderHandler(manager))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Envolve o mux com o middleware ForceHTTPSMiddleware
	finalHandler := middlewares.ForceHTTPSMiddleware(mux)

	fmt.Println("Servidor iniciado na porta " + port)
	if err := http.ListenAndServe(":"+port, finalHandler); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
