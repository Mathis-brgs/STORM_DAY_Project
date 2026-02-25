# Rapport : branche `feat/get_message` et résumé du service message

## 1. Analyse des changements sur la branche

### 1.1 Périmètre

Les changements **liés au message** sur cette branche concernent :

- **Service message** : schéma, CRUD complet, point d’entrée, migrations
- **Gateway** : module HTTP qui proxy les requêtes vers le message-service via NATS
- **Makefile** : commandes `migrate-message` et intégration des migrations 003/004

*Note : la comparaison avec `develop` inclut aussi d’autres évolutions (infra, user, media). Le présent rapport ne détaille que le périmètre message.*

### 1.2 Commits (message)

| Commit     | Objet |
|-----------|--------|
| `3e6466de` | **refactor(message)** — Schéma unifié (id int, sender_id uuid, group_id int), CRUD NATS (GET/LIST/UPDATE/DELETE), attachment, entrée `cmd/main.go` |
| `25730d3f` | **feat(gateway)** — Module `internal/modules/message`, routes REST `/api/messages`, modèles gateway |
| `94903e56` | **chore(migrations)** — 001–004 (messages + groups.user_id uuid), Makefiles |
| `7df3de9b` | **docs(message)** — README à jour, prochaines étapes (groupes, rôles) |

### 1.3 Fichiers impactés (résumé)

- **26 fichiers** modifiés/ajoutés/supprimés dans le périmètre message + gateway.
- **+1992 / -400 lignes** (hors autres changements de la branche).

| Zone | Changements principaux |
|------|------------------------|
| **message/api** | Proto étendu (sender_id string/UUID, attachment, Get/List/Update/Delete), `message.pb.go` régénéré |
| **message/internal** | `ChatMessage` (ID int, SenderID uuid, GroupID int), repo (int/uuid), service, handler NATS (5 sujets) |
| **message/cmd** | Un seul point d’entrée `main.go` (suppression `cmd/message-service/`) |
| **message/migrations** | 001 (groups + messages), 002 seed UUID, 003 messages, 004 groups.user_id uuid |
| **gateway** | Nouveau `internal/modules/message/handler.go`, `internal/models/message.go`, routes dans `cmd/main.go`, suppression ancien `api/messages.go` / `models/messages.go` |

### 1.4 Points de vigilance pour la lisibilité

- **NATS handler** (~308 lignes) : un handler par opération (Send, Get, List, Update, Delete) + helpers de réponse ; possible découpage en sous-handlers si la taille augmente.
- **Gateway handler** (~398 lignes) : même idée ; mapping proto ↔ JSON répétitif, envisageable d’extraire des helpers.
- **Proto** : bien commenté (PK, UUID) ; `cursor`/`limit` dans ListMessages prêts pour une pagination future.

---

## 2. Résumé du service message (lisible et compréhensible)

### 2.1 Rôle du service

Le **message-service** est un microservice qui :

- **Persiste les messages** de discussion (contenu, pièce jointe optionnelle).
- **Expose aucune API HTTP** : il communique uniquement via **NATS** (request/reply, protobuf).
- **Ne gère pas** : utilisateurs, auth, création de groupes, rôles, médias (dédiés à d’autres services).

Le gateway reçoit les requêtes HTTP, les transforme en requêtes NATS vers le message-service, puis renvoie les réponses en JSON.

### 2.2 Modèle de données

**Message (table `messages`)**

| Champ        | Type     | Rôle |
|-------------|----------|------|
| `id`        | int (PK) | Identifiant unique du message (généré par la base). |
| `sender_id` | UUID     | Identifiant de l’utilisateur qui envoie (référence user-service). |
| `group_id`  | int      | Identifiant de la conversation/groupe. |
| `content`   | string   | Texte du message (max 10 000 caractères). |
| `attachment`| text     | Optionnel (ex. URL ou référence vers media-service). |
| `created_at` / `updated_at` | timestamptz | Horodatage. |

**Groupe / membership (table `groups`)**

- Utilisée par les seeds ; la **création de groupes et la gestion des rôles** sont prévues dans une prochaine étape.
- Schéma actuel : `id`, `user_id` (UUID), `group_id` (int), `role`, timestamps.

### 2.3 Architecture interne (couches)

```
  NATS (sujets)          →   nats.Handler   →   service.MessageService   →   repo.MessageRepo
  NEW_MESSAGE, etc.           (désérialise         (règles métier,             (memory ou postgres)
                               proto, appelle       validation)
                               le service)
```

- **`cmd/main.go`** : connexion NATS, choix du repo (memory ou postgres), création du service et du handler, enregistrement des abonnements NATS.
- **`internal/nats/handler.go`** : reçoit les messages NATS (protobuf), appelle le service, renvoie une réponse protobuf (ou une erreur structurée).
- **`internal/service/message_service.go`** : logique métier (validation sender_id, group_id, contenu, longueur), délégation au repo.
- **`internal/repo/`** : interface `MessageRepo` + implémentations **memory** (slice + compteur) et **postgres** (requêtes SQL).

Le flux est **linéaire** : NATS → Handler → Service → Repo → DB (ou mémoire).

### 2.4 Contrat NATS (protobuf)

- **Sujets** : `NEW_MESSAGE`, `GET_MESSAGE`, `LIST_MESSAGES`, `UPDATE_MESSAGE`, `DELETE_MESSAGE`.
- **Format** : requête et réponse en protobuf (définition dans `api/v1/message.proto`).
- **Convention** : chaque requête attend une **réponse** (request/reply) ; les réponses incluent `ok`, `data` (si succès) ou `error` (code + message).

Le gateway fait un `Request()` NATS avec un timeout, reçoit la réponse, puis la convertit en JSON pour le client HTTP.

### 2.5 Structure des fichiers (service message)

```
services/message/
├── api/v1/
│   ├── message.proto      # Contrat (requêtes/réponses, ChatMessage)
│   └── message.pb.go      # Généré (ne pas modifier à la main)
├── cmd/
│   └── main.go            # Point d’entrée (NATS, repo, service, handler)
├── internal/
│   ├── models/
│   │   ├── chat_message.go  # Struct ChatMessage (id, sender_id, group_id, content, attachment, dates)
│   │   └── event.go         # Constantes d’événements (optionnel / futur)
│   ├── nats/
│   │   └── handler.go       # Écoute des 5 sujets, appel service, réponse proto
│   ├── repo/
│   │   ├── message_repo.go       # Interface MessageRepo
│   │   ├── memory/message_repo.go
│   │   └── postgres/message_repo.go
│   └── service/
│       └── message_service.go   # SendMessage, GetMessageById, GetMessagesByGroupId, Update, Delete
├── migrations/
│   ├── 001_create_tables.sql
│   ├── 002_seed_data.sql
│   ├── 003_messages_uuid.sql
│   └── 004_groups_user_id_uuid.sql
├── Makefile
├── Dockerfile
└── README.md
```

### 2.6 Résumé en une phrase

Le **message-service** est un service sans API HTTP qui écoute NATS, valide et persiste les messages (id int, sender_id UUID, group_id int) via une couche service et un repo (mémoire ou PostgreSQL), et répond en protobuf ; le **gateway** expose le CRUD en REST et fait le pont JSON ↔ NATS.

---

## 3. Prochaines étapes (hors rapport de code)

- **Création de groupes** et **gestion des rôles** (table `groups` déjà en place avec `user_id` UUID).
- Intégration avec le user-service (vérification du `sender_id`, membership).
- Pagination réelle sur `LIST_MESSAGES` (utilisation de `limit`/`cursor` côté repo).
