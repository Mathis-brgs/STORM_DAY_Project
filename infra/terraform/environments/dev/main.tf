# ==============================================================================
# ENVIRONNEMENT DEV - Point d'entrée Terraform Azure
# ==============================================================================
#
# Pour déployer :
#   cd infra/terraform/environments/dev
#   tofu init      # Première fois seulement
#   tofu plan      # Voir ce qui va être créé
#   tofu apply     # Créer l'infrastructure
#
# Prérequis :
#   az login
#   az account set --subscription <SUBSCRIPTION_ID>
#
# ==============================================================================

terraform {
  required_version = ">= 1.0.0"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }

  # Backend Azure Blob Storage pour stocker l'état Terraform (à activer plus tard)
  # backend "azurerm" {
  #   resource_group_name  = "storm-terraform-state"
  #   storage_account_name = "stormtfstate"
  #   container_name       = "tfstate"
  #   key                  = "dev/terraform.tfstate"
  # }
}

provider "azurerm" {
  features {}
  subscription_id = var.subscription_id
}

# ------------------------------------------------------------------------------
# Resource Group
# ------------------------------------------------------------------------------

resource "azurerm_resource_group" "main" {
  name     = "${var.project_name}-${var.environment}"
  location = var.location

  tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}

# ------------------------------------------------------------------------------
# Module VNet
# ------------------------------------------------------------------------------

module "vnet" {
  source = "../../modules/vnet"

  project_name        = var.project_name
  environment         = var.environment
  location            = var.location
  resource_group_name = azurerm_resource_group.main.name
  vnet_cidr           = var.vnet_cidr
  public_subnet_cidrs = var.public_subnet_cidrs
  private_subnet_cidr = var.private_subnet_cidr
}

# ------------------------------------------------------------------------------
# Module NSG
# ------------------------------------------------------------------------------

module "nsg" {
  source = "../../modules/nsg"

  project_name        = var.project_name
  environment         = var.environment
  location            = var.location
  resource_group_name = azurerm_resource_group.main.name
  vnet_cidr           = module.vnet.vnet_cidr
}

# ------------------------------------------------------------------------------
# Module PostgreSQL Flexible Server
# ------------------------------------------------------------------------------

module "postgresql" {
  source = "../../modules/postgresql"

  project_name        = var.project_name
  environment         = var.environment
  location            = var.location
  resource_group_name = azurerm_resource_group.main.name

  subnet_id           = module.vnet.private_postgresql_subnet_id
  private_dns_zone_id = module.vnet.postgresql_private_dns_zone_id

  postgres_version = "15"
  sku_name         = "B_Standard_B1ms"  # ~$12/mois
  storage_mb       = 32768

  db_username = var.db_username
  db_password = var.db_password
}

# ------------------------------------------------------------------------------
# Module Redis
# ------------------------------------------------------------------------------

module "redis" {
  source = "../../modules/redis"

  project_name        = var.project_name
  environment         = var.environment
  location            = var.location
  resource_group_name = azurerm_resource_group.main.name

  capacity      = 0      # C0 = 250 MB
  family        = "C"
  sku_name      = "Basic"  # ~$16/mois
  redis_version = "7"
}

# ------------------------------------------------------------------------------
# Module Storage (Azure Blob Storage)
# ------------------------------------------------------------------------------

module "storage" {
  source = "../../modules/storage"

  project_name         = var.project_name
  environment          = var.environment
  location             = var.location
  resource_group_name  = azurerm_resource_group.main.name
  storage_account_name = var.storage_account_name
  cors_allowed_origins = ["*"]
}

# ------------------------------------------------------------------------------
# Module Managed Identity
# ------------------------------------------------------------------------------

module "managed_identity" {
  source = "../../modules/managed-identity"

  project_name        = var.project_name
  environment         = var.environment
  location            = var.location
  resource_group_name = azurerm_resource_group.main.name
  storage_account_id  = module.storage.storage_account_id
}

# ------------------------------------------------------------------------------
# Module Budget (alertes de coût - GRATUIT)
# ------------------------------------------------------------------------------

module "budget" {
  source = "../../modules/budget"

  project_name         = var.project_name
  environment          = var.environment
  resource_group_id    = azurerm_resource_group.main.id
  monthly_budget_limit = var.monthly_budget_limit
  alert_emails         = var.alert_emails
  budget_start_date    = var.budget_start_date
}
