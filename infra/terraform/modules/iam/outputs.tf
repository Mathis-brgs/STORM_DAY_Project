# ==============================================================================
# OUTPUTS DU MODULE IAM
# ==============================================================================

output "app_role_arn" {
  description = "ARN du rôle pour les applications"
  value       = aws_iam_role.app.arn
}

output "app_role_name" {
  description = "Nom du rôle pour les applications"
  value       = aws_iam_role.app.name
}

output "app_instance_profile_arn" {
  description = "ARN de l'instance profile"
  value       = aws_iam_instance_profile.app.arn
}

output "app_instance_profile_name" {
  description = "Nom de l'instance profile"
  value       = aws_iam_instance_profile.app.name
}
