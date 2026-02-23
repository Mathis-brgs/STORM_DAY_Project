# Message Service (Go)

## Scope

Ce micro-service gère uniquement les **messages**.

- `group_id` : identifiant de la conversation (1-to-1 ou groupe)
- Pas de gestion des users (auth/JWT), des groupes/memberships, des medias (géré par media-service), ni des notifications
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
│   ├── nats/               # Handler NATS (NEW_MESSAGE)
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
| `make migrate` | Applique les migrations SQL |
| `make seed` | Insère les données de test |
| `make test` | Exécute les tests |
| `make clean` | Supprime binaires et artefacts protoc |

## Local dev

### Prérequis

- [Docker](https://docs.docker.com/get-docker/) (Postgres, NATS)
- [Go 1.25+](https://go.dev/dl/)

### 1. Démarrer l'infra

```bash
# Depuis la racine du projet
docker compose up -d postgres-chat nats
```

### 2. Migrations et seed

```bash
cd services/message
make migrate
make seed
```

### 3. Lancer le service

```bash
make run-postgres
```

### 4. Tester

**Via Postman** (passer par le gateway) :

- `POST http://localhost:8080/api/messages`
- Body JSON : `{ "group_id": 3, "sender_id": 2, "content": "Hello" }`

> Le gateway doit tourner sur le port 8080 pour les requêtes Postman.

## Variables d'environnement

| Variable | Description | Défaut |
|----------|-------------|--------|
| `STORAGE` | `postgres` ou (vide = mémoire) | mémoire |
| `NATS_URL` | URL du broker NATS | `nats://localhost:4222` |
| `DB_HOST` | Hôte PostgreSQL | `localhost` |
| `DB_PORT` | Port PostgreSQL | `5433` |
| `DB_USER` | Utilisateur | `storm` |
| `DB_PASSWORD` | Mot de passe | `password` |
| `DB_NAME` | Base de données | `storm_message_db` |

## NATS

- **Sujet** : `NEW_MESSAGE`
- **Format** : protobuf (`SendMessageRequest` / `SendMessageResponse`)
- Le gateway convertit JSON ↔ protobuf et fait le request/reply.

## Fichiers à ne pas committer

Ignorés par `.gitignore` à la racine :

- `message-service` (binaire compilé)
- `services/message/github.com/` (sortie temporaire de protoc)
- `services/message/_local/` – code temporaire lié gateway/user (MembershipRepo, Group) – à réintégrer quand ces services seront prêts

## Status

| Composant | Status |
|-----------|--------|
| NATS `NEW_MESSAGE` (protobuf) | OK |
| Repo mémoire | OK |
| Repo PostgreSQL | OK |
| Migrations + seed | OK |
| Gateway POST /api/messages (Postman) | OK |
| CRUD (get/list/edit/delete) | À faire |
| Pagination, idempotency, events | À faire |

## Prochaines étapes

Voir [FUTURE_STEPS.md](./FUTURE_STEPS.md) pour la roadmap détaillée.
