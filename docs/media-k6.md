# Media Service — Notes + test k6

## Ce qui a été fait
- Media Service passe uniquement par NATS (pas de routes HTTP côté media).
- Subscribers NATS pour:
  - media.upload.requested
  - media.delete.requested
- Service métier avec upload/delete via MinIO (local) / Azure Blob Storage (production).
- Suppression des anciens handlers/routes HTTP du media service.
- Variables d’environnement pour NATS/MinIO.

## Fichiers principaux
- services/media/cmd/main.go
- services/media/internal/subscribers/media_subscriber.go
- services/media/internal/service/media_service.go
- services/media/internal/storage/s3.go

## Prérequis
- NATS démarré.
- MinIO démarré.
- Variables d’environnement MinIO définies.

## Variables d’environnement
- NATS_URL (optionnel)
  - Local: nats://localhost:4222
  - Docker/K8s: nats://nats:4222
- MINIO_ENDPOINT (ex: localhost:9000)
- MINIO_ACCESS_KEY (ex: minioadmin)
- MINIO_SECRET_KEY (ex: minioadmin)
- MINIO_BUCKET (ex: media)

## Lancer le Media Service
1) Exporter les variables MinIO.
2) Lancer le service: go run ./services/media/cmd/main.go

## Lancer le test k6 (media upload)
Commande:
- k6 run tests/k6/load/media_upload_test.js

## Résultat attendu
- checks_failed = 0
- status is 200

## Erreurs fréquentes
- “lookup nats: no such host” → NATS_URL incorrect (utiliser localhost en local).
- “MINIO_ENDPOINT manquant” → variables MinIO non définies.
