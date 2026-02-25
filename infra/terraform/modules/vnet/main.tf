# ==============================================================================
# MODULE VNET - Azure Virtual Network
# ==============================================================================
#
# Ce module crée le réseau privé Azure avec :
# - 1 VNet (réseau privé Azure)
# - 2 subnets publics (pour l'AKS Load Balancer)
# - 1 subnet privé (pour PostgreSQL Flexible Server - délégation obligatoire)
# - 1 zone DNS privée pour PostgreSQL
#
# ==============================================================================

resource "azurerm_virtual_network" "main" {
  name                = "${var.project_name}-vnet-${var.environment}"
  address_space       = [var.vnet_cidr]
  location            = var.location
  resource_group_name = var.resource_group_name

  tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}

# Subnets publics (AKS nodes, Load Balancer)
resource "azurerm_subnet" "public" {
  count                = length(var.public_subnet_cidrs)
  name                 = "${var.project_name}-public-${count.index}-${var.environment}"
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = [var.public_subnet_cidrs[count.index]]
}

# Subnet privé pour PostgreSQL Flexible Server
# Azure exige une délégation explicite sur ce subnet
resource "azurerm_subnet" "private_postgresql" {
  name                 = "${var.project_name}-private-postgresql-${var.environment}"
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = [var.private_subnet_cidr]

  delegation {
    name = "postgresql-delegation"
    service_delegation {
      name = "Microsoft.DBforPostgreSQL/flexibleServers"
      actions = [
        "Microsoft.Network/virtualNetworks/subnets/join/action",
      ]
    }
  }
}

# Zone DNS privée pour PostgreSQL Flexible Server
# Azure génère un nom DNS interne : <server>.private.postgres.database.azure.com
resource "azurerm_private_dns_zone" "postgresql" {
  name                = "privatelink.postgres.database.azure.com"
  resource_group_name = var.resource_group_name

  tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}

# Lier la zone DNS privée au VNet
resource "azurerm_private_dns_zone_virtual_network_link" "postgresql" {
  name                  = "${var.project_name}-postgresql-dns-link-${var.environment}"
  private_dns_zone_name = azurerm_private_dns_zone.postgresql.name
  resource_group_name   = var.resource_group_name
  virtual_network_id    = azurerm_virtual_network.main.id
  registration_enabled  = false
}
