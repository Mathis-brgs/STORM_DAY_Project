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

variable "acr_name" {
  description = "Nom de l'ACR (globalement unique, alphanum, 5-50 chars)"
  type        = string
}

variable "sku" {
  description = "SKU de l'ACR : Basic, Standard, Premium"
  type        = string
  default     = "Basic"
}
