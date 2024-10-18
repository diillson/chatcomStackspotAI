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

func indexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles(filepath.Join("templates", "index.html"))
		if err != nil {
			http.Error(w, "Erro ao carregar o template", http.StatusInternalServerError)
			return
		}
		// Obter o modelName da variável de ambiente ou usar um valor padrão
		modelName := os.Getenv("OPENAI_MODEL")
		if modelName == "" {
			modelName = "gpt-4o-mini" // Modelo padrão
		}
		tmpl.Execute(w, map[string]string{
			"ModelName": modelName,
		})
	}
}

func main() {
	// Carrega variáveis de ambiente
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

	// Inicializa o ResponseStore
	responseStore := handlers.NewResponseStore()

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler())
	mux.HandleFunc("/send", handlers.SendMessageHandler(manager, responseStore))
	mux.HandleFunc("/get-response", handlers.GetResponseHandler(responseStore))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	finalHandler := middlewares.ForceHTTPSMiddleware(mux)

	fmt.Println("Servidor iniciado na porta " + port)
	if err := http.ListenAndServe(":"+port, finalHandler); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
