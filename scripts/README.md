# Development Scripts

This directory contains helpful scripts for developing and testing the Terraform Provider for Audit Log Filters.

## 📁 Directory Structure

```
scripts/
├── README.md           # This file
├── .env.example        # Environment variables template
├── .env               # Your local environment (gitignored)
├── dev-setup.sh       # Development environment setup
├── release.sh         # Release automation script
├── test-local.sh      # Local testing script
├── personal-*         # Your personal scripts (gitignored)
└── tmp/               # Temporary files (gitignored)
```

## 🚀 Quick Start

1. **Set up development environment:**
   ```bash
   ./scripts/dev-setup.sh
   ```

2. **Configure your environment:**
   ```bash
   cp scripts/.env.example scripts/.env
   # Edit scripts/.env with your MySQL connection details
   ```

3. **Run local tests:**
   ```bash
   ./scripts/test-local.sh --with-mysql --acceptance
   ```

4. **Create a release:**
   ```bash
   ./scripts/release.sh v1.0.0 --dry-run  # Test first
   ./scripts/release.sh v1.0.0           # Actually release
   ```

## 📋 Available Scripts

### 🔧 `dev-setup.sh`
**Purpose:** Set up your development environment with all required tools.

**What it does:**
- Checks Go version (requires 1.21+)
- Installs Terraform, GoReleaser, golangci-lint
- Downloads Go dependencies
- Builds the provider to verify everything works
- Creates `.env.example` template
- Sets up Terraform dev overrides
- Runs basic tests

**Usage:**
```bash
./scripts/dev-setup.sh
```

### 🧪 `test-local.sh`
**Purpose:** Comprehensive local testing with optional MySQL container.

**Features:**
- Starts Percona Server 8.4 container
- Installs audit_log_filter component
- Runs unit and acceptance tests
- Tests provider with real Terraform configurations
- Automatic cleanup

**Usage:**
```bash
# Basic testing (no MySQL)
./scripts/test-local.sh

# With MySQL container
./scripts/test-local.sh --with-mysql

# With acceptance tests (requires MySQL)
./scripts/test-local.sh --acceptance

# Full testing
./scripts/test-local.sh --with-mysql --acceptance
```

### 🚀 `release.sh`
**Purpose:** Automated release process with safety checks.

**Features:**
- Version validation (semantic versioning)
- Pre-release checks (build, test, lint)
- Git status validation
- GoReleaser configuration check
- Changelog reminder
- Dry-run support
- Automatic tag creation and push

**Usage:**
```bash
# Test release process
./scripts/release.sh v1.0.0 --dry-run

# Create actual release
./scripts/release.sh v1.0.0

# Create pre-release
./scripts/release.sh v1.0.0-beta1
```

## 🔧 Environment Configuration

### Required Variables
Set these in `scripts/.env`:

```bash
# MySQL connection for testing
MYSQL_ENDPOINT=localhost:3306
MYSQL_USERNAME=root
MYSQL_PASSWORD=your_password
MYSQL_DATABASE=mysql
MYSQL_TLS=false

# Testing
TF_ACC=1
TF_LOG=INFO
```

### Optional Variables
```bash
# Debugging
TF_LOG=DEBUG
TF_LOG_PROVIDER=DEBUG

# Release (if using GitHub CLI or custom workflows)
GITHUB_TOKEN=your_token
GPG_FINGERPRINT=your_gpg_key
```

## 🎯 Common Workflows

### Setting up for the first time:
```bash
./scripts/dev-setup.sh
cp scripts/.env.example scripts/.env
# Edit .env file with your settings
./scripts/test-local.sh --with-mysql
```

### Daily development:
```bash
make build
make test
./scripts/test-local.sh --acceptance  # If you have MySQL running
```

### Before releasing:
```bash
./scripts/test-local.sh --with-mysql --acceptance
./scripts/release.sh v1.x.x --dry-run
# If dry-run passes:
./scripts/release.sh v1.x.x
```

### Testing with different MySQL versions:
```bash
# Edit the Docker image in test-local.sh
# Change: percona/percona-server:8.4
# To:     percona/percona-server:8.0
./scripts/test-local.sh --with-mysql --acceptance
```

## 🎨 Personal Customization

You can create personal scripts that won't be committed to git:

```bash
# These files are automatically ignored:
scripts/personal-my-workflow.sh
scripts/personal-shortcuts.sh
scripts/.env
scripts/tmp/
scripts/*.local
```

Example personal script:
```bash
#!/bin/bash
# scripts/personal-quick-test.sh

# Your custom testing workflow
source scripts/.env
export TF_LOG=DEBUG

echo "Running my custom test workflow..."
make build
go test ./internal/provider/ -run TestSpecificFunction -v
```

## 🔍 Troubleshooting

### Script won't run:
```bash
chmod +x scripts/script-name.sh
```

### MySQL connection issues:
```bash
# Check if container is running
docker ps | grep percona-test

# Check logs
docker logs percona-test

# Connect manually
docker exec -it percona-test mysql -u root -p
```

### GoReleaser issues:
```bash
# Validate config
goreleaser check

# Test build
goreleaser build --single-target --clean
```

### Environment issues:
```bash
# Check your environment
source scripts/.env
env | grep MYSQL
```

## 🤝 Contributing

If you create useful scripts that could benefit other developers:

1. **Remove personal information** (passwords, tokens, etc.)
2. **Add proper documentation** and error handling
3. **Test with clean environment**
4. **Submit a pull request**

Scripts that are generally useful should be committed to git. Personal workflow scripts should use the `personal-*` naming pattern.

## 🛠️ Post-Setup Configuration

After running the development setup script, you may need to add Go's binary directory to your PATH permanently:

### Adding Go binaries to PATH

If tools installed via `go install` (like GoReleaser) are not found in your terminal, add the following to your shell profile:

```bash
# Add to ~/.bashrc (for bash) or ~/.zshrc (for zsh)
echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.bashrc

# Reload your shell configuration
source ~/.bashrc
```

### Verify PATH configuration:
```bash
# Check if Go binary directory is in PATH
echo $PATH | grep go/bin

# Test if tools are accessible
goreleaser version
golangci-lint version
```

**Note:** The dev-setup script will automatically add this to your PATH for the current session, but you may need to add it permanently to your shell profile for future sessions.

