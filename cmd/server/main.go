// Package main is the entry point for the minstant-ai server.
package main

import (
	"context"
	"log"

	appai "github.com/FPT-OJT/minstant-ai.git/internal/ai"
	"github.com/FPT-OJT/minstant-ai.git/internal/ai/flow"
	"github.com/FPT-OJT/minstant-ai.git/internal/ai/tool"
	"github.com/FPT-OJT/minstant-ai.git/internal/config"
	"github.com/FPT-OJT/minstant-ai.git/internal/repository"
	"github.com/FPT-OJT/minstant-ai.git/internal/router"
	"github.com/FPT-OJT/minstant-ai.git/internal/service"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	// ---------- Echo server ----------
	e := echo.New()

	// Global middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Register routes
	router.Setup(e, chatSvc)

	// Start server
	log.Printf("Starting server on :%s", cfg.Port)
	if err := e.Start(":" + cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
