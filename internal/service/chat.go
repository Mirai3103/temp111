// Package service contains the business logic layer for the application.
package service

import (
	"context"

	"github.com/FPT-OJT/minstant-ai.git/internal/ai/flow"
)

type ChatService interface {
	GenerateResponse(ctx context.Context, sessionID, message string) (<-chan string, <-chan error)
}

type GenkitChatService struct{}

func NewGenkitChatService() *GenkitChatService {
	return &GenkitChatService{}
}

func (s *GenkitChatService) GenerateResponse(ctx context.Context, sessionID, message string) (<-chan string, <-chan error) {
	chunks := make(chan string)
	errCh := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errCh)

		input := flow.ChatFlowInput{
			SessionID: sessionID,
			Message:   message,
		}

		for val, err := range flow.SmartWalletFlow.Stream(ctx, input) {
			if err != nil {
				errCh <- err
				return
			}
			if val.Done {
				break
			}
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			case chunks <- val.Stream:
			}
		}
	}()

	return chunks, errCh
}
