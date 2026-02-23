# ==============================================================================
# VARIABLES DU MODULE IAM
# ==============================================================================

variable "project_name" {
  description = "Nom du projet"
  type        = string
}

variable "environment" {
  description = "Environnement (dev, staging, prod)"
  type        = string
}

variable "s3_bucket_arns" {
  description = "ARNs des buckets S3 auxquels donner accès"
  type        = list(string)
  default     = []
}
