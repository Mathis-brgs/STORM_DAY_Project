# ==============================================================================
# MODULE IAM - Permissions AWS
# ==============================================================================
#
# IAM (Identity and Access Management) définit QUI peut faire QUOI sur AWS.
#
# Concepts :
# - Role : identité que peuvent assumer les services (EKS, Lambda, etc.)
# - Policy : ensemble de permissions (lire S3, écrire RDS, etc.)
#
# Principe : Moindre privilège - ne donner que les permissions nécessaires
#
# ==============================================================================

# ------------------------------------------------------------------------------
# Role pour les applications (EKS pods)
# ------------------------------------------------------------------------------
# Ce rôle sera assumé par les pods EKS pour accéder aux ressources AWS.

resource "aws_iam_role" "app" {
  name = "${var.project_name}-app-role"

  # Trust policy : qui peut assumer ce rôle
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "eks.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name        = "${var.project_name}-app-role"
    Environment = var.environment
  }
}

# ------------------------------------------------------------------------------
# Policy S3 - Accès aux buckets
# ------------------------------------------------------------------------------
# Permet aux apps de lire/écrire dans les buckets S3.

resource "aws_iam_policy" "s3_access" {
  name        = "${var.project_name}-s3-access"
  description = "Permet l'accès aux buckets S3 du projet"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "ListBuckets"
        Effect = "Allow"
        Action = [
          "s3:ListBucket",
          "s3:GetBucketLocation"
        ]
        Resource = var.s3_bucket_arns
      },
      {
        Sid    = "ReadWriteObjects"
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:GetObjectVersion"
        ]
        Resource = [for arn in var.s3_bucket_arns : "${arn}/*"]
      }
    ]
  })

  tags = {
    Name        = "${var.project_name}-s3-access"
    Environment = var.environment
  }
}

# Attacher la policy au rôle
resource "aws_iam_role_policy_attachment" "app_s3" {
  role       = aws_iam_role.app.name
  policy_arn = aws_iam_policy.s3_access.arn
}

# ------------------------------------------------------------------------------
# Policy Secrets Manager (optionnel)
# ------------------------------------------------------------------------------
# Pour stocker les secrets (DB password, API keys) de façon sécurisée.

resource "aws_iam_policy" "secrets_access" {
  name        = "${var.project_name}-secrets-access"
  description = "Permet l'accès aux secrets du projet"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "ReadSecrets"
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret"
        ]
        Resource = "arn:aws:secretsmanager:*:*:secret:${var.project_name}/*"
      }
    ]
  })

  tags = {
    Name        = "${var.project_name}-secrets-access"
    Environment = var.environment
  }
}

resource "aws_iam_role_policy_attachment" "app_secrets" {
  role       = aws_iam_role.app.name
  policy_arn = aws_iam_policy.secrets_access.arn
}

# ------------------------------------------------------------------------------
# Instance Profile (pour EC2/EKS)
# ------------------------------------------------------------------------------
# Permet d'attacher le rôle à des instances EC2 ou nodes EKS.

resource "aws_iam_instance_profile" "app" {
  name = "${var.project_name}-app-profile"
  role = aws_iam_role.app.name

  tags = {
    Name        = "${var.project_name}-app-profile"
    Environment = var.environment
  }
}
