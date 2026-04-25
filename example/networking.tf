# DDoS Protection Plan
resource "azurerm_network_ddos_protection_plan" "platform" {
  name                = "${local.resource_prefix}-ddos"
  location            = azurerm_resource_group.connectivity.location
  resource_group_name = azurerm_resource_group.connectivity.name
  tags                = local.common_tags

  depends_on = [azurerm_resource_group.connectivity]
}

# Hub Virtual Network
resource "azurerm_virtual_network" "hub" {
  name                = "${local.resource_prefix}-hub-vnet"
  location            = azurerm_resource_group.connectivity.location
  resource_group_name = azurerm_resource_group.connectivity.name
  address_space       = [var.hub_address_space]
  tags                = local.common_tags

  ddos_protection_plan {
    id     = azurerm_network_ddos_protection_plan.platform.id
    enable = true
  }

  depends_on = [
    azurerm_resource_group.connectivity,
    azurerm_network_ddos_protection_plan.platform,
  ]
}

# Management Virtual Network
resource "azurerm_virtual_network" "management" {
  name                = "${local.resource_prefix}-management-vnet"
  location            = azurerm_resource_group.management.location
  resource_group_name = azurerm_resource_group.management.name
  address_space       = [var.management_address_space]
  tags                = local.common_tags

  depends_on = [
    azurerm_resource_group.management,
    azurerm_log_analytics_workspace.platform,
  ]
}

# Production Spoke Virtual Network
resource "azurerm_virtual_network" "prod" {
  name                = "${local.resource_prefix}-prod-vnet"
  location            = azurerm_resource_group.prod.location
  resource_group_name = azurerm_resource_group.prod.name
  address_space       = [var.prod_address_space]
  tags                = local.common_tags

  depends_on = [azurerm_resource_group.prod]
}

# Development Spoke Virtual Network
resource "azurerm_virtual_network" "dev" {
  name                = "${local.resource_prefix}-dev-vnet"
  location            = azurerm_resource_group.dev.location
  resource_group_name = azurerm_resource_group.dev.name
  address_space       = [var.dev_address_space]
  tags                = local.common_tags

  depends_on = [azurerm_resource_group.dev]
}

# Hub Subnets
resource "azurerm_subnet" "firewall" {
  name                 = "AzureFirewallSubnet"
  resource_group_name  = azurerm_resource_group.connectivity.name
  virtual_network_name = azurerm_virtual_network.hub.name
  address_prefixes     = ["10.0.1.0/24"]

  depends_on = [azurerm_virtual_network.hub]
}

resource "azurerm_subnet" "bastion" {
  name                 = "AzureBastionSubnet"
  resource_group_name  = azurerm_resource_group.connectivity.name
  virtual_network_name = azurerm_virtual_network.hub.name
  address_prefixes     = ["10.0.2.0/24"]

  depends_on = [azurerm_virtual_network.hub]
}

resource "azurerm_subnet" "gateway" {
  name                 = "GatewaySubnet"
  resource_group_name  = azurerm_resource_group.connectivity.name
  virtual_network_name = azurerm_virtual_network.hub.name
  address_prefixes     = ["10.0.3.0/24"]

  depends_on = [azurerm_virtual_network.hub]
}

resource "azurerm_subnet" "management_services" {
  name                 = "management-services-subnet"
  resource_group_name  = azurerm_resource_group.connectivity.name
  virtual_network_name = azurerm_virtual_network.hub.name
  address_prefixes     = ["10.0.4.0/24"]

  depends_on = [azurerm_virtual_network.hub]
}

resource "azurerm_subnet" "management_ops" {
  name                 = "ops-subnet"
  resource_group_name  = azurerm_resource_group.management.name
  virtual_network_name = azurerm_virtual_network.management.name
  address_prefixes     = ["10.10.1.0/24"]

  depends_on = [azurerm_virtual_network.management]
}

# Production Subnets
resource "azurerm_subnet" "prod_web" {
  name                 = "web-subnet"
  resource_group_name  = azurerm_resource_group.prod.name
  virtual_network_name = azurerm_virtual_network.prod.name
  address_prefixes     = ["10.20.1.0/24"]

  depends_on = [azurerm_virtual_network.prod]
}

resource "azurerm_subnet" "prod_app" {
  name                 = "app-subnet"
  resource_group_name  = azurerm_resource_group.prod.name
  virtual_network_name = azurerm_virtual_network.prod.name
  address_prefixes     = ["10.20.2.0/24"]
  delegation {
    name = "app-service-delegation"

    service_delegation {
      actions = ["Microsoft.Network/virtualNetworks/subnets/action"]
      name    = "Microsoft.Web/serverFarms"
    }
  }

  depends_on = [azurerm_virtual_network.prod]
}

resource "azurerm_subnet" "prod_data" {
  name                 = "data-subnet"
  resource_group_name  = azurerm_resource_group.prod.name
  virtual_network_name = azurerm_virtual_network.prod.name
  address_prefixes     = ["10.20.3.0/24"]

  depends_on = [azurerm_virtual_network.prod]
}

resource "azurerm_subnet" "prod_private_endpoints" {
  name                 = "private-endpoints-subnet"
  resource_group_name  = azurerm_resource_group.prod.name
  virtual_network_name = azurerm_virtual_network.prod.name
  address_prefixes     = ["10.20.4.0/24"]

  depends_on = [azurerm_virtual_network.prod]
}

# Development Subnets
resource "azurerm_subnet" "dev_web" {
  name                 = "web-subnet"
  resource_group_name  = azurerm_resource_group.dev.name
  virtual_network_name = azurerm_virtual_network.dev.name
  address_prefixes     = ["10.30.1.0/24"]

  depends_on = [azurerm_virtual_network.dev]
}

resource "azurerm_subnet" "dev_app" {
  name                 = "app-subnet"
  resource_group_name  = azurerm_resource_group.dev.name
  virtual_network_name = azurerm_virtual_network.dev.name
  address_prefixes     = ["10.30.2.0/24"]
  delegation {
    name = "app-service-delegation"

    service_delegation {
      actions = ["Microsoft.Network/virtualNetworks/subnets/action"]
      name    = "Microsoft.Web/serverFarms"
    }
  }

  depends_on = [azurerm_virtual_network.dev]
}

resource "azurerm_subnet" "dev_data" {
  name                 = "data-subnet"
  resource_group_name  = azurerm_resource_group.dev.name
  virtual_network_name = azurerm_virtual_network.dev.name
  address_prefixes     = ["10.30.3.0/24"]

  depends_on = [azurerm_virtual_network.dev]
}

# Firewall and Bastion
resource "azurerm_public_ip" "firewall" {
  name                = "${local.resource_prefix}-firewall-pip"
  location            = azurerm_resource_group.connectivity.location
  resource_group_name = azurerm_resource_group.connectivity.name
  allocation_method   = "Static"
  sku                 = "Standard"
  tags                = local.common_tags

  depends_on = [azurerm_resource_group.connectivity]
}

resource "azurerm_firewall_policy" "hub" {
  name                = "${local.resource_prefix}-fw-policy"
  resource_group_name = azurerm_resource_group.connectivity.name
  location            = azurerm_resource_group.connectivity.location
  tags                = local.common_tags

  depends_on = [
    azurerm_resource_group.connectivity,
    azurerm_log_analytics_workspace.platform,
  ]
}

resource "azurerm_firewall" "hub" {
  name                = "${local.resource_prefix}-firewall"
  location            = azurerm_resource_group.connectivity.location
  resource_group_name = azurerm_resource_group.connectivity.name
  sku_name            = "AZFW_VNet"
  sku_tier            = "Standard"
  firewall_policy_id  = azurerm_firewall_policy.hub.id
  tags                = local.common_tags

  ip_configuration {
    name                 = "hub-config"
    subnet_id            = azurerm_subnet.firewall.id
    public_ip_address_id = azurerm_public_ip.firewall.id
  }

  depends_on = [
    azurerm_firewall_policy.hub,
    azurerm_subnet.firewall,
    azurerm_public_ip.firewall,
  ]
}

resource "azurerm_public_ip" "bastion" {
  name                = "${local.resource_prefix}-bastion-pip"
  location            = azurerm_resource_group.connectivity.location
  resource_group_name = azurerm_resource_group.connectivity.name
  allocation_method   = "Static"
  sku                 = "Standard"
  tags                = local.common_tags

  depends_on = [azurerm_resource_group.connectivity]
}

resource "azurerm_bastion_host" "platform" {
  name                = "${local.resource_prefix}-bastion"
  location            = azurerm_resource_group.connectivity.location
  resource_group_name = azurerm_resource_group.connectivity.name
  tags                = local.common_tags

  ip_configuration {
    name                 = "configuration"
    subnet_id            = azurerm_subnet.bastion.id
    public_ip_address_id = azurerm_public_ip.bastion.id
  }

  depends_on = [
    azurerm_subnet.bastion,
    azurerm_public_ip.bastion,
    azurerm_firewall.hub,
  ]
}

# Route Tables
resource "azurerm_route_table" "prod" {
  name                = "${local.resource_prefix}-prod-rt"
  location            = azurerm_resource_group.prod.location
  resource_group_name = azurerm_resource_group.prod.name
  tags                = local.common_tags

  depends_on = [
    azurerm_resource_group.prod,
    azurerm_firewall.hub,
  ]
}

resource "azurerm_route" "prod_default" {
  name                   = "default-to-firewall"
  resource_group_name    = azurerm_resource_group.prod.name
  route_table_name       = azurerm_route_table.prod.name
  address_prefix         = "0.0.0.0/0"
  next_hop_type          = "VirtualAppliance"
  next_hop_in_ip_address = "10.0.1.4"

  depends_on = [azurerm_route_table.prod]
}

resource "azurerm_route_table" "dev" {
  name                = "${local.resource_prefix}-dev-rt"
  location            = azurerm_resource_group.dev.location
  resource_group_name = azurerm_resource_group.dev.name
  tags                = local.common_tags

  depends_on = [
    azurerm_resource_group.dev,
    azurerm_firewall.hub,
  ]
}

resource "azurerm_route" "dev_default" {
  name                   = "default-to-firewall"
  resource_group_name    = azurerm_resource_group.dev.name
  route_table_name       = azurerm_route_table.dev.name
  address_prefix         = "0.0.0.0/0"
  next_hop_type          = "VirtualAppliance"
  next_hop_in_ip_address = "10.0.1.4"

  depends_on = [azurerm_route_table.dev]
}

resource "azurerm_subnet_route_table_association" "prod_app" {
  subnet_id      = azurerm_subnet.prod_app.id
  route_table_id = azurerm_route_table.prod.id

  depends_on = [
    azurerm_subnet.prod_app,
    azurerm_route.prod_default,
  ]
}

resource "azurerm_subnet_route_table_association" "prod_data" {
  subnet_id      = azurerm_subnet.prod_data.id
  route_table_id = azurerm_route_table.prod.id

  depends_on = [
    azurerm_subnet.prod_data,
    azurerm_route.prod_default,
  ]
}

resource "azurerm_subnet_route_table_association" "dev_app" {
  subnet_id      = azurerm_subnet.dev_app.id
  route_table_id = azurerm_route_table.dev.id

  depends_on = [
    azurerm_subnet.dev_app,
    azurerm_route.dev_default,
  ]
}

resource "azurerm_subnet_route_table_association" "dev_data" {
  subnet_id      = azurerm_subnet.dev_data.id
  route_table_id = azurerm_route_table.dev.id

  depends_on = [
    azurerm_subnet.dev_data,
    azurerm_route.dev_default,
  ]
}

# VNet Peering
resource "azurerm_virtual_network_peering" "hub_to_management" {
  name                      = "hub-to-management"
  resource_group_name       = azurerm_resource_group.connectivity.name
  virtual_network_name      = azurerm_virtual_network.hub.name
  remote_virtual_network_id = azurerm_virtual_network.management.id
  allow_forwarded_traffic   = true

  depends_on = [
    azurerm_virtual_network.hub,
    azurerm_virtual_network.management,
  ]
}

resource "azurerm_virtual_network_peering" "management_to_hub" {
  name                      = "management-to-hub"
  resource_group_name       = azurerm_resource_group.management.name
  virtual_network_name      = azurerm_virtual_network.management.name
  remote_virtual_network_id = azurerm_virtual_network.hub.id
  allow_forwarded_traffic   = true

  depends_on = [
    azurerm_virtual_network.hub,
    azurerm_virtual_network.management,
  ]
}

resource "azurerm_virtual_network_peering" "hub_to_prod" {
  name                      = "hub-to-prod"
  resource_group_name       = azurerm_resource_group.connectivity.name
  virtual_network_name      = azurerm_virtual_network.hub.name
  remote_virtual_network_id = azurerm_virtual_network.prod.id
  allow_forwarded_traffic   = true
  allow_gateway_transit     = true

  depends_on = [
    azurerm_virtual_network.hub,
    azurerm_virtual_network.prod,
  ]
}

resource "azurerm_virtual_network_peering" "prod_to_hub" {
  name                      = "prod-to-hub"
  resource_group_name       = azurerm_resource_group.prod.name
  virtual_network_name      = azurerm_virtual_network.prod.name
  remote_virtual_network_id = azurerm_virtual_network.hub.id
  allow_forwarded_traffic   = true

  depends_on = [
    azurerm_virtual_network.hub,
    azurerm_virtual_network.prod,
  ]
}

resource "azurerm_virtual_network_peering" "hub_to_dev" {
  name                      = "hub-to-dev"
  resource_group_name       = azurerm_resource_group.connectivity.name
  virtual_network_name      = azurerm_virtual_network.hub.name
  remote_virtual_network_id = azurerm_virtual_network.dev.id
  allow_forwarded_traffic   = true

  depends_on = [
    azurerm_virtual_network.hub,
    azurerm_virtual_network.dev,
  ]
}

resource "azurerm_virtual_network_peering" "dev_to_hub" {
  name                      = "dev-to-hub"
  resource_group_name       = azurerm_resource_group.dev.name
  virtual_network_name      = azurerm_virtual_network.dev.name
  remote_virtual_network_id = azurerm_virtual_network.hub.id
  allow_forwarded_traffic   = true

  depends_on = [
    azurerm_virtual_network.hub,
    azurerm_virtual_network.dev,
  ]
}

# Private DNS
resource "azurerm_private_dns_zone" "blob" {
  name                = "privatelink.blob.core.windows.net"
  resource_group_name = azurerm_resource_group.shared.name

  depends_on = [azurerm_resource_group.shared]
}

resource "azurerm_private_dns_zone" "keyvault" {
  name                = "privatelink.vaultcore.azure.net"
  resource_group_name = azurerm_resource_group.shared.name

  depends_on = [azurerm_resource_group.shared]
}

resource "azurerm_private_dns_zone_virtual_network_link" "blob_prod" {
  name                  = "blob-prod-link"
  resource_group_name   = azurerm_resource_group.shared.name
  private_dns_zone_name = azurerm_private_dns_zone.blob.name
  virtual_network_id    = azurerm_virtual_network.prod.id

  depends_on = [
    azurerm_private_dns_zone.blob,
    azurerm_virtual_network.prod,
  ]
}

resource "azurerm_private_dns_zone_virtual_network_link" "blob_dev" {
  name                  = "blob-dev-link"
  resource_group_name   = azurerm_resource_group.shared.name
  private_dns_zone_name = azurerm_private_dns_zone.blob.name
  virtual_network_id    = azurerm_virtual_network.dev.id

  depends_on = [
    azurerm_private_dns_zone.blob,
    azurerm_virtual_network.dev,
  ]
}

resource "azurerm_private_dns_zone_virtual_network_link" "keyvault_prod" {
  name                  = "keyvault-prod-link"
  resource_group_name   = azurerm_resource_group.shared.name
  private_dns_zone_name = azurerm_private_dns_zone.keyvault.name
  virtual_network_id    = azurerm_virtual_network.prod.id

  depends_on = [
    azurerm_private_dns_zone.keyvault,
    azurerm_virtual_network.prod,
  ]
}

resource "azurerm_private_dns_zone_virtual_network_link" "keyvault_dev" {
  name                  = "keyvault-dev-link"
  resource_group_name   = azurerm_resource_group.shared.name
  private_dns_zone_name = azurerm_private_dns_zone.keyvault.name
  virtual_network_id    = azurerm_virtual_network.dev.id

  depends_on = [
    azurerm_private_dns_zone.keyvault,
    azurerm_virtual_network.dev,
  ]
}
