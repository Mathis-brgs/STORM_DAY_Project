# P4 — Media · Notification · Infra K8s · Monitoring

> Récap de tout ce qui a été fait côté P4 sur le projet STORM.

---

## Ce qui a été livré

### 1. Media Service (`services/media/`)

Service Go qui gère l'upload et la suppression de fichiers via **NATS + MinIO/S3**.

**Architecture :**
- `cmd/main.go` — démarre NATS subscribers + serveur HTTP sur `:8080`
- `internal/storage/s3.go` — client MinIO (AWS SDK v2, path-style)
- `internal/service/media_service.go` — logique métier (validation MIME, upload base64, upload multipart)
- `internal/subscribers/media_subscriber.go` — écoute NATS `media.upload.requested` / `media.delete.requested`
- `internal/handlers/media_handler.go` — endpoints HTTP `POST /media/upload` et `GET /media/{key}`

**Types de fichiers acceptés :** `image/jpeg`, `image/png`, `image/gif`, `image/webp`, `video/mp4`, `video/webm`, `video/avi`

**Variables d'environnement :**
```
NATS_URL          nats://localhost:4222
MINIO_ENDPOINT    localhost:9000
MINIO_ACCESS_KEY  admin
MINIO_SECRET_KEY  password
MINIO_BUCKET      media
PORT              8080
```

**Test rapide :**
```bash
# Upload via HTTP multipart
curl -F "file=@image.png" http://localhost:8080/media/upload

# Récupérer un fichier
curl http://localhost:8080/media/media/123456789_image.png
```

---

### 2. Notification Service (`services/notification/`)

Service Go qui stocke et distribue les notifications utilisateurs via **NATS + Redis**.

**Architecture :**
- `cmd/main.go` — connexion NATS + Redis
- `internal/service/notification_service.go` — `Send`, `GetPending`, `MarkRead` (Redis list par user, TTL 7j, max 100 notifs)
- `internal/subscribers/notification_subscriber.go` — subscribers NATS

**Subjects NATS écoutés :**

| Subject | Payload | Description |
|---|---|---|
| `notification.send` | `{userId, type, payload}` | Envoyer une notif |
| `notification.get` | `{userId}` | Récupérer notifs non lues |
| `notification.read` | `{userId}` | Marquer tout comme lu |
| `message.sent` | `{recipientId, senderUsername, conversationId}` | Auto-notif à la réception d'un message |

**Variables d'environnement :**
```
NATS_URL      nats://localhost:4222
REDIS_ADDR    localhost:6379
REDIS_PASSWORD (vide par défaut)
```

---

### 3. K8s Manifests (`infra/k8s/base/`)

Tous les services sont déclarés dans Kustomize.

**Fichiers ajoutés / modifiés :**

| Fichier | Contenu |
|---|---|
| `notification-service.yaml` | Deployment + Service K8s pour le notification service |
| `monitoring.yaml` | Prometheus (NodePort **30090**) + Grafana (NodePort **30030**) |
| `hpa.yaml` | HPA autoscaling pour gateway, media, message, notification, user |
| `kustomization.yaml` | Inclut tous les nouveaux fichiers |

**HPA résumé :**

| Service | Min | Max | CPU trigger |
|---|---|---|---|
| gateway | 2 | 10 | 60% |
| media | 1 | 5 | 70% |
| message | 1 | 5 | 65% |
| notification | 1 | 3 | 70% |
| user | 1 | 5 | 60% |

> **Prérequis HPA :** `metrics-server` installé sur le cluster.

**Déployer tout en local (k3d) :**
```bash
kubectl apply -k infra/k8s/base/
```

---

### 4. Monitoring (`infra/monitoring/`)

**Prometheus** scrape automatiquement : NATS, gateway, media, notification, message, user service, MinIO.

**Grafana** est pré-configuré avec :
- Datasource Prometheus (auto-provisioning)
- Dashboard "STORM — Vue d'ensemble" : services UP/DOWN, requêtes/s, latence p95, erreurs 5xx, mémoire pods

**Alertes configurées :**

| Alerte | Condition | Sévérité |
|---|---|---|
| `ServiceDown` | `up == 0` pendant 1 min | critical |
| `HighLatency` | p95 > 500ms pendant 2 min | warning |
| `HighErrorRate` | erreurs 5xx > 1% pendant 2 min | warning |
| `PodCrashLooping` | > 3 redémarrages/min pendant 5 min | critical |
| `PodHighMemory` | mémoire > 90% limite pendant 5 min | warning |

**Accès en local :**
```
Grafana    → http://localhost:30030   (admin / storm2024)
Prometheus → http://localhost:30090
```

---

### 5. Load Testing K6 (`tests/k6/`)

| Script | Description | Charge max |
|---|---|---|
| `smoke.js` | Vérification rapide (1 VU, 1 itération) | 1 VU |
| `auth.js` | Flow complet register → login → refresh → logout | 100 VUs |
| `gateway-ws.js` | Connexions WebSocket, join room, envoi 5 msgs | 200 VUs |
| `media-upload.js` | Upload PNG 1×1px via gateway | 30 VUs |

**Usage :**
```bash
# Smoke test (vérif que tout tourne)
k6 run tests/k6/smoke.js

# Load test auth
k6 run tests/k6/auth.js

# Load test WebSocket
k6 run tests/k6/gateway-ws.js

# Load test media upload
k6 run tests/k6/media-upload.js

# Avec une URL custom
k6 run --env BASE_URL=http://localhost:8080 tests/k6/auth.js
```

**Seuils (thresholds) :**
- Auth : `p95 < 500ms`, `erreurs < 1%`
- WebSocket : `connect p95 < 1000ms`
- Media upload : `p95 < 2000ms`, `erreurs < 5%`

---

## Ce qui reste à faire

### Faisable en local (sans Azure)

```
□ NATS cluster 3 nœuds (Helm chart)
□ Azure K8s overlay (remplacer l'overlay AWS)
□ Ajouter /metrics aux services Go (prometheus/client_golang)
□ K8s : affiner ConfigMaps et Secrets
□ Load testing 10K → 20K users (une fois le cluster up)
□ Chaos engineering scripts (kill-pods.sh, inject-latency.sh, fill-disk.sh)
□ Alerting Slack/Discord (Grafana alertmanager)
```

### Nécessite Azure (accès école)

```
□ Terraform : AKS cluster (avec P1)
□ K8s services (LoadBalancer Azure)
□ Test deploy sur AKS
□ Terraform : Azure Application Gateway / Ingress
□ Autoscaling tuning AKS (metrics, thresholds)
□ Spot/Preemptible nodes Azure (économies)
□ Multi-zone Azure
□ Backup strategy Azure
□ Disaster recovery plan
□ Load testing 30K → 50K users
```

### Dernière ligne droite (semaines finales)

```
□ Dashboards STORM DAY dédiés
□ Tests chaos tous services
□ STORM DAY : orchestration tests + monitoring live
□ Documentation K8s (deployments, scaling)
□ Documentation Monitoring (dashboards, alertes)
□ Runbooks (que faire si X tombe)
□ Cleanup ressources inutilisées (coûts Azure)
□ Slides présentation (partie infra K8s)
□ Répétition soutenance
```