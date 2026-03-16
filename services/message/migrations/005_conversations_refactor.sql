-- Migration 005: refactor non-destructif vers conversations / conversations_users.
-- Objectifs:
--   - groups (membership) -> conversations_users
--   - conversation_members -> conversations_users (compat ancien renommage)
--   - messages.group_id -> messages.conversation_id
--   - ajout avatar_url sur conversations
--   - durcissement FK/contraintes/index

BEGIN;

-- 1) Table conversations (métadonnées conversation)
CREATE TABLE IF NOT EXISTS conversations (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(120) NOT NULL DEFAULT 'Untitled conversation',
    avatar_url TEXT,
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Backward compatibility: si la table existait déjà sans avatar_url.
ALTER TABLE conversations
    ADD COLUMN IF NOT EXISTS avatar_url TEXT;

-- 2) Renommages table de liaison -> conversations_users (si nécessaire)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public'
          AND table_name = 'groups'
    )
    AND NOT EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public'
          AND table_name = 'conversations_users'
    ) THEN
        ALTER TABLE groups RENAME TO conversations_users;
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public'
          AND table_name = 'conversation_members'
    )
    AND NOT EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public'
          AND table_name = 'conversations_users'
    ) THEN
        ALTER TABLE conversation_members RENAME TO conversations_users;
    END IF;
END $$;

-- Si aucune table n'existe (cas atypique), créer la table cible.
CREATE TABLE IF NOT EXISTS conversations_users (
    id              SERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    user_id         UUID NOT NULL,
    conversation_id INTEGER NOT NULL,
    role            INTEGER NOT NULL
);

-- Si groups existe encore (ex: anciens runs), migrer les données puis supprimer.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public'
          AND table_name = 'groups'
    ) THEN
        INSERT INTO conversations_users (created_at, deleted_at, user_id, conversation_id, role)
        SELECT created_at, deleted_at, user_id, group_id, role
        FROM groups
        ON CONFLICT DO NOTHING;

        DROP TABLE groups;
    END IF;
END $$;

-- Si conversation_members existe encore, migrer les données puis supprimer.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public'
          AND table_name = 'conversation_members'
    ) THEN
        IF EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_schema = 'public'
              AND table_name = 'conversation_members'
              AND column_name = 'conversation_id'
        ) THEN
            INSERT INTO conversations_users (created_at, deleted_at, user_id, conversation_id, role)
            SELECT created_at, deleted_at, user_id, conversation_id, role
            FROM conversation_members
            ON CONFLICT DO NOTHING;
        ELSIF EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_schema = 'public'
              AND table_name = 'conversation_members'
              AND column_name = 'group_id'
        ) THEN
            INSERT INTO conversations_users (created_at, deleted_at, user_id, conversation_id, role)
            SELECT created_at, deleted_at, user_id, group_id, role
            FROM conversation_members
            ON CONFLICT DO NOTHING;
        END IF;

        DROP TABLE conversation_members;
    END IF;
END $$;

-- 3) Renommer les colonnes group_id -> conversation_id (si nécessaire)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'messages'
          AND column_name = 'group_id'
    )
    AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'messages'
          AND column_name = 'conversation_id'
    ) THEN
        ALTER TABLE messages RENAME COLUMN group_id TO conversation_id;
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'conversations_users'
          AND column_name = 'group_id'
    )
    AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'conversations_users'
          AND column_name = 'conversation_id'
    ) THEN
        ALTER TABLE conversations_users RENAME COLUMN group_id TO conversation_id;
    END IF;
END $$;

-- Backward compatibility: ajouter l'accusé global legacy si absent.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public'
          AND table_name = 'messages'
    ) THEN
        ALTER TABLE messages
            ADD COLUMN IF NOT EXISTS received_at TIMESTAMPTZ;
    END IF;
END $$;

-- Accusés de réception par utilisateur (source de vérité).
CREATE TABLE IF NOT EXISTS message_receipts (
    message_id   INTEGER NOT NULL,
    user_id      UUID NOT NULL,
    received_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (message_id, user_id)
);

-- 4) Backfill conversations depuis les memberships/messages existants.
WITH source_ids AS (
    SELECT DISTINCT conversation_id
    FROM conversations_users
    UNION
    SELECT DISTINCT conversation_id
    FROM messages
),
owners AS (
    SELECT DISTINCT ON (conversation_id) conversation_id, user_id
    FROM conversations_users
    WHERE role = 2
    ORDER BY conversation_id, id
),
fallback_members AS (
    SELECT DISTINCT ON (conversation_id) conversation_id, user_id
    FROM conversations_users
    ORDER BY conversation_id, id
)
INSERT INTO conversations (id, name, created_by, created_at, updated_at)
SELECT
    s.conversation_id,
    'Conversation ' || s.conversation_id,
    COALESCE(o.user_id, f.user_id),
    NOW(),
    NOW()
FROM source_ids s
LEFT JOIN owners o ON o.conversation_id = s.conversation_id
LEFT JOIN fallback_members f ON f.conversation_id = s.conversation_id
WHERE s.conversation_id IS NOT NULL
ON CONFLICT (id) DO NOTHING;

-- Recaler la sequence SERIAL si des IDs explicites ont été insérés.
SELECT setval(
    pg_get_serial_sequence('conversations', 'id'),
    COALESCE((SELECT MAX(id) FROM conversations), 1),
    true
);

-- 5) Nettoyage des doublons actifs avant index unique partiel.
WITH ranked AS (
    SELECT
        id,
        ROW_NUMBER() OVER (
            PARTITION BY conversation_id, user_id
            ORDER BY id
        ) AS rn
    FROM conversations_users
    WHERE deleted_at IS NULL
)
UPDATE conversations_users cu
SET deleted_at = NOW()
FROM ranked r
WHERE cu.id = r.id
  AND r.rn > 1
  AND cu.deleted_at IS NULL;

-- 6) Contraintes d'intégrité
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
        WHERE conname = 'fk_message_receipts_messages'
    ) THEN
        ALTER TABLE message_receipts
            ADD CONSTRAINT fk_message_receipts_messages
            FOREIGN KEY (message_id)
            REFERENCES messages (id)
            ON DELETE CASCADE;
    END IF;
END $$;

-- Supprimer l'ancien nom de FK si présent.
ALTER TABLE conversations_users
    DROP CONSTRAINT IF EXISTS fk_conversation_members_conversations;

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

-- Harmoniser check constraint rôle.
ALTER TABLE conversations_users
    DROP CONSTRAINT IF EXISTS chk_conversation_member_role;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_conversations_users_role'
    ) THEN
        ALTER TABLE conversations_users
            ADD CONSTRAINT chk_conversations_users_role
            CHECK (role IN (0, 1, 2));
    END IF;
END $$;

-- Nettoyage contraintes FK users (non valides en DB-per-service).
ALTER TABLE messages
    DROP CONSTRAINT IF EXISTS fk_messages_sender_user;
ALTER TABLE conversations_users
    DROP CONSTRAINT IF EXISTS fk_conversations_users_user;

-- 7) Index de perfs
DROP INDEX IF EXISTS uq_conversation_members_active;
DROP INDEX IF EXISTS idx_conversation_members_conversation_user;

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

COMMIT;
