#!/bin/bash
set -euo pipefail

# Percona Server 8.4 Docker Management Script (SSL enabled)
# This script manages a Docker container running Percona Server 8.4 with SSL and audit_log_filter component

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT_NAME=$(basename "$0")
CONTAINER_NAME="percona-server-ssl-test"
MYSQL_ROOT_PASSWORD="t00r"
MYSQL_DATABASE="test"
MYSQL_PORT="3306"
CERT_DIR="${SCRIPT_DIR}/.mysql-ssl"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

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

show_usage() {
  cat <<USAGE
Usage: $SCRIPT_NAME [COMMAND] [OPTIONS]

COMMANDS:
    start       Start Percona Server container with SSL and install audit_log_filter
    stop        Stop the container
    destroy     Stop and remove the container completely
    status      Show container status
    logs        Show container logs
    shell       Connect to MySQL shell with SSL
    install     Install audit_log_filter component (if container is running)
    verify      Verify audit_log_filter component installation

OPTIONS:
    -n, --name NAME       Container name (default: $CONTAINER_NAME)
    -p, --password PWD    MySQL root password (default: $MYSQL_ROOT_PASSWORD)
    -P, --port PORT       MySQL port (default: $MYSQL_PORT)
    -d, --database DB     Database name (default: $MYSQL_DATABASE)
    -c, --cert-dir DIR    Certificate directory (default: $CERT_DIR)
    -h, --help            Show this help message

EXAMPLES:
    $SCRIPT_NAME start                              # Start container with default settings
    $SCRIPT_NAME start -n mydb -p secret -P 3310     # Start with custom name, password, and port
    $SCRIPT_NAME shell                              # Connect to MySQL shell with SSL
    $SCRIPT_NAME destroy                            # Stop and remove container
    $SCRIPT_NAME status                             # Show container status
USAGE
}

check_docker() {
  if ! command -v docker &>/dev/null; then
    log_error "Docker is not installed or not in PATH"
    exit 1
  fi
}

check_openssl() {
  if ! command -v openssl &>/dev/null; then
    log_error "OpenSSL is required to generate SSL certificates"
    exit 1
  fi
}

ensure_certs() {
  local cert_dir="$1"
  local ca_key="${cert_dir}/ca-key.pem"
  local ca_cert="${cert_dir}/ca.pem"
  local server_key="${cert_dir}/server-key.pem"
  local server_cert="${cert_dir}/server-cert.pem"
  local client_key="${cert_dir}/client-key.pem"
  local client_cert="${cert_dir}/client-cert.pem"

  if [[ -f "$ca_cert" && -f "$server_cert" && -f "$server_key" && -f "$client_cert" && -f "$client_key" ]]; then
    log_info "Using existing SSL certificates in $cert_dir"
    return 0
  fi

  check_openssl
  log_info "Generating SSL certificates in $cert_dir"
  mkdir -p "$cert_dir"
  chmod 700 "$cert_dir"

  (
    cd "$cert_dir"

    # Generate CA private key
    openssl genpkey -algorithm RSA -out ca-key.pem -pkeyopt rsa_keygen_bits:4096

    # Generate CA certificate
    openssl req -new -x509 -key ca-key.pem -out ca.pem -days 3650 -subj '/C=US/ST=CA/L=San Francisco/O=Percona SSL CA/CN=Percona SSL CA'

    # Generate server private key
    openssl genpkey -algorithm RSA -out server-key.pem -pkeyopt rsa_keygen_bits:4096

    # Generate server certificate request
    openssl req -new -key server-key.pem -out server-req.pem -subj '/C=US/ST=CA/L=San Francisco/O=Percona Server/CN=percona-ssl'

    # Add SANs for localhost access
    cat >server-ext.cnf <<'EOF'
[v3_req]
subjectAltName=DNS:percona-ssl,DNS:localhost,IP:127.0.0.1
EOF

    # Generate server certificate
    openssl x509 -req -in server-req.pem -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -days 3650 -extfile server-ext.cnf -extensions v3_req

    # Generate client private key
    openssl genpkey -algorithm RSA -out client-key.pem -pkeyopt rsa_keygen_bits:4096

    # Generate client certificate request
    openssl req -new -key client-key.pem -out client-req.pem -subj '/C=US/ST=CA/L=San Francisco/O=Percona Client/CN=client'

    # Generate client certificate
    openssl x509 -req -in client-req.pem -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out client-cert.pem -days 3650

    # Set permissions
    chmod 600 *.pem

    # Clean up
    rm -f *-req.pem server-ext.cnf
  )

  log_success "SSL certificates generated"
}

mysql_ssl_args() {
  echo "--ssl-mode=REQUIRED --ssl-ca=/etc/mysql/ssl/ca.pem --ssl-cert=/etc/mysql/ssl/client-cert.pem --ssl-key=/etc/mysql/ssl/client-key.pem"
}

wait_for_mysql() {
  local container_name="$1"
  local password="$2"
  local max_attempts=60
  local attempt=1

  log_info "Waiting for MySQL to be ready..."
  while [ $attempt -le $max_attempts ]; do
    if docker exec "$container_name" mysqladmin ping -u root -p"$password" $(mysql_ssl_args) --silent 2>/dev/null; then
      log_success "MySQL is ready after $attempt seconds!"
      return 0
    fi

    if [ $attempt -eq $max_attempts ]; then
      log_error "MySQL failed to start within $max_attempts seconds"
      return 1
    fi

    echo -n "."
    sleep 1
    ((attempt++))
  done
}

install_audit_component() {
  local container_name="$1"
  local password="$2"

  log_info "Installing audit_log_filter component using official installation script..."

  if ! docker exec "$container_name" test -f /usr/share/percona-server/audit_log_filter_linux_install.sql; then
    log_error "Percona audit_log_filter installation script not found in container"
    return 1
  fi

  sleep 5
  if docker exec "$container_name" sh -c "mysql -u root -p$password $(mysql_ssl_args) < /usr/share/percona-server/audit_log_filter_linux_install.sql" 2>/dev/null; then
    log_success "Audit log filter installation script executed successfully"
    sleep 5
  else
    log_error "Failed to execute audit_log_filter installation script"
    return 1
  fi

  log_info "Verifying audit_log_filter component installation..."
  local component_count
  component_count=$(docker exec "$container_name" mysql -u root -p"$password" \
    $(mysql_ssl_args) \
    -e "SELECT COUNT(*) FROM mysql.component WHERE component_urn LIKE '%audit_log_filter%';" \
    -B -N 2>/dev/null || echo "0")

  if [[ "$component_count" -gt 0 ]]; then
    log_success "audit_log_filter component verified successfully"
    log_info "Installed components:"
    docker exec "$container_name" mysql -u root -p"$password" \
      $(mysql_ssl_args) \
      -e "SELECT component_id, component_group_id, component_urn FROM mysql.component;" \
      2>/dev/null || log_warning "Could not display component details"
    return 0
  else
    log_error "audit_log_filter component verification failed"
    return 1
  fi
}

create_ssl_user() {
  local container_name="$1"
  local password="$2"

  log_info "Creating tfuser with SSL requirement..."
  if docker exec "$container_name" mysql -u root -p"$password" \
    $(mysql_ssl_args) \
    -e "CREATE USER IF NOT EXISTS 'tfuser'@'%' IDENTIFIED BY 'tfpass'; \
            ALTER USER 'tfuser'@'%' IDENTIFIED BY 'tfpass' REQUIRE SSL; \
            GRANT SELECT ON mysql.component TO 'tfuser'@'%'; \
            GRANT SELECT, INSERT, UPDATE, DELETE ON mysql.audit_log_filter TO 'tfuser'@'%'; \
            GRANT SELECT, INSERT, UPDATE, DELETE ON mysql.audit_log_user TO 'tfuser'@'%'; \
            GRANT AUDIT_ADMIN ON *.* TO 'tfuser'@'%';"; then
    log_success "User tfuser created or updated with limited audit_log_filter privileges and REQUIRE SSL"
  else
    log_error "Failed to create tfuser"
    return 1
  fi
}

start_container() {
  local container_name="$1"
  local password="$2"
  local database="$3"
  local port="$4"
  local cert_dir="$5"

  ensure_certs "$cert_dir"

  if docker ps -a --format '{{.Names}}' | grep -q "^${container_name}$"; then
    if docker ps --format '{{.Names}}' | grep -q "^${container_name}$"; then
      log_warning "Container '$container_name' is already running"
      return 0
    else
      log_info "Removing existing stopped container '$container_name'..."
      docker rm "$container_name" >/dev/null 2>&1
    fi
  fi

  log_info "Starting Percona Server 8.4 SSL container '$container_name'..."

  local container_id
  container_id=$(docker run --name "$container_name" \
    -e MYSQL_ROOT_PASSWORD="$password" \
    -e MYSQL_DATABASE="$database" \
    -p "$port:3306" \
    -v "$cert_dir:/etc/mysql/ssl:ro" \
    -d percona/percona-server:8.4 \
    --ssl-ca=/etc/mysql/ssl/ca.pem \
    --ssl-cert=/etc/mysql/ssl/server-cert.pem \
    --ssl-key=/etc/mysql/ssl/server-key.pem \
    --require_secure_transport=ON)

  if [ $? -eq 0 ]; then
    log_success "Container started with ID: ${container_id:0:12}"
  else
    log_error "Failed to start container"
    return 1
  fi

  if wait_for_mysql "$container_name" "$password"; then
    if install_audit_component "$container_name" "$password"; then
      if ! create_ssl_user "$container_name" "$password"; then
        return 1
      fi
      log_success "Percona Server 8.4 with SSL and audit_log_filter is ready!"
      echo
      log_info "Connection details:"
      echo "  Host: localhost"
      echo "  Port: $port"
      echo "  Username: root"
      echo "  Password: $password"
      echo "  Database: $database"
      echo "  CA Cert: $cert_dir/ca.pem"
      echo
      log_info "Connect using:"
      echo "  mysql -h localhost -P $port -u root -p'$password' --ssl-mode=REQUIRED --ssl-ca=$cert_dir/ca.pem"
    else
      log_error "Failed to install audit_log_filter component"
      return 1
    fi
  else
    log_error "MySQL failed to become ready"
    return 1
  fi
}

stop_container() {
  local container_name="$1"

  if docker ps --format '{{.Names}}' | grep -q "^${container_name}$"; then
    log_info "Stopping container '$container_name'..."
    if docker stop "$container_name" >/dev/null 2>&1; then
      log_success "Container '$container_name' stopped"
    else
      log_error "Failed to stop container '$container_name'"
      return 1
    fi
  else
    log_warning "Container '$container_name' is not running"
  fi
}

destroy_container() {
  local container_name="$1"

  stop_container "$container_name"

  if docker ps -a --format '{{.Names}}' | grep -q "^${container_name}$"; then
    log_info "Removing container '$container_name'..."
    if docker rm "$container_name" >/dev/null 2>&1; then
      log_success "Container '$container_name' removed"
    else
      log_error "Failed to remove container '$container_name'"
      return 1
    fi
  else
    log_info "Container '$container_name' does not exist"
  fi
}

show_status() {
  local container_name="$1"

  echo "=== Container Status ==="
  if docker ps -a --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}' | grep -E "(NAMES|${container_name})"; then
    echo
    if docker ps --format '{{.Names}}' | grep -q "^${container_name}$"; then
      echo "=== Container Logs (last 10 lines) ==="
      docker logs --tail 10 "$container_name" 2>/dev/null || echo "No logs available"
    fi
  else
    log_info "Container '$container_name' does not exist"
  fi
}

show_logs() {
  local container_name="$1"

  if docker ps -a --format '{{.Names}}' | grep -q "^${container_name}$"; then
    log_info "Showing logs for container '$container_name':"
    docker logs "$container_name"
  else
    log_error "Container '$container_name' does not exist"
    return 1
  fi
}

mysql_shell() {
  local container_name="$1"
  local password="$2"

  if docker ps --format '{{.Names}}' | grep -q "^${container_name}$"; then
    log_info "Connecting to MySQL shell in container '$container_name' with SSL..."
    docker exec -it "$container_name" mysql -u root -p"$password" $(mysql_ssl_args)
  else
    log_error "Container '$container_name' is not running"
    return 1
  fi
}

verify_component() {
  local container_name="$1"
  local password="$2"

  if ! docker ps --format '{{.Names}}' | grep -q "^${container_name}$"; then
    log_error "Container '$container_name' is not running"
    return 1
  fi

  log_info "Verifying audit_log_filter component in container '$container_name'..."

  local component_info
  component_info=$(docker exec "$container_name" mysql -u root -p"$password" \
    $(mysql_ssl_args) \
    -e "SELECT component_id, component_group_id, component_urn FROM mysql.component;" \
    2>/dev/null)

  if echo "$component_info" | grep -q "audit_log_filter"; then
    log_success "audit_log_filter component is installed"
    echo "$component_info"

    log_info "Available audit_log_filter functions:"
    docker exec "$container_name" mysql -u root -p"$password" \
      $(mysql_ssl_args) \
      -e "SELECT name FROM mysql.func WHERE name LIKE 'audit_log%' ORDER BY name;" \
      2>/dev/null || log_warning "Could not list audit functions"
  else
    log_warning "audit_log_filter component is not installed"
    echo "$component_info"
    return 1
  fi
}

# Parse command line arguments
COMMAND=""
while [[ $# -gt 0 ]]; do
  case $1 in
  start | stop | destroy | status | logs | shell | install | verify)
    COMMAND="$1"
    shift
    ;;
  -n | --name)
    CONTAINER_NAME="$2"
    shift 2
    ;;
  -p | --password)
    MYSQL_ROOT_PASSWORD="$2"
    shift 2
    ;;
  -P | --port)
    MYSQL_PORT="$2"
    shift 2
    ;;
  -d | --database)
    MYSQL_DATABASE="$2"
    shift 2
    ;;
  -c | --cert-dir)
    CERT_DIR="$2"
    shift 2
    ;;
  -h | --help)
    show_usage
    exit 0
    ;;
  *)
    log_error "Unknown option: $1"
    show_usage
    exit 1
    ;;
  esac
done

if [[ -z "$COMMAND" ]]; then
  log_error "No command specified"
  show_usage
  exit 1
fi

check_docker

case $COMMAND in
start)
  start_container "$CONTAINER_NAME" "$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" "$MYSQL_PORT" "$CERT_DIR"
  ;;
stop)
  stop_container "$CONTAINER_NAME"
  ;;
destroy)
  destroy_container "$CONTAINER_NAME"
  ;;
status)
  show_status "$CONTAINER_NAME"
  ;;
logs)
  show_logs "$CONTAINER_NAME"
  ;;
shell)
  mysql_shell "$CONTAINER_NAME" "$MYSQL_ROOT_PASSWORD"
  ;;
install)
  if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    install_audit_component "$CONTAINER_NAME" "$MYSQL_ROOT_PASSWORD"
  else
    log_error "Container '$CONTAINER_NAME' is not running. Start it first with: $SCRIPT_NAME start"
    exit 1
  fi
  ;;
verify)
  verify_component "$CONTAINER_NAME" "$MYSQL_ROOT_PASSWORD"
  ;;
*)
  log_error "Unknown command: $COMMAND"
  show_usage
  exit 1
  ;;
esac
