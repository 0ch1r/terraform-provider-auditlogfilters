# Changelog

All notable changes to the Audit Log Filter Terraform provider will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **GitHub Actions Workflows**: Comprehensive CI/CD with lint, build, and acceptance tests
- **Multi-platform Releases**: Automated releases with GoReleaser for multiple OS/arch combinations
- **Code Quality**: golangci-lint configuration with comprehensive rule set
- **Issue Templates**: Bug report and feature request templates for better issue management
- **PR Template**: Structured pull request template with testing and documentation checklists
- **Dependabot**: Automated dependency updates for Go modules and GitHub Actions
- **Enhanced Update Behavior**: Filter updates now use remove-then-recreate pattern with automatic user assignment restoration
- **Comprehensive Warnings**: Added detailed warnings about filter recreation impact on active sessions
- **User Assignment Protection**: Automatically preserve and restore all user assignments during filter updates
- Initial implementation of Audit Log Filter provider for Percona Server 8.4+
- `auditlogfilter_filter` resource for managing audit log filters
- `auditlogfilter_user_assignment` resource for managing user-to-filter assignments
- Provider configuration with MySQL connection management
- JSON validation for filter definitions
- Import support for both resources
- Comprehensive documentation and examples
- Acceptance tests for local Percona Server testing
- Support for environment variable configuration
- TLS connection support

### Features
- **Filter Management**: Create, update, and delete audit log filters using MySQL audit_log_filter functions
- **User Assignments**: Assign specific filters to users or set default assignments
- **JSON Validation**: Validate filter definitions before applying
- **Import Support**: Import existing filters and assignments into Terraform state
- **Connection Management**: Robust MySQL connection handling with proper error handling
- **Environment Variables**: Support for configuration via environment variables
- **TLS Support**: Configurable TLS options for secure connections

### Security
- Sensitive password field marked as sensitive in provider configuration
- Proper secret management recommendations in documentation
- Connection validation and component verification

### Documentation
- Complete README with usage examples and best practices
- Development guide with setup and testing instructions
- Comprehensive filter definition examples
- Troubleshooting guide with common issues and solutions

## [1.0.0] - TBD

### Added
- **GitHub Actions Workflows**: Comprehensive CI/CD with lint, build, and acceptance tests
- **Multi-platform Releases**: Automated releases with GoReleaser for multiple OS/arch combinations
- **Code Quality**: golangci-lint configuration with comprehensive rule set
- **Issue Templates**: Bug report and feature request templates for better issue management
- **PR Template**: Structured pull request template with testing and documentation checklists
- **Dependabot**: Automated dependency updates for Go modules and GitHub Actions
- **Enhanced Update Behavior**: Filter updates now use remove-then-recreate pattern with automatic user assignment restoration
- **Comprehensive Warnings**: Added detailed warnings about filter recreation impact on active sessions
- **User Assignment Protection**: Automatically preserve and restore all user assignments during filter updates
- Initial stable release
- Full feature set for audit log filter management
- Complete test coverage
- Production-ready documentation

### Changed
- N/A (Initial release)

### Deprecated
- N/A (Initial release)

### Removed
- N/A (Initial release)

### Fixed
- N/A (Initial release)

### Security
- N/A (Initial release)
