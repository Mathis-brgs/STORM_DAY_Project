# ==============================================================================
# OUTPUTS DU MODULE VPC
# ==============================================================================
# Ces valeurs sont exportées pour être utilisées par d'autres modules.
# Exemple : le module RDS a besoin des IDs des subnets privés.

output "vpc_id" {
  description = "ID du VPC créé"
  value       = aws_vpc.main.id
}

output "vpc_cidr" {
  description = "CIDR block du VPC"
  value       = aws_vpc.main.cidr_block
}

output "public_subnet_ids" {
  description = "Liste des IDs des subnets publics"
  value       = aws_subnet.public[*].id
}

output "private_subnet_ids" {
  description = "Liste des IDs des subnets privés"
  value       = aws_subnet.private[*].id
}

output "nat_gateway_id" {
  description = "ID du NAT Gateway"
  value       = aws_nat_gateway.main.id
}
