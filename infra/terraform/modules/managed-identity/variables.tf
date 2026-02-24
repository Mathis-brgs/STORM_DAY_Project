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

variable "storage_account_id" {
  description = "ID du storage account auquel donner accès"
  type        = string
}
