# ==============================================================================
# OUTPUTS DU MODULE SECURITY GROUPS
# ==============================================================================

output "app_security_group_id" {
  description = "ID du security group pour les applications"
  value       = aws_security_group.app.id
}

output "rds_security_group_id" {
  description = "ID du security group pour RDS"
  value       = aws_security_group.rds.id
}

output "redis_security_group_id" {
  description = "ID du security group pour Redis"
  value       = aws_security_group.redis.id
}

output "alb_security_group_id" {
  description = "ID du security group pour ALB"
  value       = aws_security_group.alb.id
}
