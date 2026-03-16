# ==============================================================================
# OUTPUTS DE L'ENVIRONNEMENT DEV - Azure
# ==============================================================================
# Ces valeurs sont affichées après tofu apply.
# Copie-les dans tes secrets K8s (infra/k8s/overlays/azure/secrets-azure.yaml).

output "resource_group_name" {
  description = "Nom du resource group"
  value       = azurerm_resource_group.main.name
}

output "postgresql_host" {
  description = "Hostname PostgreSQL (à mettre dans DB_HOST)"
  value       = module.postgresql.postgresql_fqdn
}

output "user_db_name" {
  description = "Nom de la base de données users"
  value       = module.postgresql.user_db_name
}

output "message_db_name" {
  description = "Nom de la base de données messages"
  value       = module.postgresql.message_db_name
}

output "redis_hostname" {
  description = "Hostname Redis"
  value       = module.redis.redis_hostname
}

output "redis_connection_string" {
  description = "String de connexion Redis (à mettre dans REDIS_URL)"
  value       = module.redis.redis_connection_string
}

output "storage_account_name" {
  description = "Nom du storage account"
  value       = module.storage.storage_account_name
}

output "primary_blob_endpoint" {
  description = "URL de base Blob Storage"
  value       = module.storage.primary_blob_endpoint
}

output "avatars_container" {
  description = "Nom du container avatars"
  value       = module.storage.avatars_container_name
}

output "media_container" {
  description = "Nom du container media"
  value       = module.storage.media_container_name
}

output "managed_identity_client_id" {
  description = "Client ID de la Managed Identity (pour configurer AKS)"
  value       = module.managed_identity.identity_client_id
}

output "acr_login_server" {
  description = "URL de login ACR (ex: stormdevacr.azurecr.io)"
  value       = module.acr.acr_login_server
}

output "aks_cluster_name" {
  description = "Nom du cluster AKS"
  value       = module.aks.cluster_name
}

output "oidc_issuer_url" {
  description = "URL OIDC Issuer AKS"
  value       = module.aks.oidc_issuer_url
}
