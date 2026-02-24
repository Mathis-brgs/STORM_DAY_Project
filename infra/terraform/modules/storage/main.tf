# ==============================================================================
# MODULE STORAGE - Azure Blob Storage
# ==============================================================================
#
# Azure Blob Storage est l'équivalent d'AWS S3.
# Structure :
#   Storage Account (≈ compte global)
#   └── Container "avatars" (≈ bucket S3)
#   └── Container "media"   (≈ bucket S3)
#
# IMPORTANT : Le nom du storage account doit être globalement unique sur Azure
# et uniquement des minuscules alphanumériques (3-24 chars).
# Si le nom est déjà pris, change la variable `storage_account_name`.
#
# ==============================================================================

resource "azurerm_storage_account" "main" {
  # Nom unique : storm + env (ex: "stormdev") — peut nécessiter un suffix unique
  name                     = var.storage_account_name
  resource_group_name      = var.resource_group_name
  location                 = var.location
  account_tier             = "Standard"
  account_replication_type = "LRS"  # Localement redondant — GRS pour prod

  # Chiffrement activé par défaut dans Azure (pas besoin de le configurer)

  blob_properties {
    cors_rule {
      allowed_origins    = var.cors_allowed_origins
      allowed_methods    = ["GET", "POST", "PUT", "DELETE"]
      allowed_headers    = ["*"]
      exposed_headers    = ["ETag"]
      max_age_in_seconds = 3600
    }

    # Supprimer les uploads incomplets après 7 jours
    delete_retention_policy {
      days = 7
    }
  }

  tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}

# Container pour les avatars utilisateurs
resource "azurerm_storage_container" "avatars" {
  name                  = "avatars"
  storage_account_name  = azurerm_storage_account.main.name
  container_access_type = "private"  # Accès uniquement via SAS token ou Managed Identity
}

# Container pour les médias uploadés
resource "azurerm_storage_container" "media" {
  name                  = "media"
  storage_account_name  = azurerm_storage_account.main.name
  container_access_type = "private"
}
