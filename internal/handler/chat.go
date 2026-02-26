package handler

import (
	"fmt"
	"net/http"

	"github.com/FPT-OJT/minstant-ai.git/internal/service"
	"github.com/labstack/echo/v4"
)

// ChatRequest is the expected JSON body for the chat endpoint.
type ChatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"sessionId"`
}

// ChatHandler handles chat-related HTTP requests.
type ChatHandler struct {
	chatService service.ChatService
}

// NewChatHandler creates a new ChatHandler with the given ChatService.
func NewChatHandler(cs service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: cs}
}

// HandleChat processes POST /api/chat. It validates the request, calls the
// ChatService to generate a streaming response, and writes each chunk back
// to the client as a Server-Sent Event.
func (h *ChatHandler) HandleChat(c echo.Context) error {
	var req ChatRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	if req.Message == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "message is required",
		})
	}

	if req.SessionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "sessionId is required",
		})
	}

	// Set SSE headers.
	w := c.Response()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	chunks, errCh := h.chatService.GenerateResponse(c.Request().Context(), req.SessionID, req.Message)

	flusher, ok := w.Writer.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}

	for chunk := range chunks {
		if _, err := fmt.Fprintf(w, "data: %s\n\n", chunk); err != nil {
			return err
		}
		flusher.Flush()
	}

	// Check if the generation ended with an error.
	if err := <-errCh; err != nil {
		fmt.Fprintf(w, "data: [ERROR] %s\n\n", err.Error())
		flusher.Flush()
		return nil
	}

	// Signal completion.
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()

	return nil
}
