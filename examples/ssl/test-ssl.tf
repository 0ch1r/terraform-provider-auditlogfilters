terraform {
  required_providers {
    auditlogfilters = {
      source = "0ch1r/auditlogfilters"
    }
  }
}

provider "auditlogfilters" {
  endpoint        = "localhost:3306"
  username        = "tfuser"
  password        = var.mysql_password
  database        = "mysql"
  tls_ca_file     = var.mysql_tls_ca
  tls_cert_file   = var.mysql_tls_cert
  tls_key_file    = var.mysql_tls_key
  tls_server_name = var.mysql_tls_server_name
  tls_skip_verify = var.mysql_tls_skip_verify
}

variable "mysql_password" {
  description = "MySQL password"
  type        = string
  sensitive   = true
  default     = "tfpass"
}

variable "mysql_tls_ca" {
  description = "Path to the CA certificate for MySQL TLS"
  type        = string
  default     = "../../scripts/.mysql-ssl/ca.pem"
}

variable "mysql_tls_cert" {
  description = "Path to the client certificate for MySQL TLS (optional)"
  type        = string
  default     = ""
}

variable "mysql_tls_key" {
  description = "Path to the client key for MySQL TLS (optional)"
  type        = string
  default     = ""
}

variable "mysql_tls_server_name" {
  description = "Server name for TLS verification (SNI)"
  type        = string
  default     = "percona-ssl"
}

variable "mysql_tls_skip_verify" {
  description = "Whether to skip TLS verification"
  type        = string
  default     = "false"
}

resource "auditlogfilters_filter" "connection_audit" {
  # Audits connection lifecycle events for all users.
  name = "connection_events_ssl"
  definition = jsonencode({
    filter = {
      class = {
        name = "connection"
        event = {
          name = ["connect", "disconnect", "change_user"]
        }
      }
    }
  })
}

resource "auditlogfilters_filter" "admin_audit" {
  # Audits connection and general events for admin/root users.
  name = "admin_activities_ssl"
  definition = jsonencode({
    filter = {
      class = [
        {
          name = "connection"
        },
        {
          name = "general"
          user = {
            name = ["admin", "root"]
          }
        }
      ]
    }
  })
}

resource "auditlogfilters_filter" "log_disabled" {
  # Disables audit logging when this filter is assigned.
  name = "log_disabled_ssl"
  definition = jsonencode({
    filter = {
      log = false
    }
  })
}

resource "auditlogfilters_filter" "droptable_audit" {
  # Audits drop table events.
  name = "drop_table_audit_ssl"
  definition = jsonencode({
    filter = {
      class = {
        name = "table_access"
        event = {
          name = ["drop"]
        }
      }
    }
  })
}

resource "auditlogfilters_user_assignment" "admin_connection" {
  username    = "admin"
  userhost    = "%.example.com"
  filter_name = auditlogfilters_filter.connection_audit.name
}

resource "auditlogfilters_user_assignment" "default_filter" {
  username    = "%"
  userhost    = "%"
  filter_name = auditlogfilters_filter.droptable_audit.name
}

output "connection_filter" {
  description = "Information about the connection audit filter"
  value = {
    id        = auditlogfilters_filter.connection_audit.id
    name      = auditlogfilters_filter.connection_audit.name
    filter_id = auditlogfilters_filter.connection_audit.filter_id
  }
}

output "admin_filter" {
  description = "Information about the admin audit filter"
  value = {
    id        = auditlogfilters_filter.admin_audit.id
    name      = auditlogfilters_filter.admin_audit.name
    filter_id = auditlogfilters_filter.admin_audit.filter_id
  }
}
