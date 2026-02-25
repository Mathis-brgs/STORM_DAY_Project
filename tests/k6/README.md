# Load Tests — Auth Service

Tests de charge k6 pour le service d'authentification de STORM.
Ciblent le **Gateway** (`http://localhost:8080`) qui relaie vers le **User Service** via NATS.

---

## Prérequis

### 1. Installer k6

```bash
brew install k6        # macOS
# ou
sudo snap install k6   # Ubuntu/Debian
```

### 2. Démarrer l'infrastructure locale

```bash
# PostgreSQL + NATS
docker compose up -d postgres-user nats

# User Service (NestJS)
cd services/user && npm run start:dev

# Gateway (Go)
cd services/gateway && go run cmd/main.go
```

Vérifier que tout répond :

```bash
curl http://localhost:8080/        # → OK
curl http://localhost:8222/        # → NATS dashboard
```

### 3. Créer les comptes de test k6 (une seule fois)

Chaque script utilise un compte dédié pour éviter les interférences entre tests.

```bash
# Compte pour auth-login.js
curl -s -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"k6-login","email":"k6-login@storm.dev","password":"k6password123"}' | jq

# Compte pour auth-validate.js
curl -s -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"k6-validate","email":"k6-validate@storm.dev","password":"k6password123"}' | jq

# Compte pour auth-refresh.js
curl -s -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"k6-refresh","email":"k6-refresh@storm.dev","password":"k6password123"}' | jq
```

---

## Scripts disponibles

### `auth-login.js` — Throughput JWT en rafale

Simule **50 VUs** qui font chacun `POST /auth/login` en continu pendant 1 min.
Mesure la capacité du service à signer des JWT sous charge.

| Étape | Durée | VUs |
|-------|-------|-----|
| Montée | 30s | 0 → 50 |
| Plateau | 1min | 50 |
| Descente | 10s | 50 → 0 |

**Seuils :** p(95) < 500ms — taux erreur < 1%

```bash
k6 run tests/k6/auth-login.js
```

---

### `auth-validate.js` — Pic de validation tokens

Simule **100 VUs** qui font `GET /users/:id` avec un Bearer token.
À chaque requête, le Gateway appelle `auth.validate` via NATS → mesure la latence de validation JWT bout en bout.

| Étape | Durée | VUs |
|-------|-------|-----|
| Pic | 20s | 0 → 100 |
| Plateau | 40s | 100 |
| Descente | 10s | 100 → 0 |

**Seuils :** p(95) < 300ms — taux erreur < 1%

```bash
k6 run tests/k6/auth-validate.js
```

---

### `auth-refresh.js` — Rotation de tokens en parallèle

Simule **20 VUs** qui font `POST /auth/refresh` en boucle.
Chaque VU maintient son propre `refresh_token` courant (rotation réelle — chaque token est à usage unique).
Si un refresh échoue, le VU se re-logue automatiquement.

| Étape | Durée | VUs |
|-------|-------|-----|
| Montée | 20s | 0 → 20 |
| Plateau | 1min | 20 |
| Descente | 10s | 20 → 0 |

**Seuils :** p(95) < 400ms — taux erreur < 1%

```bash
k6 run tests/k6/auth-refresh.js
```

---

## Variables d'environnement

Tous les scripts acceptent ces variables pour pointer vers un autre environnement :

| Variable | Défaut | Description |
|----------|--------|-------------|
| `BASE_URL` | `http://localhost:8080` | URL du Gateway |
| `TEST_EMAIL` | *(voir script)* | Email du compte de test |
| `TEST_PASSWORD` | `k6password123` | Mot de passe du compte de test |

Exemple — cibler un environnement de staging :

```bash
k6 run -e BASE_URL=https://staging.storm.dev \
       -e TEST_EMAIL=k6-login@storm.dev \
       -e TEST_PASSWORD=monmotdepasse \
       tests/k6/auth-login.js
```

---

## Résultats de référence (local, MacBook)

| Script | VUs | req/s | p(95) | avg | Erreurs |
|--------|-----|-------|-------|-----|---------|
| auth-login | 50 | 36.7 | 102ms | 85ms | 0% |
| auth-validate | 100 | 154.7 | 15ms | 7ms | 0% |
| auth-refresh | 20 | 8.5 | 27ms | 19ms | 0% |

> Résultats de référence — local macOS, services en docker-compose + NestJS dev.
