# Write-Only Secret Example
#
# This example demonstrates storing generated credentials in gopass
# without exposing them in Terraform state.

terraform {
  required_version = ">= 1.11.0"

  required_providers {
    gopass = {
      source  = "registry.opentofu.org/istr/gopass"
      version = "~> 0.1"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }
}

provider "gopass" {}

# -----------------------------------------------------------------------------
# Example 1: Store a randomly generated password
# -----------------------------------------------------------------------------

# Generate a random password (ephemeral - not in state)
ephemeral "random_password" "db_admin" {
  length           = 32
  special          = true
  override_special = "!@#$%&*()-_=+"
}

# Store it in gopass (value_wo is write-only - not in state)
resource "gopass_secret" "db_admin_password" {
  path             = "infrastructure/database/admin_password"
  value_wo         = ephemeral.random_password.db_admin.result
  value_wo_version = 1  # Increment to trigger update

  # Keep secret in gopass even if resource is destroyed
  delete_on_remove = false
}

# -----------------------------------------------------------------------------
# Example 2: Store an API key from another provider
# -----------------------------------------------------------------------------

# Uncomment when using Scaleway provider:
#
# resource "scaleway_iam_application" "infra_manager" {
#   name        = "terraform-infra-manager"
#   description = "Infrastructure automation"
# }
#
# resource "scaleway_iam_api_key" "infra_manager" {
#   application_id = scaleway_iam_application.infra_manager.id
#   description    = "API key for Terraform"
# }
#
# resource "gopass_secret" "scw_access_key" {
#   path             = "env/terraform/scaleway/infra-manager/SCW_ACCESS_KEY"
#   value_wo         = scaleway_iam_api_key.infra_manager.access_key
#   value_wo_version = 1
# }
#
# resource "gopass_secret" "scw_secret_key" {
#   path             = "env/terraform/scaleway/infra-manager/SCW_SECRET_KEY"
#   value_wo         = scaleway_iam_api_key.infra_manager.secret_key
#   value_wo_version = 1
# }

# -----------------------------------------------------------------------------
# Example 3: Bidirectional - read existing, write generated
# -----------------------------------------------------------------------------

# Read existing database credentials
ephemeral "gopass_env" "db_credentials" {
  path = "infrastructure/database"
}

# Output shows how to use existing credentials
# (values are ephemeral, only available during apply)
output "db_host" {
  value       = ephemeral.gopass_env.db_credentials.values["host"]
  ephemeral   = true
  description = "Database host from gopass"
}

# The stored password path (not the value!)
output "stored_password_path" {
  value       = gopass_secret.db_admin_password.path
  description = "Path where the generated password is stored in gopass"
}
