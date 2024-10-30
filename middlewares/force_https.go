package middlewares

import (
	"go.uber.org/zap"
	"net/http"
	"os"
)

// ForceHTTPSMiddleware redireciona todas as requisições HTTP para HTTPS somente em produção
func ForceHTTPSMiddleware(next http.Handler, logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env := os.Getenv("ENV")
		logger.Info("Recebendo requisição",
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("method", r.Method),
			zap.String("url", r.URL.Path),
		)

		if env != "prod" {
			// Não forçar HTTPS em ambientes não-produção
			next.ServeHTTP(w, r)
			return
		}

		// Verifica o cabeçalho X-Forwarded-Proto e X-Forwarded-Ssl
		if r.Header.Get("X-Forwarded-Proto") != "https" && r.Header.Get("X-Forwarded-Ssl") != "on" {
			// Constrói a URL de redirecionamento para HTTPS
			target := "https://" + r.Host + r.URL.RequestURI()
			logger.Info("Redirecionando para HTTPS", zap.String("target", target))
			http.Redirect(w, r, target, http.StatusMovedPermanently)
			return
		}

		// Adiciona o cabeçalho HSTS
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Continua para o próximo handler se já for HTTPS
		next.ServeHTTP(w, r)
	})
}
