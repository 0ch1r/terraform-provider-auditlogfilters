# Development Guide

This guide covers development setup, testing, and contribution guidelines for the Audit Log Filter Terraform provider.

## Development Environment Setup

### Prerequisites

1. **Go 1.21+**: Required for building the provider
2. **Terraform 1.0+**: For testing provider functionality
3. **Percona Server 8.4+**: With audit_log_filter component enabled
4. **Make**: For build automation (optional but recommended)

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/0ch1r/terraform-provider-auditlogfilter
cd terraform-provider-auditlogfilter

# Initialize Go modules
go mod tidy

# Install development dependencies
go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Project Structure

```
terraform-provider-auditlogfilter/
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

Acceptance tests require a running Percona Server instance:

```bash
# Set environment variables
export MYSQL_ENDPOINT=localhost:3306
export MYSQL_USERNAME=root
export MYSQL_PASSWORD=""

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

#### Code Generation

```bash
# Generate documentation and formatting
make generate

# Or manually
go generate ./...
terraform fmt -recursive ./examples/
```

## Local Development Testing

### Manual Testing Setup

1. **Build the provider locally:**
   ```bash
   go build -o terraform-provider-auditlogfilter
   ```

2. **Create a dev override file** (`~/.terraformrc`):
   ```hcl
   provider_installation {
     dev_overrides {
       "0ch1r/auditlogfilter" = "/path/to/terraform-provider-auditlogfilter"
     }
     direct {}
   }
   ```

3. **Test with example configuration:**
   ```bash
   cd examples
   terraform init
   terraform plan
   terraform apply
   ```

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
                    resource.TestCheckResourceAttr("auditlogfilter_filter.test", "name", "test_filter"),
                    resource.TestCheckResourceAttrSet("auditlogfilter_filter.test", "filter_id"),
                ),
            },
            // Import
            {
                ResourceName:      "auditlogfilter_filter.test",
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
    })
}
```

## Debugging

### Enable Debug Logging

```bash
TF_LOG=DEBUG terraform apply
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

## Release Process

### Version Management

1. Update version in relevant files
2. Update CHANGELOG.md
3. Tag the release
4. Build and publish artifacts

### Publishing to Registry

1. **Create GitHub release** with proper semver tag
2. **Registry will auto-publish** based on GitHub releases
3. **Update documentation** links if needed

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

### Runtime Issues

**Issue**: TLS connection problems
**Solution**: Verify TLS configuration matches server setup

**Issue**: Permission denied errors
**Solution**: Ensure MySQL user has required privileges for audit functions

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
