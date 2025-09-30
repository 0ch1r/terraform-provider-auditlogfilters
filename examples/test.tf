# Simple test configuration for local development
terraform {
  required_providers {
    auditlogfilter = {
      source = "0ch1r/auditlogfilter"
    }
  }
}

provider "auditlogfilter" {
  endpoint = "localhost:3306"
  username = "root"
  password = ""
  database = "mysql"
  tls      = "preferred"
}

# Test filter
resource "auditlogfilter_filter" "test_connection" {
  name = "test_connection_events"
  definition = jsonencode({
    filter = {
      class = {
        name = "connection"
        event = {
          name = ["connect", "disconnect", "change_user", "query_start", "query_end"]
        }
      }
    }
  })
}

# Test user assignment
resource "auditlogfilter_user_assignment" "test_assignment" {
  username    = "testuser"
  userhost    = "%"
  filter_name = auditlogfilter_filter.test_connection.name
}

output "test_results" {
  value = {
    filter_id   = auditlogfilter_filter.test_connection.filter_id
    filter_name = auditlogfilter_filter.test_connection.name
    assignment_id = auditlogfilter_user_assignment.test_assignment.id
  }
}

# Add another user assignment to test multi-user restoration
resource "auditlogfilter_user_assignment" "admin_assignment" {
  username    = "admin"
  userhost    = "localhost"
  filter_name = auditlogfilter_filter.test_connection.name
}
