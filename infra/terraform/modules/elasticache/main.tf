# ==============================================================================
# MODULE ELASTICACHE - Redis managé par AWS
# ==============================================================================
#
# ElastiCache c'est Redis (ou Memcached) géré par AWS.
#
# Utilisations dans STORM :
# - Cache des données fréquemment lues
# - Sessions utilisateurs
# - Rate limiting
# - Pub/Sub pour les notifications temps réel
#
# ==============================================================================

# ------------------------------------------------------------------------------
# Subnet Group
# ------------------------------------------------------------------------------
# ElastiCache doit savoir dans quels subnets il peut être déployé.

resource "aws_elasticache_subnet_group" "main" {
  name        = "${var.project_name}-redis-subnet-group"
  description = "Subnet group pour ElastiCache Redis"
  subnet_ids  = var.private_subnet_ids

  tags = {
    Name        = "${var.project_name}-redis-subnet-group"
    Environment = var.environment
  }
}

# ------------------------------------------------------------------------------
# ElastiCache Redis Cluster
# ------------------------------------------------------------------------------
# Un cluster Redis avec réplication (optionnel en dev).

resource "aws_elasticache_replication_group" "main" {
  replication_group_id = "${var.project_name}-redis"
  description          = "Redis cluster pour ${var.project_name}"

  # Configuration Redis
  engine               = "redis"
  engine_version       = var.redis_version       # 7.0
  node_type            = var.node_type           # cache.t3.micro
  num_cache_clusters   = var.num_cache_clusters  # 1 en dev, 2+ en prod
  port                 = 6379

  # Réseau
  subnet_group_name    = aws_elasticache_subnet_group.main.name
  security_group_ids   = [var.security_group_id]

  # Paramètres
  parameter_group_name = "default.redis7"

  # Haute disponibilité (prod seulement)
  automatic_failover_enabled = var.num_cache_clusters > 1
  multi_az_enabled          = var.num_cache_clusters > 1

  # Maintenance
  maintenance_window = "mon:03:00-mon:04:00"
  snapshot_window    = "02:00-03:00"
  snapshot_retention_limit = var.environment == "prod" ? 7 : 0

  # Chiffrement
  at_rest_encryption_enabled = true
  transit_encryption_enabled = false  # true nécessite TLS dans le code

  tags = {
    Name        = "${var.project_name}-redis"
    Environment = var.environment
  }
}
