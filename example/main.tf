terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
}

provider "azurerm" {
  features {}
}

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

# Workload Resource Group
resource "azurerm_resource_group" "workload" {
  name     = "${local.resource_prefix}-workload-rg"
  location = var.location
  tags     = local.common_tags
}

# Log Analytics Workspace
resource "azurerm_log_analytics_workspace" "main" {
  name                = "${local.resource_prefix}-law"
  location            = azurerm_resource_group.management.location
  resource_group_name = azurerm_resource_group.management.name
  sku                 = "PerGB2018"
  retention_in_days   = 30
  tags                = local.common_tags

  depends_on = [azurerm_resource_group.management]
}

# Automation Account
resource "azurerm_automation_account" "main" {
  name                = "${local.resource_prefix}-automation"
  location            = azurerm_resource_group.management.location
  resource_group_name = azurerm_resource_group.management.name
  sku_name            = "Basic"
  tags                = local.common_tags

  depends_on = [azurerm_resource_group.management]
}
