output "postgresql_fqdn" {
  description = "FQDN du serveur PostgreSQL (host de connexion)"
  value       = azurerm_postgresql_flexible_server.main.fqdn
}

output "postgresql_host" {
  description = "Hostname du serveur PostgreSQL"
  value       = azurerm_postgresql_flexible_server.main.fqdn
}

output "user_db_name" {
  description = "Nom de la base de données users"
  value       = azurerm_postgresql_flexible_server_database.user_db.name
}

output "message_db_name" {
  description = "Nom de la base de données messages"
  value       = azurerm_postgresql_flexible_server_database.message_db.name
}

output "server_name" {
  description = "Nom du serveur PostgreSQL Azure"
  value       = azurerm_postgresql_flexible_server.main.name
}
