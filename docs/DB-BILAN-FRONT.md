# Bilan DB → front

- **storm_message_db** : `conversations`, `conversations_users`, `messages` (+ `reply_to_id`, `status`, `forward_from_id`), `message_receipts` (livré), `message_seen_by` (vu).
- **storm_user_db** : `users`, `jwt`.

**GET /api/messages** : `status`, `reply_to` { id, sender_name, content }, `seen_by` [{ user_id, display_name }], `sender_name`, `sender_username`.

**POST /api/messages** : `reply_to_id`, `forward_from_id` optionnels.

**PATCH /api/messages/:id** : `content`.

**GET /api/groups/:id/members** : `username`, `display_name`, `avatar_url`.

**WS** : `typing` (username = display_name), `delivered`, `seen` (+ `message_id`).

Voir migration `services/message/migrations/006_message_reply_status_forward_seen.sql`.
