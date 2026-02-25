# ==============================================================================
# MODULE POSTGRESQL - Azure Database for PostgreSQL Flexible Server
# ==============================================================================
#
# PostgreSQL Flexible Server est le service managé Azure (équivalent de RDS).
# Avantages :
# - Backups automatiques (7 jours en dev, 35 en prod)
# - Patch automatique
# - Haute disponibilité optionnelle
# - ~$12/mois pour B1ms (dev) vs ~$15/mois pour RDS t3.micro
#
# ==============================================================================

resource "azurerm_postgresql_flexible_server" "main" {
  name                   = "${var.project_name}-postgresql-${var.environment}"
  resource_group_name    = var.resource_group_name
  location               = var.location

  # PostgreSQL version
  version = var.postgres_version  # "15"

  # Réseau privé (via subnet délégué + DNS privé)
  delegated_subnet_id = var.subnet_id
  private_dns_zone_id = var.private_dns_zone_id

  # Credentials administrateur
  administrator_login    = var.db_username
  administrator_password = var.db_password

  # Configuration matériel
  # B_Standard_B1ms = ~$12/mois (dev), Standard_D2s_v3 = prod
  sku_name   = var.sku_name      # "B_Standard_B1ms"
  storage_mb = var.storage_mb    # 32768 = 32 Go

  # Backups
  backup_retention_days        = var.environment == "prod" ? 35 : 7
  geo_redundant_backup_enabled = false  # Coûteux, désactivé en dev

  # Zone de disponibilité
  zone = "1"

  tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}

# Base de données "storm" dans le serveur
resource "azurerm_postgresql_flexible_server_database" "storm" {
  name      = "storm"
  server_id = azurerm_postgresql_flexible_server.main.id
  collation = "en_US.utf8"
  charset   = "utf8"
}
