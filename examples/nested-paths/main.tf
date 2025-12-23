# Example: Using nested paths with gopass_env
# ==============================================
#
# This example demonstrates accessing secrets with deep/nested paths
# using dot-notation in Terraform configs.
#
# Prerequisites:
#   - gopass configured with your GPG key
#   - Secrets stored at nested paths like:
#     env/terraform/cloud/aws/
#     ├── REGION
#     ├── API/
#     │   ├── v2/
#     │   │   ├── ACCESS_KEY
#     │   │   └── SECRET_KEY
#     │   └── v1/
#     │       └── LEGACY_TOKEN
#     └── database/
#         └── prod/
#             ├── HOST
#             ├── PORT
#             └── PASSWORD

terraform {
  required_version = ">= 1.11.0" # OpenTofu 1.11+ for ephemeral support

  required_providers {
    gopass = {
      source  = "registry.opentofu.org/istr/gopass"
      version = "~> 0.1"
    }
  }
}

provider "gopass" {}

# -----------------------------------------------------------------------------
# Read nested credential structure
# -----------------------------------------------------------------------------

ephemeral "gopass_env" "cloud_aws" {
  path = "env/terraform/cloud/aws"
}

# Access flat (immediate child) secrets
output "region" {
  value       = ephemeral.gopass_env.cloud_aws.credentials.REGION
  ephemeral   = true
  description = "AWS region from gopass"
}

# Access deeply nested secrets using dot-notation
# Path: env/terraform/cloud/aws/API/v2/ACCESS_KEY
# Access: credentials.API.v2.ACCESS_KEY
output "api_v2_access_key" {
  value       = ephemeral.gopass_env.cloud_aws.credentials.API.v2.ACCESS_KEY
  ephemeral   = true
  sensitive   = true
  description = "AWS API v2 access key from gopass (nested path)"
}

output "api_v2_secret_key" {
  value       = ephemeral.gopass_env.cloud_aws.credentials.API.v2.SECRET_KEY
  ephemeral   = true
  sensitive   = true
  description = "AWS API v2 secret key from gopass (nested path)"
}

# Legacy API token (2 levels deep)
output "api_v1_legacy_token" {
  value       = ephemeral.gopass_env.cloud_aws.credentials.API.v1.LEGACY_TOKEN
  ephemeral   = true
  sensitive   = true
  description = "AWS API v1 legacy token (nested path)"
}

# Database credentials (2 levels deep)
output "db_host" {
  value       = ephemeral.gopass_env.cloud_aws.credentials.database.prod.HOST
  ephemeral   = true
  description = "Production database host from gopass (nested path)"
}

output "db_port" {
  value       = ephemeral.gopass_env.cloud_aws.credentials.database.prod.PORT
  ephemeral   = true
  description = "Production database port from gopass (nested path)"
}

output "db_password" {
  value       = ephemeral.gopass_env.cloud_aws.credentials.database.prod.PASSWORD
  ephemeral   = true
  sensitive   = true
  description = "Production database password from gopass (nested path)"
}

# -----------------------------------------------------------------------------
# Example: Use nested credentials in a provider
# -----------------------------------------------------------------------------

# Uncomment to use with AWS provider:
# provider "aws" {
#   region     = ephemeral.gopass_env.cloud_aws.credentials.REGION
#   access_key = ephemeral.gopass_env.cloud_aws.credentials.API.v2.ACCESS_KEY
#   secret_key = ephemeral.gopass_env.cloud_aws.credentials.API.v2.SECRET_KEY
# }

# -----------------------------------------------------------------------------
# Setup instructions for testing
# -----------------------------------------------------------------------------

# To set up test secrets in gopass:
#
# $ gopass insert env/terraform/cloud/aws/REGION
# # Enter: us-east-1
#
# $ gopass insert env/terraform/cloud/aws/API/v2/ACCESS_KEY
# # Enter: AKIAIOSFODNN7EXAMPLE
#
# $ gopass insert env/terraform/cloud/aws/API/v2/SECRET_KEY
# # Enter: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
#
# $ gopass insert env/terraform/cloud/aws/API/v1/LEGACY_TOKEN
# # Enter: legacy-token-value
#
# $ gopass insert env/terraform/cloud/aws/database/prod/HOST
# # Enter: db.example.com
#
# $ gopass insert env/terraform/cloud/aws/database/prod/PORT
# # Enter: 5432
#
# $ gopass insert env/terraform/cloud/aws/database/prod/PASSWORD
# # Enter: super-secret-db-password
#
# Then run:
# $ tofu init
# $ tofu plan
