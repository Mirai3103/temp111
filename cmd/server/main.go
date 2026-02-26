package main

import (
	"context"
	"log"
	"net/http"

	appai "github.com/FPT-OJT/minstant-ai.git/internal/ai"
	"github.com/FPT-OJT/minstant-ai.git/internal/ai/flow"
	"github.com/FPT-OJT/minstant-ai.git/internal/ai/tool"
	"github.com/FPT-OJT/minstant-ai.git/internal/config"
	"github.com/FPT-OJT/minstant-ai.git/internal/repository"
	"github.com/FPT-OJT/minstant-ai.git/internal/router"
	"github.com/FPT-OJT/minstant-ai.git/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()
	_ = godotenv.Load() // optional .env file

	cfg := config.Load()

	// ---------- Database ----------
	pool, err := repository.NewPool(ctx, cfg.QueryDatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	// ---------- AI / Genkit initialization ----------
	g, err := appai.NewGenkit(ctx, cfg.AI)
	if err != nil {
		log.Fatalf("failed to initialize Genkit: %v", err)
	}

	// Register AI tools and flows.
	tools := tool.RegisterTools(g, pool)
	flow.RegisterSmartWalletFlow(g, tools)

	// Choose the ChatService implementation.
	var chatSvc service.ChatService = service.NewGenkitChatService()

	// ---------- Chi server ----------
	r := chi.NewRouter()

	// Global middleware
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

	// Register routes
	router.Setup(r, chatSvc)

	// Start server
	log.Printf("Starting server on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
