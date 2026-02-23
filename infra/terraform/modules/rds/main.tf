# ==============================================================================
# MODULE RDS - PostgreSQL managé par AWS
# ==============================================================================
#
# RDS (Relational Database Service) c'est PostgreSQL géré par AWS :
# - Backups automatiques
# - Mises à jour de sécurité
# - Haute disponibilité (Multi-AZ en prod)
# - Monitoring intégré
#
# Tu n'as plus besoin de gérer un serveur PostgreSQL toi-même !
#
# ==============================================================================

# ------------------------------------------------------------------------------
# Subnet Group
# ------------------------------------------------------------------------------
# RDS doit savoir dans quels subnets il peut être déployé.
# On utilise les subnets PRIVÉS (pas d'accès internet direct).

resource "aws_db_subnet_group" "main" {
  name        = "${var.project_name}-db-subnet-group"
  description = "Subnet group pour RDS"
  subnet_ids  = var.private_subnet_ids  # Subnets privés

  tags = {
    Name        = "${var.project_name}-db-subnet-group"
    Environment = var.environment
  }
}

# ------------------------------------------------------------------------------
# RDS PostgreSQL Instance
# ------------------------------------------------------------------------------
# L'instance de base de données elle-même.

resource "aws_db_instance" "main" {
  identifier = "${var.project_name}-db"

  # Configuration PostgreSQL
  engine               = "postgres"
  engine_version       = var.postgres_version     # 15.4
  instance_class       = var.instance_class       # db.t3.micro (dev)
  allocated_storage    = var.allocated_storage    # 20 Go
  max_allocated_storage = var.max_allocated_storage # Auto-scaling jusqu'à 100 Go

  # Credentials
  db_name  = var.database_name  # storm
  username = var.master_username
  password = var.master_password

  # Réseau
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [var.security_group_id]
  publicly_accessible    = false  # JAMAIS accessible depuis internet

  # Backups
  backup_retention_period = var.backup_retention_days  # 7 jours
  backup_window          = "03:00-04:00"              # Backup à 3h du matin
  maintenance_window     = "Mon:04:00-Mon:05:00"      # Maintenance le lundi

  # Autres options
  multi_az               = var.multi_az              # false en dev, true en prod
  storage_type           = "gp3"                     # SSD nouvelle génération
  storage_encrypted      = true                      # Chiffrement au repos
  skip_final_snapshot    = var.environment == "dev" # Pas de snapshot final en dev
  deletion_protection    = var.environment == "prod" # Protection en prod

  # Performance Insights (monitoring avancé)
  performance_insights_enabled = var.environment == "prod"

  tags = {
    Name        = "${var.project_name}-db"
    Environment = var.environment
  }
}
