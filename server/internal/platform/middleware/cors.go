package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
)

// CORS allows only the exact origins from CORS_ORIGINS. No cookies are used
// (tokens travel in the Authorization header), so credentials stay off.
func CORS(origins []string) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins: origins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		MaxAge:         600,
	})
}
