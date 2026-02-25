-- Migration 004: s'assurer que groups.user_id est en UUID
-- Recrée la table groups avec user_id UUID (les données existantes sont perdues).

DROP TABLE IF EXISTS groups;

CREATE TABLE groups (
    id         SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id    UUID NOT NULL,
    group_id   INTEGER NOT NULL,
    role       INTEGER NOT NULL
);
