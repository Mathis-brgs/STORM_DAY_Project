# ==============================================================================
# VARIABLES DU MODULE ELASTICACHE
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
  description = "ID du security group Redis"
  type        = string
}

variable "redis_version" {
  description = "Version de Redis"
  type        = string
  default     = "7.0"
}

variable "node_type" {
  description = "Type d'instance Redis"
  type        = string
  default     = "cache.t3.micro"  # Le plus petit, ~$12/mois
}

variable "num_cache_clusters" {
  description = "Nombre de nodes (1 en dev, 2+ en prod pour HA)"
  type        = number
  default     = 1
}
