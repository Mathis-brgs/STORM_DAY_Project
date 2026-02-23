# Organisation des commits – Message Service

Proposition d’organisation des commits pour créer une branche propre.

## Créer la branche

```bash
git checkout -b feature/message-service-v1
```

## Commits suggérés (ordre)

### 1. Infra et configuration

```bash
git add .gitignore
git commit -m "chore: gitignore message-service et _local"
```

### 2. Message-service – base

```bash
git add services/message/api/ services/message/internal/ services/message/cmd/ services/message/go.mod services/message/go.sum services/message/Dockerfile
git commit -m "feat(message): service NATS NEW_MESSAGE avec protobuf, memory et postgres"
```

### 3. Migrations

```bash
git add services/message/migrations/
git commit -m "feat(message): migrations et seed SQL"
```

### 4. Tooling

```bash
git add services/message/Makefile services/message/buf*.yaml services/message/api/buf.yaml
git commit -m "chore(message): Makefile et config buf"
```

### 5. Documentation

```bash
git add services/message/README.md services/message/FUTURE_STEPS.md services/message/COMMITS.md
git commit -m "docs(message): README, roadmap et guide des commits"
```

### 6. Changements gateway (si committés)

```bash
git add services/gateway/
git commit -m "feat(gateway): POST /api/messages vers message-service (protobuf)"
```

## Variante : tout en un commit

```bash
git checkout -b feature/message-service-v1
git add services/message/ .gitignore
# Exclure services/gateway si géré par l'équipe
git commit -m "feat(message): service NATS protobuf, postgres, migrations, docs"
```

## Pousser la branche

```bash
git push -u origin feature/message-service-v1
```
