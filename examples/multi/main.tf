terraform {
  required_providers {
    auditlogfilters = {
      source  = "0ch1r/auditlogfilters"
      version = ">= 0.1.0"
    }
  }
}

locals {
  # Two local MySQL instances
  instances = {
    instance1 = { endpoint = "localhost:3306" }
    instance2 = { endpoint = "localhost:3307" }
  }

  # Shared credentials and DB
  creds = {
    username = "root"
    password = "t00r"
    database = "mysql"
    tls      = "preferred"
  }

  # Simple connection filter
  filter_definition = {
    filter = {
      class = {
        name  = "connection"
        event = { name = ["connect", "disconnect"] }
      }
    }
  }

  # Single user assignment
  user_assignments = {
    appuser = {
      username = "appuser"
      userhost = "%"
    }
  }
}

# Provider aliases (Terraform requires these to be static)
provider "auditlogfilters" {
  alias    = "instance1"
  endpoint = local.instances.instance1.endpoint
  username = local.creds.username
  password = local.creds.password
  database = local.creds.database
  tls      = local.creds.tls
}

provider "auditlogfilters" {
  alias    = "instance2"
  endpoint = local.instances.instance2.endpoint
  username = local.creds.username
  password = local.creds.password
  database = local.creds.database
  tls      = local.creds.tls
}

# Reuse the audit-config module for each instance
module "instance1_audit" {
  source    = "./modules/audit-config"
  providers = { auditlogfilters = auditlogfilters.instance1 }

  filter_name       = "connection_events"
  filter_definition = local.filter_definition
  user_assignments  = local.user_assignments
}

module "instance2_audit" {
  source    = "./modules/audit-config"
  providers = { auditlogfilters = auditlogfilters.instance2 }

  filter_name       = "connection_events"
  filter_definition = local.filter_definition
  user_assignments  = local.user_assignments
}

output "filters" {
  value = {
    instance1 = {
      endpoint  = local.instances.instance1.endpoint
      filter_id = module.instance1_audit.filter_id
    }
    instance2 = {
      endpoint  = local.instances.instance2.endpoint
      filter_id = module.instance2_audit.filter_id
    }
  }
}
