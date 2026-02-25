-- Migration 001: tables groups (membership), messages
-- Schéma aligné : id = int (PK), user_id/sender_id = uuid, group_id = int

-- Table groups (table pivot user <-> group)
CREATE TABLE IF NOT EXISTS groups (
    id         SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id    UUID NOT NULL,
    group_id   INTEGER NOT NULL,
    role       INTEGER NOT NULL
);

-- Table messages : id = int (PK), sender_id = uuid, group_id = int
CREATE TABLE IF NOT EXISTS messages (
    id         SERIAL PRIMARY KEY,
    sender_id  UUID NOT NULL,
    content    VARCHAR(10000) NOT NULL,
    group_id   INTEGER NOT NULL,
    attachment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_messages_group_id ON messages (group_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages (created_at DESC);
