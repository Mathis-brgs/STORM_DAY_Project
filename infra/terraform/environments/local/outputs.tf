# ==============================================================================
# OUTPUTS - Environnement Local
# ==============================================================================

output "s3_avatars_bucket" {
  description = "Nom du bucket avatars"
  value       = aws_s3_bucket.avatars.id
}

output "s3_media_bucket" {
  description = "Nom du bucket media"
  value       = aws_s3_bucket.media.id
}

output "app_role_arn" {
  description = "ARN du rôle IAM"
  value       = aws_iam_role.app.arn
}

output "s3_endpoint" {
  description = "Endpoint S3 LocalStack"
  value       = "http://localhost:4566"
}
