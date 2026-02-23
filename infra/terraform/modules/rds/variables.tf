# ==============================================================================
# VARIABLES DU MODULE RDS
# ==============================================================================

variable "project_name" {
  description = "Nom du projet"
  type        = string
}

variable "environment" {
  description = "Environnement (dev, staging, prod)"
  type        = string
}

variable "private_subnet_ids" {
  description = "Liste des IDs des subnets privés"
  type        = list(string)
}

variable "security_group_id" {
  description = "ID du security group RDS"
  type        = string
}

# Configuration PostgreSQL
variable "postgres_version" {
  description = "Version de PostgreSQL"
  type        = string
  default     = "15.4"
}

variable "instance_class" {
  description = "Type d'instance RDS"
  type        = string
  default     = "db.t3.micro"  # Le plus petit, ~$15/mois
}

variable "allocated_storage" {
  description = "Stockage initial en Go"
  type        = number
  default     = 20
}

variable "max_allocated_storage" {
  description = "Stockage maximum (auto-scaling)"
  type        = number
  default     = 100
}

# Credentials
variable "database_name" {
  description = "Nom de la base de données"
  type        = string
  default     = "storm"
}

variable "master_username" {
  description = "Nom d'utilisateur admin"
  type        = string
  default     = "storm_admin"
}

variable "master_password" {
  description = "Mot de passe admin (à mettre dans terraform.tfvars)"
  type        = string
  sensitive   = true  # Ne s'affiche pas dans les logs
}

# Options
variable "backup_retention_days" {
  description = "Nombre de jours de rétention des backups"
  type        = number
  default     = 7
}

variable "multi_az" {
  description = "Activer Multi-AZ (haute disponibilité)"
  type        = bool
  default     = false  # true en prod
}
