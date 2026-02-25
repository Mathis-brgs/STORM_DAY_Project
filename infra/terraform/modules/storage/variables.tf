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

# IMPORTANT : doit être globalement unique sur Azure, 3-24 chars, minuscules alphanum
variable "storage_account_name" {
  description = "Nom du storage account (globalement unique, ex: stormdev42)"
  type        = string
}

variable "cors_allowed_origins" {
  description = "Origines autorisées pour CORS (URLs du frontend)"
  type        = list(string)
  default     = ["*"]  # À restreindre en prod
}
