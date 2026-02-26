// Package service contains the business logic layer for the application.
package service

import (
	"context"

	"github.com/FPT-OJT/minstant-ai.git/internal/ai/flow"
)

type ChatInput struct {
	ChatInput string   `json:"chatInput"`
	SessionID string   `json:"sessionId"`
	FullName  *string  `json:"fullName"`
	Lat       *float64 `json:"lat"`
	Long      *float64 `json:"long"`
	UserId    string   `json:"userId"`
}

type ChatService interface {
	GenerateResponse(ctx context.Context, input ChatInput) (<-chan string, <-chan error)
}

type GenkitChatService struct{}

func NewGenkitChatService() ChatService {
	return &GenkitChatService{}
}

func (s *GenkitChatService) GenerateResponse(ctx context.Context, input ChatInput) (<-chan string, <-chan error) {
	chunks := make(chan string)
	errCh := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errCh)

		input := flow.ChatFlowInput{
			SessionID: input.SessionID,
			Message:   input.ChatInput,
			FullName:  input.FullName,
			Lat:       input.Lat,
			Long:      input.Long,
			UserId:    input.UserId,
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
