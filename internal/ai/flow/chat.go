// Package flow defines Genkit AI flow definitions for the application.
package flow

import (
	"context"

	"github.com/FPT-OJT/minstant-ai.git/internal/constants"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/core/x/session"
	"github.com/firebase/genkit/go/genkit"
)

// ChatFlowInput is the input schema for the SmartWallet chat flow.
type ChatFlowInput struct {
	SessionID string   `json:"sessionId"`
	Message   string   `json:"message"`
	FullName  *string  `json:"fullName"`
	Lat       *float64 `json:"lat"`
	Long      *float64 `json:"long"`
	UserId    string   `json:"userId"`
}

// SmartWalletFlow is the streaming Genkit flow for AI-powered chat.
var SmartWalletFlow *core.Flow[ChatFlowInput, string, string]

// RegisterSmartWalletFlow defines and registers the SmartWallet streaming flow.
// It uses the session store to persist conversation history across requests.
func RegisterSmartWalletFlow(g *genkit.Genkit, tools []ai.Tool, store session.Store[ChatState]) {
	toolRefs := make([]ai.ToolRef, len(tools))
	for i, t := range tools {
		toolRefs[i] = t
	}

	SmartWalletFlow = genkit.DefineStreamingFlow(g, "smartWalletFlow",
		func(ctx context.Context, input ChatFlowInput, sendChunk core.StreamCallback[string]) (string, error) {
			// --- Session: load or create ---
			ctxWithUser := context.WithValue(ctx, constants.UserContextKey{}, &input.UserId)
			sess, err := session.Load(ctxWithUser, store, input.SessionID)
			if err != nil {
				// Session not found — create a new one.
				sess, err = session.New(ctx,
					session.WithID[ChatState](input.SessionID),
					session.WithStore(store),
					session.WithInitialState(ChatState{History: []*ai.Message{}}),
				)
				if err != nil {
					return "", err
				}
			}

			state := sess.State()

			// Build the user message.
			userMsg := ai.NewUserMessage(ai.NewTextPart(input.Message))

			// Prepare generate options.
			opts := []ai.GenerateOption{
				ai.WithSystem(GeneratePrompt(input.UserId, input.FullName, input.Lat, input.Long)),
				ai.WithMessages(append(state.History, userMsg)...),
				ai.WithTools(toolRefs...),
			}

			stream := genkit.GenerateStream(ctx, g, opts...)

			var fullResponse string
			for result, err := range stream {
				if err != nil {
					return "", err
				}
				if result.Done {
					fullResponse = result.Response.Text()
					break
				}
				chunk := result.Chunk.Text()
				sendChunk(ctx, chunk)
			}

			// --- Session: save updated history ---
			assistantMsg := ai.NewModelMessage(ai.NewTextPart(fullResponse))
			state.History = append(state.History, userMsg, assistantMsg)
			if err := sess.UpdateState(ctx, state); err != nil {
				return fullResponse, err
			}

			return fullResponse, nil
		},
	)
}
