# ==============================================================================
# MODULE REDIS - Azure Cache for Redis
# ==============================================================================
#
# Azure Cache for Redis — service Redis managé par Azure.
#
# Utilisations dans STORM :
# - Cache des données fréquemment lues
# - Sessions utilisateurs
# - Rate limiting
#
# Tiers :
# - Basic C0 (250MB) : ~$16/mois — suffisant pour dev
# - Standard C0      : ~$52/mois — haute dispo (prod)
#
# ==============================================================================

resource "azurerm_redis_cache" "main" {
  name                = "${var.project_name}-redis-${var.environment}"
  location            = var.location
  resource_group_name = var.resource_group_name

  # Taille (C0 = 250 MB, C1 = 1 GB, etc.)
  capacity = var.capacity  # 0 = C0

  # Famille : C = Standard/Basic, P = Premium
  family = var.family  # "C"

  # SKU : Basic (dev), Standard (prod HA), Premium (clustering)
  sku_name = var.sku_name  # "Basic"

  # Redis version
  redis_version = var.redis_version  # "7"

  # Pas de TLS obligatoire (le code n'est pas configuré pour)
  enable_non_ssl_port = true
  minimum_tls_version = "1.0"

  redis_configuration {
    maxmemory_policy = "allkeys-lru"
  }

  tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}
