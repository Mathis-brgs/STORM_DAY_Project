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

variable "redis_version" {
  description = "Version Redis"
  type        = string
  default     = "7"
}

# capacity : 0=C0 250MB, 1=C1 1GB, 2=C2 6GB...
variable "capacity" {
  description = "Taille du cache (0=C0 250MB pour dev)"
  type        = number
  default     = 0
}

# family : C = Basic/Standard, P = Premium
variable "family" {
  description = "Famille Redis (C = Basic/Standard)"
  type        = string
  default     = "C"
}

# Basic = dev (~$16/mois), Standard = prod HA, Premium = clustering
variable "sku_name" {
  description = "SKU Redis (Basic pour dev, Standard pour prod)"
  type        = string
  default     = "Basic"
}
