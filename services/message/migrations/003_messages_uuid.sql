-- Migration 003: recréer messages avec schéma id (int), sender_id (uuid), group_id (int)
-- À exécuter si la table messages existait avec un ancien schéma.

DROP TABLE IF EXISTS messages;

CREATE TABLE messages (
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
