# Vérification contrat WebSocket / API — front sans reload

Ce document recoupe le contrat attendu par le front avec l’état actuel du backend.

---

## 1. Schéma DB (message-service)

| Élément | Statut | Détail |
|--------|--------|--------|
| `messages.reply_to_id` | ✅ | Migration 006, FK `messages(id)` |
| `messages.status` | ✅ | `sent` \| `delivered` \| `seen`, défaut `sent` |
| `messages.forward_from_id` | ✅ | Migration 006, FK nullable |
| `message_seen_by` | ✅ | `message_id`, `user_id`, `display_name`, `seen_at` |

---

## 2. API REST

### 2.1 Réponse à un message (reply-to)

| Action | Statut | Détail |
|--------|--------|--------|
| POST /api/messages avec `reply_to_id` | ✅ | Gateway + message-service (proto `reply_to_id`) |
| GET /api/messages avec `reply_to` { id, sender_name, content } | ✅ | `toSendMessageData` + `enrichMessageListSenderNames` (ReplyTo.SenderName) |

### 2.2 Édition de message

| Action | Statut | Détail |
|--------|--------|--------|
| PATCH /api/messages/:id (body `content`) | ✅ | Gateway → NATS UPDATE_MESSAGE |
| Broadcast WS `message_updated` après PATCH réussi | ✅ | Gateway publie sur `message.broadcast.conversation:<id>` |

### 2.3 Transfert (forward)

| Action | Statut | Détail |
|--------|--------|--------|
| POST /api/messages avec `forward_from_id` | ✅ | Gateway + message-service (proto `forward_from_id`) |

### 2.4 Accusés de réception

| Action | Statut | Détail |
|--------|--------|--------|
| GET /api/messages : `status`, `seen_by` [{ user_id, display_name }] | ✅ | Proto + `toSendMessageData`, table `message_seen_by` |
| WS `delivered` : persistance + broadcast | ✅ | MESSAGE_SET_STATUS + broadcast |
| WS `seen` : persistance + broadcast | ✅ | MESSAGE_MARK_SEEN + `message_seen_by` + broadcast avec `seen_user_id`, `seen_display_name` |

### 2.5 Members enrichis

| Action | Statut | Détail |
|--------|--------|--------|
| GET /api/groups/:id/members avec username, display_name, avatar_url | ✅ | `fetchUserInfo` dans `ListGroupMembers` |

---

## 3. WebSocket — événements reçus par le front

| `action` | Backend | Payload broadcasté |
|----------|---------|--------------------|
| `typing` | ✅ | Réception + broadcast avec `user`, `username` (= display_name via `displayNameForUser`) |
| `delivered` | ✅ | MESSAGE_SET_STATUS + broadcast `action`, `room`, `message_id` |
| `seen` | ✅ | MESSAGE_MARK_SEEN + broadcast `action`, `room`, `message_id`, `seen_user_id`, `seen_display_name` |
| `message` | ✅ | NEW_MESSAGE → broadcast avec `user`, `username`, `content`, etc. |
| `message_updated` | ✅ | Après PATCH réussi : `action`, `room`, `message_id`, `content` (front accepte aussi message_edited, message_edit, updated) |
| `conversation_created` | ✅ | Après CreateGroup → `user:<actor_id>` ; après AddGroupMember → `user:<added_user_id>` avec `group_id`, `conversation_id`, `id`, `name` (optionnel) |

---

## 4. Room

- Backend utilise `conversation:<id>` pour les broadcasts de conversation.
- Le front accepte aussi le préfixe `group:` pour parser l’id.
- Room utilisateur (multi‑onglets / notifs) : `user:<uuid>` ; le hub écoute `message.broadcast.>` donc `message.broadcast.user:<uuid>` est bien routé.

---

## 5. Récap champs utilisés par le front

| Endpoint / Event | Champs | Statut backend |
|------------------|--------|-----------------|
| GET /api/messages | id, sender_id, sender_name, sender_username, content, created_at, status, reply_to { id, sender_name, content }, seen_by [{ user_id, display_name }] | ✅ |
| POST /api/messages | conversation_id, content, reply_to_id, forward_from_id | ✅ |
| PATCH /api/messages/:id | content (body) | ✅ + broadcast message_updated |
| GET /api/groups/:id/members | user_id, username, display_name, avatar_url, role, created_at | ✅ |
| WS message | action, room, user, username, content, **id**, **message_id**, **reply_to_id** (optionnel), **reply_to** { id, sender_id, sender_name, content } (optionnel) | ✅ |
| WS typing | action, room, user, username (= display_name) | ✅ |
| WS delivered | action, room, message_id | ✅ |
| WS seen | action, room, message_id, seen_user_id, seen_display_name | ✅ |
| WS message_updated | action, room, message_id, content | ✅ |
| WS conversation_created | action, group_id / conversation_id / id, name (optionnel) | ✅ |

---

## 6. Ordre de mise en œuvre (déjà en place)

1. DB (migration 006) ✅  
2. Reply-to (POST + GET) ✅  
3. Edit (PATCH + broadcast message_updated) ✅  
4. Forward (POST forward_from_id) ✅  
5. Members enrichis ✅  
6. Typing (username = display_name) ✅  
7. Accusés (GET status/seen_by, WS delivered/seen) ✅  
8. Nouvelle conversation (CreateGroup + AddGroupMember → conversation_created) ✅  

Tout le contrat décrit est couvert côté backend. Les broadcasts `message_updated` et `conversation_created` ont été ajoutés pour être alignés avec le contrat front.
