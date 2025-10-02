#!/bin/bash
set -euo pipefail

# Local testing script for Terraform Provider Audit Log Filters
# Usage: ./scripts/test-local.sh [--with-mysql] [--acceptance]

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

# Parse arguments
WITH_MYSQL=false
RUN_ACCEPTANCE=false

for arg in "$@"; do
    case $arg in
        --with-mysql)
            WITH_MYSQL=true
            shift
            ;;
        --acceptance)
            RUN_ACCEPTANCE=true
            shift
            ;;
        *)
            echo "Usage: $0 [--with-mysql] [--acceptance]"
            echo "  --with-mysql    Start MySQL container for testing"
            echo "  --acceptance    Run acceptance tests (requires MySQL)"
            exit 1
            ;;
    esac
done

cd "$PROJECT_DIR"

log_info "Starting local testing for Terraform Provider..."

# Load environment variables if available
if [[ -f "$SCRIPT_DIR/.env" ]]; then
    log_info "Loading environment variables from scripts/.env"
    source "$SCRIPT_DIR/.env"
else
    log_warning "No scripts/.env file found. Using defaults."
    export MYSQL_ENDPOINT=${MYSQL_ENDPOINT:-"localhost:3306"}
    export MYSQL_USERNAME=${MYSQL_USERNAME:-"root"}
    export MYSQL_PASSWORD=${MYSQL_PASSWORD:-"test123"}
    export MYSQL_DATABASE=${MYSQL_DATABASE:-"mysql"}
    export MYSQL_TLS=${MYSQL_TLS:-"false"}
fi

# Start MySQL container if requested
if [[ "$WITH_MYSQL" == true ]]; then
    log_info "Starting Percona Server container for testing..."
    
    # Stop existing container if running
    if docker ps -q -f name=percona-test 2>/dev/null | grep -q .; then
        log_warning "Stopping existing percona-test container..."
        docker stop percona-test >/dev/null 2>&1
    fi
    
    # Remove existing container
    if docker ps -aq -f name=percona-test 2>/dev/null | grep -q .; then
        docker rm percona-test >/dev/null 2>&1
    fi
    
    # Start new container
    log_info "Starting new Percona Server 8.4 container..."
    docker run --name percona-test \
        -e MYSQL_ROOT_PASSWORD="$MYSQL_PASSWORD" \
        -e MYSQL_DATABASE="$MYSQL_DATABASE" \
        -p 3306:3306 \
        -d percona/percona-server:8.4
    
    # Wait for MySQL to be ready
    log_info "Waiting for MySQL to be ready..."
    for i in {1..60}; do
        if docker exec percona-test mysqladmin ping -u root -p"$MYSQL_PASSWORD" --silent 2>/dev/null; then
            log_success "MySQL is ready!"
            break
        fi
        if [[ $i -eq 60 ]]; then
            log_error "MySQL failed to start within 60 seconds"
            exit 1
        fi
        echo -n "."
        sleep 1
    done
    echo
    sleep 5
    # Install audit_log_filter component
    log_info "Installing audit_log_filter component..."
    docker exec percona-test \
        sh -c "mysql -u root -p$MYSQL_PASSWORD < /usr/share/percona-server/audit_log_filter_linux_install.sql;"
    sleep 5 
    # Verify component is installed
    if docker exec percona-test mysql -u root -p"$MYSQL_PASSWORD" \
        -e "SELECT component_urn FROM mysql.component WHERE component_urn LIKE '%audit_log_filter%';" | grep -q audit_log_filter; then
        log_success "audit_log_filter component installed successfully"
    else
        log_error "Failed to install audit_log_filter component"
        exit 1
    fi
fi

# Build the provider
log_info "Building provider..."
if ! make build; then
    log_error "Provider build failed"
    exit 1
fi

# Run unit tests
log_info "Running unit tests..."
if go test ./internal/provider/ -v -short; then
    log_success "Unit tests passed"
else
    log_error "Unit tests failed"
    exit 1
fi

# Run acceptance tests if requested
if [[ "$RUN_ACCEPTANCE" == true ]]; then
    log_info "Running acceptance tests..."
    
    # Check if MySQL is accessible
    if mysqladmin ping -h "$(echo $MYSQL_ENDPOINT | cut -d: -f1)" \
        -P "$(echo $MYSQL_ENDPOINT | cut -d: -f2)" \
        -u "$MYSQL_USERNAME" -p"$MYSQL_PASSWORD" --silent 2>/dev/null; then
        log_error "Cannot connect to MySQL. Make sure MySQL is running and accessible."
        log_error "Connection details:"
        log_error "  Endpoint: $MYSQL_ENDPOINT"
        log_error "  Username: $MYSQL_USERNAME"
        log_error "  Database: $MYSQL_DATABASE"
        exit 1
    fi
    
    # Set acceptance test environment
    export TF_ACC=1
    
    # Run acceptance tests
    if go test ./internal/provider/ -v -timeout 10m; then
        log_success "Acceptance tests passed"
    else
        log_error "Acceptance tests failed"
        exit 1
    fi
fi

# Test basic provider functionality with Terraform
log_info "Testing provider with Terraform..."

# Create temporary test directory
TEST_DIR="$SCRIPT_DIR/tmp/terraform-test"
mkdir -p "$TEST_DIR"

# Create test configuration
cat > "$TEST_DIR/test.tf" << TFEOF
terraform {
  required_providers {
    auditlogfilters = {
      source = "0ch1r/auditlogfilters"
    }
  }
}

provider "auditlogfilters" {
  endpoint = "$MYSQL_ENDPOINT"
  username = "$MYSQL_USERNAME"
  password = "$MYSQL_PASSWORD"
  database = "$MYSQL_DATABASE"
  tls      = "$MYSQL_TLS"
}

resource "auditlogfilters_filter" "test" {
  name = "test_filter_\${formatdate("YYMMDDhhmmss", timestamp())}"
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


output "filter_info" {
  value = {
    name      = auditlogfilters_filter.test.name
    filter_id = auditlogfilters_filter.test.filter_id
  }
}
TFEOF

# Test Terraform workflow
cd "$TEST_DIR"

log_info "Initializing Terraform..."
# Remove lock file to avoid dependency conflicts in temporary test directory
rm -f .terraform.lock.hcl
# With dev overrides active, terraform init is not necessary and causes errors
if grep -q "dev_overrides" ~/.terraformrc 2>/dev/null; then
    log_info "Development overrides detected - skipping terraform init (as recommended by Terraform)"
else
    if ! terraform init -no-color; then
        log_error "Terraform init failed"
        exit 1
    fi
fi

log_info "Planning Terraform configuration..."
if ! terraform plan -no-color; then
    log_error "Terraform plan failed"
    exit 1
fi

# Only apply if we have MySQL connection
if [[ "$RUN_ACCEPTANCE" == true ]] || [[ "$WITH_MYSQL" == true ]]; then
    log_info "Applying Terraform configuration..."
    if terraform apply -auto-approve -no-color; then
        log_success "Terraform apply succeeded"
        
        log_info "Cleaning up resources..."
        if terraform destroy -auto-approve -no-color; then
            log_success "Terraform destroy succeeded"
        else
            log_warning "Terraform destroy failed - manual cleanup may be needed"
        fi
    else
        log_error "Terraform apply failed"
        exit 1
    fi
else
    log_warning "Skipping Terraform apply (no MySQL connection available)"
fi

# Clean up
cd "$PROJECT_DIR"
rm -rf "$TEST_DIR"
rm -f terraform-provider-auditlogfilter

# Cleanup Docker container if we started it
if [[ "$WITH_MYSQL" == true ]]; then
    log_info "Cleaning up test MySQL container..."
    docker stop percona-test >/dev/null 2>&1 || true
    docker rm percona-test >/dev/null 2>&1 || true
fi

log_success "All tests completed successfully! 🎉"

# Show summary
echo
log_info "Test Summary:"
echo "  ✅ Provider build"
echo "  ✅ Unit tests"
if [[ "$RUN_ACCEPTANCE" == true ]]; then
    echo "  ✅ Acceptance tests"
fi
echo "  ✅ Terraform integration"
if [[ "$WITH_MYSQL" == true ]]; then
    echo "  ✅ MySQL container testing"
fi

echo
log_info "Next steps:"
echo "  • Review any warnings above"
echo "  • Test with your own MySQL instance"
echo "  • Run './scripts/release.sh v1.0.0 --dry-run' to test release process"
