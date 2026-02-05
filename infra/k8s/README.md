# Manifestes Kubernetes (K8s)

Ce dossier contient l'ensemble des manifestes Kubernetes pour le deploiement du projet 

## Structure

```
infra/k8s/
├── base/                        # Manifestes de base
│   ├── kustomization.yaml       # Configuration Kustomize
│   ├── namespace.yaml           # Namespace "storm"
│   ├── secrets.yaml             # Secrets (PostgreSQL, MinIO)
│   ├── postgres-user.yaml       # BDD User Service
│   ├── postgres-message.yaml    # BDD Message Service
│   ├── nats.yaml                # Broker NATS (JetStream)
│   ├── redis.yaml               # Cache Redis
│   ├── minio.yaml               # Stockage objet MinIO
│   ├── gateway-service.yaml     # API Gateway
│   ├── user-service.yaml        # Service utilisateurs
│   ├── message-service.yaml     # Service messages
│   └── media-service.yaml       # Service medias
└── overlays/
    └── dev/                     # Overrides specifiques au dev
```

## Prerequis

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) installe
- [k3d](https://k3d.io/) installe (`brew install k3d` ou `curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash`)
- **kubectl** (`brew install kubectl` ou installe avec Docker Desktop)

## Installation et mise en place du cluster

### 1. Creer le cluster k3d

```bash
# Creer le cluster "storm" avec un seul node server
# Le port 30080 est mappe pour acceder au Gateway via NodePort
k3d cluster create storm -p "30080:30080@server:0"
```

> Le cluster utilise **k3s** (distribution Kubernetes legere) dans un container Docker via k3d.

### 2. Verifier que le cluster est pret

```bash
# Verifier le contexte actif
kubectl config current-context
# Doit afficher : k3d-storm

# Verifier que le node est Ready
kubectl get nodes
# NAME                 STATUS   ROLES                  AGE   VERSION
# k3d-storm-server-0   Ready    control-plane,master   ..    v1.33.6+k3s1
```

### 3. Builder et importer les images Docker

Avec k3d, les images locales ne sont pas directement accessibles par le cluster. Il faut les **importer** apres le build :

```bash
# Depuis la racine du projet (storm-project/)

# Builder les images
docker build -t storm/user-service:latest -f services/user/Dockerfile services/user/
docker build -t storm/gateway-service:latest -f services/gateway/Dockerfile .
docker build -t storm/message-service:latest -f services/message/Dockerfile .
docker build -t storm/media-service:latest -f services/media/Dockerfile .

# Importer les images dans le cluster k3d
k3d image import storm/user-service:latest -c storm
k3d image import storm/gateway-service:latest -c storm
k3d image import storm/message-service:latest -c storm
k3d image import storm/media-service:latest -c storm
```

> **Important** : contrairement a Docker Desktop Kubernetes, k3d utilise son propre registry interne. Il faut reimporter les images a chaque rebuild.

### 4. Deployer sur le cluster

```bash
# Appliquer tous les manifestes
kubectl apply -k infra/k8s/base/

# Verifier le deploiement
kubectl get all -n storm

# Attendre que tous les pods soient Ready
kubectl wait --for=condition=Ready pods --all -n storm --timeout=120s
```

### 5. Verifier que tout fonctionne

```bash
# Tester le health check du Gateway
curl http://localhost:30080/health
```

### Commandes k3d utiles

```bash
# Lister les clusters
k3d cluster list

# Arreter le cluster (sans le supprimer)
k3d cluster stop storm

# Redemarrer le cluster
k3d cluster start storm

# Supprimer le cluster
k3d cluster delete storm
```

## Namespace

Toutes les ressources sont deployees dans le namespace **`storm`**.

```bash
# Basculer sur le namespace storm
kubectl config set-context --current --namespace=storm
```

## Architecture deployee

### Services applicatifs

| Service | Langage | Image | Port | Type de Service |
|---------|---------|-------|------|-----------------|
| **Gateway** | Go | `storm/gateway-service:latest` | 8080 | NodePort (30080) |
| **User** | Node.js/NestJS | `storm/user-service:latest` | 3000 | ClusterIP |
| **Message** | Go | `storm/message-service:latest` | 8080 | ClusterIP |
| **Media** | Go | `storm/media-service:latest` | 8080 | ClusterIP |

### Infrastructure

| Composant | Image | Port(s) | Stockage |
|-----------|-------|---------|----------|
| **PostgreSQL User** | `postgres:15-alpine` | 5432 | PVC 1Gi |
| **PostgreSQL Message** | `postgres:15-alpine` | 5432 | PVC 1Gi |
| **Redis** | `redis:alpine` | 6379 | - |
| **NATS** (JetStream) | `nats:latest` | 4222 / 8222 | - |
| **MinIO** | `minio/minio` | 9000 / 9001 | PVC 2Gi |

## Communication inter-services

```
Client
  │
  ▼
Gateway (NodePort :30080)
  ├──► User Service ──► PostgreSQL User (user_db)
  ├──► NATS (JetStream)
  │      ├──► Message Service ──► PostgreSQL Message (message_db)
  │      ├──► Media Service ──► MinIO
  │      └──► Notification Service (a venir)
  └──► Redis (cache)
```

- Le **Gateway** est le seul point d'entree expose (NodePort `30080`)
- Les services communiquent entre eux via le DNS interne Kubernetes
- **NATS JetStream** sert de bus d'evenements entre les services
- **Redis** est utilise comme couche de cache par le Gateway

## Variables d'environnement

### Gateway Service
| Variable | Valeur |
|----------|--------|
| `NATS_URL` | `nats://nats:4222` |
| `REDIS_URL` | `redis://redis:6379` |
| `USER_SERVICE_URL` | `http://user-service:3000` |

### User Service
| Variable | Valeur |
|----------|--------|
| `DB_HOST` | `postgres-user` |
| `DB_PORT` | `5432` |
| `DB_USER` | via Secret `postgres-credentials` |
| `DB_PASSWORD` | via Secret `postgres-credentials` |
| `DB_NAME` | `user_db` |

### Message Service
| Variable | Valeur |
|----------|--------|
| `NATS_URL` | `nats://nats:4222` |
| `DB_HOST` | `postgres-message` |
| `DB_PORT` | `5432` |
| `DB_USER` | via Secret `postgres-credentials` |
| `DB_PASSWORD` | via Secret `postgres-credentials` |
| `DB_NAME` | `message_db` |

### Media Service
| Variable | Valeur |
|----------|--------|
| `NATS_URL` | `nats://nats:4222` |
| `MINIO_ENDPOINT` | `minio:9000` |
| `MINIO_ACCESS_KEY` | via Secret `minio-credentials` |
| `MINIO_SECRET_KEY` | via Secret `minio-credentials` |

## Secrets

Deux secrets Kubernetes sont definis dans `secrets.yaml` :

- **`postgres-credentials`** : identifiants PostgreSQL (`POSTGRES_USER`, `POSTGRES_PASSWORD`)
- **`minio-credentials`** : identifiants MinIO (`MINIO_ROOT_USER`, `MINIO_ROOT_PASSWORD`)

> **Attention** : les secrets actuels contiennent des valeurs en clair destinees au developpement. En production, utilisez des outils comme Sealed Secrets, Vault ou un Secret Manager cloud.

## Ressources allouees

| Service | CPU request | CPU limit | Memoire request | Memoire limit |
|---------|-------------|-----------|-----------------|---------------|
| Gateway | 50m | 200m | 64Mi | 128Mi |
| User | 100m | 250m | 128Mi | 256Mi |
| Message | 50m | 200m | 64Mi | 128Mi |
| Media | 50m | 200m | 64Mi | 128Mi |

## Health Checks

Tous les services disposent d'une **readinessProbe** :

| Service | Endpoint | Delai initial | Periode |
|---------|----------|---------------|---------|
| Gateway | `GET /health` | 5s | 5s |
| User | `GET /` | 10s | 5s |
| Message | `GET /health` | 5s | 5s |
| Media | `GET /health` | 5s | 5s |

## Volumes persistants

| PVC | Taille | Utilise par |
|-----|--------|-------------|
| `postgres-user-pvc` | 1Gi | PostgreSQL User |
| `postgres-message-pvc` | 1Gi | PostgreSQL Message |
| `minio-pvc` | 2Gi | MinIO |

## Commandes utiles

```bash
# Voir tous les pods
kubectl get pods -n storm

# Logs d'un service
kubectl logs -f deployment/gateway-service -n storm

# Acceder au Gateway depuis la machine locale
curl http://localhost:30080/health

# Verifier l'etat des PVC
kubectl get pvc -n storm

# Supprimer tout le deploiement
kubectl delete -k infra/k8s/base/
```

## Kustomize Overlays

Le dossier `overlays/dev/` permet de surcharger les manifestes de base pour un environnement specifique (dev, staging, prod). Pour creer un overlay :

1. Ajouter un `kustomization.yaml` dans `overlays/dev/`
2. Referencer la base : `resources: [../../base]`
3. Appliquer les patches ou modifications souhaitees
4. Deployer avec `kubectl apply -k infra/k8s/overlays/dev/`
