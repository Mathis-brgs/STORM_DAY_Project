# Message Service (Go)

## Scope

Ce micro-service gère les **messages** et la couche métier **conversations/memberships**.

- `conversation_id` : identifiant de la conversation (1-to-1 ou groupe)
  - compat temporaire: le contrat protobuf conserve encore le champ `group_id`
- Pas de gestion des users (auth/JWT), des medias (géré par media-service), ni des notifications
- Intégration via **NATS** (protobuf, request/reply)

## Structure

```
services/message/
├── api/v1/
│   ├── message.proto       # Contrat protobuf
│   └── message.pb.go       # Généré (ne pas éditer à la main)
├── cmd/
│   └── message-service/    # Point d'entrée du service
├── internal/
│   ├── models/             # ChatMessage, Event
│   ├── nats/               # Handlers NATS (messages + GROUP_*)
│   ├── repo/               # MessageRepo (memory + postgres)
│   │   ├── memory/         # Implémentation mémoire
│   │   └── postgres/       # Implémentation PostgreSQL
│   └── service/            # MessageService
├── migrations/             # SQL (001_create_tables, 002_seed_data)
├── Makefile
├── go.mod
├── Dockerfile
└── README.md
```

## Makefile (commandes principales)

| Commande | Description |
|----------|-------------|
| `make run` | Lance le service (stockage mémoire) |
| `make run-postgres` | Lance le service avec PostgreSQL |
| `make build` | Compile le binaire |
| `make proto` | Régénère `message.pb.go` (Docker requis) |
| `make migrate` | Crée le schéma cible sur PostgreSQL K8s (`conversations`, `conversations_users`, `messages`, `message_receipts`) |
| `make migrate-legacy` | Convertit/nettoie un ancien schéma legacy (optionnel) |
| `make seed` | Insère les données de test (K8s) |
| `make migrate-docker` | Fallback Docker Compose pour la migration |
| `make seed-docker` | Fallback Docker Compose pour le seed |
| `make test` | Exécute les tests |
| `make clean` | Supprime binaires et artefacts protoc |

## Local dev

### Prérequis

- [kubectl](https://kubernetes.io/docs/tasks/tools/) (namespace `storm`)
- Cluster K8s déployé (voir `infra/k8s/README.md`)
- [Docker](https://docs.docker.com/get-docker/) uniquement pour le fallback Docker Compose
- [Go 1.25+](https://go.dev/dl/)

### 1. Déployer l'infra (K8s)

```bash
# Depuis la racine du projet
kubectl apply -k infra/k8s/base/
kubectl wait --for=condition=Ready pods --all -n storm --timeout=120s
```

### 2. Migrations et seed

```bash
cd services/message
make migrate
make seed
```

### 3. Fallback Docker Compose (optionnel)

```bash
docker compose up -d postgres-chat nats
cd services/message
make migrate-docker
make seed-docker
make run-postgres
```

### 4. Tester

**Via Postman** (passer par le gateway) :

- `POST http://localhost:30080/api/messages` (Gateway en K8s)
- `POST http://localhost:8080/api/messages` (Gateway lancé localement)
- Body JSON : `{ "conversation_id": 3, "sender_id": "<uuid>", "content": "Hello" }`
  - compat legacy: `{ "group_id": 3, ... }` reste accepté côté gateway
- Accusé de réception : `POST /api/messages/:id/receipt` avec `actor_id` (query/body/header) et optionnel `received_at` (Unix timestamp)
  - Stockage par utilisateur dans `message_receipts` (plus de champ global partagé)
- Endpoints protégés (`GET /api/messages?conversation_id=...`, `PUT /api/messages/:id`, `DELETE /api/messages/:id`) :
  fournir `actor_id`/`user_id` en query ou `X-User-ID` en header.
- Endpoints groupes (Gateway) :
  - `POST /api/groups`, `GET /api/groups`, `GET /api/groups/:id`, `DELETE /api/groups/:id`, `POST /api/groups/:id/leave`
  - `POST /api/groups/:id/members`, `GET /api/groups/:id/members`, `PATCH /api/groups/:id/members/:user_id/role`, `DELETE /api/groups/:id/members/:user_id`

> En K8s, le Gateway est exposé en NodePort sur `30080`.

## Variables d'environnement

| Variable | Description | Défaut |
|----------|-------------|--------|
| `STORAGE` | `postgres` ou (vide = mémoire) | mémoire |
| `NATS_URL` | URL du broker NATS | `nats://localhost:4222` |
| `DB_HOST` | Hôte PostgreSQL | `localhost` |
| `DB_PORT` | Port PostgreSQL | `5433` |
| `DB_USER` | Utilisateur | `storm` |
| `DB_PASSWORD` | Mot de passe | `password` |
| `DB_NAME` | Base de données | `storm_message_db` (local) / `message_db` (K8s) |

## NATS

- **Messages** : `NEW_MESSAGE`, `GET_MESSAGE`, `LIST_MESSAGES`, `UPDATE_MESSAGE`, `DELETE_MESSAGE`, `ACK_MESSAGE`
- **Groupes/Conversations** : `GROUP_CREATE`, `GROUP_GET`, `GROUP_LIST_FOR_USER`, `GROUP_ADD_MEMBER`, `GROUP_REMOVE_MEMBER`, `GROUP_LIST_MEMBERS`, `GROUP_UPDATE_ROLE`, `GROUP_LEAVE`, `GROUP_DELETE`
- **Format** : protobuf (`services/message/api/v1/message.proto`)
- Le gateway convertit JSON ↔ protobuf et fait le request/reply.

## Fichiers à ne pas committer

Ignorés par `.gitignore` à la racine :

- `message-service` (binaire compilé)
- `services/message/github.com/` (sortie temporaire de protoc)
- `services/message/_local/` – code temporaire lié gateway/user (MembershipRepo, Group) – à réintégrer quand ces services seront prêts

## Status

| Composant | Status |
|-----------|--------|
| NATS messages (CRUD) | OK |
| NATS groupes/memberships (`GROUP_*`) | OK |
| Repo mémoire | OK |
| Repo PostgreSQL | OK |
| Migrations + seed | OK |
| Gateway POST /api/messages (Postman) | OK |
| API groupes via Gateway (REST) | OK |
| Pagination, idempotency, events messages | À faire |

## Prochaines étapes

Voir [FUTURE_STEPS.md](./FUTURE_STEPS.md) pour la roadmap détaillée.
