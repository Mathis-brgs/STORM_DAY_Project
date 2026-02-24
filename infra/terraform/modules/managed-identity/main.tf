# ==============================================================================
# MODULE MANAGED IDENTITY - Azure Managed Identity
# ==============================================================================
#
# Managed Identity est l'équivalent des IAM Roles AWS.
# Les pods AKS utilisent cette identité pour accéder à Azure Blob Storage
# sans stocker de credentials dans le code.
#
# ==============================================================================

# Identité managée assignée par l'utilisateur (User-Assigned)
resource "azurerm_user_assigned_identity" "app" {
  name                = "${var.project_name}-identity-${var.environment}"
  location            = var.location
  resource_group_name = var.resource_group_name

  tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}

# Donner accès au Blob Storage (Storage Blob Data Contributor)
resource "azurerm_role_assignment" "storage_blob" {
  scope                = var.storage_account_id
  role_definition_name = "Storage Blob Data Contributor"
  principal_id         = azurerm_user_assigned_identity.app.principal_id
}
