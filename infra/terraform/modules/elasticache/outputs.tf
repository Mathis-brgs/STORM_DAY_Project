# ==============================================================================
# OUTPUTS DU MODULE ELASTICACHE
# ==============================================================================

output "redis_endpoint" {
  description = "Endpoint de connexion Redis (primary)"
  value       = aws_elasticache_replication_group.main.primary_endpoint_address
}

output "redis_port" {
  description = "Port Redis"
  value       = aws_elasticache_replication_group.main.port
}

output "redis_connection_string" {
  description = "String de connexion Redis complète"
  value       = "redis://${aws_elasticache_replication_group.main.primary_endpoint_address}:${aws_elasticache_replication_group.main.port}"
}
