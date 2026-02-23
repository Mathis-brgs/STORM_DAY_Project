# ==============================================================================
# ENVIRONNEMENT DEV - Point d'entrée Terraform
# ==============================================================================
#
# Ce fichier assemble tous les modules pour créer l'infrastructure de dev.
#
# Pour déployer :
#   cd infra/terraform/environments/dev
#   terraform init      # Première fois seulement
#   terraform plan      # Voir ce qui va être créé
#   terraform apply     # Créer l'infrastructure
#
# ==============================================================================

# ------------------------------------------------------------------------------
# Configuration Terraform
# ------------------------------------------------------------------------------

terraform {
  required_version = ">= 1.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # Backend S3 pour stocker l'état (à activer plus tard)
  # backend "s3" {
  #   bucket         = "storm-terraform-state"
  #   key            = "dev/terraform.tfstate"
  #   region         = "eu-west-3"
  #   encrypt        = true
  #   dynamodb_table = "storm-terraform-locks"
  # }
}

# ------------------------------------------------------------------------------
# Provider AWS
# ------------------------------------------------------------------------------

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = var.project_name
      Environment = var.environment
      ManagedBy   = "Terraform"
    }
  }
}

# ------------------------------------------------------------------------------
# Module VPC
# ------------------------------------------------------------------------------

module "vpc" {
  source = "../../modules/vpc"

  project_name       = var.project_name
  environment        = var.environment
  vpc_cidr           = var.vpc_cidr
  availability_zones = var.availability_zones
}

# ------------------------------------------------------------------------------
# Module Security Groups
# ------------------------------------------------------------------------------

module "security_groups" {
  source = "../../modules/security-groups"

  project_name = var.project_name
  environment  = var.environment
  vpc_id       = module.vpc.vpc_id
  vpc_cidr     = module.vpc.vpc_cidr
}

# ------------------------------------------------------------------------------
# Module RDS PostgreSQL
# ------------------------------------------------------------------------------

module "rds" {
  source = "../../modules/rds"

  project_name       = var.project_name
  environment        = var.environment
  private_subnet_ids = module.vpc.private_subnet_ids
  security_group_id  = module.security_groups.rds_security_group_id

  # Configuration
  postgres_version  = "15.4"
  instance_class    = "db.t3.micro"  # Plus petit, ~$15/mois
  allocated_storage = 20
  database_name     = "storm"
  master_username   = var.db_username
  master_password   = var.db_password
  multi_az          = false  # Pas de HA en dev
}

# ------------------------------------------------------------------------------
# Module ElastiCache Redis
# ------------------------------------------------------------------------------

module "elasticache" {
  source = "../../modules/elasticache"

  project_name       = var.project_name
  environment        = var.environment
  private_subnet_ids = module.vpc.private_subnet_ids
  security_group_id  = module.security_groups.redis_security_group_id

  # Configuration
  redis_version      = "7.0"
  node_type          = "cache.t3.micro"  # Plus petit, ~$12/mois
  num_cache_clusters = 1                 # Pas de réplication en dev
}

# ------------------------------------------------------------------------------
# Module S3
# ------------------------------------------------------------------------------

module "s3" {
  source = "../../modules/s3"

  project_name         = var.project_name
  environment          = var.environment
  cors_allowed_origins = ["*"]  # A restreindre en prod
}

# ------------------------------------------------------------------------------
# Module IAM
# ------------------------------------------------------------------------------

module "iam" {
  source = "../../modules/iam"

  project_name   = var.project_name
  environment    = var.environment
  s3_bucket_arns = [
    module.s3.avatars_bucket_arn,
    module.s3.media_bucket_arn
  ]
}

# ------------------------------------------------------------------------------
# Module Budget (alertes de coûts - GRATUIT)
# ------------------------------------------------------------------------------

module "budget" {
  source = "../../modules/budget"

  project_name         = var.project_name
  monthly_budget_limit = var.monthly_budget_limit
  alert_emails         = var.alert_emails
}
