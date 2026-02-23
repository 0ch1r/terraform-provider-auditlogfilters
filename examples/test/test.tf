terraform {
  required_providers {
    auditlogfilters = {
      source = "0ch1r/auditlogfilters"
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

# Create a complex audit log filter
resource "auditlogfilters_filter" "complex_filter" {
  name = "complex_filter"
  definition = jsonencode({
    filter = {
      class = [
        {
          name = "query"
          event = {
            name = ["start", "status_end"]
            log = {
              or = [
                {
                  field = {
                    name  = "sql_command_id"
                    value = "create_db"
                  }
                },
                {
                  field = {
                    name  = "sql_command_id"
                    value = "drop_db"
                  }
                },
                {
                  field = {
                    name  = "sql_command_id"
                    value = "alter_user"
                  }
                }
              ]
            }
          }
        }
      ]
    }
  })
}

resource "auditlogfilters_filter" "testuser_ddl_dml" {
  name = "testuser_ddl_dml"
  definition = jsonencode({
    filter = {
      class = [
        {
          name = "table_access"
          event = {
            name = ["insert", "update", "delete"]
            log = {
              and = [
                {
                  field = {
                    name  = "table_database.str"
                    value = "prod"
                  }
                },
                {
                  or = [
                    {
                      field = {
                        name  = "table_name.str"
                        value = "employee"
                      }
                    },
                    {
                      field = {
                        name  = "table_name.str"
                        value = "projects"
                      }
                    }
                  ]
                }
              ]
            }
          }
        },
        {
          name = "general"
          event = {
            name = "status"
            log = {
              and = [
                {
                  field = {
                    name  = "general_command.str"
                    value = "Query"
                  }
                },
                {
                  or = [
                    {
                      field = {
                        name  = "general_sql_command.str"
                        value = "create_db"
                      }
                    },
                    {
                      field = {
                        name  = "general_sql_command.str"
                        value = "drop_db"
                      }
                    },
                    {
                      field = {
                        name  = "general_sql_command.str"
                        value = "alter_table"
                      }
                    },
                    {
                      field = {
                        name  = "general_sql_command.str"
                        value = "create_table"
                      }
                    },
                    {
                      field = {
                        name  = "general_sql_command.str"
                        value = "drop_table"
                      }
                    }
                  ]
                }
              ]
            }
          }
        }
      ]
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
