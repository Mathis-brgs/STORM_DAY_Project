# ==============================================================================
# VARIABLES DU MODULE SECURITY GROUPS
# ==============================================================================

variable "project_name" {
  description = "Nom du projet"
  type        = string
}

variable "environment" {
  description = "Environnement (dev, staging, prod)"
  type        = string
}

variable "vpc_id" {
  description = "ID du VPC où créer les security groups"
  type        = string
}

variable "vpc_cidr" {
  description = "CIDR du VPC (pour autoriser le trafic interne)"
  type        = string
}
