// Package router centralizes route registration for the application.
package router

import (
	"github.com/FPT-OJT/minstant-ai.git/internal/handler"
	"github.com/FPT-OJT/minstant-ai.git/internal/service"
	"github.com/labstack/echo/v4"
)

// Setup registers all application routes and wires up handlers with their
// dependencies. It receives a ChatService so the caller controls which
// implementation (Genkit or Mock) is used — keeping the router loosely coupled.
func Setup(e *echo.Echo, chatService service.ChatService) {
	// Handlers
	chatHandler := handler.NewChatHandler(chatService)

	// Routes
	e.GET("/", handler.HelloWorld)
	e.POST("/api/chat", chatHandler.HandleChat)
}
