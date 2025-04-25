package adguard

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gmichels/adguard-client-go"
	adgmodels "github.com/gmichels/adguard-client-go/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource                = &userRulesResource{}
	_ resource.ResourceWithConfigure   = &userRulesResource{}
	_ resource.ResourceWithImportState = &userRulesResource{}
)

// userRulesResource is the resource implementation
type userRulesResource struct {
	adg *adguard.ADG
}

// userRulesResourceModel maps user rules schema data
type userRulesResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Rules       types.List   `tfsdk:"rules"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

// NewUserRulesResource is a helper function to simplify the provider implementation
func NewUserRulesResource() resource.Resource {
	return &userRulesResource{}
}

// Metadata returns the resource type name
func (r *userRulesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_rules"
}

// Schema defines the schema for the resource
func (r *userRulesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier attribute",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"rules": schema.ListAttribute{
				Description: "List of user rules",
				ElementType: types.StringType,
				Required:    true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the client",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *userRulesResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.adg = req.ProviderData.(*adguard.ADG)
}

// Create creates the resource and sets the initial Terraform state
func (r *userRulesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// retrieve values from plan
	var plan userRulesResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// instantiate empty client for storing plan data
	var userRules adgmodels.SetRulesRequest

	// populate user rules from plan
	if len(plan.Rules.Elements()) > 0 {
		diags = plan.Rules.ElementsAs(ctx, &userRules.Rules, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// create user rules using plan
	err := r.adg.FilteringSetRules(userRules)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating AdGuard Home User Rules",
			"Could not create user rules, unexpected error: "+err.Error(),
		)
		return
	}

	// there can be only one entry for user rules, so hardcode the ID as 1
	plan.ID = types.StringValue("1")
	// add the last updated attribute
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data
func (r *userRulesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// get current state
	var state userRulesResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// retrieve user rules info from all filters
	allFilters, err := r.adg.FilteringStatus()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home User Rules",
			err.Error(),
		)
		return
	}
	// convert to JSON for response logging
	userRulesJson, err := json.Marshal(allFilters.UserRules)
	if err != nil {
		diags.AddError(
			"Unable to Parse AdGuard Home User Rules",
			err.Error(),
		)
		return
	}
	// log response body
	tflog.Debug(ctx, "ADG API response", map[string]interface{}{
		"object": "userRules",
		"body":   string(userRulesJson),
	})

	// overwrite user rules with refreshed state
	state.Rules, diags = types.ListValueFrom(ctx, types.StringType, allFilters.UserRules)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success
func (r *userRulesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// retrieve values from plan
	var plan userRulesResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// instantiate empty client for storing plan data
	var userRules adgmodels.SetRulesRequest

	// populate user rules from plan
	if len(plan.Rules.Elements()) > 0 {
		diags = plan.Rules.ElementsAs(ctx, &userRules.Rules, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// update user rules using plan
	err := r.adg.FilteringSetRules(userRules)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating AdGuard Home User Rules",
			"Could not update user rules, unexpected error: "+err.Error(),
		)
		return
	}

	// populate plan with computed attributes
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// update state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success
func (r *userRulesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// retrieve values from state
	var state userRulesResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// empty variable
	var userRules adgmodels.SetRulesRequest

	// delete existing user rules
	err := r.adg.FilteringSetRules(userRules)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting AdGuard Home User Rules",
			"Could not delete user rules, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *userRulesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
