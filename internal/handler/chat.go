package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/FPT-OJT/minstant-ai.git/internal/middleware"
	"github.com/FPT-OJT/minstant-ai.git/internal/service"
)

// ChatRequest is the expected JSON body for the chat endpoint.
type ChatRequest struct {
	ChatInput string   `json:"chatInput"`
	SessionID string   `json:"sessionId"`
	FullName  *string  `json:"fullName"`
	Lat       *float64 `json:"lat"`
	Long      *float64 `json:"long"`
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
func (h *ChatHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid request body",
		})
		return
	}

	if req.ChatInput == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "message is required",
		})
		return
	}

	if req.SessionID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "sessionId is required",
		})
		return
	}

	// Set SSE headers.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	userId := middleware.ExtractUserID(r)
	chatInput := service.ChatInput{
		ChatInput: req.ChatInput,
		SessionID: req.SessionID,
		FullName:  req.FullName,
		Lat:       req.Lat,
		Long:      req.Long,
		UserId:    userId,
	}

	chunks, errCh := h.chatService.GenerateResponse(r.Context(), chatInput)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	for chunk := range chunks {
		if _, err := fmt.Fprintf(w, "data: %s\n\n", chunk); err != nil {
			return
		}
		flusher.Flush()
	}

	// Check if the generation ended with an error.
	if err := <-errCh; err != nil {
		fmt.Fprintf(w, "data: [ERROR] %s\n\n", err.Error())
		flusher.Flush()
		return
	}

	// Signal completion.
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}
