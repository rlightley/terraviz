variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
}

variable "location" {
  description = "Azure region"
  type        = string
  default     = "eastus"
}

variable "organization" {
  description = "Organization name"
  type        = string
  default     = "contoso"
}

variable "hub_address_space" {
  description = "Hub virtual network address space"
  type        = string
  default     = "10.0.0.0/16"
}

variable "management_address_space" {
  description = "Management virtual network address space"
  type        = string
  default     = "10.10.0.0/16"
}

variable "prod_address_space" {
  description = "Production spoke virtual network address space"
  type        = string
  default     = "10.20.0.0/16"
}

variable "dev_address_space" {
  description = "Development spoke virtual network address space"
  type        = string
  default     = "10.30.0.0/16"
}

locals {
  common_tags = {
    Environment  = var.environment
    ManagedBy    = "Terraform"
    Owner        = var.organization
    Architecture = "landing-zone"
    CostCenter   = "platform-engineering"
  }

  resource_prefix     = "${var.organization}-${var.environment}"
  compact_prefix      = lower(replace(local.resource_prefix, "-", ""))
  diag_storage_name   = substr("${local.compact_prefix}diag", 0, 24)
  shared_storage_name = substr("${local.compact_prefix}shared", 0, 24)
}
