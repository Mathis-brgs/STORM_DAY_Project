variable "project_name" {
  description = "Nom du projet"
  type        = string
}

variable "environment" {
  description = "Environnement : dev, staging, prod"
  type        = string
}

variable "location" {
  description = "Région Azure"
  type        = string
}

variable "resource_group_name" {
  description = "Nom du resource group Azure"
  type        = string
}

variable "aks_subnet_id" {
  description = "ID du subnet AKS dans le VNet"
  type        = string
}

variable "acr_id" {
  description = "ID de l'ACR à attacher (pour AcrPull role)"
  type        = string
}

variable "kubernetes_version" {
  description = "Version Kubernetes"
  type        = string
  default     = "1.28"
}

variable "node_count" {
  description = "Nombre de nodes initial"
  type        = number
  default     = 2
}

variable "vm_size" {
  description = "Taille des VMs nodes (Standard_B2s ~$30/mois/node)"
  type        = string
  default     = "Standard_B2s"
}

variable "min_node_count" {
  description = "Nombre minimum de nodes (autoscaler)"
  type        = number
  default     = 1
}

variable "max_node_count" {
  description = "Nombre maximum de nodes (autoscaler)"
  type        = number
  default     = 3
}
