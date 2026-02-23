# ==============================================================================
# OUTPUTS DE L'ENVIRONNEMENT DEV
# ==============================================================================
# Ces valeurs sont affichées après terraform apply.
# Utile pour configurer tes applications.

# ------------------------------------------------------------------------------
# VPC
# ------------------------------------------------------------------------------

output "vpc_id" {
  description = "ID du VPC"
  value       = module.vpc.vpc_id
}

output "private_subnet_ids" {
  description = "IDs des subnets privés"
  value       = module.vpc.private_subnet_ids
}

output "public_subnet_ids" {
  description = "IDs des subnets publics"
  value       = module.vpc.public_subnet_ids
}

# ------------------------------------------------------------------------------
# Base de données
# ------------------------------------------------------------------------------

output "rds_endpoint" {
  description = "Endpoint PostgreSQL (host:port)"
  value       = module.rds.db_instance_endpoint
}

output "rds_database_name" {
  description = "Nom de la base de données"
  value       = module.rds.db_name
}

# ------------------------------------------------------------------------------
# Redis
# ------------------------------------------------------------------------------

output "redis_endpoint" {
  description = "Endpoint Redis"
  value       = module.elasticache.redis_endpoint
}

output "redis_connection_string" {
  description = "String de connexion Redis"
  value       = module.elasticache.redis_connection_string
}

# ------------------------------------------------------------------------------
# S3
# ------------------------------------------------------------------------------

output "s3_avatars_bucket" {
  description = "Nom du bucket avatars"
  value       = module.s3.avatars_bucket_name
}

output "s3_media_bucket" {
  description = "Nom du bucket media"
  value       = module.s3.media_bucket_name
}

# ------------------------------------------------------------------------------
# IAM
# ------------------------------------------------------------------------------

output "app_role_arn" {
  description = "ARN du rôle IAM pour les apps"
  value       = module.iam.app_role_arn
}
