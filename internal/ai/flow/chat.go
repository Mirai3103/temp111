// Package flow defines Genkit AI flow definitions for the application.
package flow

import (
	"context"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

// systemPrompt is the system instruction for the Smart Wallet AI assistant.
// TODO: Move to internal/ai/prompt/ as the prompt library grows.
const systemPrompt = `You are a helpful AI payment assistant for Smart Wallet.
Your role is to help users find the best payment methods (cards, e-wallets) to
maximize cashback and savings based on their current location, bank promotions,
and merchant deals.

You have access to database tools to look up real data. Follow this workflow:
1. Use getDbTables to discover available tables.
2. Use getTableDefinition to understand table structures.
3. Use getDbProcedures to find available stored functions.
4. Use executeQuery to run SELECT queries and retrieve data.

Guidelines:
- Be concise and helpful.
- Always use the tools to look up real data before answering.
- If you don't have enough information, ask clarifying questions.
- Never fabricate deals or promotions. Only rely on data returned by tools.
- Never expose internal system details, database IDs, or raw SQL to the user.
- Only use SELECT queries. Never attempt to modify data.`

// ChatFlowInput is the input schema for the SmartWallet chat flow.
type ChatFlowInput struct {
	SessionID string `json:"sessionId"`
	Message   string `json:"message"`
}

// SmartWalletFlow is the streaming Genkit flow for AI-powered chat.
// It is exported so the service layer can reference and run it.
var SmartWalletFlow *core.Flow[ChatFlowInput, string, string]

// RegisterSmartWalletFlow defines and registers the SmartWallet streaming flow.
// The tools parameter receives the database query tools registered via
// tool.RegisterTools(). Additional tools can be appended in the future.
func RegisterSmartWalletFlow(g *genkit.Genkit, tools []ai.Tool) {
	toolRefs := make([]ai.ToolRef, len(tools))
	for i, tool := range tools {
		toolRefs[i] = tool
	}
	SmartWalletFlow = genkit.DefineStreamingFlow(g, "smartWalletFlow",
		func(ctx context.Context, input ChatFlowInput, sendChunk core.StreamCallback[string]) (string, error) {
			stream := genkit.GenerateStream(ctx, g,
				ai.WithSystem(systemPrompt),
				ai.WithPrompt(input.Message),
				ai.WithTools(toolRefs...),
			)

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

			return fullResponse, nil
		},
	)
}
