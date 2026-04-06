# Rapport Technique — Projet STORM

> Plateforme de messagerie temps-réel haute performance
> **Date de rendu** : 7 avril 2026

---

## 1. Contexte et objectifs

STORM est une plateforme de messagerie instantanée construite pour démontrer la capacité à concevoir, développer et opérer un système distribué à grande échelle. Les objectifs techniques cibles du Storm Day sont :

- **100 000 connexions WebSocket simultanées**
- **500 000 messages par seconde**
- Résilience aux pannes partielles (redémarrage automatique, HPA)
- Observabilité complète (Prometheus + Grafana)

---

## 2. Architecture système

### 2.1 Vue d'ensemble

L'architecture suit un pattern **microservices événementiels** où NATS JetStream sert de bus de communication central. Le Gateway est le seul point d'entrée exposé aux clients.

```
Internet
    │
    ▼
[Gateway Go]  ←──── JWT validation (local, sans round-trip)
    │
    ├─── NATS JetStream ───┬── [User Service NestJS] ── PostgreSQL (users)
    │                      ├── [Message Service Go]  ── PostgreSQL (messages)
    │                      ├── [Media Service Go]    ── MinIO / Azure Blob
    │                      └── [Notification Service Go] ── Redis
    │
    └─── WebSocket Hub (gws) ── broadcast par room
```

### 2.2 Flux d'un message

1. Client envoie `POST /api/messages` avec JWT
2. Gateway valide le JWT localement (clé publique partagée)
3. Gateway publie `message.create` sur NATS avec un timeout 10s
4. Message Service reçoit l'événement, insère via `BatchWriter` (500 msgs/50ms)
5. Message Service publie `message.broadcast.{conversation_id}` sur NATS
6. Gateway Hub reçoit le broadcast, fan-out vers tous les WebSocket de la room

### 2.3 Décisions d'architecture

| Décision | Choix | Justification |
|----------|-------|---------------|
| Broker | NATS JetStream | Latence sub-ms, persistence optionnelle, clustering natif |
| Gateway WebSocket | `lxzan/gws` | Performance supérieure à gorilla/websocket pour >10k conn |
| Batch writes | BatchWriter 500/50ms | Réduit la pression sur PostgreSQL de 500k→1k req/s |
| Auth JWT | Validation locale Gateway | Élimine le round-trip NATS pour chaque requête |
| Infra as Code | OpenTofu (Terraform fork) | Reproductibilité, diff des changements infra |
| Orchestration | Kubernetes (k3d local, AKS prod) | HPA, rolling deployments, health checks |

---

## 3. Services détaillés

### 3.1 Gateway Service (Go)

- **Port** : 8080 (NodePort 30080 en K8s)
- **WebSocket** : bibliothèque `gws` avec SO_REUSEPORT pour multi-listener
- **Optimisation** : N goroutines listener = NumCPU, évite les bottlenecks réseau
- **Auth** : Validation JWT locale en Go (pas de dépendance au User Service à chaud)
- **Métriques Prometheus** : connexions actives, messages/s, latence par endpoint

### 3.2 User Service (NestJS)

- **Auth** : bcrypt (coût 10) + JWT RS256 (access 15min + refresh 7j)
- **Refresh tokens** : stockés en base, révoqués à la déconnexion
- **Transport** : HTTP REST exposé par Gateway + NATS microservice (auth.validate)
- **Tests** : E2E avec PostgreSQL réel en CI GitHub Actions

### 3.3 Message Service (Go)

- **Stockage** : PostgreSQL avec 4 migrations versionnées
- **BatchWriter** : accumule jusqu'à 500 messages ou flush toutes les 50ms
- **Sécurité** : vérification d'appartenance à la conversation avant toute opération
- **Métriques** : `msg_batch_size`, `msg_insert_duration`, `msg_insert_errors`

### 3.4 Media Service (Go)

- **Stockage local** : MinIO (S3-compatible)
- **Stockage Azure** : Azure Blob Storage via SDK officiel
- **Validation** : MIME type, taille maximale
- **NATS** : `media.upload`, `media.get`, `media.delete`

### 3.5 Notification Service (Go)

- **Stockage** : Redis (TTL 30 jours)
- **Transport** : NATS `notification.send`, `notification.list`, `notification.mark_read`
- **Patterns** : pub/sub Redis pour les notifications en temps réel

---

## 4. Infrastructure

### 4.1 Kubernetes (K8s)

Structure Kustomize avec overlay par environnement :
- `base/` : configuration par défaut (dev)
- `overlays/azure/` : surcharges pour AKS (secrets Azure, service monitors)
- `overlays/load-test/` : ressources augmentées pour le Storm Day (10 replicas gateway, NATS 4GB mem)

**HPA configuré pour tous les services :**

| Service | Min | Max | Trigger CPU |
|---------|-----|-----|-------------|
| Gateway | 2 (10 en load-test) | 20 | 60% |
| Message | 1 | 5 | 65% |
| User | 1 | 5 | 60% |
| Notification | 1 | 3 | 70% |

### 4.2 Azure (Production)

11 modules Terraform/OpenTofu :
- **AKS** : cluster managé avec node pools séparés (system + workload)
- **PostgreSQL Flexible Server** : high availability, backup automatique 7 jours
- **Azure Cache for Redis** : Premium tier pour clustering
- **Azure Blob Storage** : médias utilisateurs
- **ACR** : registry Docker privé
- **Managed Identity** : accès aux ressources Azure sans credentials hardcodés
- **Budget** : alertes à 50%, 75%, 90% du budget mensuel

### 4.3 CI/CD (GitHub Actions)

**`ci.yml`** (sur push/PR vers main/develop) :
- Build Docker (Makefile)
- Tests unitaires NestJS + lint
- Tests E2E NestJS avec PostgreSQL
- Lint Go (golangci-lint) sur gateway + pkg
- Tests Go avec couverture
- Scan de vulnérabilités Trivy

**`deploy-azure.yml`** (manuel/workflow_dispatch) :
- Build et push des 4 images Docker vers ACR
- Authentification OIDC Azure (pas de secrets statiques)
- Deploy sur AKS avec vérification de rollout

---

## 5. Tests

### 5.1 Tests unitaires

| Service | Fichiers de test | Couverture estimée |
|---------|-----------------|-------------------|
| Gateway | 7 fichiers (`handler_test.go`, `hub_test.go`...) | ~70% |
| Message | 6 fichiers (service, repo, NATS handlers) | ~75% |
| User | Jest unit + E2E | ~80% |
| Media | — | Non couvert |
| Notification | — | Non couvert |

### 5.2 Tests de charge k6

12 scripts dans `tests/k6/` :

| Script | Scénario | VUs max |
|--------|----------|---------|
| `smoke.js` | Sanity check | 1 |
| `auth.js` | Flow complet auth | 100 |
| `ws-connections.js` | Connexions WS simultanées | 30 000 |
| `ws-messages.js` | Throughput messages | 2 000 × 5 msg/s |

**Tests distribués k6 operator (Storm Day)** :
- `testrun-100k-connections.yaml` : 50 pods × 2 000 VUs = 100k connexions
- `testrun-500k-messages.yaml` : 50 pods × 2 000 VUs × 5 msg/s = 500k msg/s

---

## 6. Résultats de performance

### 6.1 Baseline (stack locale, 1 réplique)

| Scénario | VUs | p95 | Taux erreur | Débit |
|----------|-----|-----|-------------|-------|
| Auth login | 50 | ~100ms | < 1% | ~36 req/s |
| JWT validation | 100 | ~15ms | < 0.1% | ~154 req/s |
| WS connexions | 1 000 | ~200ms | < 1% | — |

### 6.2 Objectifs Storm Day (avec overlay load-test)

| Scénario | Objectif | Configuration |
|----------|----------|---------------|
| Connexions WS | 100 000 simultanées | 10 replicas Gateway, 2Gi/pod |
| Throughput messages | 500 000 msg/s | 10 replicas Message, BatchWriter |
| Latence p95 WS | < 200ms handshake | NATS 6Gi, tuning TCP |

---

## 7. Goulots d'étranglement identifiés

### 7.1 PostgreSQL (goulot principal)

À 500k msg/s avec BatchWriter (500 msgs/50ms) → ~1 000 bulk inserts/s.
PostgreSQL avec `synchronous_commit=off` et `shared_buffers=2GB` peut théoriquement tenir 50-100k rows/s.

**Mitigation** : tuning PostgreSQL dans l'overlay load-test, connection pooling.

### 7.2 NATS JetStream

Un seul nœud NATS avec 128Mi était insuffisant. Upgrade à 6Gi + JetStream 4GB mémoire dans l'overlay.

### 7.3 File descriptors OS

100k connexions WebSocket = 100k+ file descriptors. Tuning via init container (`net.core.somaxconn=65535`) et `ulimit -n 65536` sur les nodes.

### 7.4 Gateway mémoire

~100KB/connexion WS en Go → 100k conn = 10GB. Splité sur 10 pods × 2Gi chacun.

---

## 8. Sécurité

- **Auth** : JWT RS256 + bcrypt(10), pas de credentials en dur en production
- **Transport** : TLS sur Azure (AKS Ingress), mTLS entre services optionnel
- **Secrets K8s** : Azure Key Vault via Managed Identity
- **CI** : OIDC Azure pour le déploiement (pas de secrets statiques dans GitHub)
- **Scan** : Trivy dans le pipeline CI (images Docker + IaC)

---

## 9. Limites et améliorations futures

### Limites actuelles

1. **PostgreSQL single-node** : pas de réplication ni sharding (limite à ~100k msg/s en écriture réaliste)
2. **Media service** : pas de tests unitaires
3. **NATS cluster** : déployé en single-node en base (cluster 3 nœuds uniquement en overlay load-test)
4. **Pas d'Ingress** : exposition via NodePort, pas de load balancer Azure Ingress configuré

### Améliorations prioritaires

1. Migration vers **TimescaleDB** ou **Cassandra** pour les messages (writes massifs)
2. **NATS cluster 3 nœuds** en configuration de base
3. **CDN** pour les médias (Azure CDN devant Blob Storage)
4. **Rate limiting** sur le Gateway (par user, par IP)
5. **End-to-end encryption** des messages (optionnel)

---

## 10. Équipe

| Membre | Responsabilité principale |
|--------|--------------------------|
| Mathis | User Service (NestJS) + Infrastructure Azure (Terraform) |
| — | Gateway Service (Go) + Lead Technique |
| — | Message & Conversation Services (Go) |
| — | Media + Notification Services + K8s + Monitoring |
