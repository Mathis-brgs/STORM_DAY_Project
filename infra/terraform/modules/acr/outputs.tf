output "acr_id" {
  description = "ID de l'ACR (utilisé pour attacher à AKS)"
  value       = azurerm_container_registry.main.id
}

output "acr_login_server" {
  description = "URL de login ACR (ex: stormdevacr.azurecr.io)"
  value       = azurerm_container_registry.main.login_server
}

output "acr_admin_username" {
  description = "Username admin ACR"
  value       = azurerm_container_registry.main.admin_username
}

output "acr_admin_password" {
  description = "Mot de passe admin ACR"
  value       = azurerm_container_registry.main.admin_password
  sensitive   = true
}
