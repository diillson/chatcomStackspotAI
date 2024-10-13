// middleware.go
package middlewares

import (
	"log"
	"net/http"
	"os"
)

// ForceHTTPSMiddleware redireciona todas as requisições HTTP para HTTPS somente em produção
func ForceHTTPSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env := os.Getenv("ENV")
		log.Printf("Recebendo requisição de %s %s %s", r.RemoteAddr, r.Method, r.URL.Path)

		if env != "prod" {
			// Não forçar HTTPS em ambientes não-producao
			next.ServeHTTP(w, r)
			return
		}

		// Verifica o cabeçalho X-Forwarded-Proto
		if r.Header.Get("X-Forwarded-Proto") != "https" {
			// Constrói a URL de redirecionamento para HTTPS
			target := "https://" + r.Host + r.URL.RequestURI()
			log.Printf("Redirecionando para HTTPS: %s", target)
			http.Redirect(w, r, target, http.StatusMovedPermanently)
			return
		}

		// Adiciona o cabeçalho HSTS
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Continua para o próximo handler se já for HTTPS
		next.ServeHTTP(w, r)
	})
}
