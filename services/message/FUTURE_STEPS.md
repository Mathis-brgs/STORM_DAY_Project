# Prochaines étapes – Message Service

Roadmap des fonctionnalités à développer après le premier commit.

## Priorité 1 – CRUD complet

| Feature | Description | Notes |
|---------|-------------|-------|
| `message.get` | Récupérer un message par ID | Sujet NATS, proto `GetMessageRequest` / `GetMessageResponse` |
| `message.list` | Lister les messages d'un groupe (pagination) | Keyset pagination sur `(created_at, id)` |
| `message.edit` | Modifier un message (soft/hard) | `updated_at`, champ `edited_at` en base |
| `message.delete` | Supprimer un message (soft delete) | `deleted_at` |

## Priorité 2 – Qualité production

| Feature | Description |
|---------|-------------|
| Idempotency | `client_msg_id` (UUID) + contrainte unique pour éviter les doublons sur retry |
| Events pub/sub | Publier `message.created`, `message.edited`, `message.deleted` pour le gateway |
| Health check | Endpoint `GET /health` pour K8s readiness/liveness |
| Observabilité | Logs structurés, métriques (Prometheus) |

## Priorité 3 – Intégration équipe

| Feature | Description |
|---------|-------------|
| Membership | Réintégrer `_local/` (MembershipRepo) quand gateway/user-service gère l'auth |
| Validation membership | Appeler `IsMember(sender_id, group_id)` avant d'accepter un message |
| Cache Redis | Cache des derniers N messages par groupe (optionnel) |

## Contrats à ajouter (proto)

- `GetMessageRequest` / `GetMessageResponse`
- `ListMessagesRequest` / `ListMessagesResponse` (items + next_cursor)
- `EditMessageRequest` / `EditMessageResponse`
- `DeleteMessageRequest` / `DeleteMessageResponse`

## Migrations à prévoir

- Index `idx_messages_group_created_id_desc` sur `(group_id, created_at DESC, id DESC)`
- Colonne `client_msg_id` (UUID) + contrainte unique pour l'idempotency
- Colonne `edited_at` si non présente
