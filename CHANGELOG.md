# Changelog

All notable changes to the Audit Log Filter Terraform provider will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1] - 2026-02-17

### Added (2026-01-04)

- **Security and Code Quality Improvements**: Addressed errorlint/gosec findings with tightened database and TLS handling
- **Enhanced Linter Configuration**: Optimized .golangci.yml with minimum required linters for improved code quality
- **Issue Process Improvements**: Created comprehensive issue template and clarified supported releases
- **Windows and FreeBSD Support Removal**: Focused on Linux/macOS deployment targets for better maintainability
- **Documentation Consolidation**: Consolidated LOCAL_DEVELOPMENT.md into docs/DEVELOPMENT.md for improved organization
- **Contribution Guidelines Updated**: Fixed Code of Conduct reference and issue tracker links in CONTRIBUTING.md
- **Comprehensive Examples**: Added detailed documentation examples and clarified licensing

### Added (2025-12-31)

- **Simplified CI/CD Workflow Example**: Updated Terraform CI/CD workflow template for improved production usage
- **Multi-Instance Example Module**: Added example demonstrating multiple audit log filter instances

### Added (2025-12-26)

- **Comprehensive TLS/SSL Support**: Full support for secure MySQL connections with:
  - Custom CA certificates (tls_ca_file)
  - Client certificates and keys (tls_cert_file, tls_key_file)
  - Server name verification (tls_server_name)
  - Insecure connection bypass (tls_skip_verify)
  - New tls.go module with robust certificate handling
  - Complete SSL testing example with certificates

### Added (2025-12-23)

- **Enhanced Database Operations**: Hardening of DB connections and improved stability

### Added (2025-12-08)

- **Improved Docker Health Checks**: Standardized test scripts and enhanced MySQL container health monitoring

### Added (2025-10-04)

- **Documentation Alignment**: Updated README.md make commands to match actual Makefile targets
- **Test Infrastructure Improvements**: Resolved acceptance test connectivity issues with MySQL

### Added (2025-10-03)

- **MySQL Docker Development Environment**: Comprehensive Docker management script for development
- **User Assignment Resources**: Enabled user assignment resources in test configurations
- **Provider Naming Consistency**: Resolved provider naming inconsistencies to use plural form

### Added (2025-09-30)

- **Complete Terraform Provider**: Initial full-featured provider for Percona Server audit log filters
- **Development Scripts Infrastructure**: Comprehensive scripts for local development, testing, and MySQL management
- **Simplified Development Workflow**: Streamlined local development and testing processes

### Core Features

- `auditlogfilters_filter` resource for managing audit log filters
- `auditlogfilters_user_assignment` resource for managing user-to-filter assignments
- Provider configuration with MySQL connection management
- JSON validation for filter definitions
- Import support for both resources
- Comprehensive documentation and examples
- Acceptance tests for local Percona Server testing
- Support for environment variable configuration
- GitHub Actions Workflows for CI/CD with lint, build, and acceptance tests
- Multi-platform releases with GoReleaser for multiple OS/arch combinations
- golangci-lint configuration with comprehensive rule set
- Issue Templates and PR templates for contribution management
- Dependabot for automated dependency updates
- Enhanced Update Behavior: Filter updates now use remove-then-recreate pattern with automatic user assignment restoration
- Comprehensive Warnings: Added detailed warnings about filter recreation impact on active sessions
- User Assignment Protection: Automatically preserve and restore all user assignments during filter updates

### Features

- **Filter Management**: Create, update, and delete audit log filters using MySQL audit_log_filter functions
- **User Assignments**: Assign specific filters to users or set default assignments
- **JSON Validation**: Validate filter definitions before applying
- **Import Support**: Import existing filters and assignments into Terraform state
- **Connection Management**: Robust MySQL connection handling with proper error handling
- **Environment Variables**: Support for configuration via environment variables
- **TLS Support**: Full featured TLS/SSL support with custom certificates and secure connections
- **CI/CD Integration**: Complete workflow templates for Terraform infrastructure as code deployments
- **Docker Development**: Integrated MySQL Docker environment for testing and development

### Security

- Sensitive password field marked as sensitive in provider configuration
- Proper secret management recommendations in documentation
- Connection validation and component verification
- TLS/SSL encryption support for data in transit
- Certificate-based authentication support

### Documentation

- Complete README with usage examples and best practices
- Development guide with setup and testing instructions
- Comprehensive filter definition examples
- Troubleshooting guide with common issues and solutions
- SSL/TLS configuration examples
- CI/CD workflow templates
- Development environment setup scripts

### Other

- Various bug fixes and improvements for production readiness
- Enhanced development environment with better Docker integration
- Improved testing infrastructure and CI/CD workflows
