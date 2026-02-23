# Test Terraform en local (LocalStack)

Permet de tester l'infrastructure Terraform **gratuitement** sur ta machine
sans accès AWS.

## Prérequis

```bash
# 1. OpenTofu (fork open source de Terraform, ARM64 natif)
curl -L https://github.com/opentofu/opentofu/releases/download/v1.8.8/tofu_1.8.8_darwin_arm64.zip -o /tmp/tofu.zip
unzip /tmp/tofu.zip -d /tmp/tofu-bin
sudo mv /tmp/tofu-bin/tofu /usr/local/bin/tofu

# 2. AWS CLI (pour vérifier les ressources)
brew install awscli

# 3. Configurer des fausses credentials (LocalStack accepte n'importe quoi)
aws configure
# Access Key ID     : test
# Secret Access Key : test
# Region            : eu-west-3
# Output format     : json
```

## Démarrage

```bash
# 1. Lancer LocalStack
cd <racine du projet>
docker compose up -d localstack

# 2. Vérifier que LocalStack est prêt
docker logs storm-localstack
# → doit afficher "Ready."

# 3. Aller dans l'environnement local
cd infra/terraform/environments/local

# 4. Initialiser
tofu init

# 5. Voir ce qui va être créé
tofu plan

# 6. Créer les ressources
tofu apply
# → taper "yes" pour confirmer
```

## Vérifier que ça marche

```bash
# Lister les buckets S3 créés
aws --endpoint-url=http://localhost:4566 s3 ls

# Résultat attendu :
# storm-avatars-local
# storm-media-local
```

## Ressources créées

| Ressource | Nom | Rôle |
|-----------|-----|------|
| S3 Bucket | `storm-avatars-local` | Photos de profil |
| S3 Bucket | `storm-media-local` | Fichiers uploadés |
| IAM Role | `storm-app-role` | Permissions des services |
| IAM Policy | `storm-s3-access` | Droits lecture/écriture S3 |

## Nettoyer

```bash
# Supprimer toutes les ressources LocalStack
tofu destroy

# Arrêter LocalStack
docker compose stop localstack
```

## Note

L'environnement local teste uniquement **S3 + IAM**.
VPC, RDS, ElastiCache et les alertes budget nécessitent les vrais accès AWS
(voir `infra/k8s/overlays/aws/README.md`).
