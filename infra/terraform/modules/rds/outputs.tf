# ==============================================================================
# OUTPUTS DU MODULE RDS
# ==============================================================================

output "db_instance_endpoint" {
  description = "Endpoint de connexion à la base (host:port)"
  value       = aws_db_instance.main.endpoint
}

output "db_instance_address" {
  description = "Adresse DNS de la base"
  value       = aws_db_instance.main.address
}

output "db_instance_port" {
  description = "Port PostgreSQL"
  value       = aws_db_instance.main.port
}

output "db_name" {
  description = "Nom de la base de données"
  value       = aws_db_instance.main.db_name
}

output "db_instance_id" {
  description = "ID de l'instance RDS"
  value       = aws_db_instance.main.id
}
