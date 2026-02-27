package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestLoadRawConfig(t *testing.T) {
	t.Setenv("MYSQL_ENDPOINT", "env-endpoint:3306")
	t.Setenv("MYSQL_USERNAME", "env-user")
	t.Setenv("MYSQL_PASSWORD", "env-pass")
	t.Setenv("MYSQL_DATABASE", "env-db")
	t.Setenv("MYSQL_TLS", "env-tls")
	t.Setenv("MYSQL_TLS_CA", "/env/ca.pem")
	t.Setenv("MYSQL_TLS_CERT", "/env/cert.pem")
	t.Setenv("MYSQL_TLS_KEY", "/env/key.pem")
	t.Setenv("MYSQL_TLS_SERVER_NAME", "env-server")
	t.Setenv("MYSQL_TLS_SKIP_VERIFY", "true")
	t.Setenv("MYSQL_CONN_MAX_LIFETIME", "30s")
	t.Setenv("MYSQL_MAX_OPEN_CONNS", "7")
	t.Setenv("MYSQL_MAX_IDLE_CONNS", "3")
	t.Setenv("MYSQL_WAIT_TIMEOUT", "120")
	t.Setenv("MYSQL_INNODB_LOCK_WAIT_TIMEOUT", "5")
	t.Setenv("MYSQL_LOCK_WAIT_TIMEOUT", "40")

	model := AuditLogFilterProviderModel{
		Endpoint:      types.StringValue("cfg-endpoint:3307"),
		Username:      types.StringNull(),
		Password:      types.StringNull(),
		Database:      types.StringNull(),
		TLS:           types.StringNull(),
		TLSCAFile:     types.StringNull(),
		TLSCertFile:   types.StringNull(),
		TLSKeyFile:    types.StringNull(),
		TLSServerName: types.StringNull(),
	}

	raw := loadRawConfig(model)

	if raw.endpoint != "cfg-endpoint:3307" {
		t.Fatalf("expected endpoint from config override, got %q", raw.endpoint)
	}
	if raw.username != "env-user" {
		t.Fatalf("expected username from env fallback, got %q", raw.username)
	}
	if raw.waitTimeoutEnv != "120" {
		t.Fatalf("expected MYSQL_WAIT_TIMEOUT in raw config, got %q", raw.waitTimeoutEnv)
	}
}

func TestParseAndValidateProviderConfigDefaults(t *testing.T) {
	t.Parallel()

	raw := providerRawConfig{}
	var diagnostics diag.Diagnostics

	validated, ok := parseAndValidateProviderConfig(raw, &diagnostics)
	if !ok {
		t.Fatalf("expected parse to succeed, diagnostics: %+v", diagnostics)
	}
	if diagnostics.HasError() {
		t.Fatalf("expected no diagnostics errors, got: %+v", diagnostics)
	}

	if validated.mysqlConfig.Addr != "localhost:3306" {
		t.Fatalf("unexpected default endpoint: %q", validated.mysqlConfig.Addr)
	}
	if validated.mysqlConfig.User != "root" {
		t.Fatalf("unexpected default username: %q", validated.mysqlConfig.User)
	}
	if validated.mysqlConfig.DBName != "mysql" {
		t.Fatalf("unexpected default database: %q", validated.mysqlConfig.DBName)
	}
	if validated.mysqlConfig.TLSConfig != "preferred" {
		t.Fatalf("unexpected default tls config: %q", validated.mysqlConfig.TLSConfig)
	}
	if validated.mysqlConfig.Params["wait_timeout"] != "10000" {
		t.Fatalf("unexpected wait_timeout: %q", validated.mysqlConfig.Params["wait_timeout"])
	}
	if validated.mysqlConfig.Params["innodb_lock_wait_timeout"] != "1" {
		t.Fatalf("unexpected innodb_lock_wait_timeout: %q", validated.mysqlConfig.Params["innodb_lock_wait_timeout"])
	}
	if validated.mysqlConfig.Params["lock_wait_timeout"] != "60" {
		t.Fatalf("unexpected lock_wait_timeout: %q", validated.mysqlConfig.Params["lock_wait_timeout"])
	}
}

func TestParseAndValidateProviderConfigInvalidWaitTimeout(t *testing.T) {
	t.Parallel()

	raw := providerRawConfig{
		waitTimeoutEnv: "0",
	}
	var diagnostics diag.Diagnostics

	_, ok := parseAndValidateProviderConfig(raw, &diagnostics)
	if ok {
		t.Fatalf("expected parse to fail for invalid wait timeout")
	}
	if !diagnostics.HasError() {
		t.Fatalf("expected diagnostic error for invalid wait timeout")
	}
	if diagnostics.ErrorsCount() != 1 {
		t.Fatalf("expected one diagnostic error, got %d", diagnostics.ErrorsCount())
	}
	if diagnostics[0].Summary() != "Invalid MySQL Wait Timeout" {
		t.Fatalf("unexpected diagnostic summary: %q", diagnostics[0].Summary())
	}
}

func TestParseAndValidateProviderConfigTLSConflict(t *testing.T) {
	t.Parallel()

	raw := providerRawConfig{
		tlsConfig: "false",
		tlsCAFile: "/tmp/ca.pem",
	}
	var diagnostics diag.Diagnostics

	_, ok := parseAndValidateProviderConfig(raw, &diagnostics)
	if ok {
		t.Fatalf("expected parse to fail for TLS conflict")
	}
	if !diagnostics.HasError() {
		t.Fatalf("expected diagnostic error for TLS conflict")
	}
	if diagnostics[0].Summary() != "TLS Configuration Conflict" {
		t.Fatalf("unexpected diagnostic summary: %q", diagnostics[0].Summary())
	}
}

func TestParseAndValidateProviderConfigInvalidTLSSkipVerifyEnv(t *testing.T) {
	t.Parallel()

	raw := providerRawConfig{
		tlsSkipVerifyEnv: "not-a-bool",
	}
	var diagnostics diag.Diagnostics

	_, ok := parseAndValidateProviderConfig(raw, &diagnostics)
	if ok {
		t.Fatalf("expected parse to fail for invalid tls skip verify env")
	}
	if !diagnostics.HasError() {
		t.Fatalf("expected diagnostic error for invalid tls skip verify env")
	}
	if diagnostics[0].Summary() != "Invalid TLS Skip Verify" {
		t.Fatalf("unexpected diagnostic summary: %q", diagnostics[0].Summary())
	}
}
