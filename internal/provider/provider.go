// Copyright (c) 0ch1r
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure AuditLogFilterProvider satisfies various provider interfaces.
var _ provider.Provider = &AuditLogFilterProvider{}

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
	Endpoint types.String `tfsdk:"endpoint"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Database types.String `tfsdk:"database"`
	TLS      types.String `tfsdk:"tls"`
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

	// Configuration values
	endpoint := os.Getenv("MYSQL_ENDPOINT")
	username := os.Getenv("MYSQL_USERNAME")
	password := os.Getenv("MYSQL_PASSWORD")
	database := os.Getenv("MYSQL_DATABASE")
	tlsConfig := os.Getenv("MYSQL_TLS")
	connMaxLifetime := os.Getenv("MYSQL_CONN_MAX_LIFETIME")
	maxOpenConns := os.Getenv("MYSQL_MAX_OPEN_CONNS")
	maxIdleConns := os.Getenv("MYSQL_MAX_IDLE_CONNS")

	if !data.Endpoint.IsNull() {
		endpoint = data.Endpoint.ValueString()
	}

	if !data.Username.IsNull() {
		username = data.Username.ValueString()
	}

	if !data.Password.IsNull() {
		password = data.Password.ValueString()
	}

	if !data.Database.IsNull() {
		database = data.Database.ValueString()
	}

	if !data.TLS.IsNull() {
		tlsConfig = data.TLS.ValueString()
	}

	// Default values
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

	maxLifetime := 5 * time.Minute
	if connMaxLifetime != "" {
		parsed, err := time.ParseDuration(connMaxLifetime)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid MySQL Connection Lifetime",
				"MYSQL_CONN_MAX_LIFETIME must be a valid duration (e.g. 5m, 30s, 1h): "+err.Error(),
			)
			return
		}
		maxLifetime = parsed
	}

	maxOpen := 5
	if maxOpenConns != "" {
		parsed, err := strconv.Atoi(maxOpenConns)
		if err != nil || parsed < 0 {
			resp.Diagnostics.AddError(
				"Invalid MySQL Max Open Conns",
				"MYSQL_MAX_OPEN_CONNS must be a non-negative integer: "+err.Error(),
			)
			return
		}
		maxOpen = parsed
	}

	maxIdle := 5
	if maxIdleConns != "" {
		parsed, err := strconv.Atoi(maxIdleConns)
		if err != nil || parsed < 0 {
			resp.Diagnostics.AddError(
				"Invalid MySQL Max Idle Conns",
				"MYSQL_MAX_IDLE_CONNS must be a non-negative integer: "+err.Error(),
			)
			return
		}
		maxIdle = parsed
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

	// Create MySQL connection
	config := mysql.Config{
		User:              username,
		Passwd:            password,
		Net:               "tcp",
		Addr:              endpoint,
		DBName:            database,
		AllowNativePasswords: true,
		ParseTime:         true,
		TLSConfig:         tlsConfig,
		InterpolateParams: true,
	}

	db, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create MySQL Client",
			"An unexpected error occurred when creating the MySQL client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"MySQL Client Error: "+err.Error(),
		)
		return
	}

	db.SetConnMaxLifetime(maxLifetime)
	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		resp.Diagnostics.AddError(
			"Unable to Connect to MySQL",
			"An unexpected error occurred when connecting to MySQL. "+
				"Please verify the connection configuration.\n\n"+
				"MySQL Connection Error: "+err.Error(),
		)
		return
	}

	// Verify audit_log_filter component is available
	var componentExists int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM mysql.component WHERE component_urn = 'file://component_audit_log_filter'").Scan(&componentExists)
	if err != nil || componentExists == 0 {
		_ = db.Close()
		resp.Diagnostics.AddError(
			"Audit Log Filter Component Not Available",
			"The audit_log_filter component is not installed or enabled on this MySQL server. "+
				"Please install and enable the component before using this provider.\n\n"+
				"Error: "+fmt.Sprintf("Component check error: %v, exists: %d", err, componentExists),
		)
		return
	}

	p.db = db
	resp.DataSourceData = db
	resp.ResourceData = db
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
