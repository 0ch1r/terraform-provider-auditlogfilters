// Copyright (c) 0ch1r
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AuditLogUserAssignmentResource{}
var _ resource.ResourceWithImportState = &AuditLogUserAssignmentResource{}

func NewAuditLogUserAssignmentResource() resource.Resource {
	return &AuditLogUserAssignmentResource{}
}

// AuditLogUserAssignmentResource defines the resource implementation.
type AuditLogUserAssignmentResource struct {
	db *sql.DB
}

// AuditLogUserAssignmentResourceModel describes the resource data model.
type AuditLogUserAssignmentResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Username   types.String `tfsdk:"username"`
	Userhost   types.String `tfsdk:"userhost"`
	FilterName types.String `tfsdk:"filter_name"`
}

func (r *AuditLogUserAssignmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_assignment"
}

func (r *AuditLogUserAssignmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages user assignments to audit log filters using the audit_log_filter component.\n\n" +
			"This resource allows you to assign users to specific audit log filters, enabling targeted " +
			"auditing for different users. The user can be specified with a username and host pattern, " +
			"or use '%' for the default filter assignment.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the user assignment (username@userhost).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Description: "MySQL username to assign the filter to. Use '%' for default assignment.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"userhost": schema.StringAttribute{
				Description: "Host pattern for the user assignment. Use '%' to match any host. " +
					"This is combined with username to form the complete user specification.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"filter_name": schema.StringAttribute{
				Description: "Name of the audit log filter to assign to the user. The filter must exist.",
				Required:    true,
			},
		},
	}
}

func (r *AuditLogUserAssignmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	db, ok := req.ProviderData.(*sql.DB)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sql.DB, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.db = db
}

// buildUserSpec constructs the user specification for MySQL functions
func (r *AuditLogUserAssignmentResource) buildUserSpec(username, userhost string) string {
	if username == "%" {
		return "%"
	}
	if userhost == "" {
		userhost = "%"
	}
	return fmt.Sprintf("%s@%s", username, userhost)
}

// parseUserSpec parses a user specification into username and userhost components
func (r *AuditLogUserAssignmentResource) parseUserSpec(userSpec string) (username, userhost string) {
	if userSpec == "%" {
		return "%", ""
	}
	parts := strings.SplitN(userSpec, "@", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], "%"
}

func (r *AuditLogUserAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AuditLogUserAssignmentResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Set default userhost if not provided
	userhost := data.Userhost.ValueString()
	if userhost == "" {
		userhost = "%"
		data.Userhost = types.StringValue(userhost)
	}

	username := data.Username.ValueString()
	filterName := data.FilterName.ValueString()

	// Verify the filter exists
	var filterCount int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM mysql.audit_log_filter WHERE name = ?", filterName).Scan(&filterCount)
	if err != nil {
		resp.Diagnostics.AddError("Database Error", "Failed to check filter existence: "+err.Error())
		return
	}

	if filterCount == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("filter_name"),
			"Filter Not Found",
			fmt.Sprintf("No audit log filter found with name '%s'", filterName),
		)
		return
	}

	// Check if assignment already exists
	var existingCount int
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM mysql.audit_log_user WHERE username = ? AND userhost = ?", username, userhost).Scan(&existingCount)
	if err != nil {
		resp.Diagnostics.AddError("Database Error", "Failed to check existing assignment: "+err.Error())
		return
	}

	if existingCount > 0 {
		resp.Diagnostics.AddError(
			"Assignment Already Exists",
			fmt.Sprintf("User assignment already exists for '%s@%s'", username, userhost),
		)
		return
	}

	// Create the user assignment using the MySQL function - use direct query due to Go driver issues
	userSpec := r.buildUserSpec(username, userhost)
	var result string
	err = r.db.QueryRowContext(ctx, "SELECT audit_log_filter_set_user(?, ?)", userSpec, filterName).Scan(&result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create User Assignment",
			"Could not create audit log user assignment: "+err.Error(),
		)
		return
	}

	if result != "OK" {
		resp.Diagnostics.AddError(
			"User Assignment Creation Failed",
			"MySQL returned an error: "+result,
		)
		return
	}

	// Set computed values
	data.ID = types.StringValue(fmt.Sprintf("%s@%s", username, userhost))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AuditLogUserAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AuditLogUserAssignmentResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	username := data.Username.ValueString()
	userhost := data.Userhost.ValueString()
	if userhost == "" {
		userhost = "%"
	}

	// Query the user assignment from the database
	var filterName string
	err := r.db.QueryRowContext(ctx, "SELECT filtername FROM mysql.audit_log_user WHERE username = ? AND userhost = ?", username, userhost).Scan(&filterName)
	if err != nil {
		if err == sql.ErrNoRows {
			// Assignment no longer exists, remove from state
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Database Error", "Failed to read user assignment: "+err.Error())
		return
	}

	// Update the model with current database values
	data.FilterName = types.StringValue(filterName)
	data.Userhost = types.StringValue(userhost)
	data.ID = types.StringValue(fmt.Sprintf("%s@%s", username, userhost))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AuditLogUserAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AuditLogUserAssignmentResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	username := data.Username.ValueString()
	userhost := data.Userhost.ValueString()
	if userhost == "" {
		userhost = "%"
		data.Userhost = types.StringValue(userhost)
	}
	filterName := data.FilterName.ValueString()

	// Verify the new filter exists
	var filterCount int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM mysql.audit_log_filter WHERE name = ?", filterName).Scan(&filterCount)
	if err != nil {
		resp.Diagnostics.AddError("Database Error", "Failed to check filter existence: "+err.Error())
		return
	}

	if filterCount == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("filter_name"),
			"Filter Not Found",
			fmt.Sprintf("No audit log filter found with name '%s'", filterName),
		)
		return
	}

	// Update the user assignment using the MySQL function - use direct query
	userSpec := r.buildUserSpec(username, userhost)
	var result string
	err = r.db.QueryRowContext(ctx, "SELECT audit_log_filter_set_user(?, ?)", userSpec, filterName).Scan(&result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Update User Assignment",
			"Could not update audit log user assignment: "+err.Error(),
		)
		return
	}

	if result != "OK" {
		resp.Diagnostics.AddError(
			"User Assignment Update Failed",
			"MySQL returned an error: "+result,
		)
		return
	}

	// Update computed values
	data.ID = types.StringValue(fmt.Sprintf("%s@%s", username, userhost))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AuditLogUserAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AuditLogUserAssignmentResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	username := data.Username.ValueString()
	userhost := data.Userhost.ValueString()
	if userhost == "" {
		userhost = "%"
	}

	// Remove the user assignment using the MySQL function - use direct query
	userSpec := r.buildUserSpec(username, userhost)
	var result string
	err := r.db.QueryRowContext(ctx, "SELECT audit_log_filter_remove_user(?)", userSpec).Scan(&result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Delete User Assignment",
			"Could not delete audit log user assignment: "+err.Error(),
		)
		return
	}

	if result != "OK" {
		resp.Diagnostics.AddError(
			"User Assignment Deletion Failed",
			"MySQL returned an error: "+result,
		)
		return
	}
}

func (r *AuditLogUserAssignmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by user specification (username@userhost)
	userSpec := req.ID
	username, userhost := r.parseUserSpec(userSpec)

	// Validate that the assignment exists
	var filterName string
	err := r.db.QueryRowContext(ctx, "SELECT filtername FROM mysql.audit_log_user WHERE username = ? AND userhost = ?", username, userhost).Scan(&filterName)
	if err != nil {
		if err == sql.ErrNoRows {
			resp.Diagnostics.AddError(
				"User Assignment Not Found",
				fmt.Sprintf("No user assignment found for '%s'", userSpec),
			)
			return
		}
		resp.Diagnostics.AddError("Database Error", "Failed to query user assignment: "+err.Error())
		return
	}

	// Set the state
	data := AuditLogUserAssignmentResourceModel{
		ID:         types.StringValue(userSpec),
		Username:   types.StringValue(username),
		Userhost:   types.StringValue(func() string {
			if userhost == "" {
				return "%"
			}
			return userhost
		}()),
		FilterName: types.StringValue(filterName),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
