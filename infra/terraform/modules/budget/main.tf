# ==============================================================================
# MODULE BUDGET - Alertes de coût AWS
# ==============================================================================
#
# Ce module crée des alertes automatiques quand tu dépasses un certain % du budget.
# Tu reçois un email quand :
# - 50% du budget est atteint
# - 75% du budget est atteint
# - 90% du budget est atteint
# - 100% du budget est atteint
#
# SERVICE GRATUIT - AWS Budgets ne coûte rien
#
# ==============================================================================

resource "aws_budgets_budget" "monthly" {
  name         = "${var.project_name}-monthly-budget"
  budget_type  = "COST"
  limit_amount = var.monthly_budget_limit  # ex: 100 ($)
  limit_unit   = "USD"
  time_unit    = "MONTHLY"

  # Alerte à 50%
  notification {
    comparison_operator       = "GREATER_THAN"
    threshold                 = 50
    threshold_type            = "PERCENTAGE"
    notification_type         = "ACTUAL"
    subscriber_email_addresses = var.alert_emails
  }

  # Alerte à 75%
  notification {
    comparison_operator       = "GREATER_THAN"
    threshold                 = 75
    threshold_type            = "PERCENTAGE"
    notification_type         = "ACTUAL"
    subscriber_email_addresses = var.alert_emails
  }

  # Alerte à 90%
  notification {
    comparison_operator       = "GREATER_THAN"
    threshold                 = 90
    threshold_type            = "PERCENTAGE"
    notification_type         = "ACTUAL"
    subscriber_email_addresses = var.alert_emails
  }

  # Alerte à 100% (DÉPASSEMENT)
  notification {
    comparison_operator       = "GREATER_THAN"
    threshold                 = 100
    threshold_type            = "PERCENTAGE"
    notification_type         = "ACTUAL"
    subscriber_email_addresses = var.alert_emails
  }
}
