package middleware

import (
	"net/http"

	"github.com/rs/cors"
)

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Client-Id"},
		AllowCredentials: true,
	})
	return c.Handler
}
