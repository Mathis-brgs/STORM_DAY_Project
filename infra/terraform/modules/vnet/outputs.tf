output "vnet_id" {
  description = "ID du VNet"
  value       = azurerm_virtual_network.main.id
}

output "vnet_name" {
  description = "Nom du VNet"
  value       = azurerm_virtual_network.main.name
}

output "vnet_cidr" {
  description = "CIDR du VNet"
  value       = azurerm_virtual_network.main.address_space[0]
}

output "public_subnet_ids" {
  description = "IDs des subnets publics"
  value       = azurerm_subnet.public[*].id
}

output "private_postgresql_subnet_id" {
  description = "ID du subnet privé pour PostgreSQL"
  value       = azurerm_subnet.private_postgresql.id
}

output "postgresql_private_dns_zone_id" {
  description = "ID de la zone DNS privée PostgreSQL"
  value       = azurerm_private_dns_zone.postgresql.id
}
