terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.110"
    }
  }
}

provider "azurerm" {
  features {}
}

data "azurerm_client_config" "current" {}

# Management Resource Group
resource "azurerm_resource_group" "management" {
  name     = "${local.resource_prefix}-management-rg"
  location = var.location
  tags     = local.common_tags
}

# Connectivity Resource Group
resource "azurerm_resource_group" "connectivity" {
  name     = "${local.resource_prefix}-connectivity-rg"
  location = var.location
  tags     = local.common_tags
}

# Identity Resource Group
resource "azurerm_resource_group" "identity" {
  name     = "${local.resource_prefix}-identity-rg"
  location = var.location
  tags     = local.common_tags
}

# Shared Services Resource Group
resource "azurerm_resource_group" "shared" {
  name     = "${local.resource_prefix}-shared-rg"
  location = var.location
  tags     = local.common_tags
}

# Production Workload Resource Group
resource "azurerm_resource_group" "prod" {
  name     = "${local.resource_prefix}-prod-rg"
  location = var.location
  tags     = local.common_tags
}

# Development Workload Resource Group
resource "azurerm_resource_group" "dev" {
  name     = "${local.resource_prefix}-dev-rg"
  location = var.location
  tags     = local.common_tags
}

# Log Analytics Workspace
resource "azurerm_log_analytics_workspace" "platform" {
  name                = "${local.resource_prefix}-law"
  location            = azurerm_resource_group.management.location
  resource_group_name = azurerm_resource_group.management.name
  sku                 = "PerGB2018"
  retention_in_days   = 90
  tags                = local.common_tags

  depends_on = [azurerm_resource_group.management]
}

# Automation Account
resource "azurerm_automation_account" "platform" {
  name                = "${local.resource_prefix}-automation"
  location            = azurerm_resource_group.management.location
  resource_group_name = azurerm_resource_group.management.name
  sku_name            = "Basic"
  tags                = local.common_tags

  depends_on = [
    azurerm_resource_group.management,
    azurerm_log_analytics_workspace.platform,
  ]
}

# Platform Diagnostics Storage Account
resource "azurerm_storage_account" "diagnostics" {
  name                     = local.diag_storage_name
  resource_group_name      = azurerm_resource_group.management.name
  location                 = azurerm_resource_group.management.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  min_tls_version          = "TLS1_2"
  tags                     = local.common_tags

  depends_on = [azurerm_resource_group.management]
}

# Recovery Services Vault
resource "azurerm_recovery_services_vault" "platform" {
  name                = "${local.resource_prefix}-rsv"
  location            = azurerm_resource_group.management.location
  resource_group_name = azurerm_resource_group.management.name
  sku                 = "Standard"
  soft_delete_enabled = true
  tags                = local.common_tags

  depends_on = [
    azurerm_resource_group.management,
    azurerm_storage_account.diagnostics,
  ]
}

# Backup Policy
resource "azurerm_backup_policy_vm" "platform" {
  name                = "${local.resource_prefix}-daily-vm"
  resource_group_name = azurerm_resource_group.management.name
  recovery_vault_name = azurerm_recovery_services_vault.platform.name
  timezone            = "UTC"

  backup {
    frequency = "Daily"
    time      = "23:00"
  }

  retention_daily {
    count = 30
  }

  depends_on = [azurerm_recovery_services_vault.platform]
}

# Platform Operations Identity
resource "azurerm_user_assigned_identity" "platform_ops" {
  name                = "${local.resource_prefix}-platform-ops-id"
  location            = azurerm_resource_group.identity.location
  resource_group_name = azurerm_resource_group.identity.name
  tags                = local.common_tags

  depends_on = [azurerm_resource_group.identity]
}

# Workload Identity
resource "azurerm_user_assigned_identity" "workload_runtime" {
  name                = "${local.resource_prefix}-runtime-id"
  location            = azurerm_resource_group.identity.location
  resource_group_name = azurerm_resource_group.identity.name
  tags                = local.common_tags

  depends_on = [
    azurerm_resource_group.identity,
    azurerm_user_assigned_identity.platform_ops,
  ]
}

# Shared Observability
resource "azurerm_application_insights" "shared" {
  name                = "${local.resource_prefix}-appi"
  location            = azurerm_resource_group.shared.location
  resource_group_name = azurerm_resource_group.shared.name
  workspace_id        = azurerm_log_analytics_workspace.platform.id
  application_type    = "web"
  tags                = local.common_tags

  depends_on = [
    azurerm_resource_group.shared,
    azurerm_log_analytics_workspace.platform,
  ]
}

# Shared Alerting
resource "azurerm_monitor_action_group" "platform" {
  name                = "${local.resource_prefix}-alerts"
  resource_group_name = azurerm_resource_group.management.name
  short_name          = "platops"
  tags                = local.common_tags

  email_receiver {
    name          = "platform-email"
    email_address = "platformops@contoso.example"
  }

  depends_on = [azurerm_resource_group.management]
}

# Shared Container Registry
resource "azurerm_container_registry" "shared" {
  name                          = substr("${local.compact_prefix}acr", 0, 50)
  resource_group_name           = azurerm_resource_group.shared.name
  location                      = azurerm_resource_group.shared.location
  sku                           = "Premium"
  admin_enabled                 = false
  public_network_access_enabled = false
  tags                          = local.common_tags

  depends_on = [azurerm_resource_group.shared]
}

# Shared Storage Account
resource "azurerm_storage_account" "shared" {
  name                     = local.shared_storage_name
  resource_group_name      = azurerm_resource_group.shared.name
  location                 = azurerm_resource_group.shared.location
  account_tier             = "Standard"
  account_replication_type = "ZRS"
  min_tls_version          = "TLS1_2"
  tags                     = local.common_tags

  depends_on = [
    azurerm_resource_group.shared,
    azurerm_log_analytics_workspace.platform,
  ]
}

# Production App Service Plan
resource "azurerm_service_plan" "prod" {
  name                = "${local.resource_prefix}-prod-plan"
  resource_group_name = azurerm_resource_group.prod.name
  location            = azurerm_resource_group.prod.location
  os_type             = "Linux"
  sku_name            = "P1v3"
  tags                = local.common_tags

  depends_on = [
    azurerm_resource_group.prod,
    azurerm_application_insights.shared,
  ]
}

# Development App Service Plan
resource "azurerm_service_plan" "dev" {
  name                = "${local.resource_prefix}-dev-plan"
  resource_group_name = azurerm_resource_group.dev.name
  location            = azurerm_resource_group.dev.location
  os_type             = "Linux"
  sku_name            = "S1"
  tags                = local.common_tags

  depends_on = [
    azurerm_resource_group.dev,
    azurerm_application_insights.shared,
  ]
}

# Production Web Application
resource "azurerm_linux_web_app" "prod_portal" {
  name                = "${local.resource_prefix}-prod-portal"
  resource_group_name = azurerm_resource_group.prod.name
  location            = azurerm_resource_group.prod.location
  service_plan_id     = azurerm_service_plan.prod.id
  https_only          = true
  tags                = local.common_tags

  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.workload_runtime.id]
  }

  site_config {
    always_on              = true
    ftps_state             = "Disabled"
    vnet_route_all_enabled = true
  }

  app_settings = {
    APPINSIGHTS_INSTRUMENTATIONKEY = azurerm_application_insights.shared.instrumentation_key
    CONTAINER_REGISTRY_URL         = azurerm_container_registry.shared.login_server
    SHARED_STORAGE_ACCOUNT         = azurerm_storage_account.shared.name
  }

  depends_on = [
    azurerm_service_plan.prod,
    azurerm_application_insights.shared,
    azurerm_user_assigned_identity.workload_runtime,
    azurerm_container_registry.shared,
    azurerm_storage_account.shared,
    azurerm_subnet.prod_app,
    azurerm_key_vault.platform,
  ]
}

# Development Web Application
resource "azurerm_linux_web_app" "dev_portal" {
  name                = "${local.resource_prefix}-dev-portal"
  resource_group_name = azurerm_resource_group.dev.name
  location            = azurerm_resource_group.dev.location
  service_plan_id     = azurerm_service_plan.dev.id
  https_only          = true
  tags                = local.common_tags

  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.workload_runtime.id]
  }

  site_config {
    always_on              = true
    ftps_state             = "Disabled"
    vnet_route_all_enabled = true
  }

  app_settings = {
    APPINSIGHTS_INSTRUMENTATIONKEY = azurerm_application_insights.shared.instrumentation_key
    SHARED_STORAGE_ACCOUNT         = azurerm_storage_account.shared.name
  }

  depends_on = [
    azurerm_service_plan.dev,
    azurerm_application_insights.shared,
    azurerm_user_assigned_identity.workload_runtime,
    azurerm_storage_account.shared,
    azurerm_subnet.dev_app,
    azurerm_key_vault.platform,
  ]
}

# App Service VNet Integration
resource "azurerm_app_service_virtual_network_swift_connection" "prod" {
  app_service_id = azurerm_linux_web_app.prod_portal.id
  subnet_id      = azurerm_subnet.prod_app.id

  depends_on = [
    azurerm_linux_web_app.prod_portal,
    azurerm_subnet.prod_app,
  ]
}

resource "azurerm_app_service_virtual_network_swift_connection" "dev" {
  app_service_id = azurerm_linux_web_app.dev_portal.id
  subnet_id      = azurerm_subnet.dev_app.id

  depends_on = [
    azurerm_linux_web_app.dev_portal,
    azurerm_subnet.dev_app,
  ]
}
