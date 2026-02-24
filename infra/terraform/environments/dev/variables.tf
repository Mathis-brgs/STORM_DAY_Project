# ==============================================================================
# VARIABLES DE L'ENVIRONNEMENT DEV - Azure
# ==============================================================================

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

variable "subscription_id" {
  description = "ID de la subscription Azure (az account show --query id)"
  type        = string
}

variable "location" {
  description = "Région Azure (francecentral = Paris)"
  type        = string
  default     = "francecentral"
}

# Réseau
variable "vnet_cidr" {
  description = "CIDR du VNet"
  type        = string
  default     = "10.0.0.0/16"
}

variable "public_subnet_cidrs" {
  description = "CIDRs des subnets publics (AKS nodes)"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24"]
}

variable "private_subnet_cidr" {
  description = "CIDR du subnet privé (PostgreSQL)"
  type        = string
  default     = "10.0.10.0/24"
}

# Base de données
variable "db_username" {
  description = "Username admin PostgreSQL"
  type        = string
  default     = "storm_admin"
}

variable "db_password" {
  description = "Mot de passe admin PostgreSQL"
  type        = string
  sensitive   = true
}

# Storage
# IMPORTANT : doit être globalement unique sur Azure, 3-24 chars, minuscules alphanum
variable "storage_account_name" {
  description = "Nom du storage account Azure (globalement unique)"
  type        = string
  default     = "stormdev"
}

# Budget
variable "monthly_budget_limit" {
  description = "Budget mensuel max en EUR"
  type        = number
  default     = 100
}

variable "alert_emails" {
  description = "Emails pour les alertes budget"
  type        = list(string)
  default     = ["mathis@example.com"]
}

variable "budget_start_date" {
  description = "Date de début du budget (premier du mois courant, format RFC3339)"
  type        = string
  default     = "2026-02-01T00:00:00Z"
}
