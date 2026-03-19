-- Migration 006: reply_to, status, forward_from, message_seen_by
-- À exécuter après 001/005. Idempotent.

ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS reply_to_id INTEGER REFERENCES messages(id) ON DELETE SET NULL;

ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'sent';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'chk_messages_status'
    ) THEN
        ALTER TABLE messages
            ADD CONSTRAINT chk_messages_status
            CHECK (status IN ('sent', 'delivered', 'seen'));
    END IF;
END $$;

ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS forward_from_id INTEGER REFERENCES messages(id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS message_seen_by (
    message_id   INTEGER NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL,
    display_name VARCHAR(255) NOT NULL DEFAULT '',
    seen_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (message_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_message_seen_by_message_id
    ON message_seen_by (message_id);

CREATE INDEX IF NOT EXISTS idx_message_seen_by_user_id
    ON message_seen_by (user_id);

CREATE INDEX IF NOT EXISTS idx_messages_reply_to_id ON messages (reply_to_id) WHERE reply_to_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_messages_forward_from_id ON messages (forward_from_id) WHERE forward_from_id IS NOT NULL;
