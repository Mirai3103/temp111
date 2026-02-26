package flow

import "github.com/firebase/genkit/go/ai"

// ChatState holds the persistent state for a chat session.
// It is serialized as JSON and stored in the chat_sessions table.
type ChatState struct {
	// History stores the full conversation history (user + model messages)
	// for multi-turn context.
	History []*ai.Message `json:"history"`
}
