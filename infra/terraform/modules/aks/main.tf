# ==============================================================================
# MODULE AKS - Azure Kubernetes Service
# ==============================================================================
#
# Cluster AKS pour les services STORM.
# - 2 nodes Standard_B2s (~$60/mois)
# - Kubernetes 1.28
# - Réseau : subnet AKS dédié
# - Managed Identity system-assigned
# - ACR attach : pull images sans secret supplémentaire
# - OIDC + Workload Identity activé pour Workload Identity future
# - Autoscaler : min 1, max 3
#
# ==============================================================================

resource "azurerm_kubernetes_cluster" "main" {
  name                = "${var.project_name}-aks-${var.environment}"
  location            = var.location
  resource_group_name = var.resource_group_name
  dns_prefix          = "${var.project_name}-${var.environment}"
  kubernetes_version  = var.kubernetes_version != "" ? var.kubernetes_version : null

  # Node pool par défaut
  default_node_pool {
    name                = "default"
    vm_size             = var.vm_size
    vnet_subnet_id      = var.aks_subnet_id

    # Autoscaler activé
    enable_auto_scaling = true
    min_count            = var.min_node_count
    max_count            = var.max_node_count

    # Upgrade automatique des nodes désactivé (contrôle manuel en dev)
    upgrade_settings {
      max_surge = "10%"
    }
  }

  # Managed Identity system-assigned (plus simple que user-assigned pour dev)
  identity {
    type = "SystemAssigned"
  }

  # Réseau : kubenet (plus simple, suffisant pour dev)
  network_profile {
    network_plugin    = "kubenet"
    load_balancer_sku = "standard"
    service_cidr      = "10.96.0.0/16"
    dns_service_ip    = "10.96.0.10"
  }

  # OIDC Issuer + Workload Identity (pour Workload Identity future si besoin)
  oidc_issuer_enabled       = true
  workload_identity_enabled = true

  tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}

# Attacher l'ACR au cluster AKS
# Permet au kubelet de pull les images sans secret Docker supplémentaire
resource "azurerm_role_assignment" "aks_acr_pull" {
  principal_id                     = azurerm_kubernetes_cluster.main.kubelet_identity[0].object_id
  role_definition_name             = "AcrPull"
  scope                            = var.acr_id
  skip_service_principal_aad_check = true
}
