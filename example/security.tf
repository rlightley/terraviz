# Production Network Security
resource "azurerm_network_security_group" "prod_web" {
  name                = "${local.resource_prefix}-prod-web-nsg"
  location            = azurerm_resource_group.prod.location
  resource_group_name = azurerm_resource_group.prod.name
  tags                = local.common_tags

  depends_on = [azurerm_resource_group.prod]
}

resource "azurerm_network_security_rule" "prod_allow_https" {
  name                        = "AllowHTTPS"
  priority                    = 100
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "443"
  source_address_prefix       = "AzureFrontDoor.Backend"
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.prod.name
  network_security_group_name = azurerm_network_security_group.prod_web.name

  depends_on = [azurerm_network_security_group.prod_web]
}

resource "azurerm_subnet_network_security_group_association" "prod_web" {
  subnet_id                 = azurerm_subnet.prod_web.id
  network_security_group_id = azurerm_network_security_group.prod_web.id

  depends_on = [
    azurerm_subnet.prod_web,
    azurerm_network_security_rule.prod_allow_https,
  ]
}

resource "azurerm_network_security_group" "prod_app" {
  name                = "${local.resource_prefix}-prod-app-nsg"
  location            = azurerm_resource_group.prod.location
  resource_group_name = azurerm_resource_group.prod.name
  tags                = local.common_tags

  depends_on = [azurerm_resource_group.prod]
}

resource "azurerm_network_security_rule" "prod_allow_web_to_app" {
  name                        = "AllowWebToApp"
  priority                    = 100
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "8080"
  source_address_prefix       = "10.20.1.0/24"
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.prod.name
  network_security_group_name = azurerm_network_security_group.prod_app.name

  depends_on = [azurerm_network_security_group.prod_app]
}

resource "azurerm_subnet_network_security_group_association" "prod_app" {
  subnet_id                 = azurerm_subnet.prod_app.id
  network_security_group_id = azurerm_network_security_group.prod_app.id

  depends_on = [
    azurerm_subnet.prod_app,
    azurerm_network_security_rule.prod_allow_web_to_app,
  ]
}

resource "azurerm_network_security_group" "prod_data" {
  name                = "${local.resource_prefix}-prod-data-nsg"
  location            = azurerm_resource_group.prod.location
  resource_group_name = azurerm_resource_group.prod.name
  tags                = local.common_tags

  depends_on = [azurerm_resource_group.prod]
}

resource "azurerm_network_security_rule" "prod_allow_app_to_data" {
  name                        = "AllowAppToData"
  priority                    = 100
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_ranges     = ["1433", "5432"]
  source_address_prefix       = "10.20.2.0/24"
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.prod.name
  network_security_group_name = azurerm_network_security_group.prod_data.name

  depends_on = [azurerm_network_security_group.prod_data]
}

resource "azurerm_subnet_network_security_group_association" "prod_data" {
  subnet_id                 = azurerm_subnet.prod_data.id
  network_security_group_id = azurerm_network_security_group.prod_data.id

  depends_on = [
    azurerm_subnet.prod_data,
    azurerm_network_security_rule.prod_allow_app_to_data,
  ]
}

# Development Network Security
resource "azurerm_network_security_group" "dev_web" {
  name                = "${local.resource_prefix}-dev-web-nsg"
  location            = azurerm_resource_group.dev.location
  resource_group_name = azurerm_resource_group.dev.name
  tags                = local.common_tags

  depends_on = [azurerm_resource_group.dev]
}

resource "azurerm_network_security_rule" "dev_allow_https" {
  name                        = "AllowHTTPS"
  priority                    = 100
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "443"
  source_address_prefix       = "Internet"
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.dev.name
  network_security_group_name = azurerm_network_security_group.dev_web.name

  depends_on = [azurerm_network_security_group.dev_web]
}

resource "azurerm_subnet_network_security_group_association" "dev_web" {
  subnet_id                 = azurerm_subnet.dev_web.id
  network_security_group_id = azurerm_network_security_group.dev_web.id

  depends_on = [
    azurerm_subnet.dev_web,
    azurerm_network_security_rule.dev_allow_https,
  ]
}

resource "azurerm_network_security_group" "dev_app" {
  name                = "${local.resource_prefix}-dev-app-nsg"
  location            = azurerm_resource_group.dev.location
  resource_group_name = azurerm_resource_group.dev.name
  tags                = local.common_tags

  depends_on = [azurerm_resource_group.dev]
}

resource "azurerm_network_security_rule" "dev_allow_web_to_app" {
  name                        = "AllowWebToApp"
  priority                    = 100
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "8080"
  source_address_prefix       = "10.30.1.0/24"
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.dev.name
  network_security_group_name = azurerm_network_security_group.dev_app.name

  depends_on = [azurerm_network_security_group.dev_app]
}

resource "azurerm_subnet_network_security_group_association" "dev_app" {
  subnet_id                 = azurerm_subnet.dev_app.id
  network_security_group_id = azurerm_network_security_group.dev_app.id

  depends_on = [
    azurerm_subnet.dev_app,
    azurerm_network_security_rule.dev_allow_web_to_app,
  ]
}

# Key Vault and Private Access
resource "azurerm_key_vault" "platform" {
  name                       = "${local.resource_prefix}-kv"
  location                   = azurerm_resource_group.shared.location
  resource_group_name        = azurerm_resource_group.shared.name
  tenant_id                  = data.azurerm_client_config.current.tenant_id
  sku_name                   = "standard"
  soft_delete_retention_days = 14
  purge_protection_enabled   = true
  tags                       = local.common_tags

  network_acls {
    default_action = "Deny"
    bypass         = "AzureServices"
  }

  depends_on = [
    azurerm_resource_group.shared,
    azurerm_private_dns_zone.keyvault,
  ]
}

resource "azurerm_key_vault_access_policy" "platform_ops" {
  key_vault_id = azurerm_key_vault.platform.id
  tenant_id    = data.azurerm_client_config.current.tenant_id
  object_id    = azurerm_user_assigned_identity.platform_ops.principal_id

  secret_permissions = ["Get", "List", "Set"]
  key_permissions    = ["Get", "List"]

  depends_on = [
    azurerm_key_vault.platform,
    azurerm_user_assigned_identity.platform_ops,
  ]
}

resource "azurerm_key_vault_secret" "shared_storage_connection" {
  name         = "shared-storage-connection-string"
  value        = azurerm_storage_account.shared.primary_connection_string
  key_vault_id = azurerm_key_vault.platform.id

  depends_on = [
    azurerm_key_vault_access_policy.platform_ops,
    azurerm_storage_account.shared,
  ]
}

resource "azurerm_private_endpoint" "keyvault" {
  name                = "${local.resource_prefix}-kv-pe"
  location            = azurerm_resource_group.prod.location
  resource_group_name = azurerm_resource_group.prod.name
  subnet_id           = azurerm_subnet.prod_private_endpoints.id
  tags                = local.common_tags

  private_service_connection {
    name                           = "keyvault-connection"
    private_connection_resource_id = azurerm_key_vault.platform.id
    is_manual_connection           = false
    subresource_names              = ["vault"]
  }

  private_dns_zone_group {
    name                 = "keyvault-dns"
    private_dns_zone_ids = [azurerm_private_dns_zone.keyvault.id]
  }

  depends_on = [
    azurerm_key_vault.platform,
    azurerm_subnet.prod_private_endpoints,
    azurerm_private_dns_zone_virtual_network_link.keyvault_prod,
  ]
}

resource "azurerm_private_endpoint" "shared_storage" {
  name                = "${local.resource_prefix}-shared-storage-pe"
  location            = azurerm_resource_group.prod.location
  resource_group_name = azurerm_resource_group.prod.name
  subnet_id           = azurerm_subnet.prod_private_endpoints.id
  tags                = local.common_tags

  private_service_connection {
    name                           = "storage-connection"
    private_connection_resource_id = azurerm_storage_account.shared.id
    is_manual_connection           = false
    subresource_names              = ["blob"]
  }

  private_dns_zone_group {
    name                 = "blob-dns"
    private_dns_zone_ids = [azurerm_private_dns_zone.blob.id]
  }

  depends_on = [
    azurerm_storage_account.shared,
    azurerm_subnet.prod_private_endpoints,
    azurerm_private_dns_zone_virtual_network_link.blob_prod,
  ]
}

# Monitoring Alerts
resource "azurerm_monitor_metric_alert" "prod_http_5xx" {
  name                = "${local.resource_prefix}-prod-http-5xx"
  resource_group_name = azurerm_resource_group.management.name
  scopes              = [azurerm_linux_web_app.prod_portal.id]
  description         = "Alert when the production portal returns excessive 5xx responses"
  severity            = 2
  frequency           = "PT5M"
  window_size         = "PT15M"
  tags                = local.common_tags

  criteria {
    metric_namespace = "Microsoft.Web/sites"
    metric_name      = "Http5xx"
    aggregation      = "Total"
    operator         = "GreaterThan"
    threshold        = 5
  }

  action {
    action_group_id = azurerm_monitor_action_group.platform.id
  }

  depends_on = [
    azurerm_linux_web_app.prod_portal,
    azurerm_monitor_action_group.platform,
    azurerm_application_insights.shared,
  ]
}

resource "azurerm_monitor_metric_alert" "dev_cpu_high" {
  name                = "${local.resource_prefix}-dev-cpu-high"
  resource_group_name = azurerm_resource_group.management.name
  scopes              = [azurerm_linux_web_app.dev_portal.id]
  description         = "Alert when the development portal sustains high CPU"
  severity            = 3
  frequency           = "PT5M"
  window_size         = "PT15M"
  tags                = local.common_tags

  criteria {
    metric_namespace = "Microsoft.Web/sites"
    metric_name      = "CpuTime"
    aggregation      = "Average"
    operator         = "GreaterThan"
    threshold        = 80
  }

  action {
    action_group_id = azurerm_monitor_action_group.platform.id
  }

  depends_on = [
    azurerm_linux_web_app.dev_portal,
    azurerm_monitor_action_group.platform,
    azurerm_application_insights.shared,
  ]
}
