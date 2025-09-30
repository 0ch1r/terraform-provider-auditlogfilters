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
    log_warning "Terraform is not installed. Installing via package manager..."
    
    # Try different package managers
    if command -v brew &> /dev/null; then
        brew install terraform
    elif command -v apt-get &> /dev/null; then
        curl -fsSL https://apt.releases.hashicorp.com/gpg | sudo apt-key add -
        sudo apt-add-repository "deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main"
        sudo apt-get update && sudo apt-get install terraform
    elif command -v yum &> /dev/null; then
        sudo yum install -y yum-utils
        sudo yum-config-manager --add-repo https://rpm.releases.hashicorp.com/RHEL/hashicorp.repo
        sudo yum -y install terraform
    else
        log_warning "Could not auto-install Terraform. Please install it manually from https://terraform.io/downloads"
    fi
else
    TF_VERSION=$(terraform version | head -n1 | grep -o 'v[0-9]\+\.[0-9]\+\.[0-9]\+')
    log_success "Terraform $TF_VERSION detected"
fi

# Check GoReleaser installation  
if ! command -v goreleaser &> /dev/null; then
    log_warning "GoReleaser is not installed. Installing..."
    
    if command -v brew &> /dev/null; then
        brew install goreleaser/tap/goreleaser
    else
        log_info "Installing GoReleaser via script..."
        curl -sfL https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh | sh
        sudo mv ./bin/goreleaser /usr/local/bin/
    fi
else
    log_success "GoReleaser detected"
fi

# Check golangci-lint installation
if ! command -v golangci-lint &> /dev/null; then
    log_warning "golangci-lint is not installed. Installing..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
    export PATH=$PATH:$(go env GOPATH)/bin
else
    log_success "golangci-lint detected"
fi

# Install Go dependencies
log_info "Installing Go dependencies..."
go mod download
go mod tidy

# Build the provider to verify everything works
log_info "Building provider to verify setup..."
if make build; then
    log_success "Provider built successfully"
    rm -f terraform-provider-auditlogfilters
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

# Setup dev overrides
log_info "Setting up Terraform dev overrides..."
make dev-override

# Run basic tests
log_info "Running basic tests to verify setup..."
if go test ./internal/provider/ -v -short; then
    log_success "Basic tests passed"
else
    log_warning "Some tests failed - this might be due to missing MySQL connection"
fi

log_success "Development environment setup complete! 🎉"
echo
log_info "Next steps:"
echo "  1. Configure scripts/.env with your MySQL connection details"
echo "  2. Start a Percona Server 8.4+ instance for testing"
echo "  3. Install the audit_log_filter component in MySQL"
echo "  4. Run 'make test' to verify everything works"
echo "  5. Check out the examples/ directory for usage examples"
echo
log_info "Useful commands:"
echo "  make help           - Show all available commands"
echo "  make build          - Build the provider"
echo "  make test           - Run tests"
echo "  make testacc        - Run acceptance tests (requires MySQL)"
echo "  ./scripts/release.sh v1.0.0 --dry-run  - Test release process"
