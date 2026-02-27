# Cursor Commit Order (1 seule PR)

Ce guide permet a Cursor d'executer les commits/push dans le bon ordre et de livrer **une seule PR** vers `develop`.

## 0) Toujours partir d'un `develop` a jour

```bash
git fetch origin
git checkout develop
git pull --ff-only origin develop
```

Creer ensuite une seule branche de travail:

```bash
git checkout -b feat/message-receipts-per-user
```

## 1) Commit 1 - Feature (receipt par utilisateur)

Objectif: supprimer la logique d'accuse global et passer a un stockage par `(message_id, user_id)`.

Fichiers inclus:
- `services/message/internal/models/message_receipt.go`
- `services/message/internal/models/chat_message.go`
- `services/message/internal/repo/message_repo.go`
- `services/message/internal/repo/memory/message_repo.go`
- `services/message/internal/repo/postgres/message_repo.go`
- `services/message/internal/service/message_service.go`
- `services/message/internal/nats/handler.go`
- `services/message/migrations/001_create_tables.sql`
- `services/message/migrations/005_conversations_refactor.sql`
- `services/message/api/v1/message.proto`
- `services/gateway/internal/models/message.go`

Commandes:

```bash
git add \
  services/message/internal/models/message_receipt.go \
  services/message/internal/models/chat_message.go \
  services/message/internal/repo/message_repo.go \
  services/message/internal/repo/memory/message_repo.go \
  services/message/internal/repo/postgres/message_repo.go \
  services/message/internal/service/message_service.go \
  services/message/internal/nats/handler.go \
  services/message/migrations/001_create_tables.sql \
  services/message/migrations/005_conversations_refactor.sql \
  services/message/api/v1/message.proto \
  services/gateway/internal/models/message.go
git commit -m "feat(message): store message receipts per user instead of global message received_at"
```

## 2) Commit 2 - Tests

Objectif: adapter les tests au nouveau modele (receipt par utilisateur).

Fichiers inclus:
- `services/message/internal/service/message_service_test.go`
- `services/message/internal/nats/handler_lot6_test.go`

Commandes:

```bash
git add \
  services/message/internal/service/message_service_test.go \
  services/message/internal/nats/handler_lot6_test.go
git commit -m "test(message): update ack receipt tests for per-user receipts"
```

## 3) Commit 3 - Documentation

Fichiers inclus:
- `services/message/README.md`
- `docs/cursor-commit-order.md`

Commandes:

```bash
git add services/message/README.md docs/cursor-commit-order.md
git commit -m "docs(message): document per-user receipts and commit workflow"
```

## 4) Verification locale avant push

```bash
go test ./services/message/...
go test ./services/gateway/...
```

## 5) Push de la branche

```bash
git push -u origin feat/message-receipts-per-user
```

## 6) Rester a jour avec `develop` avant d'ouvrir la PR

```bash
git fetch origin
git rebase origin/develop
git push --force-with-lease
```

Puis ouvrir **une seule PR**:
- base: `develop`
- compare: `feat/message-receipts-per-user`

