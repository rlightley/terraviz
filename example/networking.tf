# Hub Virtual Network
resource "azurerm_virtual_network" "hub" {
  name                = "${local.resource_prefix}-hub-vnet"
  location            = azurerm_resource_group.connectivity.location
  resource_group_name = azurerm_resource_group.connectivity.name
  address_space       = ["10.0.0.0/16"]
  tags                = local.common_tags

  depends_on = [azurerm_resource_group.connectivity]
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

# Spoke Virtual Network - Workload
resource "azurerm_virtual_network" "workload" {
  name                = "${local.resource_prefix}-workload-vnet"
  location            = azurerm_resource_group.workload.location
  resource_group_name = azurerm_resource_group.workload.name
  address_space       = ["10.1.0.0/16"]
  tags                = local.common_tags

  depends_on = [azurerm_resource_group.workload]
}

# Workload Subnets
resource "azurerm_subnet" "web" {
  name                 = "web-subnet"
  resource_group_name  = azurerm_resource_group.workload.name
  virtual_network_name = azurerm_virtual_network.workload.name
  address_prefixes     = ["10.1.1.0/24"]

  depends_on = [azurerm_virtual_network.workload]
}

resource "azurerm_subnet" "app" {
  name                 = "app-subnet"
  resource_group_name  = azurerm_resource_group.workload.name
  virtual_network_name = azurerm_virtual_network.workload.name
  address_prefixes     = ["10.1.2.0/24"]

  depends_on = [azurerm_virtual_network.workload]
}

resource "azurerm_subnet" "data" {
  name                 = "data-subnet"
  resource_group_name  = azurerm_resource_group.workload.name
  virtual_network_name = azurerm_virtual_network.workload.name
  address_prefixes     = ["10.1.3.0/24"]

  depends_on = [azurerm_virtual_network.workload]
}

# VNet Peering - Hub to Spoke
resource "azurerm_virtual_network_peering" "hub_to_workload" {
  name                      = "hub-to-workload"
  resource_group_name       = azurerm_resource_group.connectivity.name
  virtual_network_name      = azurerm_virtual_network.hub.name
  remote_virtual_network_id = azurerm_virtual_network.workload.id
  allow_forwarded_traffic   = true
  allow_gateway_transit     = true

  depends_on = [azurerm_virtual_network.hub, azurerm_virtual_network.workload]
}

# VNet Peering - Spoke to Hub
resource "azurerm_virtual_network_peering" "workload_to_hub" {
  name                      = "workload-to-hub"
  resource_group_name       = azurerm_resource_group.workload.name
  virtual_network_name      = azurerm_virtual_network.workload.name
  remote_virtual_network_id = azurerm_virtual_network.hub.id
  allow_forwarded_traffic   = true
  use_remote_gateways       = false

  depends_on = [azurerm_virtual_network.hub, azurerm_virtual_network.workload]
}

# Public IP for Bastion
resource "azurerm_public_ip" "bastion" {
  name                = "${local.resource_prefix}-bastion-pip"
  location            = azurerm_resource_group.connectivity.location
  resource_group_name = azurerm_resource_group.connectivity.name
  allocation_method   = "Static"
  sku                 = "Standard"
  tags                = local.common_tags

  depends_on = [azurerm_resource_group.connectivity]
}

# Bastion Host
resource "azurerm_bastion_host" "main" {
  name                = "${local.resource_prefix}-bastion"
  location            = azurerm_resource_group.connectivity.location
  resource_group_name = azurerm_resource_group.connectivity.name
  tags                = local.common_tags

  ip_configuration {
    name                 = "configuration"
    subnet_id            = azurerm_subnet.bastion.id
    public_ip_address_id = azurerm_public_ip.bastion.id
  }

  depends_on = [azurerm_subnet.bastion, azurerm_public_ip.bastion]
}
