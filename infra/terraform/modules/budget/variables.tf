variable "project_name" {
  description = "Nom du projet"
  type        = string
}

variable "environment" {
  description = "Environnement : dev, staging, prod"
  type        = string
}

variable "resource_group_id" {
  description = "ID du resource group à surveiller"
  type        = string
}

variable "monthly_budget_limit" {
  description = "Budget mensuel maximum en EUR/USD"
  type        = number
  default     = 100
}

variable "alert_emails" {
  description = "Emails qui reçoivent les alertes budget"
  type        = list(string)
}

# Format : "2026-02-01T00:00:00Z" — premier jour du mois courant
variable "budget_start_date" {
  description = "Date de début du budget (premier du mois, format RFC3339)"
  type        = string
  default     = "2026-02-01T00:00:00Z"
}
