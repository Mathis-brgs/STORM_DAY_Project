# Déploiement AWS - Guide

## Prérequis

1. **Accès AWS** (fournis par l'école)
2. **AWS CLI configuré** : `aws configure`
3. **Terraform** : `brew install terraform`
4. **kubectl** : déjà installé

## Étape 1 : Créer l'infrastructure avec Terraform

```bash
cd infra/terraform/environments/dev

# Configurer les variables
cp terraform.tfvars.example terraform.tfvars
# Éditer terraform.tfvars avec les vrais mots de passe

# Déployer
terraform init
terraform plan       # Vérifier
terraform apply      # Créer (~5 min)
```

Résultat : VPC, RDS, ElastiCache, S3, IAM créés.

## Étape 2 : Créer les repos ECR (registry Docker)

```bash
# Créer un repo par service
for SERVICE in user gateway media message; do
  aws ecr create-repository --repository-name storm-${SERVICE} --region eu-west-3
done
```

## Étape 3 : Configurer les secrets K8s

```bash
# Récupérer les outputs Terraform
terraform output

# Éditer les secrets avec les vraies valeurs
vim infra/k8s/overlays/aws/secrets-aws.yaml
# Remplacer les REPLACE_WITH_... par les outputs Terraform
```

## Étape 4 : Déployer sur EKS

```bash
# Configurer kubectl pour EKS
aws eks update-kubeconfig --name storm-cluster --region eu-west-3

# Déployer via Kustomize
kubectl apply -k infra/k8s/overlays/aws

# Vérifier
kubectl get pods -n storm
kubectl get svc -n storm
```

## Étape 5 : Configurer GitHub Actions

Ajouter ces secrets dans GitHub (Settings > Secrets > Actions) :

| Secret | Valeur |
|--------|--------|
| `AWS_ACCOUNT_ID` | Ton ID de compte AWS |
| `AWS_DEPLOY_ROLE_ARN` | ARN du rôle IAM pour le déploiement |

## Commandes utiles

```bash
# Voir les pods
kubectl get pods -n storm

# Logs d'un service
kubectl logs -f deployment/user-service -n storm

# Redéployer un service
kubectl rollout restart deployment/user-service -n storm

# Voir les coûts
aws ce get-cost-and-usage \
  --time-period Start=2026-02-01,End=2026-02-28 \
  --granularity MONTHLY \
  --metrics BlendedCost

# Détruire tout (pour économiser)
terraform destroy
```

## Architecture sur AWS

```
Local (k3d)                    →    AWS
─────────────────────────────────────────────────
postgres-user (pod)            →    RDS PostgreSQL
redis (pod)                    →    ElastiCache Redis
minio (pod)                    →    S3 buckets
storm/user-service (image)     →    ECR + EKS
storm/gateway-service (image)  →    ECR + EKS
nats (pod)                     →    NATS (reste dans EKS)
```
