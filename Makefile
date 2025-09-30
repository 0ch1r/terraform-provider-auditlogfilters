HOSTNAME=github.com
NAMESPACE=0ch1r
NAME=auditlogfilter
BINARY=terraform-provider-${NAME}
VERSION?=0.1.0
OS_ARCH?=linux_amd64

default: install

build:
	go build -v ./...

install: build
	go install .

lint:
	golangci-lint run

# Run unit tests
test:
	go test -v -cover -timeout=120s -parallel=4 ./...

# Run acceptance tests
testacc:
	TF_ACC=1 go test -v -cover -timeout 10m ./...

# Run tests with race detection
testrace:
	go test -v -race -cover -timeout=120s -parallel=4 ./...

# Generate documentation
generate:
	go generate ./...

# Format code
fmt:
	gofmt -s -w .
	go mod tidy

# Check formatting
fmtcheck:
	@gofmt -s -l . | grep -q '.*' && echo "Files not formatted correctly:" && gofmt -s -l . || echo "All files formatted correctly"

# Clean build artifacts
clean:
	rm -f ${BINARY}

# Release preparation
release-snapshot:
	goreleaser release --snapshot --rm-dist

# Development installation
dev-install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	cp ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

# Check for security vulnerabilities
security:
	go list -json -deps ./... | nancy sleuth

# Validate Terraform files
tf-validate:
	terraform fmt -recursive -check examples/
	terraform validate examples/

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the provider binary"
	@echo "  install       - Install the provider"
	@echo "  lint          - Run golangci-lint"
	@echo "  test          - Run unit tests"
	@echo "  testacc       - Run acceptance tests"
	@echo "  testrace      - Run tests with race detection"
	@echo "  generate      - Generate code and documentation"
	@echo "  fmt           - Format code and tidy modules"
	@echo "  fmtcheck      - Check code formatting"
	@echo "  clean         - Clean build artifacts"
	@echo "  dev-install   - Install provider for local development"
	@echo "  security      - Check for security vulnerabilities"
	@echo "  tf-validate   - Validate Terraform examples"
	@echo "  help          - Show this help message"

.PHONY: build install lint test testacc testrace generate fmt fmtcheck clean dev-install security tf-validate help
