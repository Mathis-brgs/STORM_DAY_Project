# ==============================================================================
# VARIABLES DU MODULE S3
# ==============================================================================

variable "project_name" {
  description = "Nom du projet"
  type        = string
}

variable "environment" {
  description = "Environnement (dev, staging, prod)"
  type        = string
}

variable "cors_allowed_origins" {
  description = "Origines autorisées pour CORS (URLs du frontend)"
  type        = list(string)
  default     = ["*"]  # A restreindre en prod !
}
