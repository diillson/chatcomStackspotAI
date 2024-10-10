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

func indexHandler(llmProvider, modelName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	llmProvider := os.Getenv("LLM_PROVIDER")
	if llmProvider == "" {
		log.Fatal("LLM_PROVIDER deve estar definido nas variáveis de ambiente")
	}

	var client llm.LLMClient
	var modelName string

	switch llmProvider {
	case "STACKSPOT":
		clientID := os.Getenv("CLIENT_ID")
		clientSecret := os.Getenv("CLIENT_SECRET")
		slug := os.Getenv("SLUG_NAME")
		if clientID == "" || clientSecret == "" || slug == "" {
			log.Fatal("CLIENT_ID, CLIENT_SECRET, e SLUG_NAME devem estar definidos nas variáveis de ambiente")
		}
		tokenManager := llm.NewTokenManager(clientID, clientSecret)
		client = llm.NewStackSpotClient(tokenManager, slug)
		modelName = "StackSpotAI"

	case "OPENAI":
		apiKey := os.Getenv("OPENAI_API_KEY")
		model := os.Getenv("MODEL_NAME")
		if apiKey == "" {
			log.Fatal("OPENAI_API_KEY deve estar definido nas variáveis de ambiente")
		}
		if model == "" {
			model = "gpt-4o"
		}
		client = llm.NewOpenAIClient(apiKey, model)
		modelName = model

	default:
		log.Fatalf("LLM_PROVIDER '%s' não suportado", llmProvider)
	}

	http.HandleFunc("/", indexHandler(llmProvider, modelName))
	http.HandleFunc("/send", handlers.SendMessageHandler(client))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Servidor iniciado na porta 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
