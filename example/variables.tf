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

locals {
  common_tags = {
    Environment = var.environment
    ManagedBy   = "Terraform"
    Owner       = var.organization
  }

  resource_prefix = "${var.organization}-${var.environment}"
}
