package middleware

import (
	"crypto/rsa"
	"os"

	"github.com/FPT-OJT/minstant-ai.git/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-jwt/jwt/v5"
)

func SetupMiddleware(r *chi.Mux, cfg *config.Config) error {
	pubKey, err := loadPublicKey(cfg.PublicKeyPath)
	if err != nil {
		return err
	}
	authMiddleware := JWTAuth(pubKey)

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(authMiddleware)

	return nil
}

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return jwt.ParseRSAPublicKeyFromPEM(bytes)
}
