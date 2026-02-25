variable "project_name" {
  description = "Nom du projet (utilisé pour nommer les ressources)"
  type        = string
}

variable "environment" {
  description = "Environnement : dev, staging, prod"
  type        = string
}

variable "location" {
  description = "Région Azure (ex: francecentral, westeurope)"
  type        = string
  default     = "francecentral"
}

variable "resource_group_name" {
  description = "Nom du resource group Azure"
  type        = string
}

variable "vnet_cidr" {
  description = "Plage d'adresses IP du VNet (ex: 10.0.0.0/16)"
  type        = string
  default     = "10.0.0.0/16"
}

variable "public_subnet_cidrs" {
  description = "CIDRs des subnets publics (AKS nodes)"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24"]
}

variable "private_subnet_cidr" {
  description = "CIDR du subnet privé (PostgreSQL Flexible Server)"
  type        = string
  default     = "10.0.10.0/24"
}
