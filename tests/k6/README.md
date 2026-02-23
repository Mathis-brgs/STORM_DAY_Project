# Load Tests (k6)

## Prerequis

Installer k6 : https://k6.io/docs/get-started/installation/

```bash
# macOS
brew install k6
```

## Tests disponibles

| Script | Service | Description |
|--------|---------|-------------|
| `load/media_upload_test.js` | Media Service | Upload multipart vers MinIO via HTTP |
| `load/gateway_smoke.js` | Gateway | Smoke test health + root endpoints |
| `load/auth_test.js` | User Service | Register + login flow |

## Lancer les tests

```bash
# Media upload (necessite: NATS + MinIO + Media Service)
k6 run tests/k6/load/media_upload_test.js

# Gateway smoke test
k6 run tests/k6/load/gateway_smoke.js

# Auth test (necessite: PostgreSQL + User Service)
k6 run tests/k6/load/auth_test.js
```

## Variables d'environnement

```bash
# Changer l'URL cible
k6 run -e MEDIA_URL=http://media-service:8080 tests/k6/load/media_upload_test.js
k6 run -e GATEWAY_URL=http://gateway:8080 tests/k6/load/gateway_smoke.js
k6 run -e USER_URL=http://user-service:3000 tests/k6/load/auth_test.js
```
