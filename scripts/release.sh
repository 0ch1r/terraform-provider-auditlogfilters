#!/bin/bash
set -euo pipefail

# Release script for Terraform Provider Audit Log Filters
# Usage: ./scripts/release.sh [version] [--dry-run]
#
# Examples:
#   ./scripts/release.sh v1.0.0          # Create and push release
#   ./scripts/release.sh v1.0.1 --dry-run # Test without pushing

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
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

# Check if we're in the right directory
if [[ ! -f "$PROJECT_DIR/go.mod" ]] || [[ ! -f "$PROJECT_DIR/.goreleaser.yml" ]]; then
    log_error "This doesn't appear to be the root of a Go project with GoReleaser"
    log_error "Expected files: go.mod, .goreleaser.yml"
    exit 1
fi

# Parse arguments
VERSION="${1:-}"
DRY_RUN="${2:-}"

if [[ -z "$VERSION" ]]; then
    log_error "Usage: $0 <version> [--dry-run]"
    log_error "Example: $0 v1.0.0"
    exit 1
fi

# Validate version format
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)*$ ]]; then
    log_error "Invalid version format. Expected: vX.Y.Z or vX.Y.Z-suffix"
    log_error "Examples: v1.0.0, v1.2.3-beta1"
    exit 1
fi

cd "$PROJECT_DIR"

log_info "Starting release process for version: $VERSION"

# Check if we're on the right branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [[ "$CURRENT_BRANCH" != "master" ]] && [[ "$CURRENT_BRANCH" != "main" ]]; then
    log_warning "You're not on master/main branch (currently on: $CURRENT_BRANCH)"
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Aborting release"
        exit 0
    fi
fi

# Check for uncommitted changes
if [[ -n $(git status --porcelain) ]]; then
    log_error "You have uncommitted changes. Please commit or stash them first."
    git status --short
    exit 1
fi

# Check if tag already exists
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    log_error "Tag $VERSION already exists"
    exit 1
fi

# Fetch latest changes
log_info "Fetching latest changes from remote..."
git fetch origin

# Check if we're behind remote
LOCAL_COMMIT=$(git rev-parse HEAD)
REMOTE_COMMIT=$(git rev-parse origin/$CURRENT_BRANCH 2>/dev/null || echo "")

if [[ -n "$REMOTE_COMMIT" ]] && [[ "$LOCAL_COMMIT" != "$REMOTE_COMMIT" ]]; then
    log_warning "Your local branch is not up to date with remote"
    log_info "Local:  $LOCAL_COMMIT"
    log_info "Remote: $REMOTE_COMMIT"
    read -p "Pull latest changes? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git pull origin "$CURRENT_BRANCH"
    else
        log_warning "Continuing with local changes..."
    fi
fi

# Run pre-release checks
log_info "Running pre-release checks..."

# Check if Go modules are tidy
log_info "Checking Go modules..."
go mod tidy
if [[ -n $(git status --porcelain go.mod go.sum 2>/dev/null) ]]; then
    log_error "go.mod or go.sum files need to be updated. Run 'go mod tidy' and commit."
    exit 1
fi

# Build and test
log_info "Building the provider..."
if ! go build -o terraform-provider-auditlogfilters; then
    log_error "Build failed"
    exit 1
fi
rm -f terraform-provider-auditlogfilters

# Run basic tests
log_info "Running tests..."
if ! go test ./internal/provider/ -v -short; then
    log_error "Tests failed"
    exit 1
fi

# Check GoReleaser configuration
log_info "Validating GoReleaser configuration..."
if ! goreleaser check; then
    log_error "GoReleaser configuration is invalid"
    exit 1
fi

# Update changelog if it exists
if [[ -f "CHANGELOG.md" ]]; then
    log_info "Please ensure CHANGELOG.md is updated with changes for $VERSION"
    read -p "Is CHANGELOG.md ready for $VERSION? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_warning "Please update CHANGELOG.md before releasing"
        exit 1
    fi
fi

# Show what will be released
log_info "Release summary:"
echo "  Version: $VERSION"
echo "  Branch: $CURRENT_BRANCH"
echo "  Commit: $(git rev-parse --short HEAD)"
echo "  Message: $(git log -1 --pretty=format:'%s')"

if [[ "$DRY_RUN" == "--dry-run" ]]; then
    log_warning "DRY RUN MODE - No changes will be made"
    
    # Test GoReleaser without publishing
    log_info "Testing GoReleaser build (dry run)..."
    if goreleaser release --snapshot --clean; then
        log_success "Dry run successful! Release would work."
    else
        log_error "Dry run failed"
        exit 1
    fi
    
    log_info "Dry run completed. No tag was created."
    exit 0
fi

# Final confirmation
echo
log_warning "This will create and push tag $VERSION, triggering a GitHub release."
read -p "Are you sure you want to proceed? (y/N): " -n 1 -r
echo

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    log_info "Release cancelled"
    exit 0
fi

# Create and push the tag
log_info "Creating tag $VERSION..."
git tag -a "$VERSION" -m "Release $VERSION"

log_info "Pushing tag to remote..."
git push origin "$VERSION"

log_success "Release tag $VERSION has been created and pushed!"
log_info "GitHub Actions will now build and publish the release."
log_info "You can monitor the progress at:"
log_info "  https://github.com/0ch1r/terraform-provider-auditlogfilters/actions"
log_info "  https://github.com/0ch1r/terraform-provider-auditlogfilters/releases"

# Clean up any local build artifacts
rm -f terraform-provider-auditlogfilters
rm -rf dist/

log_success "Release process completed successfully! 🎉"
