-- Migration 001: schéma cible conversations/messages.
-- Cette migration est idempotente et crée directement le modèle final.

-- Table conversations (métadonnées conversation)
CREATE TABLE IF NOT EXISTS conversations (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(120) NOT NULL DEFAULT 'Untitled conversation',
    avatar_url TEXT,
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Table de liaison users <-> conversations
CREATE TABLE IF NOT EXISTS conversations_users (
    id              SERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    user_id         UUID NOT NULL,
    conversation_id INTEGER NOT NULL,
    role            INTEGER NOT NULL CHECK (role IN (0, 1, 2))
);

-- Table messages
CREATE TABLE IF NOT EXISTS messages (
    id              SERIAL PRIMARY KEY,
    sender_id       UUID NOT NULL,
    content         VARCHAR(10000) NOT NULL,
    conversation_id INTEGER NOT NULL,
    attachment      TEXT,
    -- Champ legacy: accusé global (déprécié, conservé pour compat).
    received_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- Backward compatibility: si la table existait déjà sans accusé global.
ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS received_at TIMESTAMPTZ;

-- Accusés de réception par utilisateur (source de vérité).
CREATE TABLE IF NOT EXISTS message_receipts (
    message_id   INTEGER NOT NULL,
    user_id      UUID NOT NULL,
    received_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (message_id, user_id)
);

-- FK (nommées) + index.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_messages_conversations'
    ) THEN
        ALTER TABLE messages
            ADD CONSTRAINT fk_messages_conversations
            FOREIGN KEY (conversation_id)
            REFERENCES conversations (id)
            ON DELETE CASCADE;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_conversations_users_conversations'
    ) THEN
        ALTER TABLE conversations_users
            ADD CONSTRAINT fk_conversations_users_conversations
            FOREIGN KEY (conversation_id)
            REFERENCES conversations (id)
            ON DELETE CASCADE;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_message_receipts_messages'
    ) THEN
        ALTER TABLE message_receipts
            ADD CONSTRAINT fk_message_receipts_messages
            FOREIGN KEY (message_id)
            REFERENCES messages (id)
            ON DELETE CASCADE;
    END IF;
END $$;

-- Pas de FK SQL vers users: architecture DB-per-service.
-- sender_id/user_id sont validés applicativement via les services (gateway + user-service).

CREATE UNIQUE INDEX IF NOT EXISTS uq_conversations_users_active
    ON conversations_users (conversation_id, user_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_conversations_users_conversation_user
    ON conversations_users (conversation_id, user_id);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_created_id_desc
    ON messages (conversation_id, created_at DESC, id DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_message_receipts_user_received
    ON message_receipts (user_id, received_at DESC);
