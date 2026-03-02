// Package router centralizes route registration for the application.
package router

import (
	"github.com/FPT-OJT/minstant-ai.git/internal/handler"
	"github.com/FPT-OJT/minstant-ai.git/internal/middleware"
	"github.com/FPT-OJT/minstant-ai.git/internal/service"
	"github.com/go-chi/chi/v5"
)

// Setup registers all application routes and wires up handlers with their
// dependencies. It receives a ChatService so the caller controls which
// implementation (Genkit or Mock) is used — keeping the router loosely coupled.
func Setup(r *chi.Mux, chatService service.ChatService) {
	// Handlers
	chatHandler := handler.NewChatHandler(chatService)

	// Routes
	aiRoute := chi.NewRouter()
	aiRoute.Use(middleware.RequireAuth())
	aiRoute.Post("/chat", chatHandler.HandleChat)
	r.Mount("/", aiRoute)
}
