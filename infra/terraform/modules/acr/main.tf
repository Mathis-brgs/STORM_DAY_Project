# ==============================================================================
# MODULE ACR - Azure Container Registry
# ==============================================================================
#
# Stocke les images Docker des services STORM.
# SKU Basic ~$5/mois suffit pour un projet dev.
# Admin account activé pour que AKS puisse pull les images.
#
# ==============================================================================

resource "azurerm_container_registry" "main" {
  name                = var.acr_name
  resource_group_name = var.resource_group_name
  location            = var.location
  sku                 = var.sku

  # Admin activé pour credentials simples (pull depuis AKS via secret ou attach)
  admin_enabled = true

  tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}
