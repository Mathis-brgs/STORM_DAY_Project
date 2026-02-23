-- Migration 001: tables groups (membership), messages
-- Base: storm_message_db
-- Schema conforme au diagramme

-- Table groups (table pivot user <-> group)
CREATE TABLE IF NOT EXISTS groups (
    id         SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id    INTEGER NOT NULL,
    group_id   INTEGER NOT NULL,
    role       INTEGER NOT NULL
);

-- Table messages (attachments gérés par media-service)
CREATE TABLE IF NOT EXISTS messages (
    id         SERIAL PRIMARY KEY,
    sender_id  INTEGER NOT NULL,
    content    VARCHAR(255) NOT NULL,
    group_id   INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
