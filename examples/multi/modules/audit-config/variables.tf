variable "filter_name" {
  type        = string
  description = "Name of the audit log filter"
}

variable "filter_definition" {
  type        = any
  description = "JSON object for the filter; will be jsonencoded"
}

variable "user_assignments" {
  type = map(object({
    username = string
    userhost = string
  }))
  description = "Map of user assignments"
  default     = {}
}
