output "identity_id" {
  description = "ID de la Managed Identity"
  value       = azurerm_user_assigned_identity.app.id
}

output "identity_principal_id" {
  description = "Principal ID (pour les role assignments)"
  value       = azurerm_user_assigned_identity.app.principal_id
}

output "identity_client_id" {
  description = "Client ID (pour configurer les pods AKS)"
  value       = azurerm_user_assigned_identity.app.client_id
}
