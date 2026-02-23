# ==============================================================================
# VARIABLES DU MODULE VPC
# ==============================================================================
# Les variables permettent de réutiliser ce module pour dev, staging, prod
# avec des valeurs différentes.

variable "project_name" {
  description = "Nom du projet (utilisé pour nommer les ressources)"
  type        = string
}

variable "environment" {
  description = "Environnement : dev, staging, prod"
  type        = string
}

variable "vpc_cidr" {
  description = "Plage d'adresses IP du VPC (ex: 10.0.0.0/16 = 65536 IPs)"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "Zones de disponibilité AWS (2 minimum pour haute dispo)"
  type        = list(string)
  default     = ["eu-west-3a", "eu-west-3b"]  # Paris
}
