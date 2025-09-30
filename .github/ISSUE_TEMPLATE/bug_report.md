---
name: Bug Report
about: Report a bug in the audit log filter provider
title: '[BUG] '
labels: 'bug'
assignees: '0ch1r'
---

## Bug Description
A clear and concise description of what the bug is.

## To Reproduce
Steps to reproduce the behavior:
1. Provider configuration: '...'
2. Resource configuration: '...'
3. Run terraform command: '...'
4. See error

## Expected Behavior
A clear and concise description of what you expected to happen.

## Actual Behavior
What actually happened, including full error messages.

```
paste error messages here
```

## Configuration Files
Please provide your Terraform configuration files (redact any sensitive information):

```hcl
# Provider configuration
provider "auditlogfilter" {
  # your configuration
}

# Resource configuration
resource "auditlogfilter_filter" "example" {
  # your configuration
}
```

## Environment
- Provider Version: [e.g. v1.0.0]
- Terraform Version: [e.g. 1.5.0]
- Percona Server Version: [e.g. 8.4.6-6]
- Operating System: [e.g. Ubuntu 22.04]

## Debug Information
Please run with `TF_LOG=DEBUG` and provide relevant debug output:

```
paste debug output here (redact sensitive information)
```

## Additional Context
Add any other context about the problem here, such as:
- Recent changes to your infrastructure
- Similar configurations that work
- Workarounds you've tried
