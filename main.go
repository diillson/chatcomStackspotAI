package main

import (
	"fmt"
	"github.com/chatcomStackspotAI/handlers"
	"github.com/chatcomStackspotAI/llm"
	"github.com/chatcomStackspotAI/middlewares"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func indexHandler(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles(filepath.Join("templates", "index.html"))
		if err != nil {
			logger.Error("Erro ao carregar o template", zap.Error(err))
			http.Error(w, "Erro ao carregar o template", http.StatusInternalServerError)
			return
		}
		// Definir modelo com base no provedor selecionado
		provider := r.URL.Query().Get("provider")
		var modelName string

		switch provider {
		case "CLAUDEAI":
			modelName = os.Getenv("CLAUDEAI_MODEL")
			if modelName == "" {
				modelName = "claude-3-5-sonnet-20241022"
			}
		case "OPENAI":
			modelName = os.Getenv("OPENAI_MODEL")
			if modelName == "" {
				modelName = "gpt-3.5-turbo"
			}
		default:
			modelName = "stackspot-default"
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
		fmt.Println("Nenhum arquivo .env encontrado, continuando sem ele")
	}

	// Configurar o logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	manager, err := llm.NewLLMManager(logger)
	if err != nil {
		logger.Fatal("Erro ao inicializar o LLMManager", zap.Error(err))
	}

	// Inicializa o ResponseStore
	responseStore := handlers.NewResponseStore()

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler(logger))
	mux.HandleFunc("/send", handlers.SendMessageHandler(manager, responseStore, logger))
	mux.HandleFunc("/get-response", handlers.GetResponseHandler(responseStore, logger))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	finalHandler := middlewares.ForceHTTPSMiddleware(mux, logger)

	logger.Info("Servidor iniciado", zap.String("port", port))

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      finalHandler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Erro ao iniciar o servidor", zap.Error(err))
	}
}
