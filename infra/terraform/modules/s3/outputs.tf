# ==============================================================================
# OUTPUTS DU MODULE S3
# ==============================================================================

output "avatars_bucket_name" {
  description = "Nom du bucket avatars"
  value       = aws_s3_bucket.avatars.id
}

output "avatars_bucket_arn" {
  description = "ARN du bucket avatars"
  value       = aws_s3_bucket.avatars.arn
}

output "media_bucket_name" {
  description = "Nom du bucket media"
  value       = aws_s3_bucket.media.id
}

output "media_bucket_arn" {
  description = "ARN du bucket media"
  value       = aws_s3_bucket.media.arn
}
