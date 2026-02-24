# ==============================================================================
# MODULE BUDGET - Azure Cost Management
# ==============================================================================
#
# Ce module crée des alertes de coût Azure quand le budget est dépassé.
# Tu reçois un email quand : 50%, 75%, 90%, 100% du budget est atteint.
#
# Équivalent d'AWS Budgets.
# SERVICE GRATUIT - Azure Cost Management ne coûte rien.
#
# ==============================================================================

resource "azurerm_consumption_budget_resource_group" "main" {
  name              = "${var.project_name}-budget-${var.environment}"
  resource_group_id = var.resource_group_id

  amount     = var.monthly_budget_limit
  time_grain = "Monthly"

  time_period {
    start_date = var.budget_start_date  # Format : "2026-02-01T00:00:00Z"
  }

  # Alerte à 50%
  notification {
    enabled        = true
    threshold      = 50
    operator       = "GreaterThan"
    threshold_type = "Actual"
    contact_emails = var.alert_emails
  }

  # Alerte à 75%
  notification {
    enabled        = true
    threshold      = 75
    operator       = "GreaterThan"
    threshold_type = "Actual"
    contact_emails = var.alert_emails
  }

  # Alerte à 90%
  notification {
    enabled        = true
    threshold      = 90
    operator       = "GreaterThan"
    threshold_type = "Actual"
    contact_emails = var.alert_emails
  }

  # Alerte à 100% (DÉPASSEMENT)
  notification {
    enabled        = true
    threshold      = 100
    operator       = "GreaterThan"
    threshold_type = "Actual"
    contact_emails = var.alert_emails
  }
}
