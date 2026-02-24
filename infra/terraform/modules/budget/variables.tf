# ==============================================================================
# VARIABLES DU MODULE BUDGET
# ==============================================================================

variable "project_name" {
  description = "Nom du projet"
  type        = string
}

variable "monthly_budget_limit" {
  description = "Budget mensuel maximum en USD"
  type        = string
  default     = "100"
}

variable "alert_emails" {
  description = "Emails qui reçoivent les alertes budget"
  type        = list(string)
}
