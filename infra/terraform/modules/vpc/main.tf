# ==============================================================================
# MODULE VPC - Réseau principal AWS
# ==============================================================================
#
# Ce module crée ton réseau privé dans AWS avec :
# - 1 VPC (ton réseau isolé)
# - 2 subnets publics (pour Load Balancer, NAT)
# - 2 subnets privés (pour RDS, Redis, EKS - plus sécurisé)
# - 1 Internet Gateway (permet l'accès internet aux subnets publics)
# - 1 NAT Gateway (permet aux subnets privés d'accéder à internet)
# - Tables de routage (définissent où va le trafic)
#
# ==============================================================================

# ------------------------------------------------------------------------------
# VPC Principal
# ------------------------------------------------------------------------------
# CIDR 10.0.0.0/16 = 65,536 adresses IP disponibles
# C'est ton espace réseau privé, isolé des autres clients AWS

resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr           # 10.0.0.0/16
  enable_dns_hostnames = true                   # Noms DNS pour les instances
  enable_dns_support   = true                   # Support DNS activé

  tags = {
    Name        = "${var.project_name}-vpc"
    Environment = var.environment
    Project     = var.project_name
  }
}

# ------------------------------------------------------------------------------
# Internet Gateway
# ------------------------------------------------------------------------------
# Passerelle vers internet. Sans ça, rien ne peut accéder à internet.
# Attaché au VPC, utilisé par les subnets PUBLICS.

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name        = "${var.project_name}-igw"
    Environment = var.environment
  }
}

# ------------------------------------------------------------------------------
# Subnets Publics (2 zones pour haute disponibilité)
# ------------------------------------------------------------------------------
# Ressources accessibles depuis internet : Load Balancer, Bastion, NAT Gateway
# map_public_ip_on_launch = true → les instances reçoivent une IP publique

resource "aws_subnet" "public" {
  count = length(var.availability_zones)

  vpc_id                  = aws_vpc.main.id
  cidr_block              = cidrsubnet(var.vpc_cidr, 8, count.index + 1)  # 10.0.1.0/24, 10.0.2.0/24
  availability_zone       = var.availability_zones[count.index]
  map_public_ip_on_launch = true

  tags = {
    Name        = "${var.project_name}-public-${var.availability_zones[count.index]}"
    Environment = var.environment
    Type        = "public"
    # Tags pour EKS (si tu l'utilises plus tard)
    "kubernetes.io/role/elb" = "1"
  }
}

# ------------------------------------------------------------------------------
# Subnets Privés (2 zones pour haute disponibilité)
# ------------------------------------------------------------------------------
# Ressources NON accessibles depuis internet : RDS, Redis, workers EKS
# Plus sécurisé car pas d'accès direct depuis internet

resource "aws_subnet" "private" {
  count = length(var.availability_zones)

  vpc_id            = aws_vpc.main.id
  cidr_block        = cidrsubnet(var.vpc_cidr, 8, count.index + 10)  # 10.0.10.0/24, 10.0.11.0/24
  availability_zone = var.availability_zones[count.index]

  tags = {
    Name        = "${var.project_name}-private-${var.availability_zones[count.index]}"
    Environment = var.environment
    Type        = "private"
    # Tags pour EKS
    "kubernetes.io/role/internal-elb" = "1"
  }
}

# ------------------------------------------------------------------------------
# Elastic IP pour NAT Gateway
# ------------------------------------------------------------------------------
# IP publique fixe pour le NAT Gateway

resource "aws_eip" "nat" {
  domain = "vpc"

  tags = {
    Name        = "${var.project_name}-nat-eip"
    Environment = var.environment
  }

  depends_on = [aws_internet_gateway.main]
}

# ------------------------------------------------------------------------------
# NAT Gateway
# ------------------------------------------------------------------------------
# Permet aux ressources PRIVÉES d'accéder à internet (mises à jour, API externes)
# SANS être accessibles depuis internet. Trafic sortant uniquement.
#
# Note: Coûte ~$35/mois. Pour économiser en dev, tu peux commenter ce bloc.

resource "aws_nat_gateway" "main" {
  allocation_id = aws_eip.nat.id
  subnet_id     = aws_subnet.public[0].id  # Doit être dans un subnet PUBLIC

  tags = {
    Name        = "${var.project_name}-nat"
    Environment = var.environment
  }

  depends_on = [aws_internet_gateway.main]
}

# ------------------------------------------------------------------------------
# Table de routage - Publique
# ------------------------------------------------------------------------------
# Règles de routage pour les subnets publics :
# - Trafic vers 10.0.0.0/16 → reste dans le VPC (automatique)
# - Trafic vers 0.0.0.0/0 (tout le reste) → Internet Gateway

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name        = "${var.project_name}-public-rt"
    Environment = var.environment
  }
}

# Associer les subnets publics à cette table
resource "aws_route_table_association" "public" {
  count          = length(aws_subnet.public)
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

# ------------------------------------------------------------------------------
# Table de routage - Privée
# ------------------------------------------------------------------------------
# Règles de routage pour les subnets privés :
# - Trafic vers 10.0.0.0/16 → reste dans le VPC (automatique)
# - Trafic vers 0.0.0.0/0 → NAT Gateway (accès internet sortant seulement)

resource "aws_route_table" "private" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.main.id
  }

  tags = {
    Name        = "${var.project_name}-private-rt"
    Environment = var.environment
  }
}

# Associer les subnets privés à cette table
resource "aws_route_table_association" "private" {
  count          = length(aws_subnet.private)
  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private.id
}
