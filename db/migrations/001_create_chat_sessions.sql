-- Migration: Create chat_sessions table for persistent AI chat state.
-- This table is used by the Genkit session store to persist conversation history.

CREATE TABLE IF NOT EXISTS chat_sessions (
    session_id  TEXT        PRIMARY KEY,
    data        JSONB       NOT NULL DEFAULT '{}',
    user_id     UUID        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_sessions_user_id ON chat_sessions(user_id);
