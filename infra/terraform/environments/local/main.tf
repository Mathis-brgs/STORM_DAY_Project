# ==============================================================================
# ENVIRONNEMENT LOCAL - Terraform avec LocalStack
# ==============================================================================
#
# Test simplifié avec les services supportés par LocalStack Community :
# - S3 (buckets)
# - IAM (roles, policies)
#
# ==============================================================================

terraform {
  required_version = ">= 1.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# ------------------------------------------------------------------------------
# Provider AWS configuré pour LocalStack
# ------------------------------------------------------------------------------

provider "aws" {
  region                      = "eu-west-3"
  access_key                  = "test"
  secret_key                  = "test"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  endpoints {
    s3             = "http://localhost:4566"
    iam            = "http://localhost:4566"
    sts            = "http://localhost:4566"
  }

  # Nécessaire pour S3 avec LocalStack
  s3_use_path_style = true
}

# ------------------------------------------------------------------------------
# Variables locales
# ------------------------------------------------------------------------------

locals {
  project_name = "storm"
  environment  = "local"
}

# ------------------------------------------------------------------------------
# S3 Buckets
# ------------------------------------------------------------------------------

resource "aws_s3_bucket" "avatars" {
  bucket = "${local.project_name}-avatars-${local.environment}"

  tags = {
    Name        = "${local.project_name}-avatars"
    Environment = local.environment
    Purpose     = "User avatars"
  }
}

resource "aws_s3_bucket" "media" {
  bucket = "${local.project_name}-media-${local.environment}"

  tags = {
    Name        = "${local.project_name}-media"
    Environment = local.environment
    Purpose     = "User uploaded media"
  }
}

# ------------------------------------------------------------------------------
# IAM Role pour les applications
# ------------------------------------------------------------------------------

resource "aws_iam_role" "app" {
  name = "${local.project_name}-app-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name        = "${local.project_name}-app-role"
    Environment = local.environment
  }
}

resource "aws_iam_policy" "s3_access" {
  name        = "${local.project_name}-s3-access"
  description = "Accès aux buckets S3"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.avatars.arn,
          "${aws_s3_bucket.avatars.arn}/*",
          aws_s3_bucket.media.arn,
          "${aws_s3_bucket.media.arn}/*"
        ]
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "app_s3" {
  role       = aws_iam_role.app.name
  policy_arn = aws_iam_policy.s3_access.arn
}
