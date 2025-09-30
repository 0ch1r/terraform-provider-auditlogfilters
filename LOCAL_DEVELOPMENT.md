# Local Development Guide

This guide explains how to develop and test the Terraform Provider for Audit Log Filters locally.

## Prerequisites

- Go 1.21+
- Terraform 1.0+
- GoReleaser (for releases)
- Percona Server 8.4+ with audit_log_filter component
- golangci-lint (for code quality)

## Quick Start

1. **Build the provider:**
   ```bash
   make build
   ```

2. **Set up dev overrides for local testing:**
   ```bash
   make dev-override
   ```

3. **Test with your Terraform configuration:**
   ```hcl
   terraform {
     required_providers {
       auditlogfilters = {
         source = "localhost/providers/auditlogfilters"
         version = "1.0.0"
       }
     }
   }

   provider "auditlogfilters" {
     endpoint = "localhost:3306"
     username = "root"
     password = "your_password"
     database = "mysql"
     tls      = "preferred"
   }
   ```

## Development Workflow

### Building and Testing

```bash
# Build the provider
make build

# Run unit tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Lint code
make lint

# Tidy dependencies
make tidy
```

### Acceptance Testing

**Important:** Acceptance tests require a running Percona Server 8.4+ instance with the audit_log_filter component installed.

1. **Start Percona Server 8.4+ with Docker:**
   ```bash
   docker run --name percona-test -e MYSQL_ROOT_PASSWORD=test123 \
     -p 3306:3306 -d percona/percona-server:8.4
   ```

2. **Install the audit_log_filter component:**
   ```bash
   mysql -h localhost -u root -ptest123 \
     -e "INSTALL COMPONENT 'file://component_audit_log_filter';"
   ```

3. **Run acceptance tests:**
   ```bash
   # Set environment variables
   export MYSQL_ENDPOINT="localhost:3306"
   export MYSQL_USERNAME="root"
   export MYSQL_PASSWORD="test123"
   export MYSQL_DATABASE="mysql"
   export MYSQL_TLS="false"

   # Run acceptance tests
   make testacc
   ```

### Local Testing with Terraform

1. **Set up dev overrides:**
   ```bash
   make dev-override
   ```

2. **Create a test configuration (test.tf):**
   ```hcl
   terraform {
     required_providers {
       auditlogfilters = {
         source = "localhost/providers/auditlogfilters"
       }
     }
   }

   provider "auditlogfilters" {
     endpoint = "localhost:3306"
     username = "root" 
     password = "test123"
     database = "mysql"
     tls      = "false"
   }

   resource "auditlogfilters_filter" "test" {
     name = "test_filter"
     definition = jsonencode({
       filter = {
         class = {
           name = "connection"
           event = {
             name = ["connect", "disconnect"]
           }
         }
       }
     })
   }
   ```

3. **Test the provider:**
   ```bash
   terraform init
   terraform plan
   terraform apply
   terraform destroy
   ```

4. **Clean up dev overrides when done:**
   ```bash
   make dev-clean
   ```

## Creating a Release

### Manual Release (Recommended)

1. **Create and push a tag:**
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. **The GitHub Actions workflow will automatically:**
   - Build binaries for multiple platforms
   - Sign the release with GPG
   - Create a GitHub release with assets
   - Include the Terraform Registry manifest

### Local Release Testing

1. **Test the release build locally:**
   ```bash
   make release-test
   ```

2. **Build for a specific platform:**
   ```bash
   make release-build
   ```

## Directory Structure

```
terraform-provider-auditlogfilters/
├── .github/workflows/     # GitHub Actions workflows
├── docs/                  # Generated documentation
├── examples/              # Example Terraform configurations  
├── internal/provider/     # Provider implementation
├── templates/             # Documentation templates
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── main.go                # Provider entry point
├── Makefile               # Development automation
└── LOCAL_DEVELOPMENT.md   # This file
```

## Useful Commands

```bash
# See all available make targets
make help

# Generate documentation
make docs

# Clean build artifacts
make clean

# Install provider locally
make install
```

## Debugging

### Enable Provider Logging
```bash
export TF_LOG=DEBUG
export TF_LOG_PROVIDER=DEBUG
terraform plan
```

### Test Specific Functions
```bash
# Test only filter resource
go test ./internal/provider/ -run TestAccAuditLogFilterResource -v

# Test with specific environment
MYSQL_ENDPOINT=localhost:3306 go test ./internal/provider/ -run TestProvider -v
```

## Troubleshooting

### Common Issues

1. **"Provider not found"** - Make sure dev overrides are set up correctly
2. **MySQL connection errors** - Verify MySQL is running and audit_log_filter component is installed
3. **Permission denied errors** - Check MySQL user privileges for audit log filter functions
4. **Test failures** - Ensure test database is clean before running acceptance tests

### Getting Help

- Check the [README.md](README.md) for general information
- Review the [docs/](docs/) directory for usage examples
- Look at [examples/](examples/) for working configurations
