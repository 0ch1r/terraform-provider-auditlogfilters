output "filter_id" {
  value       = auditlogfilters_filter.this.filter_id
  description = "Filter ID"
}

output "filter_name" {
  value       = auditlogfilters_filter.this.name
  description = "Filter name"
}

output "assignments" {
  value = {
    for k, v in auditlogfilters_user_assignment.this : k => {
      id          = v.id
      username    = v.username
      userhost    = v.userhost
      filter_name = v.filter_name
    }
  }
  description = "Assignment details"
}
