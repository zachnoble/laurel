package rpc

import (
	"net/http"

	connectcors "connectrpc.com/cors"
	"github.com/rs/cors"
)

// CORS allows browser Connect clients from origins, with credentials (cookies).
func CORS(origins []string) func(http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   connectcors.AllowedMethods(),
		AllowedHeaders:   connectcors.AllowedHeaders(),
		ExposedHeaders:   connectcors.ExposedHeaders(),
		AllowCredentials: true,
		MaxAge:           7200,
	}).Handler
}
