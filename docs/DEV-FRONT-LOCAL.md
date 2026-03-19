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

## k8s : Postgres message corrompu

```bash
make k8s-reset-postgres-message
# puis quand le pod est Ready :
make migrate-message && make migrate-message-legacy && make migrate-message-006
```
