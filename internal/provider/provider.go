package provider

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure AuditLogFilterProvider satisfies various provider interfaces.
var _ provider.Provider = &AuditLogFilterProvider{}

var errNonPositiveInt64 = errors.New("value must be a positive integer (seconds)")

// AuditLogFilterProvider defines the provider implementation.
type AuditLogFilterProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
	db      *sql.DB
}

// AuditLogFilterProviderModel describes the provider data model.
type AuditLogFilterProviderModel struct {
	Endpoint              types.String `tfsdk:"endpoint"`
	Username              types.String `tfsdk:"username"`
	Password              types.String `tfsdk:"password"`
	Database              types.String `tfsdk:"database"`
	TLS                   types.String `tfsdk:"tls"`
	TLSCAFile             types.String `tfsdk:"tls_ca_file"`
	TLSCertFile           types.String `tfsdk:"tls_cert_file"`
	TLSKeyFile            types.String `tfsdk:"tls_key_file"`
	TLSServerName         types.String `tfsdk:"tls_server_name"`
	TLSSkipVerify         types.Bool   `tfsdk:"tls_skip_verify"`
	WaitTimeout           types.Int64  `tfsdk:"wait_timeout"`
	InnodbLockWaitTimeout types.Int64  `tfsdk:"innodb_lock_wait_timeout"`
	LockWaitTimeout       types.Int64  `tfsdk:"lock_wait_timeout"`
}

type providerRawConfig struct {
	endpoint                 string
	username                 string
	password                 string
	database                 string
	tlsConfig                string
	tlsCAFile                string
	tlsCertFile              string
	tlsKeyFile               string
	tlsServerName            string
	tlsSkipVerifyEnv         string
	connMaxLifetimeEnv       string
	maxOpenConnsEnv          string
	maxIdleConnsEnv          string
	waitTimeoutEnv           string
	innodbLockWaitTimeoutEnv string
	lockWaitTimeoutEnv       string
	tlsSkipVerify            types.Bool
	waitTimeout              types.Int64
	innodbLockWaitTimeout    types.Int64
	lockWaitTimeout          types.Int64
}

type providerValidatedConfig struct {
	mysqlConfig  mysql.Config
	maxLifetime  time.Duration
	maxOpenConns int
	maxIdleConns int
}

func (p *AuditLogFilterProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "auditlogfilters"
	resp.Version = p.version
}

func (p *AuditLogFilterProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description: "MySQL server endpoint (host:port). May also be provided via MYSQL_ENDPOINT environment variable.",
				Optional:    true,
			},
			"username": schema.StringAttribute{
				Description: "MySQL username. May also be provided via MYSQL_USERNAME environment variable.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "MySQL password. May also be provided via MYSQL_PASSWORD environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"database": schema.StringAttribute{
				Description: "MySQL database name to connect to. Defaults to 'mysql'. May also be provided via MYSQL_DATABASE environment variable.",
				Optional:    true,
			},
			"tls": schema.StringAttribute{
				Description: "TLS configuration for the MySQL connection. Options: 'true', 'false', 'skip-verify', 'preferred'. Defaults to 'preferred'. May also be provided via MYSQL_TLS environment variable.",
				Optional:    true,
			},
			"tls_ca_file": schema.StringAttribute{
				Description: "Path to a PEM-encoded CA certificate file for MySQL TLS. May also be provided via MYSQL_TLS_CA environment variable.",
				Optional:    true,
			},
			"tls_cert_file": schema.StringAttribute{
				Description: "Path to a PEM-encoded client certificate file for MySQL TLS. May also be provided via MYSQL_TLS_CERT environment variable.",
				Optional:    true,
			},
			"tls_key_file": schema.StringAttribute{
				Description: "Path to a PEM-encoded client key file for MySQL TLS. May also be provided via MYSQL_TLS_KEY environment variable.",
				Optional:    true,
			},
			"tls_server_name": schema.StringAttribute{
				Description: "Server name for TLS verification (SNI). May also be provided via MYSQL_TLS_SERVER_NAME environment variable.",
				Optional:    true,
			},
			"tls_skip_verify": schema.BoolAttribute{
				Description: "Skip TLS certificate verification. May also be provided via MYSQL_TLS_SKIP_VERIFY environment variable.",
				Optional:    true,
			},
			"wait_timeout": schema.Int64Attribute{
				Description: "MySQL session wait_timeout in seconds (idle connection timeout). Defaults to 10000. May also be provided via MYSQL_WAIT_TIMEOUT environment variable.",
				Optional:    true,
			},
			"innodb_lock_wait_timeout": schema.Int64Attribute{
				Description: "MySQL session innodb_lock_wait_timeout in seconds. Defaults to 1. May also be provided via MYSQL_INNODB_LOCK_WAIT_TIMEOUT environment variable.",
				Optional:    true,
			},
			"lock_wait_timeout": schema.Int64Attribute{
				Description: "MySQL session lock_wait_timeout in seconds (metadata lock timeout). Defaults to 60. May also be provided via MYSQL_LOCK_WAIT_TIMEOUT environment variable.",
				Optional:    true,
			},
		},
		MarkdownDescription: "The Audit Log Filter provider manages Percona Server 8.4+ audit log filters and user assignments. " +
			"It provides resources to create, modify, and remove audit log filters using the audit_log_filter component functions.",
	}
}

func (p *AuditLogFilterProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data AuditLogFilterProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Close any prior connection on reconfigure to avoid leaks.
	if p.db != nil {
		_ = p.db.Close()
		p.db = nil
	}

	rawConfig := loadRawConfig(data)
	validatedConfig, ok := parseAndValidateProviderConfig(rawConfig, &resp.Diagnostics)
	if !ok {
		return
	}

	// Allow empty password for testing
	// if password == "" {
	//	resp.Diagnostics.AddAttributeError(
	//		path.Root("password"),
	//		"Missing MySQL Password",
	//		"The provider cannot create the MySQL client as there is a missing or empty value for the MySQL password. "+
	//			"Set the password value in the configuration or use the MYSQL_PASSWORD environment variable. "+
	//			"If the password is intentionally empty, explicitly set it to an empty string.",
	//	)
	// }

	if resp.Diagnostics.HasError() {
		return
	}
	db, ok := connectAndVerify(ctx, validatedConfig, &resp.Diagnostics)
	if !ok {
		return
	}

	p.db = db
	resp.DataSourceData = db
	resp.ResourceData = db
}

func loadRawConfig(data AuditLogFilterProviderModel) providerRawConfig {
	return providerRawConfig{
		endpoint:                 configStringOrEnv(data.Endpoint, os.Getenv("MYSQL_ENDPOINT")),
		username:                 configStringOrEnv(data.Username, os.Getenv("MYSQL_USERNAME")),
		password:                 configStringOrEnv(data.Password, os.Getenv("MYSQL_PASSWORD")),
		database:                 configStringOrEnv(data.Database, os.Getenv("MYSQL_DATABASE")),
		tlsConfig:                configStringOrEnv(data.TLS, os.Getenv("MYSQL_TLS")),
		tlsCAFile:                configStringOrEnv(data.TLSCAFile, os.Getenv("MYSQL_TLS_CA")),
		tlsCertFile:              configStringOrEnv(data.TLSCertFile, os.Getenv("MYSQL_TLS_CERT")),
		tlsKeyFile:               configStringOrEnv(data.TLSKeyFile, os.Getenv("MYSQL_TLS_KEY")),
		tlsServerName:            configStringOrEnv(data.TLSServerName, os.Getenv("MYSQL_TLS_SERVER_NAME")),
		tlsSkipVerifyEnv:         os.Getenv("MYSQL_TLS_SKIP_VERIFY"),
		connMaxLifetimeEnv:       os.Getenv("MYSQL_CONN_MAX_LIFETIME"),
		maxOpenConnsEnv:          os.Getenv("MYSQL_MAX_OPEN_CONNS"),
		maxIdleConnsEnv:          os.Getenv("MYSQL_MAX_IDLE_CONNS"),
		waitTimeoutEnv:           os.Getenv("MYSQL_WAIT_TIMEOUT"),
		innodbLockWaitTimeoutEnv: os.Getenv("MYSQL_INNODB_LOCK_WAIT_TIMEOUT"),
		lockWaitTimeoutEnv:       os.Getenv("MYSQL_LOCK_WAIT_TIMEOUT"),
		tlsSkipVerify:            data.TLSSkipVerify,
		waitTimeout:              data.WaitTimeout,
		innodbLockWaitTimeout:    data.InnodbLockWaitTimeout,
		lockWaitTimeout:          data.LockWaitTimeout,
	}
}

func parseAndValidateProviderConfig(raw providerRawConfig, diagnostics *diag.Diagnostics) (providerValidatedConfig, bool) {
	endpoint := raw.endpoint
	username := raw.username
	password := raw.password
	database := raw.database
	tlsConfig := raw.tlsConfig

	if endpoint == "" {
		endpoint = "localhost:3306"
	}
	if username == "" {
		username = "root"
	}
	if database == "" {
		database = "mysql"
	}
	if tlsConfig == "" {
		tlsConfig = "preferred"
	}

	tlsSkipVerify := false
	tlsSkipVerifySet := false
	if raw.tlsSkipVerifyEnv != "" {
		parsed, err := strconv.ParseBool(raw.tlsSkipVerifyEnv)
		if err != nil {
			diagnostics.AddError(
				"Invalid TLS Skip Verify",
				"MYSQL_TLS_SKIP_VERIFY must be a boolean: "+err.Error(),
			)
			return providerValidatedConfig{}, false
		}
		tlsSkipVerify = parsed
		tlsSkipVerifySet = true
	}
	if !raw.tlsSkipVerify.IsNull() {
		tlsSkipVerify = raw.tlsSkipVerify.ValueBool()
		tlsSkipVerifySet = true
	}

	customTLSRequested := raw.tlsCAFile != "" || raw.tlsCertFile != "" || raw.tlsKeyFile != "" || raw.tlsServerName != "" || tlsSkipVerifySet
	if customTLSRequested {
		if strings.EqualFold(tlsConfig, "false") {
			diagnostics.AddError(
				"TLS Configuration Conflict",
				"TLS is disabled via tls=\"false\" or MYSQL_TLS=false, but custom TLS settings were provided.",
			)
			return providerValidatedConfig{}, false
		}
		registeredName, err := registerTLSConfig(raw.tlsCAFile, raw.tlsCertFile, raw.tlsKeyFile, raw.tlsServerName, tlsSkipVerify)
		if err != nil {
			diagnostics.AddError(
				"Invalid TLS Configuration",
				"Failed to configure TLS settings: "+err.Error(),
			)
			return providerValidatedConfig{}, false
		}
		tlsConfig = registeredName
	}

	maxLifetime := 5 * time.Minute
	if raw.connMaxLifetimeEnv != "" {
		parsed, err := parseDuration(raw.connMaxLifetimeEnv)
		if err != nil {
			diagnostics.AddError(
				"Invalid MySQL Connection Lifetime",
				"MYSQL_CONN_MAX_LIFETIME must be a valid duration (e.g. 5m, 30s, 1h): "+err.Error(),
			)
			return providerValidatedConfig{}, false
		}
		maxLifetime = parsed
	}

	maxOpen := 5
	if raw.maxOpenConnsEnv != "" {
		parsed, err := parseNonNegativeInt(raw.maxOpenConnsEnv)
		if err != nil {
			diagnostics.AddError(
				"Invalid MySQL Max Open Conns",
				"MYSQL_MAX_OPEN_CONNS must be a non-negative integer: "+err.Error(),
			)
			return providerValidatedConfig{}, false
		}
		maxOpen = parsed
	}

	maxIdle := 5
	if raw.maxIdleConnsEnv != "" {
		parsed, err := parseNonNegativeInt(raw.maxIdleConnsEnv)
		if err != nil {
			diagnostics.AddError(
				"Invalid MySQL Max Idle Conns",
				"MYSQL_MAX_IDLE_CONNS must be a non-negative integer: "+err.Error(),
			)
			return providerValidatedConfig{}, false
		}
		maxIdle = parsed
	}

	waitTimeout := int64(10000)
	if raw.waitTimeoutEnv != "" {
		parsed, err := parsePositiveInt64(raw.waitTimeoutEnv)
		if err != nil {
			if errors.Is(err, errNonPositiveInt64) {
				diagnostics.AddError(
					"Invalid MySQL Wait Timeout",
					"MYSQL_WAIT_TIMEOUT must be a positive integer (seconds).",
				)
				return providerValidatedConfig{}, false
			}
			diagnostics.AddError(
				"Invalid MySQL Wait Timeout",
				"MYSQL_WAIT_TIMEOUT must be a positive integer (seconds): "+err.Error(),
			)
			return providerValidatedConfig{}, false
		}
		waitTimeout = parsed
	}
	if !raw.waitTimeout.IsNull() {
		parsed := raw.waitTimeout.ValueInt64()
		if err := validatePositiveInt64(parsed); err != nil {
			diagnostics.AddError(
				"Invalid MySQL Wait Timeout",
				"wait_timeout must be a positive integer (seconds).",
			)
			return providerValidatedConfig{}, false
		}
		waitTimeout = parsed
	}

	innodbLockWaitTimeout := int64(1)
	if raw.innodbLockWaitTimeoutEnv != "" {
		parsed, err := parsePositiveInt64(raw.innodbLockWaitTimeoutEnv)
		if err != nil {
			if errors.Is(err, errNonPositiveInt64) {
				diagnostics.AddError(
					"Invalid InnoDB Lock Wait Timeout",
					"MYSQL_INNODB_LOCK_WAIT_TIMEOUT must be a positive integer (seconds).",
				)
				return providerValidatedConfig{}, false
			}
			diagnostics.AddError(
				"Invalid InnoDB Lock Wait Timeout",
				"MYSQL_INNODB_LOCK_WAIT_TIMEOUT must be a positive integer (seconds): "+err.Error(),
			)
			return providerValidatedConfig{}, false
		}
		innodbLockWaitTimeout = parsed
	}
	if !raw.innodbLockWaitTimeout.IsNull() {
		parsed := raw.innodbLockWaitTimeout.ValueInt64()
		if err := validatePositiveInt64(parsed); err != nil {
			diagnostics.AddError(
				"Invalid InnoDB Lock Wait Timeout",
				"innodb_lock_wait_timeout must be a positive integer (seconds).",
			)
			return providerValidatedConfig{}, false
		}
		innodbLockWaitTimeout = parsed
	}

	lockWaitTimeout := int64(60)
	if raw.lockWaitTimeoutEnv != "" {
		parsed, err := parsePositiveInt64(raw.lockWaitTimeoutEnv)
		if err != nil {
			if errors.Is(err, errNonPositiveInt64) {
				diagnostics.AddError(
					"Invalid Lock Wait Timeout",
					"MYSQL_LOCK_WAIT_TIMEOUT must be a positive integer (seconds).",
				)
				return providerValidatedConfig{}, false
			}
			diagnostics.AddError(
				"Invalid Lock Wait Timeout",
				"MYSQL_LOCK_WAIT_TIMEOUT must be a positive integer (seconds): "+err.Error(),
			)
			return providerValidatedConfig{}, false
		}
		lockWaitTimeout = parsed
	}
	if !raw.lockWaitTimeout.IsNull() {
		parsed := raw.lockWaitTimeout.ValueInt64()
		if err := validatePositiveInt64(parsed); err != nil {
			diagnostics.AddError(
				"Invalid Lock Wait Timeout",
				"lock_wait_timeout must be a positive integer (seconds).",
			)
			return providerValidatedConfig{}, false
		}
		lockWaitTimeout = parsed
	}

	return providerValidatedConfig{
		mysqlConfig: mysql.Config{
			User:                 username,
			Passwd:               password,
			Net:                  "tcp",
			Addr:                 endpoint,
			DBName:               database,
			AllowNativePasswords: true,
			ParseTime:            true,
			TLSConfig:            tlsConfig,
			InterpolateParams:    true,
			Params: map[string]string{
				"wait_timeout":             strconv.FormatInt(waitTimeout, 10),
				"innodb_lock_wait_timeout": strconv.FormatInt(innodbLockWaitTimeout, 10),
				"lock_wait_timeout":        strconv.FormatInt(lockWaitTimeout, 10),
			},
		},
		maxLifetime:  maxLifetime,
		maxOpenConns: maxOpen,
		maxIdleConns: maxIdle,
	}, true
}

func connectAndVerify(ctx context.Context, validated providerValidatedConfig, diagnostics *diag.Diagnostics) (*sql.DB, bool) {
	db, err := sql.Open("mysql", validated.mysqlConfig.FormatDSN())
	if err != nil {
		diagnostics.AddError(
			"Unable to Create MySQL Client",
			"An unexpected error occurred when creating the MySQL client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"MySQL Client Error: "+err.Error(),
		)
		return nil, false
	}

	db.SetConnMaxLifetime(validated.maxLifetime)
	db.SetMaxOpenConns(validated.maxOpenConns)
	db.SetMaxIdleConns(validated.maxIdleConns)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		diagnostics.AddError(
			"Unable to Connect to MySQL",
			"An unexpected error occurred when connecting to MySQL. "+
				"Please verify the connection configuration.\n\n"+
				"MySQL Connection Error: "+err.Error(),
		)
		return nil, false
	}

	var componentExists int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM mysql.component WHERE component_urn = 'file://component_audit_log_filter'").Scan(&componentExists)
	if err != nil || componentExists == 0 {
		_ = db.Close()
		diagnostics.AddError(
			"Audit Log Filter Component Not Available",
			"The audit_log_filter component is not installed or enabled on this MySQL server. "+
				"Please install and enable the component before using this provider.\n\n"+
				"Error: "+fmt.Sprintf("Component check error: %v, exists: %d", err, componentExists),
		)
		return nil, false
	}

	return db, true
}

func (p *AuditLogFilterProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAuditLogFilterResource,
		NewAuditLogUserAssignmentResource,
	}
}

func (p *AuditLogFilterProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Data sources can be added here if needed
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AuditLogFilterProvider{
			version: version,
		}
	}
}

func parseNonNegativeInt(value string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if parsed < 0 {
		return 0, fmt.Errorf("value must be a non-negative integer")
	}
	return parsed, nil
}

func parseDuration(value string) (time.Duration, error) {
	return time.ParseDuration(value)
}

func parsePositiveInt64(value string) (int64, error) {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}
	if parsed < 1 {
		return 0, errNonPositiveInt64
	}
	return parsed, nil
}

func validatePositiveInt64(value int64) error {
	if value < 1 {
		return errNonPositiveInt64
	}
	return nil
}

func configStringOrEnv(attr types.String, envValue string) string {
	if !attr.IsNull() {
		return attr.ValueString()
	}
	return envValue
}
