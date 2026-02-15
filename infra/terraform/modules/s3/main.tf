# ==============================================================================
# MODULE S3 - Stockage de fichiers
# ==============================================================================
#
# S3 (Simple Storage Service) pour stocker les fichiers :
# - Avatars des utilisateurs
# - Médias uploadés (images, vidéos, documents)
#
# S3 est hors du VPC, accessible via URLs publiques ou signées.
#
# ==============================================================================

# ------------------------------------------------------------------------------
# Bucket Avatars
# ------------------------------------------------------------------------------
# Photos de profil des utilisateurs

resource "aws_s3_bucket" "avatars" {
  bucket = "${var.project_name}-avatars-${var.environment}"

  tags = {
    Name        = "${var.project_name}-avatars"
    Environment = var.environment
    Purpose     = "User avatars"
  }
}

# Bloquer l'accès public par défaut (sécurité)
resource "aws_s3_bucket_public_access_block" "avatars" {
  bucket = aws_s3_bucket.avatars.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Activer le versioning (garde les anciennes versions)
resource "aws_s3_bucket_versioning" "avatars" {
  bucket = aws_s3_bucket.avatars.id

  versioning_configuration {
    status = "Enabled"
  }
}

# Chiffrement au repos
resource "aws_s3_bucket_server_side_encryption_configuration" "avatars" {
  bucket = aws_s3_bucket.avatars.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# CORS pour permettre l'upload depuis le frontend
resource "aws_s3_bucket_cors_configuration" "avatars" {
  bucket = aws_s3_bucket.avatars.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "PUT", "POST"]
    allowed_origins = var.cors_allowed_origins
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }
}

# ------------------------------------------------------------------------------
# Bucket Médias
# ------------------------------------------------------------------------------
# Fichiers uploadés par les utilisateurs (images, vidéos, documents)

resource "aws_s3_bucket" "media" {
  bucket = "${var.project_name}-media-${var.environment}"

  tags = {
    Name        = "${var.project_name}-media"
    Environment = var.environment
    Purpose     = "User uploaded media"
  }
}

resource "aws_s3_bucket_public_access_block" "media" {
  bucket = aws_s3_bucket.media.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_versioning" "media" {
  bucket = aws_s3_bucket.media.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "media" {
  bucket = aws_s3_bucket.media.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_cors_configuration" "media" {
  bucket = aws_s3_bucket.media.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "PUT", "POST", "DELETE"]
    allowed_origins = var.cors_allowed_origins
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }
}

# Lifecycle : supprimer les fichiers incomplets après 7 jours
resource "aws_s3_bucket_lifecycle_configuration" "media" {
  bucket = aws_s3_bucket.media.id

  rule {
    id     = "cleanup-incomplete-uploads"
    status = "Enabled"

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }
}
