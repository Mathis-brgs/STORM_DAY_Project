# Environnement local (développement sans Azure)

> **Note** : Cet environnement ne contient plus de code Terraform actif.
> LocalStack (émulateur AWS) a été supprimé lors de la migration vers Azure.
> Il n'existe pas d'équivalent Azure complet pour tester Terraform localement.

---

## Comment tester en local ?

Le développement local utilise directement **docker-compose** à la racine du projet.

```bash
# Lancer tous les services locaux
docker compose up -d

# Services disponibles :
# - PostgreSQL user  → localhost:5432
# - PostgreSQL chat  → localhost:5433
# - Redis            → localhost:6379
# - NATS             → localhost:4222 (dashboard : http://localhost:8222)
# - MinIO            → localhost:9000 (console : http://localhost:9001)
```

---

## Correspondance local → Azure

| Local (docker-compose) | Azure (production) |
|------------------------|--------------------|
| PostgreSQL (Docker) | Azure Database for PostgreSQL Flexible Server |
| Redis (Docker) | Azure Cache for Redis |
| MinIO (Docker) | Azure Blob Storage (Storage Account) |
| NATS (Docker) | NATS (auto-hébergé sur AKS) |

---

## MinIO — équivalent local de Azure Blob Storage

MinIO remplace Azure Blob Storage en local. L'API est compatible S3/Blob.

```bash
# Accès console MinIO
open http://localhost:9001
# Login : admin / password

# Créer les buckets manuellement dans la console
# ou via la CLI mc :
brew install minio/stable/mc
mc alias set local http://localhost:9000 admin password
mc mb local/storm-avatars
mc mb local/storm-media
mc ls local
```

---

## Valider la syntaxe Terraform (sans Azure)

Pour vérifier que le code Terraform Azure est syntaxiquement correct sans avoir accès à une subscription Azure :

```bash
cd infra/terraform/environments/dev

# Copier les variables (mettre n'importe quelle valeur pour valider)
cp terraform.tfvars.example terraform.tfvars

# Télécharger le provider azurerm
tofu init

# Valider la syntaxe (ne crée RIEN, ne se connecte pas à Azure)
tofu validate

# Voir le plan (échouera sans vraies credentials, mais montre les erreurs de config)
tofu plan
```

---

## Déployer en vrai (quand l'école fournit les accès Azure)

```bash
az login
az account set --subscription <SUBSCRIPTION_ID>

cd infra/terraform/environments/dev
# Remplir terraform.tfvars avec les vraies valeurs
tofu apply
```

Voir [infra/terraform/README.md](../../README.md) pour la documentation complète.
