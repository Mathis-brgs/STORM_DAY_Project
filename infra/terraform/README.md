# Infrastructure AWS avec Terraform

## C'est quoi Terraform ?

**Terraform** est un outil d'Infrastructure as Code (IaC). Au lieu de cliquer dans la console AWS pour créer des ressources, tu écris du code qui décrit ton infrastructure. Avantages :

- **Reproductible** : même code = même infrastructure
- **Versionné** : ton infra est dans Git, tu peux revenir en arrière
- **Automatisé** : un seul `terraform apply` crée tout

---

## Structure du projet

```
infra/terraform/
├── README.md                    # Ce fichier
├── environments/
│   ├── dev/                     # Environnement de développement
│   │   ├── main.tf              # Point d'entrée, appelle les modules
│   │   ├── variables.tf         # Variables de l'environnement
│   │   ├── outputs.tf           # Valeurs exportées
│   │   └── terraform.tfvars     # Valeurs des variables
│   └── prod/                    # Environnement de production
└── modules/                     # Modules réutilisables
    ├── vpc/                     # Réseau AWS
    ├── security-groups/         # Pare-feu
    ├── rds/                     # Base de données PostgreSQL
    ├── elasticache/             # Redis
    ├── s3/                      # Stockage fichiers
    └── iam/                     # Permissions
```

---

## Architecture AWS cible

```
┌─────────────────────────────────────────────────────────────────┐
│                         INTERNET                                │
└─────────────────────────────────┬───────────────────────────────┘
                                  │
                    ┌─────────────▼─────────────┐
                    │   Application Load        │
                    │      Balancer (ALB)       │
                    │   (Point d'entrée HTTPS)  │
                    └─────────────┬─────────────┘
                                  │
┌─────────────────────────────────▼───────────────────────────────┐
│                         VPC (10.0.0.0/16)                       │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    SUBNETS PUBLICS                       │   │
│  │  ┌─────────────────┐         ┌─────────────────┐        │   │
│  │  │  10.0.1.0/24    │         │  10.0.2.0/24    │        │   │
│  │  │    (eu-west-3a) │         │    (eu-west-3b) │        │   │
│  │  │  NAT Gateway    │         │                 │        │   │
│  │  └─────────────────┘         └─────────────────┘        │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                  │                              │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    SUBNETS PRIVÉS                        │   │
│  │  ┌─────────────────┐         ┌─────────────────┐        │   │
│  │  │  10.0.10.0/24   │         │  10.0.11.0/24   │        │   │
│  │  │    (eu-west-3a) │         │    (eu-west-3b) │        │   │
│  │  │  ┌───────────┐  │         │  ┌───────────┐  │        │   │
│  │  │  │    EKS    │  │         │  │   Redis   │  │        │   │
│  │  │  │  Cluster  │  │         │  │           │  │        │   │
│  │  │  └───────────┘  │         │  └───────────┘  │        │   │
│  │  │  ┌───────────┐  │         │                 │        │   │
│  │  │  │    RDS    │  │         │                 │        │   │
│  │  │  │ PostgreSQL│  │         │                 │        │   │
│  │  │  └───────────┘  │         │                 │        │   │
│  │  └─────────────────┘         └─────────────────┘        │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                    S3 BUCKETS (hors VPC)                        │
│  ┌─────────────────┐         ┌─────────────────┐               │
│  │  storm-avatars  │         │  storm-media    │               │
│  └─────────────────┘         └─────────────────┘               │
└─────────────────────────────────────────────────────────────────┘
```

---

## Concepts clés

### VPC (Virtual Private Cloud)
Ton réseau privé isolé dans AWS. C'est comme ton propre datacenter virtuel.

### Subnets (Sous-réseaux)
- **Publics** : accessibles depuis internet (Load Balancer, NAT)
- **Privés** : isolés d'internet (RDS, Redis, EKS) - plus sécurisé

### Security Groups
Pare-feu virtuels. Définissent quels ports sont ouverts et depuis où.

### RDS
PostgreSQL managé par AWS (backups auto, mises à jour, haute dispo).

### ElastiCache
Redis managé par AWS (cache, sessions).

### S3
Stockage de fichiers illimité (avatars, médias).

---

## Commandes Terraform

```bash
# 1. Initialiser (télécharge les providers)
terraform init

# 2. Voir ce qui va être créé (GRATUIT)
terraform plan

# 3. Créer l'infrastructure (PAYANT)
terraform apply

# 4. Détruire tout
terraform destroy
```

---

## Coûts estimés (Paris)

| Service | Coût/mois |
|---------|-----------|
| NAT Gateway | ~$35 |
| RDS db.t3.micro | ~$15 |
| ElastiCache cache.t3.micro | ~$12 |
| S3 | ~$1 |
| **TOTAL** | **~$63/mois** |

---

## Prérequis

```bash
# Installer Terraform
brew install terraform

# Installer AWS CLI
brew install awscli

# Configurer AWS (avec les clés de l'école)
aws configure
```

---

## Démarrage

```bash
cd infra/terraform/environments/dev
terraform init
terraform plan    # Vérifier (gratuit)
terraform apply   # Créer (quand l'école paie)
```
