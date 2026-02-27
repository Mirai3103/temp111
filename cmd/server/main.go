package main

import (
	"context"
	"log"
	"net/http"

	appai "github.com/FPT-OJT/minstant-ai.git/internal/ai"
	"github.com/FPT-OJT/minstant-ai.git/internal/ai/flow"
	"github.com/FPT-OJT/minstant-ai.git/internal/ai/tool"
	"github.com/FPT-OJT/minstant-ai.git/internal/config"
	"github.com/FPT-OJT/minstant-ai.git/internal/middleware"
	"github.com/FPT-OJT/minstant-ai.git/internal/repository"
	"github.com/FPT-OJT/minstant-ai.git/internal/router"
	"github.com/FPT-OJT/minstant-ai.git/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()
	_ = godotenv.Load() // optional .env file

	cfg := config.Load()

	// ---------- Query Database ----------
	log.Println("Connecting to query database...")
	queryPool, err := repository.NewPool(ctx, cfg.QueryDatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to query database: %v", err)
	}
	defer queryPool.Close()

	// ---------- Chat Database ----------
	log.Println("Connecting to chat database...")
	chatPool, err := repository.NewPool(ctx, cfg.ChatDatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to chat database: %v", err)
	}
	defer chatPool.Close()

	log.Println("Running database migrations...")
	if err := repository.RunMigrations(ctx, chatPool); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("Database migrations applied successfully")

	sessionStore := repository.NewPgSessionStore(chatPool)

	// ---------- AI / Genkit initialization ----------
	g, err := appai.NewGenkit(ctx, cfg.AI)
	if err != nil {
		log.Fatalf("failed to initialize Genkit: %v", err)
	}

	// Register AI tools and flows.
	tools := tool.RegisterTools(g, queryPool)
	flow.RegisterSmartWalletFlow(g, tools, sessionStore)

	// Choose the ChatService implementation.
	var chatSvc service.ChatService = service.NewGenkitChatService()

	// ---------- Chi server ----------
	r := chi.NewRouter()
	if err := middleware.SetupMiddleware(r, cfg); err != nil {
		log.Fatalf("failed to setup middleware: %v", err)
	}

	// Register routes
	router.Setup(r, chatSvc)

	// Start server
	log.Printf("Starting server on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
