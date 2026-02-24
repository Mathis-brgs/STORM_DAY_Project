output "storage_account_name" {
  description = "Nom du storage account"
  value       = azurerm_storage_account.main.name
}

output "storage_account_id" {
  description = "ID du storage account"
  value       = azurerm_storage_account.main.id
}

output "primary_blob_endpoint" {
  description = "URL de base pour accéder au Blob Storage"
  value       = azurerm_storage_account.main.primary_blob_endpoint
}

output "storage_account_primary_access_key" {
  description = "Clé d'accès primaire (pour le media service)"
  value       = azurerm_storage_account.main.primary_access_key
  sensitive   = true
}

output "avatars_container_name" {
  description = "Nom du container avatars"
  value       = azurerm_storage_container.avatars.name
}

output "media_container_name" {
  description = "Nom du container media"
  value       = azurerm_storage_container.media.name
}
