default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./internal/provider/ -v $(TESTARGS) -timeout 120m

# Build provider for local development
.PHONY: build
build:
	go build -o terraform-provider-auditlogfilters

# Install provider locally for development
.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/localhost/providers/auditlogfilters/1.0.0/linux_amd64
	cp terraform-provider-auditlogfilters ~/.terraform.d/plugins/localhost/providers/auditlogfilters/1.0.0/linux_amd64/

# Create dev overrides for local testing
.PHONY: dev-override
dev-override:
	@echo 'provider_installation {' > ~/.terraformrc
	@echo '  dev_overrides {' >> ~/.terraformrc
	@echo '    "localhost/providers/auditlogfilters" = "'$(PWD)'"' >> ~/.terraformrc
	@echo '  }' >> ~/.terraformrc
	@echo '  direct {}' >> ~/.terraformrc
	@echo '}' >> ~/.terraformrc
	@echo "Dev overrides created in ~/.terraformrc"

# Remove dev overrides
.PHONY: dev-clean
dev-clean:
	@rm -f ~/.terraformrc
	@echo "Dev overrides removed"

# Generate documentation
.PHONY: docs
docs:
	tfplugindocs generate --provider-name auditlogfilters --rendered-provider-name "Audit Log Filter"

# Clean build artifacts
.PHONY: clean
clean:
	rm -f terraform-provider-auditlogfilters
	rm -rf dist/

# Test basic functionality
.PHONY: test
test:
	go test ./internal/provider/ -v

# Lint code
.PHONY: lint
lint:
	golangci-lint run

# Format code
.PHONY: fmt
fmt:
	go fmt ./...
	terraform fmt -recursive ./examples/

# Tidy dependencies
.PHONY: tidy
tidy:
	go mod tidy

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	go test ./internal/provider/ -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Build for release (local testing)
.PHONY: release-build
release-build:
	goreleaser build --clean --single-target

# Create a test release (without publishing)
.PHONY: release-test
release-test:
	goreleaser release --clean --skip-publish

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          - Build the provider binary"
	@echo "  install        - Install provider locally"
	@echo "  dev-override   - Setup dev overrides for local testing"
	@echo "  dev-clean      - Remove dev overrides"
	@echo "  test           - Run unit tests"
	@echo "  testacc        - Run acceptance tests (requires MySQL)"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  docs           - Generate documentation"
	@echo "  lint           - Run linter"
	@echo "  fmt            - Format code"
	@echo "  tidy           - Tidy go modules"
	@echo "  clean          - Clean build artifacts"
	@echo "  release-build  - Build release binaries locally"
	@echo "  release-test   - Test release process without publishing"
	@echo "  help           - Show this help message"

# Scripts integration
.PHONY: dev-setup
dev-setup:
	@./scripts/dev-setup.sh

.PHONY: test-local
test-local:
	@./scripts/test-local.sh

.PHONY: test-with-mysql
test-with-mysql:
	@./scripts/test-local.sh --with-mysql --acceptance

.PHONY: release-dry-run
release-dry-run:
	@echo "Usage: make release-dry-run VERSION=v1.0.0"
	@if [ -z "$(VERSION)" ]; then echo "ERROR: VERSION is required"; exit 1; fi
	@./scripts/release.sh $(VERSION) --dry-run

.PHONY: release
release:
	@echo "Usage: make release VERSION=v1.0.0"
	@if [ -z "$(VERSION)" ]; then echo "ERROR: VERSION is required"; exit 1; fi
	@./scripts/release.sh $(VERSION)
