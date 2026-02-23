# Roadmaps Individuelles STORM

> **Mise à jour : Jeudi Semaine 1** — Retard pris à cause des cours théoriques. L'infra K8s locale, la CI, et les bases des services sont en place. Les endpoints métier restent à implémenter.

## Stack Technique

| Service | Langage/Framework | Base de données | Broker |
|---------|-------------------|----------------|--------|
| User Service (P1) | **NestJS / TypeScript** | PostgreSQL (TypeORM) | - |
| Gateway Service (P2) | **Go** (`net/http`) | - | NATS |
| Message Service (P3) | **Go** | PostgreSQL | NATS |
| Media Service (P4) | **Go** + AWS SDK v2 | MinIO (S3) | NATS |
| Notification Service (P4) | **Go** | Redis | NATS |

**Infra locale** : k3d (K3s in Docker), Kustomize, Makefile
**CI/CD** : GitHub Actions (build Go + NestJS, lint)
**Monitoring** : Prometheus + Grafana (à venir)

---

## Roadmap cet après-midi (Jeudi S1)

### P1 (Mathis) — Auth dans User Service

```
□ Installer deps : @nestjs/jwt, @nestjs/passport, passport-jwt, bcrypt, @types/bcrypt
□ Créer AuthModule + AuthService + AuthController
□ POST /auth/register → hash password (bcrypt), créer user en DB
□ POST /auth/login → vérifier credentials, générer JWT (access + refresh)
□ Tester avec curl ou Postman
```

### P4 — Media Service endpoints

```
□ Ajouter endpoint POST /media/upload dans le media service HTTP
□ Recevoir un fichier multipart, upload via S3Client existant
□ Retourner l'URL du fichier uploadé
□ Tester avec curl : curl -F "file=@image.png" http://localhost:8080/media/upload
```

---

## P1 - User Service (NestJS) + Auth + Lead Infra AWS

### SEMAINE 1 (COURS)

**Jour 1-3 — Setup (retard cours théoriques)**

```
✅ Structure services/user/ avec NestJS + TypeORM
✅ Entities User (UUIDv7, username, email, password_hash, avatar_url)
✅ Entity JWT (token, created_at, expirated_at)
✅ Connexion PostgreSQL via TypeORM (env vars)
✅ K8s infra complète : PostgreSQL x2, Redis, NATS, MinIO
✅ CI/CD GitHub Actions (build NestJS + Go, lint)
✅ Documentation K8s (infra/k8s/README.md)
✅ .gitignore configuré
□ Endpoints Auth : register, login, JWT tokens
□ Endpoints User : GET /users/:id, PUT /users/:id
```

**Jour 4 (Jeudi) — Auth Service**

``` fait 
□ POST /auth/register (hash bcrypt, créer user)
□ POST /auth/login (vérifier credentials, générer JWT access + refresh)
□ Stockage refresh token (Redis ou entity JWT)
□ Tests unitaires Auth
```

**Jour 5 (Vendredi) — User endpoints + Intégration**

```
□ GET /users/:id
□ PUT /users/:id (update profil)
□ POST /auth/validate (pour Gateway)
□ POST /auth/refresh
□ Intégration Auth + Gateway (aider P2)
□ Demo 17h
```

---

### SEMAINES 2-3 (ENTREPRISE)

**Semaine 2 — AWS Infrastructure**

```
□ Terraform : init + structure folders
□ Terraform : VPC + Subnets
□ Terraform : Security Groups
□ Terraform : RDS PostgreSQL
□ Terraform : ElastiCache Redis
□ Terraform : S3 buckets (avatars + médias)
□ Terraform : IAM roles
□ Premier terraform apply sur AWS
```

**Semaine 3 — CI/CD + Deploy AWS**

```
□ CI/CD : Deploy sur AWS (GitHub Actions)
□ Secrets K8s (DB passwords, JWT secret)
□ Deploy Auth + User sur K8s AWS
□ Vérifier services accessibles
□ Budget AWS : Setup alertes (50%, 75%, 90%)
```

---

### SEMAINE 4 (COURS)

**Optimisations + Finitions**

```
□ Auth/User : Optimisations performance
□ Auth/User : Tests coverage >80%
□ Terraform : Outputs propres
□ Documentation Terraform (README)
□ Load testing Auth (avec P4)
□ Profiling endpoints lents
□ Fix bottlenecks identifiés
```

---

### SEMAINES 5-6 (ENTREPRISE)

**Stabilisation**

```
□ Monitoring coûts AWS
□ Optimisation ressources (downsize si possible)
□ Backup strategy RDS
□ Tests avancés Auth/User
□ Documentation API complète (OpenAPI)
□ Fix bugs mineurs
```

---

### SEMAINE 7 (COURS)

**STORM DAY Prep**

```
□ Tests de charge Auth (validation JWT haute fréquence)
□ Autoscaling test Auth pods
□ Chaos : Kill Auth pods → mesurer impact
□ Fix recovery time si trop lent
□ Documentation finale Terraform
□ Mercredi : STORM DAY test complet
```

---

### SEMAINES 8-9 (ENTREPRISE)

**Finalisation**

```
□ Post-mortem STORM DAY
□ Corrections finales
□ Architecture diagrams (Terraform)
□ README déploiement
□ Slides présentation (partie infra)
□ Répétition soutenance
```

---

## P2 (Mr Go) - Gateway Service + Lead Technique

### SEMAINE 1 (COURS)

**Jour 1-3 — Setup (retard cours théoriques)**

```
✅ Structure services/gateway/ avec Go
✅ go.mod initialisé
✅ Serveur HTTP basique (health + /)
✅ Dockerfile multi-stage
✅ Déployé sur K8s (NodePort 30080)
□ Connexion WebSocket
□ Intégration NATS
□ Validation JWT
```

**Jour 4 (Jeudi) — NATS + WebSocket**

```
□ Install gorilla/websocket ou gws
□ Connexion WebSocket basique (echo)
□ Connexion NATS client
□ Pub message sur NATS quand reçu de WebSocket
□ Tests connexion/déconnexion
```

**Jour 5 (Vendredi) — JWT + Intégration**

```
□ Valider JWT à la connexion WebSocket (appel Auth Service)
□ Rejeter connexions non authentifiées
□ Heartbeat toutes les 30s
□ Typing indicator (via NATS)
□ Intégration complète avec Auth + Message
□ Demo 17h
```

---

### SEMAINES 2-3 (ENTREPRISE)

**Optimisations Performance**

```
□ Profiling pprof (CPU, memory)
□ Optimiser allocations mémoire
□ Benchmark goroutines
□ Tests concurrency (race detector)
□ Review code tous les services (P1, P3, P4)
```

---

### SEMAINE 4 (COURS)

**Scale + Performance**

```
□ Tests de charge Gateway (10K → 20K connexions)
□ Identifier bottlenecks (profiling)
□ Optimiser broadcast NATS
□ Connection pooling NATS
□ Review architecture tous services
□ Load testing collectif vendredi
```

---

### SEMAINES 5-6 (ENTREPRISE)

**Stabilisation + Mentoring**

```
□ Tests de charge 30K → 50K
□ Optimisations avancées
□ Code review continu
□ Documentation architecture Gateway
□ Patterns Go best practices (doc partagée)
```

---

### SEMAINE 7 (COURS)

**STORM DAY Prep**

```
□ Tests 80K → 100K connexions
□ Profiling final
□ Fix memory leaks si détectés
□ Chaos : Kill Gateway pods → recovery time
□ Mercredi : STORM DAY - monitoring live Gateway
```

---

### SEMAINES 8-9 (ENTREPRISE)

**Documentation + Présentation**

```
□ Post-mortem technique
□ Documentation complète Gateway
□ Diagrammes architecture
□ Slides présentation (partie technique)
□ Répétition soutenance
```

---

## P3 - Message + Conversation Services

### SEMAINE 1 (COURS)

**Jour 1-3 — Setup (retard cours théoriques)**

```
✅ Structure services/message/ avec Go
✅ go.mod initialisé
✅ cmd/main.go créé (placeholder Hello World)
✅ Dockerfile multi-stage
✅ Déployé sur K8s (CrashLoopBackOff — normal, pas de serveur HTTP)
□ Schema PostgreSQL (conversations, messages)
□ Endpoints Message Service
```

**Jour 4 (Jeudi) — Schema DB + Conversation**

```
□ Schema PostgreSQL :
  - conversations table
  - conversation_members table
  - messages table
□ Remplacer Hello World par serveur HTTP avec /health
□ Connexion PostgreSQL
□ POST /conversations (créer 1-to-1)
□ GET /conversations (lister par user)
```

**Jour 5 (Vendredi) — Message Service**

```
□ NATS subscriber "message.send"
□ Valider user autorisé dans conversation
□ Sauvegarder message PostgreSQL
□ Publish NATS "message.broadcast.{room_id}"
□ Historique messages (GET avec pagination)
□ Demo 17h
```

---

### SEMAINES 2-3 (ENTREPRISE)

**Features Avancées**

```
□ Messages non lus (compteur par user/conversation)
□ Accusés de réception (✓✓)
□ Éditer/Supprimer message
□ Recherche messages (full-text search)
□ Cache Redis : derniers 50 messages par room
□ Tests coverage >70%
```

---

### SEMAINE 4 (COURS)

**Performance + Scale**

```
□ Tests de charge messages (1K → 10K msg/s)
□ Profiling queries PostgreSQL
□ Optimiser indexes
□ Connection pooling PostgreSQL
□ Cache strategy Redis (quoi cacher, TTL)
□ Tuning PostgreSQL config
```

---

### SEMAINES 5-6 (ENTREPRISE)

**Optimisations DB**

```
□ Tests 20K → 50K msg/s
□ Partitioning PostgreSQL si nécessaire
□ Query optimization avancée
□ Tests coverage >80%
□ Documentation API complète
```

---

### SEMAINE 7 (COURS)

**STORM DAY Prep**

```
□ Tests de charge massifs
□ Chaos : Déconnecter PostgreSQL 30s → recovery
□ Chaos : Saturer Redis → fallback DB
□ Backup/Restore test
□ Mercredi : STORM DAY monitoring DB
```

---

### SEMAINES 8-9 (ENTREPRISE)

**Finalisation**

```
□ Post-mortem
□ Documentation Message + Conversation
□ Schema DB final + migrations
□ Diagrammes flux de données
□ Slides présentation (partie data)
□ Répétition soutenance
```

---

## P4 - Media + Notification Services + Infra K8s + Monitoring

### SEMAINE 1 (COURS)

**Jour 1-3 — Setup (retard cours théoriques)**

```
✅ Structure services/media/ avec Go
✅ AWS SDK Go v2 client S3/MinIO (internal/storage/s3.go)
✅ Script test upload MinIO (cmd/media-test/main.go)
✅ go.mod avec dépendances AWS SDK
✅ cmd/main.go pour gateway, media, message (serveurs HTTP basiques)
✅ Makefile complet (up, down, clean, build, deploy, restart, status, logs)
✅ Branche git nettoyée
□ Endpoint POST /media/upload
□ Client HTML test WebSocket
□ k6 setup
```

**Jour 4 (Jeudi) — Media endpoints**

```
□ Endpoint POST /media/upload (multipart → S3)
□ Validation type fichier (image, video)
□ Retourner URL fichier uploadé
□ Endpoint GET /media/:id
□ Tests upload
```

**Jour 5 (Vendredi) — Load Testing Setup**

```
□ Client HTML test WebSocket basique
□ Setup k6 : ramp-up.js
□ Premiers tests 100 users
□ Dashboard Grafana pour tests (optionnel)
□ Demo 17h
```

---

### SEMAINES 2-3 (ENTREPRISE)

**Semaine 2 — Kubernetes sur AWS**

```
□ Terraform : EKS cluster (avec aide P1)
□ K8s deployments : tous les services
□ K8s services (ClusterIP, LoadBalancer)
□ NATS cluster (Helm install)
□ Test deploy sur EKS
```

**Semaine 3 — K8s Avancé + Notification**

```
□ Terraform : ALB Ingress
□ K8s : ConfigMaps, Secrets
□ K8s : Resource limits (CPU, RAM)
□ Autoscaling HPA (tous services)
□ Notification Service (début)
```

---

### SEMAINE 4 (COURS)

**Monitoring + Notification**

```
□ Prometheus : install sur K8s
□ Grafana : dashboards avancés
  - Gateway : connexions actives, latence
  - Messages : throughput, cache hits
  - PostgreSQL : queries, connections
  - Global : vue d'ensemble
□ Alertes : service down, latency >500ms, errors >1%
□ Notification Service : push notifications
□ Load testing 10K → 20K users
```

---

### SEMAINES 5-6 (ENTREPRISE)

**Optimisations Infra**

```
□ Autoscaling tuning (metrics, thresholds)
□ Spot instances pour économiser
□ Multi-AZ configuration
□ Backup strategy
□ Disaster recovery plan
□ Load testing 30K → 50K
```

---

### SEMAINE 7 (COURS)

**STORM DAY Prep**

```
□ Dashboards STORM DAY dédiés
□ Chaos engineering scripts :
  - kill-pods.sh
  - inject-latency.sh
  - fill-disk.sh
□ Tests chaos tous services
□ Alerting Slack/Discord
□ Mercredi : STORM DAY orchestration tests
□ Monitoring live toute la journée
```

---

### SEMAINES 8-9 (ENTREPRISE)

**Finalisation**

```
□ Post-mortem infrastructure
□ Documentation K8s (deployments, scaling)
□ Documentation Monitoring (dashboards, alertes)
□ Runbooks (que faire si X tombe)
□ Cleanup ressources inutilisées (coûts)
□ Slides présentation (partie infra K8s)
□ Répétition soutenance
```

---

## Vue d'Ensemble Chronologique

### SEMAINE 1 : MVP Local

```
P1 : User Service NestJS + Auth (register/login) + K8s infra ⏳
P2 : Gateway Go + WebSocket + NATS ⏳
P3 : Message + Conversation + PostgreSQL ⏳
P4 : Media S3/MinIO + Makefile + cmd stubs ⏳

→ Livrable : Chat basique qui marche en local (k3d)
```

### SEMAINES 2-3 : AWS Deploy

```
P1 : Terraform AWS (VPC, RDS, ElastiCache, S3)
P2 : Code review + optimisations
P3 : Features avancées messages
P4 : EKS + NATS cluster + K8s manifests

→ Livrable : Sur AWS, 5K users OK
```

### SEMAINE 4 : Scale 20K

```
Tous : Optimisations performance
Tous : Load testing vendredi

→ Livrable : 20K users stables
```

### SEMAINES 5-6 : Scale 50K

```
P1 : Infra stable
P2 : Gateway optimisé
P3 : DB tuning
P4 : Monitoring complet

→ Livrable : 30K-50K users
```

### SEMAINE 7 : STORM DAY

```
Tous : Chaos engineering
Tous : Tests 80K-100K
Mercredi : STORM DAY test
Jeudi-Vendredi : Fix + doc

→ Livrable : Projet technique complet
```

### SEMAINES 8-9 : Présentation

```
Tous : Documentation
Tous : Slides
STORM DAY final
Finitions

→ Livrable : Rendu 7 avril
```
