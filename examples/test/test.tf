terraform {
  required_providers {
    auditlogfilters = {
      source  = "0ch1r/auditlogfilters"
    }
  }
}

# Configure the Audit Log Filter Provider
provider "auditlogfilters" {
  endpoint = "localhost:3306"
  username = "root"
  password = var.mysql_password
  database = "mysql"
  tls      = "preferred"
}

# Example MySQL root password variable
variable "mysql_password" {
  description = "MySQL root password"
  type        = string
  sensitive   = true
  default     = ""
}

# Create an audit log filter for connection events
# resource "auditlogfilters_filter" "connection_audit" {
#   name = "connection_events"
#   definition = jsonencode({
#     filter = {
#       class = {
#         name = "connection"
#         event = {
#           name = ["connect", "disconnect", "change_user"]
#         }
#       }
#     }
#   })
# }

# Create an audit log filter for specific user activities
# resource "auditlogfilters_filter" "admin_audit" {
#   name = "admin_activities"
#   definition = jsonencode({
#     filter = {
#       class = [
#         {
#           name = "connection"
#         },
#         {
#           name = "general"
#           user = {
#             name = ["admin", "root"]
#           }
#         }
#       ]
#     }
#   })
# }

# Create an audit log that disables all logging
resource "auditlogfilters_filter" "log_disabled" {
  name = "logging_disabled"
  definition = jsonencode({
    filter = {
      log = false
    }
  })
}

# Assign the connection filter to a specific user
# resource "auditlogfilters_user_assignment" "admin_connection" {
#   username    = "admin"
#   userhost    = "%.example.com"
#  filter_name = auditlogfilters_filter.connection_audit.name
# }

# Set default filter for all users
# resource "auditlogfilters_user_assignment" "default_filter" {
#   username    = "%"
#   userhost    = "%"
#   filter_name = auditlogfilters_filter.admin_audit.name
# }

# Output filter information
# output "connection_filter" {
#   description = "Information about the connection audit filter"
#   value = {
#     id          = auditlogfilters_filter.connection_audit.id
#     name        = auditlogfilters_filter.connection_audit.name
#     filter_id   = auditlogfilters_filter.connection_audit.filter_id
#   }
# }
#
# output "admin_filter" {
#   description = "Information about the admin audit filter"
#   value = {
#     id          = auditlogfilters_filter.admin_audit.id
#     name        = auditlogfilters_filter.admin_audit.name
#     filter_id   = auditlogfilters_filter.admin_audit.filter_id
#   }
# }
