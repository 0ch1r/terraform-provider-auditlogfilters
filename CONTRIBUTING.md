# Contributing to terraform-provider-auditlogfilters

Thank you for your interest in contributing to the terraform-provider-auditlogfilters project! This document provides information about contributing to this project and outlines the licensing terms.

## License

By contributing to this project, you agree that your contributions will be licensed under the same license as the project. This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Code of Conduct

This project follows the [Contributor Covenant](https://www.contributor-covenant.org/version/3/0/code_of_conduct/). Please be respectful and professional in all interactions.

## Development Setup

1. Clone the repository:

   ```bash
   git clone https://github.com/yourusername/terraform-provider-auditlogfilters.git
   cd terraform-provider-auditlogfilters
   ```

2. Install dependencies:

   ```bash
   make dev-override
   go mod tidy
   ```

3. Build the provider:
   ```bash
   make build
   ```

## Running Tests

1. Run unit tests:

   ```bash
   make test
   ```

2. Run tests with coverage:

   ```bash
   make test-coverage
   ```

3. Run acceptance tests (requires MySQL with audit_log_filter component):
   ```bash
   make testacc
   ```

## Code Style

We use the following tools to maintain code quality:

- `go fmt` for standard Go formatting
- `golanci-lint` for code quality checks (`make lint`)
- 140 character line limits
- Standard Go error handling conventions

## Submitting Changes

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for your changes
5. Ensure all tests pass (`make test`)
6. Ensure code formatting is correct (`make fmt`)
7. Commit your changes with a clear commit message
8. Push to the branch (`git push origin feature/amazing-feature`)
9. Open a Pull Request

## Liability Protection and Disclaimer

This provider manages database audit log filters, which are critical security components. Contributors should understand:

1. **No Warranty**: The software is provided "AS IS", without warranty of any kind, express or implied, including but not limited to the warranties of merchantability, fitness for a particular purpose and noninfringement.

2. **Limited Liability**: In no event shall the authors or copyright holders be liable for any claim, damages or other liability, whether in an action of contract, tort or otherwise, arising from, out of or in connection with the software or the use or other dealings in the software.

3. **Security Considerations**: This provider handles database credentials and manages security-critical configurations. Contributors must:
   - Never log sensitive data
   - Follow security best practices
   - Validate all input parameters
   - Use prepared statements to prevent SQL injection

4. **Production Use**: Contributors should clearly mark any features that are experimental and not suitable for production use.

5. **User Responsibility**: Users of this provider are responsible for:
   - Validating configurations for their specific use cases
   - Maintaining appropriate database security
   - Testing in non-production environments before deployment
   - Maintaining backups and recovery procedures

## Bug Reports and Feature Requests

Bug reports and feature requests are welcome! Please use the [issue tracker](https://github.com/0ch1r/terraform-provider-auditlogfilters/issues) to submit bug reports or feature requests.

When reporting bugs, please include:

- Terraform and provider versions
- Configuration used
- Error messages and stack traces
- Expected vs actual behavior
- Steps to reproduce the issue

## Questions

If you have questions about contributing, please open an issue with the "question" label, and we'll be happy to help!

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for detail

