# Rapport Technique — Projet STORM

> Plateforme de messagerie temps-réel
> **Date de rendu** : 7 avril 2026

---

## 1. Contexte et objectifs

STORM est une application de messagerie instantanée pensée pour supporter une charge élevée. L'objectif du Storm Day est de tenir :

- **100 000 connexions WebSocket simultanées**
- **500 000 messages par seconde**
- Une reprise rapide en cas de panne d'un composant

---

## 2. Architecture

### 2.1 Vue d'ensemble

Le projet est découpé en **5 microservices** qui communiquent via **NATS JetStream** (bus de messages asynchrone). Le Gateway est l'unique point d'entrée exposé aux clients.

```
Internet
    │
    ▼
[Gateway Go]  ──── valide le JWT localement
    │
    ├─── NATS ───┬── [User Service NestJS] ── PostgreSQL (users)
    │            ├── [Message Service Go]  ── PostgreSQL (messages)
    │            ├── [Media Service Go]    ── MinIO / Azure Blob
    │            └── [Notification Go]    ── Redis
    │
    └─── WebSocket Hub ── diffusion en temps réel
```

### 2.2 Flux d'un message

1. Le client envoie `POST /api/messages` avec son token JWT
2. Le Gateway valide le token localement (sans appel réseau)
3. Il publie l'événement `message.create` sur NATS
4. Le Message Service reçoit l'événement et écrit en base (par lots de 500)
5. Il publie ensuite `message.broadcast` sur NATS
6. Le Gateway diffuse le message à tous les clients connectés à la conversation via WebSocket

### 2.3 Choix techniques

| Choix | Justification |
|-------|---------------|
| NATS JetStream comme broker | Faible latence, simple à déployer, supporte le clustering |
| Validation JWT locale au Gateway | Évite un aller-retour réseau à chaque requête |
| Écriture par lots (BatchWriter) | Réduit la pression sur PostgreSQL sous forte charge |
| Kubernetes + HPA | Montée en charge automatique selon l'utilisation CPU |
| OpenTofu (Terraform) | Infrastructure Azure reproductible et versionnable |

---

## 3. Services

### Gateway (Go)
Point d'entrée HTTP et WebSocket. Valide les tokens JWT, route les requêtes via NATS, et gère le hub de diffusion WebSocket.

### User Service (NestJS)
Gère l'inscription, la connexion et les profils. Authentification par JWT (access token 15 min + refresh token 7 jours), mots de passe hashés avec bcrypt.

### Message Service (Go)
CRUD des messages et conversations. Utilise un BatchWriter pour regrouper les insertions et tenir la charge en écriture.

### Media Service (Go)
Upload de fichiers (images, vidéos). Stockage sur MinIO en local, Azure Blob Storage en production. Validation du type MIME et de la taille (max 50 MB).

### Notification Service (Go)
Gère les notifications utilisateur via Redis. Conserve les 100 dernières notifications par utilisateur (TTL 7 jours).

---

## 4. Infrastructure

### Kubernetes
Manifests organisés avec Kustomize :
- `base/` — configuration commune
- `overlays/azure/` — déploiement production sur AKS
- `overlays/load-test/` — ressources augmentées pour le Storm Day

L'autoscaling (HPA) est configuré sur tous les services :

| Service | Répliques min | Répliques max |
|---------|--------------|--------------|
| Gateway | 2 (10 en load-test) | 20 |
| Message | 1 | 5 |
| User | 1 | 5 |
| Notification | 1 | 3 |

### Azure
L'infrastructure est entièrement définie en code (OpenTofu) : cluster AKS, bases PostgreSQL, Redis, stockage Blob, registry Docker (ACR) et réseau virtuel.

### CI/CD (GitHub Actions)
Deux pipelines :
- **`ci.yml`** — lance les tests, le lint et un scan de sécurité (Trivy) à chaque push
- **`deploy-azure.yml`** — build et déploie sur AKS, authentification sans secret statique (OIDC)

---

## 5. Tests

### Tests unitaires et d'intégration

| Service | Tests |
|---------|-------|
| Gateway | ~70% de couverture (handlers, hub WebSocket) |
| Message | ~75% de couverture (services, repo, NATS handlers) |
| User | ~80% (Jest unit + E2E avec PostgreSQL réel) |
| Media | Tests unitaires service + handler (mock storage) |
| Notification | Tests unitaires des validations métier |

### Tests de charge (k6)

Scripts dans `tests/k6/` pour couvrir les scénarios principaux :

| Script | Scénario |
|--------|----------|
| `smoke.js` | Vérification basique du fonctionnement |
| `auth.js` | Flux d'authentification complet |
| `ws-connections.js` | Connexions WebSocket simultanées |
| `ws-messages.js` | Débit de messages |

Pour le Storm Day, les tests sont distribués via le k6 Operator sur Kubernetes (50 pods en parallèle).

---

## 6. Résultats

### Mesures en local (avant Storm Day)

| Scénario | p95 | Taux d'erreur |
|----------|-----|---------------|
| Auth login (50 VUs) | ~100 ms | < 1% |
| Validation JWT (100 VUs) | ~15 ms | < 0.1% |
| Connexions WebSocket (1 000 VUs) | ~200 ms | < 1% |

### Objectifs Storm Day

| Objectif | Configuration prévue |
|----------|---------------------|
| 100 000 connexions WS simultanées | 10 répliques Gateway |
| 500 000 messages/s | 10 répliques Message Service + BatchWriter |
| Latence p95 < 200 ms | NATS dimensionné à 6 Gi de mémoire |

---

## 7. Points de vigilance identifiés

- **PostgreSQL** reste le composant le moins scalable horizontalement — le BatchWriter atténue la pression
- **NATS** configuré en single-node par défaut, cluster 3 nœuds activé uniquement dans l'overlay load-test
- **Limites OS** (file descriptors) à augmenter sur les nodes K8s pour supporter 100k connexions

---

## 8. Sécurité

- JWT RS256 + bcrypt pour l'authentification
- Aucun secret en dur dans le code (variables d'environnement, Azure Key Vault en prod)
- Scan Trivy automatique dans le pipeline CI
- Authentification OIDC pour les déploiements (pas de credentials statiques dans GitHub)

---

## 9. Limites et améliorations envisagées

- Le Media et Notification Service pourraient bénéficier de plus de tests de charge
- Pas d'Ingress configuré (exposition via NodePort)
- Une base de données orientée écriture (ex. Cassandra) serait plus adaptée à très grande échelle

---

## 10. Équipe

| Membre | Responsabilité |
|--------|---------------|
| Mathis | User Service (NestJS) + Infrastructure Azure (Terraform) + K8s + Monitoring |
| — | Gateway Service (Go) + Lead Technique |
| — | Message & Conversation Services (Go) + Front |
| — | Media + Notification Services + Monitoring |
