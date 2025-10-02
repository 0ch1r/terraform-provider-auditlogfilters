#!/bin/bash
set -euo pipefail

# Development environment setup script for Terraform Provider Audit Log Filters
# Usage: ./scripts/dev-setup.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

cd "$PROJECT_DIR"

log_info "Setting up development environment for Terraform Provider..."

# Check Go installation
if ! command -v go &> /dev/null; then
    log_error "Go is not installed. Please install Go 1.21+ first."
    exit 1
fi

GO_VERSION=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | sed 's/go//')
REQUIRED_VERSION="1.21"

if ! printf '%s\n%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V -C; then
    log_error "Go version $GO_VERSION is too old. Please install Go $REQUIRED_VERSION or later."
    exit 1
fi

log_success "Go version $GO_VERSION detected"

# Check Terraform installation
if ! command -v terraform &> /dev/null; then
    log_warning "Terraform is not installed. Please install it manually from https://terraform.io/downloads"
else
    TF_VERSION=$(terraform version | head -n1 | grep -o 'v[0-9]\+\.[0-9]\+\.[0-9]\+')
    log_success "Terraform $TF_VERSION detected"
fi

# Ensure Go bin directory is in PATH
GOPATH="${GOPATH:-$(go env GOPATH)}"
GOBIN="${GOBIN:-${GOPATH}/bin}"
if [[ ! ":$PATH:" == *":$GOBIN:"* ]]; then
    export PATH="$PATH:$GOBIN"
    log_info "Added $GOBIN to PATH for current session"
fi

# Check GoReleaser installation  
if ! command -v goreleaser &> /dev/null; then
    log_warning "GoReleaser is not installed. Installing..."
    
    if command -v brew &> /dev/null; then
        log_info "Installing GoReleaser via Homebrew..."
        brew install goreleaser/tap/goreleaser
    else
        log_info "Installing GoReleaser via go install..."
        go install github.com/goreleaser/goreleaser/v2@latest
        
        # Verify installation
        if ! command -v goreleaser &> /dev/null; then
            log_error "GoReleaser installation failed. Please ensure $GOBIN is in your PATH"
            log_info "Add this to your ~/.bashrc or ~/.profile:"
            echo "export PATH=\"\$PATH:$GOBIN\""
log_info "To make this permanent, add the above line to your ~/.bashrc or ~/.zshrc"
            exit 1
        fi
    fi
else
    log_success "GoReleaser detected"
fi

# Check golangci-lint installation
if ! command -v golangci-lint &> /dev/null; then
    log_warning "golangci-lint is not installed. Installing..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
else
    log_success "golangci-lint detected"
fi

# Download dependencies (avoid go mod tidy which might cause issues)
log_info "Downloading Go dependencies..."
go mod download

# Build the provider to verify everything works
log_info "Building provider to verify setup..."
if go build -v; then
    log_success "Provider built successfully"
    # Clean up the binary
    rm -f terraform-provider-auditlogfilter
else
    log_error "Provider build failed"
    exit 1
fi

# Create environment file template
if [[ ! -f "$SCRIPT_DIR/.env" ]]; then
    log_info "Creating environment file template..."
    cat > "$SCRIPT_DIR/.env.example" << 'ENVEOF'
# Environment variables for local development
# Copy this file to .env and fill in your values

# MySQL connection settings
MYSQL_ENDPOINT=localhost:3306
MYSQL_USERNAME=root
MYSQL_PASSWORD=your_password_here
MYSQL_DATABASE=mysql
MYSQL_TLS=false

# Test settings
TF_ACC=1
TF_LOG=DEBUG
TF_LOG_PROVIDER=DEBUG
ENVEOF

    log_info "Created .env.example template"
    log_warning "Please copy scripts/.env.example to scripts/.env and configure your settings"
fi

# Run basic tests (skip if they might fail due to missing MySQL)
log_info "Running basic unit tests..."
if go test ./internal/provider/ -short -v 2>/dev/null; then
    log_success "Basic tests passed"
else
    log_warning "Some tests failed - this might be due to missing MySQL connection"
fi

log_success "Development environment setup complete! 🎉"
echo
log_info "Next steps:"
echo "  1. Add $GOBIN to your PATH permanently:"
echo "     echo 'export PATH=\"\$PATH:$GOBIN\"' >> ~/.bashrc"
echo "  2. Configure scripts/.env with your MySQL connection details"  
echo "  3. Start a Percona Server 8.4+ instance for testing"
echo "  4. Install the audit_log_filter component in MySQL"
echo "  5. Run tests: go test ./..."
echo "  6. Check out the examples/ directory for usage examples"
echo
log_info "Useful commands:"
echo "  go build                - Build the provider"
echo "  go test ./...           - Run all tests"
echo "  go test ./... -short    - Run unit tests only"
echo "  ./scripts/test-local.sh - Run tests with Docker MySQL"
echo "  ./scripts/release.sh v1.0.0 --dry-run  - Test release process"

