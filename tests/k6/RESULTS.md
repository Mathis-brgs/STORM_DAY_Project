# K6 — Résultats Baseline STORM

> Remplir ce fichier après chaque session de tests.
> Objectif : établir les chiffres de référence **avant** le STORM DAY pour mesurer la dégradation sous charge.

---

## Prérequis avant de lancer

```bash
# 1. Démarrer l'infra locale
docker compose up -d

# 2. Démarrer les services Go
# Depuis chaque dossier service (ou via Makefile)
make up

# 3. Augmenter les file descriptors (macOS — obligatoire pour >1000 connexions)
ulimit -n 65536

# 4. Vérifier que tout répond
k6 run tests/k6/smoke.js
```

---

## Smoke Test — Vérification de base

```bash
k6 run tests/k6/smoke.js
```

| Date | Résultat | Gateway status | Auth register | WS connect |
|------|----------|---------------|---------------|------------|
| JJ/MM/AAAA | ✅ / ❌ | 200 ✅ | 201 ✅ | 101 ✅ |

---

## Auth — Baseline (register → login → refresh → logout)

```bash
k6 run tests/k6/auth.js
```

**Cible** : `p95 < 500ms`, `erreurs < 1%`, jusqu'à 100 VUs

| Date | VUs max | p50 | p95 | p99 | Taux erreur | Débit (req/s) | Notes |
|------|---------|-----|-----|-----|-------------|---------------|-------|
| _à compléter_ | 100 | - | - | - | - | - | baseline |

**Seuils de dégradation acceptable (STORM DAY) :**
- p95 < 1000ms à 500 VUs
- p95 < 2000ms à 1000 VUs

---

## WebSocket — Connexions simultanées

```bash
# Paliers recommandés (augmenter progressivement)
k6 run --vus 100  --duration 30s tests/k6/ws-connections.js
k6 run --vus 500  --duration 30s tests/k6/ws-connections.js
k6 run --vus 1000 --duration 30s tests/k6/ws-connections.js
```

**Prérequis** : un utilisateur `k6-validate@storm.dev` / `k6password123` doit exister en DB.

| Date | VUs | Handshake p95 | Erreurs connexion | Goroutines Gateway (pic) | Notes |
|------|-----|--------------|-------------------|--------------------------|-------|
| _à compléter_ | 100 | - | - | - | baseline |
| _à compléter_ | 500 | - | - | - | |
| _à compléter_ | 1000 | - | - | - | |

**Max connexions stables observées :** _à compléter_

---

## Gateway WebSocket — Load Test complet

```bash
k6 run tests/k6/gateway-ws.js
```

| Date | VUs max | Messages envoyés | p95 latence WS | Erreurs | Notes |
|------|---------|-----------------|----------------|---------|-------|
| _à compléter_ | 200 | - | - | - | baseline |

---

## Media Upload

```bash
k6 run tests/k6/media-upload.js
```

| Date | VUs | p95 upload | Erreurs | Taille fichier test | Notes |
|------|-----|-----------|---------|---------------------|-------|
| _à compléter_ | 30 | - | - | 67 bytes (PNG 1×1px) | baseline |

---

## Résumé Baseline (à remplir avant STORM DAY)

| Service | Métrique clé | Baseline | Limite acceptable |
|---------|-------------|----------|-------------------|
| Auth (register) | p95 latence | _ms | 1000ms |
| Auth (login) | p95 latence | _ms | 500ms |
| Gateway WS | handshake p95 | _ms | 200ms |
| Gateway WS | max connexions stables | _ | 10 000 |
| Media upload | p95 latence | _ms | 2000ms |

---

## Commandes de référence

```bash
# Smoke rapide (1 VU, 1 itération)
k6 run tests/k6/smoke.js

# Auth flow complet
k6 run tests/k6/auth.js

# Login uniquement
k6 run tests/k6/auth-login.js

# WebSocket connexions massives
k6 run --vus 1000 --duration 30s tests/k6/ws-connections.js

# Gateway WebSocket avec messages
k6 run tests/k6/gateway-ws.js

# Media upload
k6 run tests/k6/media-upload.js

# Avec URL custom (si services sur autre port)
k6 run --env BASE_URL=http://localhost:8080 tests/k6/auth.js
```
