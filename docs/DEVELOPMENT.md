# Development Guide

This guide covers development setup, testing, and contribution guidelines for the Audit Log Filter Terraform provider.

## Development Environment Setup

### Prerequisites

1. **Go 1.21+**: Required for building the provider
2. **Terraform 1.0+**: For testing provider functionality
3. **Percona Server 8.4+**: With audit_log_filter component enabled
4. **Make**: For build automation (optional but recommended)
5. **GoReleaser**: For releases (optional)
6. **golangci-lint**: For code quality (optional)

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/0ch1r/terraform-provider-auditlogfilters
cd terraform-provider-auditlogfilters

# Initialize Go modules
go mod tidy

# Install development dependencies
go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

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
         source = "0ch1r/auditlogfilters"
         version = "1.0.0"
       }
     }
   }

   provider "auditlogfilters" {
     endpoint = "localhost:3306"
     username = "root"
     password = var.mysql_password
     database = "mysql"
     tls      = "preferred"
   }

   variable "mysql_password" {
     description = "MySQL root password"
     type        = string
     sensitive   = true
     default     = ""
   }
   ```

## Project Structure

```
terraform-provider-auditlogfilters/
├── main.go                          # Provider main entry point
├── go.mod                           # Go module definition
├── Makefile                         # Build automation
├── README.md                        # Project documentation
├── examples/                        # Usage examples
│   └── main.tf                      # Complete example configuration
├── docs/                            # Additional documentation
│   └── DEVELOPMENT.md               # This file
└── internal/
    └── provider/                    # Provider implementation
        ├── provider.go              # Main provider logic
        ├── provider_test.go         # Provider tests
        ├── audit_log_filter_resource.go      # Filter resource
        └── audit_log_user_assignment_resource.go  # User assignment resource
```

## Building and Testing

### Build the Provider

```bash
# Using Make (recommended)
make build

# Using Go directly
go build -v ./...
```

### Running Tests

#### Unit Tests

```bash
# Using Make
make test

# Using Go directly
go test -v ./...
```

#### Acceptance Tests

**Important:** Acceptance tests require a running Percona Server 8.4+ instance with the audit_log_filter component installed.

**Using Docker:**

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

**Running Tests:**

```bash
# Set environment variables
export MYSQL_ENDPOINT="localhost:3306"
export MYSQL_USERNAME="root"
export MYSQL_PASSWORD="test123"
export MYSQL_DATABASE="mysql"
export MYSQL_TLS="false"
export MYSQL_CONN_MAX_LIFETIME="5m"
export MYSQL_MAX_OPEN_CONNS="5"
export MYSQL_MAX_IDLE_CONNS="5"
export MYSQL_WAIT_TIMEOUT="10000"
export MYSQL_INNODB_LOCK_WAIT_TIMEOUT="1"
export MYSQL_LOCK_WAIT_TIMEOUT="60"

# Run acceptance tests
make testacc

# Or using Go directly
TF_ACC=1 go test -v ./... -run="TestAcc" -timeout 10m
```

### Code Quality

#### Linting

```bash
# Using Make
make lint

# Using golangci-lint directly
golangci-lint run
```

#### Code Formatting

```bash
# Using Make
make fmt

# Manually check formatting
make fmtcheck
```

#### Code Generation

```bash
# Generate documentation and formatting
make generate

# Or manually
go generate ./...
terraform fmt -recursive ./examples/
```

## Local Development Testing

### Using Make Commands (Recommended)

```bash
# See all available make targets
make help

# Set up dev overrides for local testing
make dev-override

# Test with your configuration (see Quick Start section for example)
# ...

# Clean up dev overrides when done
make dev-clean
```

### Manual Testing Setup

1. **Build the provider locally:**
   ```bash
   go build -o terraform-provider-auditlogfilters
   ```

2. **Create a dev override file** (`~/.terraformrc`):
   ```hcl
   provider_installation {
     dev_overrides {
       "0ch1r/auditlogfilters" = "/path/to/terraform-provider-auditlogfilters"
     }
     direct {}
   }
   ```

3. **Test with example configuration:**
   ```hcl
   terraform {
     required_providers {
       auditlogfilters = {
         source = "0ch1r/auditlogfilters"
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

   resource "auditlogfilters_filter" "connection_audit" {
     name = "connection_events"
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
   ```

4. **Test the provider:**
   ```bash
   terraform init
   terraform plan
   terraform apply
   terraform destroy
   ```

### Testing with SSL/TLS

1. **Start the SSL-enabled Percona container:**
   ```bash
   scripts/mysql-ssl-docker.ssh start
   ```

2. **Use the SSL example configuration:**
   ```bash
   cd examples/ssl
   terraform init
   terraform apply
   ```

   The SSL example uses:
   - `tls_ca_file` pointing at `scripts/.mysql-ssl/ca.pem`
   - Optional client cert/key at `scripts/.mysql-ssl/client-cert.pem` and `scripts/.mysql-ssl/client-key.pem`
   - `tls_server_name = "percona-ssl"`

### Database Setup for Testing

Ensure your Percona Server has the audit_log_filter component enabled:

```sql
-- Check if component is installed
SELECT * FROM mysql.component WHERE component_urn LIKE '%audit_log_filter%';

-- Install if needed (check Percona documentation for installation script)
-- INSTALL COMPONENT 'file://component_audit_log_filter';

-- Verify tables exist
SHOW TABLES IN mysql LIKE 'audit_log_%';
```

## Contributing

### Code Standards

1. **Go Formatting**: Use `go fmt` and follow Go conventions
2. **Error Handling**: Properly handle and wrap errors
3. **Testing**: Write tests for new functionality
4. **Documentation**: Update docs for new features
5. **Commits**: Use meaningful commit messages

### Pull Request Process

1. **Fork the repository** and create a feature branch
2. **Make your changes** with appropriate tests
3. **Run the full test suite** and ensure it passes
4. **Update documentation** if needed
5. **Submit a pull request** with a clear description

### Testing Guidelines

#### Unit Tests

- Test resource schema validation
- Test CRUD operations logic
- Mock database interactions where appropriate
- Use table-driven tests for multiple scenarios

#### Acceptance Tests

- Test complete resource lifecycle (Create, Read, Update, Delete)
- Test import functionality
- Test error conditions and validation
- Clean up resources after tests

### Example Test Structure

```go
func TestAccAuditLogFilterResource_basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Create and Read
            {
                Config: testAccFilterConfig("test_filter"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("auditlogfilters_filter.test", "name", "test_filter"),
                    resource.TestCheckResourceAttrSet("auditlogfilters_filter.test", "filter_id"),
                ),
            },
            // Import
            {
                ResourceName:      "auditlogfilters_filter.test",
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
    })
}
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

### Database Query Debugging

Enable MySQL query logging to see what the provider is executing:

```sql
SET GLOBAL general_log = 'ON';
SET GLOBAL log_output = 'TABLE';

-- View logged queries
SELECT * FROM mysql.general_log
WHERE command_type = 'Query'
AND argument LIKE '%audit_log_%'
ORDER BY event_time DESC;
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

## Common Issues and Solutions

### Provider Build Issues

**Issue**: Module dependency errors
**Solution**: Run `go mod tidy` and ensure Go version compatibility

**Issue**: Missing development tools
**Solution**: Install required tools as shown in setup

### Testing Issues

**Issue**: Acceptance tests fail with connection errors
**Solution**: Verify MySQL/Percona Server is running and accessible

**Issue**: Component not available error
**Solution**: Ensure audit_log_filter component is properly installed

**Issue**: Test failures
**Solution**: Ensure test database is clean before running acceptance tests

### Runtime Issues

**Issue**: Provider not found
**Solution**: Make sure dev overrides are set up correctly

**Issue**: MySQL connection errors
**Solution**: Verify MySQL is running and audit_log_filter component is installed

**Issue**: Permission denied errors
**Solution**: Check MySQL user privileges for audit log filter functions

**Issue**: TLS connection problems
**Solution**: Verify TLS configuration matches server setup and certificate SANs. For local SSL Docker testing, use `scripts/mysql-ssl-docker.ssh` and set `tls_server_name = "percona-ssl"` or update certificates to include the hostname you connect to.

## Additional Resources

- [Terraform Plugin Framework Documentation](https://developer.hashicorp.com/terraform/plugin/framework)
- [Percona Server Audit Log Filter Documentation](https://docs.percona.com/percona-server/8.4/audit-log-filter-overview.html)
- [Go Testing Documentation](https://pkg.go.dev/testing)
- [MySQL Connector/Go Documentation](https://github.com/go-sql-driver/mysql)

## Support and Questions

For development questions or issues:

1. Check existing GitHub issues
2. Review this documentation
3. Create a new issue with details
4. Include relevant logs and configuration
