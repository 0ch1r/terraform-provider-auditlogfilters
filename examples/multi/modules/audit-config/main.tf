terraform {
  required_providers {
    auditlogfilters = {
      source  = "0ch1r/auditlogfilters"
      version = ">= 0.1.0"
    }
  }
}

resource "auditlogfilters_filter" "this" {
  name       = var.filter_name
  definition = jsonencode(var.filter_definition)
}

resource "auditlogfilters_user_assignment" "this" {
  for_each = var.user_assignments

  username    = each.value.username
  userhost    = each.value.userhost
  filter_name = auditlogfilters_filter.this.name
}
