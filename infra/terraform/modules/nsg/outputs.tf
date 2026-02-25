output "app_nsg_id" {
  description = "ID du NSG pour les applications"
  value       = azurerm_network_security_group.app.id
}

output "postgresql_nsg_id" {
  description = "ID du NSG pour PostgreSQL"
  value       = azurerm_network_security_group.postgresql.id
}

output "redis_nsg_id" {
  description = "ID du NSG pour Redis"
  value       = azurerm_network_security_group.redis.id
}

output "lb_nsg_id" {
  description = "ID du NSG pour le Load Balancer"
  value       = azurerm_network_security_group.lb.id
}
