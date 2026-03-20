# Plan d’intégration front — Storm (Gateway)

Base URL dev : **`http://localhost:8080`** (ou NodePort K8s **`http://localhost:30080`**).  
Toutes les routes API (sauf `/` et `/metrics`) attendent en général **`Authorization: Bearer <access_token>`** pour l’identité métier (messages, groupes, média).

---

## 1. Carte des fonctionnalités

| Domaine | Fonctionnalité | Routes / canal |
|--------|----------------|----------------|
| **Auth** | Inscription, login, refresh, logout | `POST /auth/*` |
| **Profil** | Lire / mettre à jour un user, recherche | `GET/PUT /users/*` |
| **Média** | Upload fichier | `POST /media/upload` |
| **Messages** | CRUD, accusés, liste paginée | `GET/POST/PUT/PATCH/DELETE /api/messages*` |
| **Conversations** | CRUD groupe, membres, rôles | `GET/POST/PATCH/DELETE /api/groups*` |
| **Temps réel** | Messages, typing, delivered, seen, édition, nouvelle conv | `GET /ws` + événements JSON |

---

## 2. Routes HTTP (référence)

### 2.1 Auth

| Méthode | Route | Body (JSON) | Réponse (schéma) |
|--------|-------|-------------|------------------|
| POST | `/auth/register` | `{ username, password, email?, display_name? }` (selon DTO user-service) | JSON Nest (via NATS) |
| POST | `/auth/login` | idem pattern login | `access_token`, `refresh_token`, user… |
| POST | `/auth/refresh` | refresh token | nouveaux tokens |
| POST | `/auth/logout` | selon handler | OK / erreur |

### 2.2 Users

| Méthode | Route | Query / params | Réponse |
|--------|-------|----------------|---------|
| GET | `/users/search` | `q=` (requis) | liste users |
| GET | `/users/{id}` | — | profil |
| PUT | `/users/{id}` | body profil | profil mis à jour |

### 2.3 Média

| Méthode | Route | Body | Réponse |
|--------|-------|------|---------|
| POST | `/media/upload` | **multipart** champ **`file`** | JSON avec identifiant média (ex. `mediaId`) — utilisable dans `attachment` des messages |

### 2.4 Messages

| Méthode | Route | Query | Body | Notes |
|--------|-------|-------|------|--------|
| POST | `/api/messages` | — | Voir §3 | **JWT obligatoire** — `sender_id` body est **écrasé** par l’utilisateur du token |
| GET | `/api/messages` | **`conversation_id`** ou **`group_id`** (legacy), optionnel **`cursor`**, **`user_id`** si ton proxy l’ajoute | — | Liste ; **`data` est toujours un tableau** (peut être `[]`) |
| GET | `/api/messages/{id}` | — | — | Détail d’un message |
| PUT | `/api/messages/{id}` | — | `content` ou `message` | Édition |
| PATCH | `/api/messages/{id}` | — | idem | Idem PUT |
| DELETE | `/api/messages/{id}` | — | — | Suppression |
| POST | `/api/messages/{id}/receipt` | — | optionnel `actor_id`, `received_at` | Accusé de réception |

### 2.5 Groupes / conversations

| Méthode | Route | Body | Notes |
|--------|-------|------|--------|
| POST | `/api/groups` | `name`, `avatar_url?` | Création ; nom vide OK → titre résolu côté API pour l’affichage |
| GET | `/api/groups` | query `user_id` souvent utilisé par le front | Liste des conv. de l’utilisateur |
| GET | `/api/groups/{id}` | — | Détail |
| DELETE | `/api/groups/{id}` | — | Suppression |
| POST | `/api/groups/{id}/leave` | — | Quitter |
| POST | `/api/groups/{id}/members` | `user_id`, `role` | Ajouter un membre |
| GET | `/api/groups/{id}/members` | — | Membres enrichis (`username`, `display_name`, `avatar_url`) |
| PATCH | `/api/groups/{id}/members/{user_id}/role` | `role` | Changer le rôle |
| DELETE | `/api/groups/{id}/members/{user_id}` | — | Retirer un membre |

### 2.6 Santé & observabilité

| Méthode | Route |
|--------|-------|
| GET | `/` → `OK` |
| GET | `/metrics` → Prometheus |

---

## 3. Modèles de données JSON (types utiles au front)

### 3.1 Enveloppe commune (messages / groupes)

```ts
type OkError<T> =
  | { ok: true; data: T; error?: undefined }
  | { ok: false; error: { code: string; message: string } };
```

Les handlers renvoient souvent **`ok` + `data` + `error`** selon le cas.

### 3.2 Message (`SendMessageData` / élément de liste)

| Champ | Type | Description |
|-------|------|-------------|
| `id` | number | PK `messages` |
| `sender_id` | string (UUID) | Auteur |
| `sender_name` | string? | Résolu par le gateway |
| `sender_username` | string? | Souvent aligné sur le display |
| `conversation_id` | number | ID conversation |
| `group_id` | number? | Alias legacy = même id |
| `content` | string | Texte |
| `attachment` | string? | Référence média |
| `created_at` / `updated_at` | number | Unix (sec) |
| `received_at` | number? | Réception (acteur) |
| `status` | string? | `sent` \| `delivered` \| `seen` |
| `reply_to` | object? | `{ id, sender_id?, sender_name, content }` |
| `seen_by` | array? | `{ user_id, display_name }[]` |

### 3.3 POST `/api/messages` — body

```json
{
  "conversation_id": 2,
  "content": "Hello",
  "attachment": "media-ref-optionnel",
  "reply_to_id": 12,
  "forward_from_id": 5
}
```

(`group_id` peut remplacer `conversation_id` en legacy.)

### 3.4 Liste messages — GET `/api/messages`

```ts
{
  ok: boolean;
  data: SendMessageData[];  // toujours présent (tableau, éventuellement vide)
  next_cursor?: string;
  error?: { code: string; message: string };
}
```

### 3.5 Conversation / groupe (`Group`)

| Champ | Type |
|-------|------|
| `id` | number |
| `name` | string |
| `avatar_url` | string? |
| `created_by` | string (UUID)? |
| `created_at` / `updated_at` | number |

### 3.6 Membre (`GroupMember`)

| Champ | Type |
|-------|------|
| `id` | number |
| `conversation_id` / `group_id` | number |
| `user_id` | string (UUID) |
| `username` | string? |
| `display_name` | string? |
| `avatar_url` | string? |
| `role` | number |
| `created_at` | number |

*(Les rôles sont des entiers côté message-service — à mapper avec ton UI.)*

---

## 4. WebSocket `GET /ws`

### 4.1 Connexion

- URL : **`/ws?token=<access_token>`** (recommandé) ou header `Authorization: Bearer …` si le client le permet.
- À l’ouverture, le serveur place le client dans la room **`user:<userId>`** automatiquement.

### 4.2 Événements **émis par le client** (JSON texte)

| `action` | Champs typiques | Rôle |
|----------|-----------------|------|
| `join` | `room`: `conversation:<id>` ou `group:<id>` | Rejoindre une room (vérif membre côté serveur) |
| `message` | `room`, `content`, pièces jointes base64 optionnelles | Envoi chat via WS (persisté via NATS) |
| `typing` | `room` | Indicateur de frappe |
| `delivered` | `room`, `message_id` | Statut livré |
| `seen` | `room`, `message_id` | Vu |

### 4.3 Événements **reçus du serveur** (broadcast)

| `action` | Champs utiles |
|----------|----------------|
| `message` | `room`, `user`, `username`, `content`, **`id`**, **`message_id`**, **`reply_to_id`** (optionnel), **`reply_to`** `{ id, sender_id, sender_name, content }` (optionnel), `attachment?` |
| `typing` | `room`, `user`, `username` (**display name**, pas l’UUID seul) |
| `delivered` | `room`, `message_id` |
| `seen` | `room`, `message_id`, `seen_user_id`, `seen_display_name` |
| `message_updated` | `room`, `message_id`, `content` (après PATCH réussi) |
| `conversation_created` | `group_id`, `conversation_id`, `id`, `name?` — sur **`user:<uuid>`** (création ou ajout membre) |
| `error` | `code`, ex. `JOIN_DENIED`, `room`, `detail` |

**Rooms :** préférer `conversation:<numeric_id>` ; `group:<id>` reste accepté côté parsing serveur.

**Déduplication :** comparer `String(ws.user) === String(currentUserId)` pour ignorer ses propres `message` si tu as déjà fait l’optimistic UI.

---

## 5. Ordre d’implémentation front suggéré

1. **Auth** + stockage tokens + intercepteur `Authorization`.
2. **Profil** + **search users** (création de conv / ajout membres).
3. **Liste conversations** `GET /api/groups` + **détail** `GET /api/groups/{id}`.
4. **Membres** `GET /api/groups/{id}/members` (avatars / noms).
5. **Messages** `GET /api/messages?conversation_id=` puis **POST** envoi.
6. **Média** `POST /media/upload` puis référence dans `attachment`.
7. **WebSocket** : connexion, `join` sur chaque conv connue, handlers `message`, `typing`, `delivered`, `seen`, `message_updated`, `conversation_created`.
8. **Édition** `PATCH /api/messages/:id` + sync via `message_updated`.
9. **Reply / forward** : champs optionnels sur POST.
10. **Accusés** : `receipt` HTTP + WS `delivered` / `seen` selon UX.

---

## 6. Docs complémentaires dans le repo

| Fichier | Contenu |
|---------|---------|
| `docs/DEV-FRONT-LOCAL.md` | Docker, services locaux, proxy Vite, dépannage WS |
| `docs/WS-CONTRACT-VERIFICATION.md` | Alignement contrat WS / API |
| `docs/DB-BILAN-FRONT.md` | Champs messages / membres |

---

## 7. Rappels erreurs HTTP fréquentes

| Code | Cause typique |
|------|----------------|
| 401 | Token manquant / invalide (messages protégés par JWT) |
| 400 | `conversation_id` / `actor` manquant, JSON invalide |
| 403 | Pas membre de la conversation |
| 502 | Message-service injoignable (NATS / pod down) |

---

*Généré à partir du gateway (`services/gateway/cmd/main.go`) et des modèles `internal/models`.*
