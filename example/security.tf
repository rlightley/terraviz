# Network Security Group - Web Tier
resource "azurerm_network_security_group" "web" {
  name                = "${local.resource_prefix}-web-nsg"
  location            = azurerm_resource_group.workload.location
  resource_group_name = azurerm_resource_group.workload.name
  tags                = local.common_tags

  depends_on = [azurerm_resource_group.workload]
}

# NSG Rule - Allow HTTPS
resource "azurerm_network_security_rule" "allow_https" {
  name                        = "AllowHTTPS"
  priority                    = 100
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "443"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.workload.name
  network_security_group_name = azurerm_network_security_group.web.name

  depends_on = [azurerm_network_security_group.web]
}

# NSG Rule - Allow HTTP
resource "azurerm_network_security_rule" "allow_http" {
  name                        = "AllowHTTP"
  priority                    = 110
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "80"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.workload.name
  network_security_group_name = azurerm_network_security_group.web.name

  depends_on = [azurerm_network_security_group.web]
}

# Associate NSG with Web Subnet
resource "azurerm_subnet_network_security_group_association" "web" {
  subnet_id                 = azurerm_subnet.web.id
  network_security_group_id = azurerm_network_security_group.web.id

  depends_on = [azurerm_subnet.web, azurerm_network_security_group.web]
}

# Network Security Group - App Tier
resource "azurerm_network_security_group" "app" {
  name                = "${local.resource_prefix}-app-nsg"
  location            = azurerm_resource_group.workload.location
  resource_group_name = azurerm_resource_group.workload.name
  tags                = local.common_tags

  depends_on = [azurerm_resource_group.workload]
}

# NSG Rule - Allow from Web Tier
resource "azurerm_network_security_rule" "allow_from_web" {
  name                        = "AllowFromWeb"
  priority                    = 100
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "8080"
  source_address_prefix       = "10.1.1.0/24"
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.workload.name
  network_security_group_name = azurerm_network_security_group.app.name

  depends_on = [azurerm_network_security_group.app]
}

# Associate NSG with App Subnet
resource "azurerm_subnet_network_security_group_association" "app" {
  subnet_id                 = azurerm_subnet.app.id
  network_security_group_id = azurerm_network_security_group.app.id

  depends_on = [azurerm_subnet.app, azurerm_network_security_group.app]
}

# Network Security Group - Data Tier
resource "azurerm_network_security_group" "data" {
  name                = "${local.resource_prefix}-data-nsg"
  location            = azurerm_resource_group.workload.location
  resource_group_name = azurerm_resource_group.workload.name
  tags                = local.common_tags

  depends_on = [azurerm_resource_group.workload]
}

# NSG Rule - Allow from App Tier
resource "azurerm_network_security_rule" "allow_from_app" {
  name                        = "AllowFromApp"
  priority                    = 100
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_ranges     = ["1433", "3306", "5432"]
  source_address_prefix       = "10.1.2.0/24"
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.workload.name
  network_security_group_name = azurerm_network_security_group.data.name

  depends_on = [azurerm_network_security_group.data]
}

# Associate NSG with Data Subnet
resource "azurerm_subnet_network_security_group_association" "data" {
  subnet_id                 = azurerm_subnet.data.id
  network_security_group_id = azurerm_network_security_group.data.id

  depends_on = [azurerm_subnet.data, azurerm_network_security_group.data]
}

# Key Vault
resource "azurerm_key_vault" "main" {
  name                       = "${local.resource_prefix}-kv"
  location                   = azurerm_resource_group.workload.location
  resource_group_name        = azurerm_resource_group.workload.name
  tenant_id                  = data.azurerm_client_config.current.tenant_id
  sku_name                   = "standard"
  soft_delete_retention_days = 7
  purge_protection_enabled   = false
  tags                       = local.common_tags

  depends_on = [azurerm_resource_group.workload]
}

data "azurerm_client_config" "current" {}
