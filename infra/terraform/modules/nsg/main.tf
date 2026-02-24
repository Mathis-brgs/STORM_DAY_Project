# ==============================================================================
# MODULE NSG - Network Security Groups Azure
# ==============================================================================
#
# Les NSG sont les pare-feu Azure (équivalent des Security Groups AWS).
# Contrôlent le trafic entrant et sortant de chaque subnet.
#
# ==============================================================================

# NSG pour les applications (pods AKS)
resource "azurerm_network_security_group" "app" {
  name                = "${var.project_name}-nsg-app-${var.environment}"
  location            = var.location
  resource_group_name = var.resource_group_name

  # Port 3000 - NestJS (user-service)
  security_rule {
    name                       = "allow-nestjs"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "3000"
    source_address_prefix      = var.vnet_cidr
    destination_address_prefix = "*"
  }

  # Port 8080 - Go Gateway
  security_rule {
    name                       = "allow-gateway"
    priority                   = 110
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "8080"
    source_address_prefix      = var.vnet_cidr
    destination_address_prefix = "*"
  }

  # Port 4222 - NATS
  security_rule {
    name                       = "allow-nats"
    priority                   = 120
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "4222"
    source_address_prefix      = var.vnet_cidr
    destination_address_prefix = "*"
  }

  tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}

# NSG pour PostgreSQL
resource "azurerm_network_security_group" "postgresql" {
  name                = "${var.project_name}-nsg-postgresql-${var.environment}"
  location            = var.location
  resource_group_name = var.resource_group_name

  # Port 5432 - PostgreSQL, seulement depuis le VNet
  security_rule {
    name                       = "allow-postgresql-from-vnet"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "5432"
    source_address_prefix      = var.vnet_cidr
    destination_address_prefix = "*"
  }

  tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}

# NSG pour Redis
resource "azurerm_network_security_group" "redis" {
  name                = "${var.project_name}-nsg-redis-${var.environment}"
  location            = var.location
  resource_group_name = var.resource_group_name

  # Port 6379 - Redis, seulement depuis le VNet
  security_rule {
    name                       = "allow-redis-from-vnet"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "6379"
    source_address_prefix      = var.vnet_cidr
    destination_address_prefix = "*"
  }

  tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}

# NSG pour le Load Balancer (point d'entrée public)
resource "azurerm_network_security_group" "lb" {
  name                = "${var.project_name}-nsg-lb-${var.environment}"
  location            = var.location
  resource_group_name = var.resource_group_name

  # Port 80 - HTTP
  security_rule {
    name                       = "allow-http"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "80"
    source_address_prefix      = "Internet"
    destination_address_prefix = "*"
  }

  # Port 443 - HTTPS
  security_rule {
    name                       = "allow-https"
    priority                   = 110
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "443"
    source_address_prefix      = "Internet"
    destination_address_prefix = "*"
  }

  tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}
