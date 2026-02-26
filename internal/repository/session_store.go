package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/firebase/genkit/go/core/x/session"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/FPT-OJT/minstant-ai.git/internal/ai/flow"
)

// PgSessionStore implements session.Store[flow.ChatState] backed by PostgreSQL.
// It persists session data in the chat_sessions table using pgx.
type PgSessionStore struct {
	pool *pgxpool.Pool
}

// NewPgSessionStore creates a new PostgreSQL-backed session store.
func NewPgSessionStore(pool *pgxpool.Pool) *PgSessionStore {
	return &PgSessionStore{pool: pool}
}

// Compile-time check that PgSessionStore implements session.Store.
var _ session.Store[flow.ChatState] = (*PgSessionStore)(nil)

// Get retrieves session data by ID. Returns nil if not found.
func (s *PgSessionStore) Get(ctx context.Context, sessionID string) (*session.Data[flow.ChatState], error) {
	var dataJSON []byte
	err := s.pool.QueryRow(ctx,
		`SELECT data FROM chat_sessions WHERE session_id = $1`, sessionID,
	).Scan(&dataJSON)
	if err != nil {
		// pgx returns no rows error; treat as not found.
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("session store get: %w", err)
	}

	var state flow.ChatState
	if err := json.Unmarshal(dataJSON, &state); err != nil {
		return nil, fmt.Errorf("session store get: failed to unmarshal state: %w", err)
	}

	return &session.Data[flow.ChatState]{
		ID:    sessionID,
		State: state,
	}, nil
}

// Save persists session data, creating or updating as needed (UPSERT).
func (s *PgSessionStore) Save(ctx context.Context, sessionID string, data *session.Data[flow.ChatState]) error {
	dataJSON, err := json.Marshal(data.State)
	if err != nil {
		return fmt.Errorf("session store save: failed to marshal state: %w", err)
	}

	_, err = s.pool.Exec(ctx,
		`INSERT INTO chat_sessions (session_id, data, user_id)
		 VALUES ($1, $2, '00000000-0000-0000-0000-000000000000')
		 ON CONFLICT (session_id) DO UPDATE
		 SET data = $2, updated_at = NOW()`,
		sessionID, dataJSON,
	)
	if err != nil {
		return fmt.Errorf("session store save: %w", err)
	}

	return nil
}
