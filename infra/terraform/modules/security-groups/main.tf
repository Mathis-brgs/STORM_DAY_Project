# ==============================================================================
# MODULE SECURITY GROUPS - Pare-feu AWS
# ==============================================================================
#
# Les Security Groups sont des pare-feu virtuels qui contrôlent :
# - INGRESS : trafic ENTRANT (qui peut se connecter)
# - EGRESS : trafic SORTANT (vers où on peut se connecter)
#
# Principe : N'ouvrir que les ports strictement nécessaires
#
# ==============================================================================

# ------------------------------------------------------------------------------
# Security Group - Applications (EKS / Services)
# ------------------------------------------------------------------------------
# Pour tes microservices (user-service, gateway, etc.)

resource "aws_security_group" "app" {
  name        = "${var.project_name}-app-sg"
  description = "Security group pour les applications"
  vpc_id      = var.vpc_id

  # Port 3000 - NestJS (user-service)
  ingress {
    description = "NestJS services"
    from_port   = 3000
    to_port     = 3000
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]  # Seulement depuis le VPC
  }

  # Port 8080 - Go Gateway
  ingress {
    description = "Go gateway"
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]
  }

  # Port 4222 - NATS (messaging inter-services)
  ingress {
    description = "NATS messaging"
    from_port   = 4222
    to_port     = 4222
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]
  }

  # Trafic sortant : tout autorisé
  egress {
    description = "Tout le trafic sortant"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"           # -1 = tous les protocoles
    cidr_blocks = ["0.0.0.0/0"]  # Vers internet
  }

  tags = {
    Name        = "${var.project_name}-app-sg"
    Environment = var.environment
  }
}

# ------------------------------------------------------------------------------
# Security Group - RDS PostgreSQL
# ------------------------------------------------------------------------------
# Seules les applications peuvent accéder à la base de données.
# PostgreSQL n'est JAMAIS accessible depuis internet.

resource "aws_security_group" "rds" {
  name        = "${var.project_name}-rds-sg"
  description = "Security group pour RDS PostgreSQL"
  vpc_id      = var.vpc_id

  # Port 5432 - PostgreSQL, UNIQUEMENT depuis les apps
  ingress {
    description     = "PostgreSQL depuis apps"
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.app.id]  # Référence au SG app
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${var.project_name}-rds-sg"
    Environment = var.environment
  }
}

# ------------------------------------------------------------------------------
# Security Group - ElastiCache Redis
# ------------------------------------------------------------------------------
# Redis pour le cache et les sessions

resource "aws_security_group" "redis" {
  name        = "${var.project_name}-redis-sg"
  description = "Security group pour ElastiCache Redis"
  vpc_id      = var.vpc_id

  # Port 6379 - Redis, UNIQUEMENT depuis les apps
  ingress {
    description     = "Redis depuis apps"
    from_port       = 6379
    to_port         = 6379
    protocol        = "tcp"
    security_groups = [aws_security_group.app.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${var.project_name}-redis-sg"
    Environment = var.environment
  }
}

# ------------------------------------------------------------------------------
# Security Group - ALB (Application Load Balancer)
# ------------------------------------------------------------------------------
# Point d'entrée public de ton application.
# Seul composant accessible depuis internet.

resource "aws_security_group" "alb" {
  name        = "${var.project_name}-alb-sg"
  description = "Security group pour Application Load Balancer"
  vpc_id      = var.vpc_id

  # Port 80 - HTTP (redirige vers HTTPS)
  ingress {
    description = "HTTP depuis internet"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]  # Tout internet
  }

  # Port 443 - HTTPS
  ingress {
    description = "HTTPS depuis internet"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${var.project_name}-alb-sg"
    Environment = var.environment
  }
}
