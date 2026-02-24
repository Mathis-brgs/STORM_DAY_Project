output "postgresql_fqdn" {
  description = "FQDN du serveur PostgreSQL (host de connexion)"
  value       = azurerm_postgresql_flexible_server.main.fqdn
}

output "postgresql_host" {
  description = "Hostname du serveur PostgreSQL"
  value       = azurerm_postgresql_flexible_server.main.fqdn
}

output "db_name" {
  description = "Nom de la base de données"
  value       = azurerm_postgresql_flexible_server_database.storm.name
}

output "server_name" {
  description = "Nom du serveur PostgreSQL Azure"
  value       = azurerm_postgresql_flexible_server.main.name
}
