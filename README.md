# STORM — Système de Messagerie Temps-Réel

Plateforme de messagerie instantanée haute performance, conçue pour supporter **100 000 connexions WebSocket simultanées** et **500 000 messages/seconde**.

## Architecture

```
┌─────────────┐     WebSocket / HTTP      ┌─────────────────┐
│   Clients   │ ──────────────────────── ▶│  Gateway (Go)   │
└─────────────┘                           └────────┬────────┘
                                                   │ NATS JetStream
                        ┌──────────────────────────┼──────────────────────┐
                        ▼                          ▼                      ▼
               ┌────────────────┐      ┌───────────────────┐   ┌──────────────────┐
               │  User Service  │      │  Message Service  │   │  Notification    │
               │   (NestJS)     │      │      (Go)         │   │  Service (Go)    │
               └───────┬────────┘      └─────────┬─────────┘   └────────┬─────────┘
                       │                         │                       │
               ┌───────▼────────┐      ┌─────────▼─────────┐   ┌────────▼─────────┐
               │  PostgreSQL    │      │    PostgreSQL      │   │      Redis       │
               │  (users DB)    │      │  (messages DB)     │   │  (notifications) │
               └────────────────┘      └───────────────────┘   └──────────────────┘
```

## Stack Technique

| Service | Technologie | Rôle |
|---------|-------------|------|
| **Gateway** | Go + chi + gws | Point d'entrée HTTP/WebSocket, proxy NATS |
| **User Service** | NestJS + TypeScript | Auth (JWT), profils utilisateurs |
| **Message Service** | Go | CRUD messages, conversations, batch writer |
| **Media Service** | Go + Azure SDK | Upload fichiers vers MinIO / Azure Blob |
| **Notification Service** | Go | Notifications push via Redis |
| **NATS JetStream** | NATS 2.x | Bus événements asynchrone |
| **PostgreSQL 15** | x2 instances | Stockage users + messages |
| **Redis** | Redis 7 | Sessions, pub/sub notifications |
| **MinIO** | S3-compatible | Stockage médias local (dev) |

## Lancement local

### Prérequis

- Docker + Docker Compose
- Go 1.23+
- Node.js 20+
- k6 (pour les tests de charge)

### Démarrer l'infrastructure

```bash
# Démarrer tous les services (PostgreSQL, NATS, Redis, MinIO)
docker compose up -d

# Vérifier que tout est sain
docker compose ps
```

### Démarrer les services

```bash
# Build toutes les images
make build

# Déployer sur k3d (local)
make deploy

# Vérifier les pods
kubectl get pods -n storm
```

### Variables d'environnement

Copier `.env.example` en `.env` et renseigner :

```bash
cp .env.example .env
```

| Variable | Description | Défaut |
|----------|-------------|--------|
| `JWT_SECRET` | Clé de signature JWT | **Obligatoire en prod** |
| `DB_PASSWORD` | Mot de passe PostgreSQL | `password` (dev) |
| `NATS_URL` | URL broker NATS | `nats://nats:4222` |

## Tests

### Tests unitaires

```bash
# Go (tous les services)
cd services/message && go test ./...
cd services/gateway && go test ./...

# NestJS (user service)
cd services/user && npm test
```

### Tests d'intégration

```bash
cd services/user && npm run test:e2e
```

### Tests de charge (k6)

```bash
# Smoke test (vérification basique)
k6 run tests/k6/smoke.js

# 100k connexions WebSocket (augmenter ulimit d'abord)
ulimit -n 65536
k6 run --vus 1000 --duration 30s tests/k6/ws-connections.js

# Voir tests/k6/RESULTS.md pour les résultats de référence
```

### Tests de charge distribués (k6 operator sur K8s)

```bash
# Installer le k6 operator
kubectl apply -f https://github.com/grafana/k6-operator/releases/latest/download/bundle.yaml

# Appliquer l'overlay load-test (ressources augmentées)
kubectl apply -k infra/k8s/overlays/load-test/

# Lancer le test 100k connexions (50 pods × 2000 VUs)
kubectl apply -f infra/k8s/k6-operator/testrun-100k-connections.yaml

# Lancer le test 500k messages/s (50 pods × 2000 VUs × 5 msg/s)
kubectl apply -f infra/k8s/k6-operator/testrun-500k-messages.yaml
```

## Déploiement Azure (Production)

### Infrastructure

```bash
cd infra/terraform/environments/dev
tofu init
tofu plan
tofu apply
```

Ressources créées : AKS, Azure Database for PostgreSQL, Azure Cache for Redis, Azure Blob Storage, ACR, VNet, NSG.

### Déploiement applicatif

Le pipeline GitHub Actions `deploy-azure.yml` se charge automatiquement du build, push vers ACR et déploiement sur AKS.

```bash
# Déploiement manuel si besoin
kubectl apply -k infra/k8s/overlays/azure/
```

## Monitoring

- **Prometheus** : métriques Gateway (connexions actives, throughput, latence)
- **Grafana** : dashboards `infra/monitoring/dashboards/`
- **Alertes** : service down, latence > 500ms, taux d'erreur > 1%

```bash
# Accès Grafana local
kubectl port-forward svc/grafana -n monitoring 3000:3000
```

## Structure du projet

```
storm-project/
├── services/
│   ├── gateway/          # Go — point d'entrée HTTP/WS
│   ├── user/             # NestJS — authentification
│   ├── message/          # Go — messages et conversations
│   ├── media/            # Go — upload de fichiers
│   └── notification/     # Go — notifications push
├── pkg/                  # Packages Go partagés (models, events, id)
├── infra/
│   ├── k8s/              # Manifests Kubernetes (Kustomize)
│   │   ├── base/         # Configuration de base
│   │   ├── overlays/     # Surcharges (azure, load-test)
│   │   └── k6-operator/  # TestRun pour tests de charge distribués
│   ├── terraform/        # Infrastructure Azure (OpenTofu)
│   ├── monitoring/       # Prometheus, Grafana, AlertManager
│   └── load-tests/       # Scripts de charge supplémentaires
├── tests/k6/             # Scripts k6 de test de charge
└── docs/                 # Documentation technique
```

## Documentation

- [Architecture détaillée](docs/architecture.md)
- [Rapport technique](docs/rapport-technique.md)
- [Tests de charge k6](tests/k6/RESULTS.md)
- [Infrastructure K8s](infra/k8s/README.md)
- [Infrastructure Azure Terraform](infra/terraform/README.md)
- [Monitoring](infra/monitoring/README.md)
