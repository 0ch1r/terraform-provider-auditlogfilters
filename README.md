# Terraform Provider: Audit Log Filter for Percona Server

A Terraform provider for managing Percona Server 8.4+ audit log filters using the `audit_log_filter` component.

![Build Status](https://img.shields.io/badge/status-development-orange)
![License](https://img.shields.io/badge/license-MPL--2.0-blue)
![Terraform](https://img.shields.io/badge/terraform-%3E%3D1.0-blueviolet)
![Go](https://img.shields.io/badge/go-%3E%3D1.21-blue)

## Overview

This provider enables Infrastructure as Code management of Percona Server audit log filters, allowing you to:

- Create, update, and delete audit log filters with JSON definitions
- Assign users to specific filters for targeted auditing
- Import existing filters and assignments into Terraform state
- Manage audit configuration through standard Terraform workflows

## Requirements

- **Terraform**: >= 1.0
- **Go**: >= 1.21 (for development)
- **Percona Server**: >= 8.4 with `audit_log_filter` component enabled
- **MySQL Driver**: Compatible with mysql 8.0 protocol

## Installation

### From Terraform Registry (Coming Soon)

```hcl
terraform {
  required_providers {
    auditlogfilters = {
      source  = "0ch1r/auditlogfilters"
      version = "~> 1.0"
    }
  }
}
```

### Local Development

```bash
git clone https://github.com/0ch1r/terraform-provider-auditlogfilters
cd terraform-provider-auditlogfilters
go build -o terraform-provider-auditlogfilters
```

## Quick Start

### Provider Configuration

```hcl
provider "auditlogfilters" {
  endpoint = "localhost:3306"
  username = "root"
  password = "your-password"
  database = "mysql"
  tls      = "preferred"
}
```

### Environment Variables

The provider supports the following environment variables:

- `MYSQL_ENDPOINT` - MySQL server endpoint
- `MYSQL_USERNAME` - MySQL username
- `MYSQL_PASSWORD` - MySQL password
- `MYSQL_DATABASE` - Database name (default: "mysql")
- `MYSQL_TLS` - TLS configuration
- `MYSQL_TLS_CA` - TLS CA certificate file path
- `MYSQL_TLS_CERT` - TLS client certificate file path
- `MYSQL_TLS_KEY` - TLS client key file path
- `MYSQL_TLS_SERVER_NAME` - TLS server name override
- `MYSQL_TLS_SKIP_VERIFY` - Skip TLS certificate verification
- `MYSQL_CONN_MAX_LIFETIME` - Maximum connection lifetime (default: "5m")
- `MYSQL_MAX_OPEN_CONNS` - Maximum open connections (default: "5")
- `MYSQL_MAX_IDLE_CONNS` - Maximum idle connections (default: "5")
- `MYSQL_WAIT_TIMEOUT` - Session wait_timeout in seconds (default: "10000")
- `MYSQL_INNODB_LOCK_WAIT_TIMEOUT` - Session innodb_lock_wait_timeout in seconds (default: "1")
- `MYSQL_LOCK_WAIT_TIMEOUT` - Session lock_wait_timeout in seconds (default: "60")

### SSL/TLS Example (Docker)

The provider includes comprehensive TLS/SSL support for secure MySQL connections. Start an SSL-enabled container:

```bash
make test-with-mysql
```

Provider configuration using the generated CA and optional client certificates:

```hcl
provider "auditlogfilters" {
  endpoint        = "localhost:3307"
  username        = "tfuser"
  password        = "tfpass"
  database        = "mysql"
  tls_ca_file     = "./scripts/.mysql-ssl/ca.pem"
  tls_cert_file   = "./scripts/.mysql-ssl/client-cert.pem"
  tls_key_file    = "./scripts/.mysql-ssl/client-key.pem"
  tls_server_name = "percona-ssl"
}
```

See `examples/ssl/` for a complete TLS configuration example.

### Basic Usage

#### Create an Audit Log Filter

```hcl
resource "auditlogfilters_filter" "login_filter" {
  name = "login_events"
  definition = jsonencode({
    "filter": {
      "class": {
        "name": "connection",
        "event": {
          "name": ["connect", "disconnect"]
        }
      }
    }
  })
}
```

#### Assign Filter to User

```hcl
resource "auditlogfilters_user_assignment" "app_user" {
  username    = "app_user"
  userhost    = "%"
  filter_name = auditlogfilters_filter.login_filter.name
}
```

## Resource Documentation

### auditlogfilters_filter

Manages audit log filters using the MySQL `audit_log_filter_set_filter()` function.

#### Arguments

- `name` (Required, String) - Unique name for the audit log filter. Changing this forces recreation.
- `definition` (Required, String) - JSON definition of the filter rules according to MySQL audit log filter syntax.

#### Attributes

- `id` (String) - Unique identifier (same as name)
- `filter_id` (Number) - Internal MySQL filter ID

#### Import

```bash
terraform import auditlogfilters_filter.example filter_name
```

### auditlogfilters_user_assignment

Manages user assignments to audit log filters using the MySQL `audit_log_filter_set_user()` function.

#### Arguments

- `username` (Required, String) - MySQL username. Use "%" for default assignment. Changing this forces recreation.
- `userhost` (Optional, String) - Host pattern. Defaults to "%". Changing this forces recreation.
- `filter_name` (Required, String) - Name of the filter to assign.

#### Attributes

- `id` (String) - Unique identifier (username@userhost)

#### Import

```bash
terraform import auditlogfilters_user_assignment.example "username@hostname"
terraform import auditlogfilters_user_assignment.default "%"
```

## Filter Definition Examples

### Connection Events Only

```json
{
  "filter": {
    "class": {
      "name": "connection",
      "event": {
        "name": ["connect", "disconnect", "change_user"]
      }
    }
  }
}
```

### Query Events with User Filter

```json
{
  "filter": {
    "class": [
      {
        "name": "connection"
      },
      {
        "name": "general",
        "user": {
          "name": ["sensitive_user", "admin"]
        }
      }
    ]
  }
}
```

### Complex Filter with Multiple Conditions

```json
{
  "filter": {
    "class": {
      "name": "general",
      "event": {
        "name": "status",
        "log": false
      }
    }
  },
  "filter": {
    "class": {
      "name": "table_access",
      "event": {
        "name": ["read", "insert", "update", "delete"]
      },
      "database": {
        "name": "sensitive_db"
      }
    }
  }
}
```

## Best Practices

### Filter Management

1. **Use descriptive names** for filters to make management easier
2. **Test filters carefully** before applying to production users
3. **Consider performance impact** of comprehensive filters
4. **Use version control** for filter definitions

### User Assignments

1. **Start with specific users** before applying default filters
2. **Use host patterns judiciously** to avoid overly broad assignments
3. **Monitor audit log size** after applying new assignments
4. **Document filter assignments** for operational teams

### State Management

1. **Import existing resources** before managing them with Terraform
2. **Use remote state** for team collaboration
3. **Plan changes carefully** as filter updates affect active sessions
4. **Test in non-production** environments first

## Examples

This repository includes several examples demonstrating different use cases:

### SSL/TLS Example

Demonstrates secure connection to MySQL using TLS certificates. Located in `examples/ssl/`.

### Multi-Instance Example

Shows how to manage audit log filters across multiple instances using a reusable module. Located in `examples/multi/`.

### CI/CD Example

Provides a complete GitHub Actions workflow example for a Terraform CI/CD pipeline. Located in `examples/cicd/.github/workflows/terraform.yml`.

## Development

### Building the Provider

```bash
go build -o terraform-provider-auditlogfilters
```

### Running Tests

```bash
# Unit tests
make test

# Acceptance tests (requires running MySQL instance)
make testacc
```

### Code Generation

```bash
go generate ./...
```

## Testing with Local Percona Server

The provider includes comprehensive acceptance tests that work with a local Percona Server instance.

### Prerequisites

1. Percona Server 8.4+ running locally
2. `audit_log_filter` component installed and enabled
3. MySQL root access without password (for testing)

### Running Acceptance Tests

```bash
export MYSQL_ENDPOINT=localhost:3306
export MYSQL_USERNAME=root
export MYSQL_PASSWORD=""
make testacc
```

## Troubleshooting

### Common Issues

**Provider fails to connect**
- Verify MySQL endpoint and credentials
- Check firewall settings
- Ensure TLS configuration matches server setup

**Component not available error**
- Verify `audit_log_filter` component is installed
- Check component status: `SELECT * FROM mysql.component;`
- Install component if needed

**Filter creation fails**
- Validate JSON syntax in filter definition
- Check MySQL error logs for detailed information
- Verify user has required privileges

**User assignment fails**
- Ensure target filter exists before assignment
- Check for existing conflicting assignments
- Verify user specification format

### Debug Mode

Run the provider in debug mode for detailed logging:

```bash
TF_LOG=DEBUG terraform apply
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

### Development Environment

```bash
# Clone and setup
git clone https://github.com/0ch1r/terraform-provider-auditlogfilters
cd terraform-provider-auditlogfilters
go mod tidy

# Install development dependencies
make dev-setup

# Build and test
make build
make test
```

## License

This project is licensed under the Mozilla Public License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- HashiCorp for the Terraform Plugin Framework
- Percona for the MySQL audit log filter functionality
- The Go and Terraform communities

## Links

- [Percona Server Audit Log Filter Documentation](https://docs.percona.com/percona-server/8.4/audit-log-filter-overview.html)
- [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
- [MySQL Audit Log Filter Reference](https://dev.mysql.com/doc/refman/8.0/en/audit-log-filter-installation.html)

## Update Behavior

### Filter Updates

⚠️ **Important**: MySQL audit log filters cannot be updated in-place. When you modify a filter's definition, the provider will:

1. **Remove** the existing filter (which also removes all user assignments)
2. **Recreate** the filter with the new definition  
3. **Restore** all user assignments that were affected

This process includes comprehensive warnings:
- **"Filter Update Requires Recreation"** - Alerts you that the filter will be recreated
- **"Restoring User Assignments"** - Shows how many user assignments are being restored

### Impact on Active Sessions

- Sessions using the filter may experience a brief interruption
- Sessions may need to reconnect to pick up the new filter rules
- The provider automatically restores all user assignments to maintain consistency

### Example Update Flow

```hcl
resource "auditlogfilters_filter" "example" {
  name = "connection_events"
  definition = jsonencode({
    filter = {
      class = {
        name = "connection"
        event = {
          name = ["connect", "disconnect"]  # Adding more events will trigger recreation
        }
      }
    }
  })
}
```

When you modify the `definition`, OpenTofu will show:
```
╷
│ Warning: Filter Update Requires Recreation
│ MySQL audit log filters cannot be updated in-place. The provider will remove 
│ the existing filter and recreate it with the new definition.
╷
│ Warning: Restoring User Assignments  
│ Restoring N user assignments that were affected by the filter update.
╵
```

## CI/CD and Workflows

### Provider Workflows

The repository includes CI/CD for provider releases and validation:

#### Release (`.github/workflows/release.yml`)
- **GoReleaser**: Automated releases with signed binaries
- **Multi-platform**: Builds for multiple OS/architecture combinations
- **GPG Signing**: Signs release artifacts for security
- Triggered on version tags (`v*`) or manual workflow dispatch

### Example Project Workflow

An example Terraform CI/CD workflow is provided for projects that use this provider:

#### Terraform CI/CD Example (`examples/cicd/.github/workflows/terraform.yml`)
- **Validate**: Checks Terraform formatting, validation, and linting
- **Plan**: Generates and uploads Terraform plan artifacts
- **Comment**: Posts plan summaries to pull requests
- **Apply**: Applies changes when merged to the `main` branch
- Uses TFLint for policy checking and validation

### Status Badges

Add these badges to your repository README:

```markdown
[![Tests](https://github.com/0ch1r/terraform-provider-auditlogfilters/actions/workflows/test.yml/badge.svg)](https://github.com/0ch1r/terraform-provider-auditlogfilters/actions/workflows/test.yml)
[![Acceptance Tests](https://github.com/0ch1r/terraform-provider-auditlogfilters/actions/workflows/acceptance.yml/badge.svg)](https://github.com/0ch1r/terraform-provider-auditlogfilters/actions/workflows/acceptance.yml)
[![Validate](https://github.com/0ch1r/terraform-provider-auditlogfilters/actions/workflows/validate.yml/badge.svg)](https://github.com/0ch1r/terraform-provider-auditlogfilters/actions/workflows/validate.yml)
[![golangci-lint](https://github.com/0ch1r/terraform-provider-auditlogfilters/actions/workflows/test.yml/badge.svg)](https://github.com/0ch1r/terraform-provider-auditlogfilters/actions/workflows/test.yml)
```

### Local Development

Use the provided Makefile for local development:

**To see all available make commands:**
```bash
make help
```

```bash
# Build the provider
make build

# Run tests
make test

# Run acceptance tests (requires MySQL)
make testacc

# Run tests with coverage report
make test-coverage

# Lint code
make lint

# Format code
make fmt

# Tidy dependencies
make tidy

# Generate documentation
make docs

# Set up dev overrides for local testing
make dev-override

# Remove dev overrides
make dev-clean

# Test with local MySQL container
make test-with-mysql

# Clean build artifacts
make clean
```

### Release Process

**Makefile Release Commands:**

```bash
# Test release process without publishing
make release-test

# Build release binaries locally
make release-build

# Create a release (requires VERSION)
make release VERSION=v1.0.0

# Test release dry run (requires VERSION)
make release-dry-run VERSION=v1.0.0
```

1. **Update Version**: Update version numbers in relevant files
2. **Update Changelog**: Add changes to CHANGELOG.md
3. **Create Tag**: Create and push a git tag (e.g., `v1.0.0`)
4. **Automated Release**: GitHub Actions will automatically build and release
5. **Registry**: Provider will be available in the Terraform Registry

### Security

- **GPG Signing**: All releases are GPG signed
- **Dependabot**: Automated dependency updates
- **Security Scanning**: Regular vulnerability scans
- **Code Review**: Required PR reviews before merge
