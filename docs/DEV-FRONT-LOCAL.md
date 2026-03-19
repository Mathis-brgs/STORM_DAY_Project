# Dev local pour le front

## Une commande (Docker Compose)

```bash
make dev-setup-docker
```

Démarre Postgres user + message, NATS, Redis et applique migrations **001 + 005 + 006** + seed user.

## Services sur la machine hôte

**Message-service :**

```bash
export STORAGE=postgres NATS_URL=nats://127.0.0.1:4222 DB_HOST=127.0.0.1 DB_PORT=5433 DB_USER=storm DB_PASSWORD=password DB_NAME=storm_message_db
cd services/message && go run ./cmd/main.go
```

**Gateway :**

```bash
export NATS_URL=nats://127.0.0.1:4222 USER_SERVICE_URL=http://127.0.0.1:3000
cd services/gateway && go run ./cmd/main.go
```

API / WS : port **8080** (NodePort K8s : **30080**).

## WebSocket (Vite / proxy)

- Le gateway exige un **JWT** : `ws://…/ws?token=<access_token>` (ou header `Authorization` si ton client le supporte).
- Dans `vite.config`, il faut un proxy **WebSocket** vers le gateway, pas seulement `/api` :

```ts
server: {
  proxy: {
    '/api': { target: 'http://localhost:8080', changeOrigin: true },
    '/ws': { target: 'ws://localhost:8080', ws: true, changeOrigin: true },
  },
},
```

*(Adapter le port si tu utilises NodePort **30080** : `target: 'ws://localhost:30080'`.)*

## Si « le WebSocket ne marche plus » après un pull

1. **Rebuild + import** `storm/gateway-service:latest` et `storm/message-service:latest` (même **protobuf** des deux côtés, sinon `NEW_MESSAGE` échoue au décodage → pas de broadcast).
2. **Migration 006** sur la DB message (`make migrate-message-006` en K8s ou Docker).
3. **`message-service` Running 1/1** (sinon pas de subscriber NATS → rien ne persiste).
4. Logs gateway : `Acces refuse a la room` → join refusé (`GROUP_GET` / pas membre) ; `Message non sauvegardé` → erreur côté message-service ou proto/DB.
5. Sur **join refusé**, le gateway envoie maintenant un message JSON `{"action":"error","code":"JOIN_DENIED",…}` pour le voir dans l’onglet Network → WS du navigateur.

## k8s : Postgres message corrompu

```bash
make k8s-reset-postgres-message
# puis quand le pod est Ready :
make migrate-message && make migrate-message-legacy && make migrate-message-006
```
