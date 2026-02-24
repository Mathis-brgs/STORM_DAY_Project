variable "project_name" {
  description = "Nom du projet"
  type        = string
}

variable "environment" {
  description = "Environnement : dev, staging, prod"
  type        = string
}

variable "location" {
  description = "Région Azure"
  type        = string
}

variable "resource_group_name" {
  description = "Nom du resource group Azure"
  type        = string
}

variable "subnet_id" {
  description = "ID du subnet délégué pour PostgreSQL Flexible Server"
  type        = string
}

variable "private_dns_zone_id" {
  description = "ID de la zone DNS privée PostgreSQL"
  type        = string
}

variable "db_username" {
  description = "Nom d'utilisateur administrateur"
  type        = string
  default     = "storm_admin"
}

variable "db_password" {
  description = "Mot de passe administrateur"
  type        = string
  sensitive   = true
}

variable "postgres_version" {
  description = "Version PostgreSQL"
  type        = string
  default     = "15"
}

# B_Standard_B1ms = ~$12/mois (dev)
# GP_Standard_D2s_v3 = prod
variable "sku_name" {
  description = "SKU du serveur (ex: B_Standard_B1ms pour dev)"
  type        = string
  default     = "B_Standard_B1ms"
}

variable "storage_mb" {
  description = "Stockage en Mo (32768 = 32 Go)"
  type        = number
  default     = 32768
}
