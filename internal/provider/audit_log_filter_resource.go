package provider

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AuditLogFilterResource{}
var _ resource.ResourceWithImportState = &AuditLogFilterResource{}

func NewAuditLogFilterResource() resource.Resource {
	return &AuditLogFilterResource{}
}

func normalizeJSON(input string) (string, error) {
	var v interface{}
	if err := json.Unmarshal([]byte(input), &v); err != nil {
		return "", err
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// AuditLogFilterResource defines the resource implementation.
type AuditLogFilterResource struct {
	db *sql.DB
}

// AuditLogFilterResourceModel describes the resource data model.
type AuditLogFilterResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Definition types.String `tfsdk:"definition"`
	FilterID   types.Int64  `tfsdk:"filter_id"`
}

func (r *AuditLogFilterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_filter"
}

func (r *AuditLogFilterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Percona Server audit log filter using the audit_log_filter component.\n\n" +
			"This resource allows you to create, update, and delete audit log filters that define which " +
			"events should be logged. The filter definition must be a valid JSON object that conforms to " +
			"the MySQL audit log filter syntax.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the audit log filter (same as name).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the audit log filter. Must be unique across all filters.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"definition": schema.StringAttribute{
				Description: "JSON definition of the audit log filter. This must be a valid JSON object " +
					"that defines the filter rules according to MySQL audit log filter syntax. **WARNING**: Changing this value will cause the filter to be recreated, temporarily affecting active sessions using this filter.",
				Required: true,
				Validators: []validator.String{
					auditLogFilterDefinitionValidator{},
				},
			},
			"filter_id": schema.Int64Attribute{
				Description: "Internal filter ID assigned by MySQL.",
				Computed:    true,
			},
		},
	}
}

func (r *AuditLogFilterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AuditLogFilterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AuditLogFilterResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate filter definition structure and syntax
	if err := validateAuditLogFilterDefinition(data.Definition.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("definition"),
			"Invalid JSON Definition",
			err.Error(),
		)
		return
	}

	normalizedDefinition, err := normalizeJSON(data.Definition.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("definition"),
			"Invalid JSON Definition",
			"The filter definition must be valid JSON: "+err.Error(),
		)
		return
	}

	// Check if filter name already exists
	var existingCount int
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM mysql.audit_log_filter WHERE name = ?", data.Name.ValueString()).Scan(&existingCount)
	if err != nil {
		resp.Diagnostics.AddError("Database Error", "Failed to check existing filter: "+err.Error())
		return
	}

	if existingCount > 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("name"),
			"Filter Already Exists",
			fmt.Sprintf("A filter with name '%s' already exists", data.Name.ValueString()),
		)
		return
	}

	// Create the audit log filter using the MySQL function - use direct query due to Go driver issues with prepared statements
	var result string
	err = r.db.QueryRowContext(ctx, "SELECT audit_log_filter_set_filter(?, ?)", data.Name.ValueString(), normalizedDefinition).Scan(&result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create Filter",
			"Could not create audit log filter: "+err.Error(),
		)
		return
	}

	if result != "OK" {
		resp.Diagnostics.AddError(
			"Filter Creation Failed",
			"MySQL returned an error: "+result,
		)
		return
	}

	// Retrieve the created filter to get the filter_id
	var filterID int64
	err = r.db.QueryRowContext(ctx, "SELECT filter_id FROM mysql.audit_log_filter WHERE name = ?", data.Name.ValueString()).Scan(&filterID)
	if err != nil {
		resp.Diagnostics.AddError("Database Error", "Failed to retrieve filter ID: "+err.Error())
		return
	}

	// Set computed values
	data.ID = data.Name
	data.FilterID = types.Int64Value(filterID)
	data.Definition = types.StringValue(normalizedDefinition)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AuditLogFilterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AuditLogFilterResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Query the filter from the database
	var filterID int64
	var definition string
	err := r.db.QueryRowContext(ctx, "SELECT filter_id, filter FROM mysql.audit_log_filter WHERE name = ?", data.Name.ValueString()).Scan(&filterID, &definition)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Filter no longer exists, remove from state
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Database Error", "Failed to read filter: "+err.Error())
		return
	}

	normalizedDefinition, err := normalizeJSON(definition)
	if err != nil {
		resp.Diagnostics.AddError("Database Error", "Failed to normalize filter definition: "+err.Error())
		return
	}

	// Update the model with current database values
	data.FilterID = types.Int64Value(filterID)
	data.Definition = types.StringValue(normalizedDefinition)
	data.ID = data.Name

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AuditLogFilterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AuditLogFilterResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate filter definition structure and syntax
	if err := validateAuditLogFilterDefinition(data.Definition.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("definition"),
			"Invalid JSON Definition",
			err.Error(),
		)
		return
	}

	normalizedDefinition, err := normalizeJSON(data.Definition.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("definition"),
			"Invalid JSON Definition",
			"The filter definition must be valid JSON: "+err.Error(),
		)
		return
	}

	// Add warning about the update process
	resp.Diagnostics.AddWarning(
		"Filter Update Requires Recreation",
		"MySQL audit log filters cannot be updated in-place. The provider will remove the existing filter "+
			"and recreate it with the new definition. This may temporarily affect active sessions that are "+
			"using this filter. Sessions may need to reconnect to pick up the new filter rules.",
	)

	filterName := data.Name.ValueString()
	newDefinition := normalizedDefinition

	// Capture existing definition for rollback if recreation fails
	var oldDefinition string
	err = r.db.QueryRowContext(ctx,
		"SELECT filter FROM mysql.audit_log_filter WHERE name = ?",
		filterName,
	).Scan(&oldDefinition)
	if err != nil {
		resp.Diagnostics.AddError("Database Error", "Failed to read existing filter definition: "+err.Error())
		return
	}

	// Check which users are currently assigned to this filter before removing it
	type userAssignment struct {
		username string
		userhost string
	}
	var assignedUsers []userAssignment

	rows, err := r.db.QueryContext(ctx, "SELECT username, userhost FROM mysql.audit_log_user WHERE filtername = ?", filterName)
	if err != nil {
		resp.Diagnostics.AddError("Database Error", "Failed to check user assignments: "+err.Error())
		return
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			resp.Diagnostics.AddError("Database Error", "Failed to close user assignment rows: "+closeErr.Error())
		}
	}()

	for rows.Next() {
		var user userAssignment
		if err := rows.Scan(&user.username, &user.userhost); err != nil {
			resp.Diagnostics.AddError("Database Error", "Failed to scan user assignments: "+err.Error())
			return
		}
		assignedUsers = append(assignedUsers, user)
	}

	if err := rows.Err(); err != nil {
		resp.Diagnostics.AddError("Database Error", "Failed to iterate user assignments: "+err.Error())
		return
	}

	// Step 1: Remove the existing filter (this will also remove all user assignments)
	var removeResult string
	err = r.db.QueryRowContext(ctx, "SELECT audit_log_filter_remove_filter(?)", filterName).Scan(&removeResult)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Remove Existing Filter",
			"Could not remove existing audit log filter during update: "+err.Error(),
		)
		return
	}

	if removeResult != "OK" {
		resp.Diagnostics.AddError(
			"Filter Removal Failed",
			"MySQL returned an error when removing existing filter: "+removeResult,
		)
		return
	}

	// Step 2: Create the filter with the new definition
	var createResult string
	err = r.db.QueryRowContext(ctx, "SELECT audit_log_filter_set_filter(?, ?)", filterName, newDefinition).Scan(&createResult)
	if err != nil {
		var rollbackResult string
		rbErr := r.db.QueryRowContext(ctx,
			"SELECT audit_log_filter_set_filter(?, ?)",
			filterName,
			oldDefinition,
		).Scan(&rollbackResult)
		if rbErr != nil || rollbackResult != "OK" {
			resp.Diagnostics.AddError(
				"Failed to Recreate Filter",
				"Could not recreate audit log filter with new definition: "+err.Error()+
					". Rollback attempt also failed; manual restoration may be required.",
			)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Recreate Filter",
			"Could not recreate audit log filter with new definition: "+err.Error()+
				". The original filter definition was restored.",
		)
		return
	}

	if createResult != "OK" {
		var rollbackResult string
		rbErr := r.db.QueryRowContext(ctx,
			"SELECT audit_log_filter_set_filter(?, ?)",
			filterName,
			oldDefinition,
		).Scan(&rollbackResult)
		if rbErr != nil || rollbackResult != "OK" {
			resp.Diagnostics.AddError(
				"Filter Recreation Failed",
				"MySQL returned an error when recreating filter: "+createResult+
					". Rollback attempt also failed; manual restoration may be required.",
			)
			return
		}
		resp.Diagnostics.AddError(
			"Filter Recreation Failed",
			"MySQL returned an error when recreating filter: "+createResult+
				". The original filter definition was restored.",
		)
		return
	}

	// Step 3: Restore user assignments that were removed when the filter was deleted
	if len(assignedUsers) > 0 {
		resp.Diagnostics.AddWarning(
			"Restoring User Assignments",
			fmt.Sprintf("Restoring %d user assignments that were affected by the filter update. "+
				"These users may experience a brief interruption in audit logging.", len(assignedUsers)),
		)

		for _, user := range assignedUsers {
			var userSpec string
			if user.username == "%" {
				userSpec = "%"
			} else {
				userSpec = fmt.Sprintf("%s@%s", user.username, user.userhost)
			}

			var assignResult string
			err = r.db.QueryRowContext(ctx, "SELECT audit_log_filter_set_user(?, ?)", userSpec, filterName).Scan(&assignResult)
			if err != nil {
				resp.Diagnostics.AddWarning(
					"Failed to Restore User Assignment",
					fmt.Sprintf("Could not restore user assignment for '%s': %s. "+
						"You may need to manually reassign this user to the filter.", userSpec, err.Error()),
				)
				continue
			}

			if assignResult != "OK" {
				resp.Diagnostics.AddWarning(
					"User Assignment Restoration Failed",
					fmt.Sprintf("MySQL returned an error when restoring user assignment for '%s': %s. "+
						"You may need to manually reassign this user to the filter.", userSpec, assignResult),
				)
			}
		}
	}

	// Step 4: Retrieve the updated filter to get the new filter_id
	var filterID int64
	err = r.db.QueryRowContext(ctx, "SELECT filter_id FROM mysql.audit_log_filter WHERE name = ?", filterName).Scan(&filterID)
	if err != nil {
		resp.Diagnostics.AddError("Database Error", "Failed to retrieve updated filter: "+err.Error())
		return
	}

	// Update computed values
	data.FilterID = types.Int64Value(filterID)
	data.ID = data.Name
	data.Definition = types.StringValue(normalizedDefinition)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AuditLogFilterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AuditLogFilterResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Remove the audit log filter using the MySQL function - use direct query
	var result string
	err := r.db.QueryRowContext(ctx, "SELECT audit_log_filter_remove_filter(?)", data.Name.ValueString()).Scan(&result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Delete Filter",
			"Could not delete audit log filter: "+err.Error(),
		)
		return
	}

	if result != "OK" {
		resp.Diagnostics.AddError(
			"Filter Deletion Failed",
			"MySQL returned an error: "+result,
		)
		return
	}
}

func (r *AuditLogFilterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by filter name
	filterName := req.ID

	// Validate that the filter exists
	var filterID int64
	var definition string
	err := r.db.QueryRowContext(ctx, "SELECT filter_id, filter FROM mysql.audit_log_filter WHERE name = ?", filterName).Scan(&filterID, &definition)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			resp.Diagnostics.AddError(
				"Filter Not Found",
				fmt.Sprintf("No audit log filter found with name '%s'", filterName),
			)
			return
		}
		resp.Diagnostics.AddError("Database Error", "Failed to query filter: "+err.Error())
		return
	}

	normalizedDefinition, err := normalizeJSON(definition)
	if err != nil {
		resp.Diagnostics.AddError("Database Error", "Failed to normalize filter definition: "+err.Error())
		return
	}

	// Set the state
	data := AuditLogFilterResourceModel{
		ID:         types.StringValue(filterName),
		Name:       types.StringValue(filterName),
		Definition: types.StringValue(normalizedDefinition),
		FilterID:   types.Int64Value(filterID),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
