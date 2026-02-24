## Storage — Client MinIO (local) / Azure Blob Storage (production)

### Commandes pour lancer :

- **Démarrer MinIO :**
  ```
  docker compose up -d minio
  ```

- **Lancer le test d'upload :**
  ```
  cd services/media && MINIO_ENDPOINT=localhost:9000 MINIO_ACCESS_KEY=admin MINIO_SECRET_KEY=password go run ./cmd/media-test
  ```

- **(Optionnel) Console MinIO :**
  ```
  http://localhost:9001  (admin / password)
  ```

### À quoi ça sert ?

Le client `MinIOClient` (`internal/storage/s3.go`) gère le stockage d'objets (images, vidéos, documents) en local via **MinIO**.

- En **local** : MinIO (docker-compose) — API compatible S3, aucun cloud nécessaire.
- En **production Azure** : Azure Blob Storage remplace MinIO. Les containers `avatars` et `media` sont provisionnés par Terraform (`infra/terraform/modules/storage/`).

### Variables d'environnement (local)

| Variable | Exemple | Description |
|----------|---------|-------------|
| `MINIO_ENDPOINT` | `localhost:9000` | Adresse MinIO |
| `MINIO_ACCESS_KEY` | `admin` | Identifiant |
| `MINIO_SECRET_KEY` | `password` | Mot de passe |
| `MINIO_BUCKET` | `media` | Bucket cible |
