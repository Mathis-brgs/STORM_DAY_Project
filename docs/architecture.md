# Architecture STORM

## Vue d'ensemble

```mermaid
graph TB
    subgraph Clients
        C1[Browser / App]
        C2[k6 Load Test]
    end

    subgraph Gateway["Gateway Service (Go :8080)"]
        GW_HTTP[HTTP Router chi]
        GW_WS[WebSocket Hub gws]
        GW_JWT[JWT Validator local]
    end

    subgraph NATS["NATS JetStream"]
        N1[auth.* subjects]
        N2[message.* subjects]
        N3[media.* subjects]
        N4[notification.* subjects]
    end

    subgraph Services
        US["User Service (NestJS :3000)"]
        MS["Message Service (Go :8080)"]
        MDS["Media Service (Go :8080)"]
        NS["Notification Service (Go :8080)"]
    end

    subgraph Storage
        PG_U[(PostgreSQL\nusers DB)]
        PG_M[(PostgreSQL\nmessages DB)]
        RD[(Redis)]
        MINIO[(MinIO /\nAzure Blob)]
    end

    C1 -->|HTTP REST| GW_HTTP
    C1 -->|WebSocket ws://| GW_WS
    C2 -->|Load Test| GW_HTTP
    C2 -->|Load Test| GW_WS

    GW_HTTP --> GW_JWT
    GW_WS --> GW_JWT

    GW_HTTP -->|NATS request| N1
    GW_HTTP -->|NATS request| N2
    GW_HTTP -->|NATS request| N3
    GW_WS -->|subscribe broadcast| N2

    N1 --> US
    N2 --> MS
    N3 --> MDS
    N4 --> NS

    MS -->|broadcast message| N2

    US --> PG_U
    MS --> PG_M
    NS --> RD
    MDS --> MINIO
```

## Flux d'un message WebSocket

```mermaid
sequenceDiagram
    participant C as Client
    participant GW as Gateway
    participant NATS as NATS JetStream
    participant MS as Message Service
    participant DB as PostgreSQL

    C->>GW: POST /api/messages (JWT)
    GW->>GW: Valide JWT localement
    GW->>NATS: Publie message.create
    NATS->>MS: Délivre message.create
    MS->>MS: BatchWriter accumule
    Note over MS: flush à 500 msgs ou 50ms
    MS->>DB: BulkInsert(batch)
    DB-->>MS: OK + IDs
    MS->>NATS: Publie message.broadcast.{conv_id}
    NATS->>GW: Délivre broadcast
    GW->>C: WebSocket push (new_message)
    GW-->>C: HTTP 201 (réponse initiale)
```

## Flux d'authentification

```mermaid
sequenceDiagram
    participant C as Client
    participant GW as Gateway
    participant NATS as NATS JetStream
    participant US as User Service
    participant DB as PostgreSQL

    C->>GW: POST /auth/login {email, password}
    GW->>NATS: Publie auth.login (timeout 10s)
    NATS->>US: Délivre auth.login
    US->>DB: SELECT user WHERE email=...
    DB-->>US: User row
    US->>US: bcrypt.Compare(password, hash)
    US->>DB: INSERT jwt_tokens (refresh)
    US->>NATS: Répond {access_token, refresh_token}
    NATS->>GW: Livre réponse
    GW-->>C: 200 {access_token, refresh_token}

    Note over C,GW: Connexion WebSocket
    C->>GW: WS /ws?token=ACCESS_TOKEN
    GW->>GW: Valide JWT localement (SANS NATS)
    GW-->>C: 101 Switching Protocols
```

## Infrastructure K8s

```mermaid
graph LR
    subgraph k8s["Kubernetes Cluster (AKS / k3d)"]
        subgraph storm["namespace: storm"]
            GW_POD["Gateway\n(2-20 pods HPA)"]
            MS_POD["Message\n(1-5 pods HPA)"]
            US_POD["User\n(1-5 pods HPA)"]
            NS_POD["Notification\n(1-3 pods HPA)"]
            MDS_POD["Media\n(1-5 pods HPA)"]
            NATS_POD["NATS\n(1-3 pods)"]
            REDIS_POD["Redis\n(1 pod)"]
            PGM_POD["PostgreSQL\nmessages"]
            PGU_POD["PostgreSQL\nusers"]
        end

        subgraph monitoring["namespace: monitoring"]
            PROM["Prometheus"]
            GRAF["Grafana"]
            ALERT["AlertManager"]
        end

        subgraph k6ns["namespace: k6"]
            K6_OP["k6 Operator"]
            K6_PODS["50 k6 Runner\npods (load test)"]
        end
    end

    NodePort30080["NodePort :30080"] --> GW_POD
    GW_POD --> NATS_POD
    NATS_POD --> MS_POD
    NATS_POD --> US_POD
    NATS_POD --> NS_POD
    NATS_POD --> MDS_POD
    MS_POD --> PGM_POD
    US_POD --> PGU_POD
    NS_POD --> REDIS_POD
    GW_POD --> PROM
    MS_POD --> PROM
    K6_OP --> K6_PODS
    K6_PODS --> GW_POD
```

## Dimensionnement pour le Storm Day

| Composant | Répliques | CPU | RAM | Justification |
|-----------|-----------|-----|-----|---------------|
| Gateway | 10 (HPA max 20) | 2 | 2Gi | 100k conn × ~100KB = 10GB total |
| Message | 10 | 1 | 512Mi | BatchWriter absorbe 50k msg/s par pod |
| User | 1-5 | 250m | 256Mi | Auth peu sollicité pendant le test |
| NATS | 1 (3 en prod) | 4 | 6Gi | 500k msg/s × ~1KB = 500MB/s |
| PostgreSQL | 1 | 4 | 6Gi | 1k bulk inserts/s, tuning aggressif |
| Redis | 1 | 1 | 2Gi | 100k sessions |
| k6 runner | 50 | 2 | 4Gi | 2k VUs × 1.5MB = 3GB/pod |
