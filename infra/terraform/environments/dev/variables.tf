# ==============================================================================
# VARIABLES DE L'ENVIRONNEMENT DEV
# ==============================================================================

# ------------------------------------------------------------------------------
# Général
# ------------------------------------------------------------------------------

variable "project_name" {
  description = "Nom du projet"
  type        = string
  default     = "storm"
}

variable "environment" {
  description = "Environnement"
  type        = string
  default     = "dev"
}

variable "aws_region" {
  description = "Région AWS"
  type        = string
  default     = "eu-west-3"  # Paris
}

# ------------------------------------------------------------------------------
# Réseau
# ------------------------------------------------------------------------------

variable "vpc_cidr" {
  description = "CIDR du VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "Zones de disponibilité"
  type        = list(string)
  default     = ["eu-west-3a", "eu-west-3b"]
}

# ------------------------------------------------------------------------------
# Base de données
# ------------------------------------------------------------------------------

variable "db_username" {
  description = "Username admin PostgreSQL"
  type        = string
  default     = "storm_admin"
}

variable "db_password" {
  description = "Mot de passe admin PostgreSQL"
  type        = string
  sensitive   = true  # Ne s'affiche pas dans les logs
}

# ------------------------------------------------------------------------------
# Budget
# ------------------------------------------------------------------------------

variable "monthly_budget_limit" {
  description = "Budget mensuel max en USD"
  type        = string
  default     = "100"
}

variable "alert_emails" {
  description = "Emails pour les alertes budget"
  type        = list(string)
  default     = ["mathis@example.com"]  # À changer
}
