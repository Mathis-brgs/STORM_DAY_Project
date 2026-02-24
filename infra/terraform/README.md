# Infrastructure Azure avec OpenTofu (Terraform)

## C'est quoi OpenTofu ?

**OpenTofu** (fork open-source de Terraform) est un outil d'Infrastructure as Code (IaC). Au lieu de cliquer dans le portail Azure pour créer des ressources, tu écris du code qui décrit ton infrastructure. Avantages :

- **Reproductible** : même code = même infrastructure
- **Versionné** : ton infra est dans Git, tu peux revenir en arrière
- **Automatisé** : un seul `tofu apply` crée tout

---

## Structure du projet

```
infra/terraform/
├── README.md                    # Ce fichier
├── environments/
│   └── dev/                     # Environnement de développement (Azure)
│       ├── main.tf              # Point d'entrée, appelle les modules
│       ├── variables.tf         # Variables de l'environnement
│       ├── outputs.tf           # Valeurs exportées après apply
│       └── terraform.tfvars.example  # Exemple de fichier de config
└── modules/                     # Modules réutilisables
    ├── vnet/                    # Réseau Azure (VNet + subnets)
    ├── nsg/                     # Pare-feu (Network Security Groups)
    ├── postgresql/              # Azure Database for PostgreSQL Flexible
    ├── redis/                   # Azure Cache for Redis
    ├── storage/                 # Azure Blob Storage (avatars + médias)
    ├── managed-identity/        # Identité managée (accès Blob Storage)
    └── budget/                  # Alertes de coût Azure
```

---

## Architecture Azure cible

```
┌─────────────────────────────────────────────────────────────────┐
│                         INTERNET                                │
└─────────────────────────────────┬───────────────────────────────┘
                                  │
                    ┌─────────────▼─────────────┐
                    │   Azure Load Balancer      │
                    │   (Point d'entrée HTTPS)   │
                    └─────────────┬─────────────┘
                                  │
┌─────────────────────────────────▼───────────────────────────────┐
│                    VNet (10.0.0.0/16)                           │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    SUBNETS PUBLICS                       │   │
│  │  ┌─────────────────┐         ┌─────────────────┐        │   │
│  │  │  10.0.1.0/24    │         │  10.0.2.0/24    │        │   │
│  │  │  AKS nodes      │         │  AKS nodes      │        │   │
│  │  └─────────────────┘         └─────────────────┘        │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │               SUBNET PRIVÉ (délégué PostgreSQL)          │   │
│  │  ┌───────────────────────────────────────────────────┐  │   │
│  │  │  10.0.10.0/24  — PostgreSQL Flexible Server       │  │   │
│  │  └───────────────────────────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│              SERVICES MANAGÉS AZURE (hors VNet)                 │
│  ┌─────────────────┐  ┌─────────────────┐  ┌───────────────┐   │
│  │  Azure Cache    │  │  Blob Storage   │  │    Blob       │   │
│  │  for Redis      │  │  storm-avatars  │  │  storm-media  │   │
│  └─────────────────┘  └─────────────────┘  └───────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Correspondance AWS → Azure

| AWS | Azure |
|-----|-------|
| VPC | Virtual Network (VNet) |
| Security Groups | Network Security Groups (NSG) |
| RDS PostgreSQL | Azure Database for PostgreSQL Flexible Server |
| ElastiCache Redis | Azure Cache for Redis |
| S3 | Azure Blob Storage (Storage Account) |
| IAM Role | Managed Identity |
| ECR | Azure Container Registry (ACR) |
| EKS | Azure Kubernetes Service (AKS) |

---

## Concepts clés Azure

### Resource Group
Conteneur logique pour toutes les ressources. Tout doit être dans un resource group.

### VNet (Virtual Network)
Réseau privé isolé dans Azure. Équivalent du VPC AWS.

### NSG (Network Security Group)
Pare-feu virtuels sur les subnets. Définissent quels ports sont ouverts et depuis où.

### PostgreSQL Flexible Server
PostgreSQL managé par Azure (backups auto, mises à jour, haute dispo optionnelle).

### Azure Cache for Redis
Redis managé par Azure (cache, sessions, rate limiting).

### Azure Blob Storage
Stockage de fichiers (avatars, médias). Organisé en Storage Account → Containers.

### Managed Identity
Identité pour les pods AKS permettant d'accéder à Blob Storage sans stocker de credentials.

---

## Commandes OpenTofu

```bash
# 1. Initialiser (télécharge les providers)
tofu init

# 2. Voir ce qui va être créé (GRATUIT)
tofu plan

# 3. Créer l'infrastructure (PAYANT)
tofu apply

# 4. Voir les outputs (endpoints, noms de ressources)
tofu output

# 5. Détruire tout
tofu destroy
```

---

## Coûts estimés (francecentral - Paris)

| Service | SKU dev | Coût/mois |
|---------|---------|-----------|
| PostgreSQL Flexible Server | B_Standard_B1ms | ~$12 |
| Azure Cache for Redis | Basic C0 (250MB) | ~$16 |
| Azure Blob Storage | Standard LRS | ~$1 |
| Managed Identity | - | Gratuit |
| Budget alerts | - | Gratuit |
| **TOTAL** | | **~$29/mois** |

> Note : Pas de NAT Gateway (~$35/mois AWS) → Azure VNet ne facture pas la sortie interne de la même façon. Économie significative vs AWS (~$63/mois).

---

## Prérequis

```bash
# Installer OpenTofu (ARM64 Mac)
brew install opentofu

# Installer Azure CLI
brew install azure-cli

# Se connecter à Azure
az login
az account set --subscription <SUBSCRIPTION_ID>
```

---

## Démarrage

```bash
cd infra/terraform/environments/dev

# Copier et remplir les variables
cp terraform.tfvars.example terraform.tfvars
# → Remplis subscription_id, db_password, storage_account_name

tofu init      # Télécharge le provider azurerm
tofu plan      # Vérifier (gratuit)
tofu apply     # Créer (quand l'école fournit les accès)
```

---

## Après le apply — copier les outputs dans K8s

```bash
# Afficher tous les outputs
tofu output

# Copier dans le fichier de secrets K8s
# infra/k8s/overlays/azure/secrets-azure.yaml
tofu output postgresql_host      # → DB_HOST
tofu output redis_connection_string  # → REDIS_URL
tofu output storage_account_name     # → AZURE_STORAGE_ACCOUNT
```
