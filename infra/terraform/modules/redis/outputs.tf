output "redis_hostname" {
  description = "Hostname de connexion Redis"
  value       = azurerm_redis_cache.main.hostname
}

output "redis_port" {
  description = "Port Redis (non-SSL)"
  value       = azurerm_redis_cache.main.port
}

output "redis_ssl_port" {
  description = "Port Redis SSL"
  value       = azurerm_redis_cache.main.ssl_port
}

output "redis_primary_access_key" {
  description = "Clé d'accès primaire Redis"
  value       = azurerm_redis_cache.main.primary_access_key
  sensitive   = true
}

output "redis_connection_string" {
  description = "String de connexion Redis"
  value       = "redis://${azurerm_redis_cache.main.hostname}:${azurerm_redis_cache.main.port}"
}
