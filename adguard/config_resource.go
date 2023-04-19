package adguard

import (
	"context"
	"time"

	"github.com/gmichels/adguard-client-go"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource                = &configResource{}
	_ resource.ResourceWithConfigure   = &configResource{}
	_ resource.ResourceWithImportState = &configResource{}
)

// configResource is the resource implementation
type configResource struct {
	adg *adguard.ADG
}

// configResourceModel maps config schema data
type configResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	LastUpdated             types.String `tfsdk:"last_updated"`
	FilteringEnabled        types.Bool   `tfsdk:"filtering_enabled"`
	FilteringUpdateInterval types.Int64  `tfsdk:"filtering_update_interval"`
}

// NewConfigResource is a helper function to simplify the provider implementation
func NewConfigResource() resource.Resource {
	return &configResource{}
}

// Metadata returns the resource type name
func (r *configResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config"
}

// Schema defines the schema for the resource
func (r *configResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internal identifier for this config",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the config",
				Computed:    true,
			},
			"filtering_enabled": schema.BoolAttribute{
				Description: "Whether DNS filtering is enabled",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(true),
			},
			"filtering_update_interval": schema.Int64Attribute{
				Description: "Update interval for all list-based filters, in hours",
				Computed:    true,
				Optional:    true,
				Default:     int64default.StaticInt64(24),
				Validators: []validator.Int64{
					int64validator.OneOf([]int64{1, 12, 24, 72, 168}...),
				},
			},
		},
	}
}

// Configure adds the provider configured config to the resource
func (r *configResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.adg = req.ProviderData.(*adguard.ADG)
}

// Create creates the resource and sets the initial Terraform state
func (r *configResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// retrieve values from plan
	var plan configResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// instantiate empty objects for storing plan data
	var filterConfig adguard.FilterConfig

	// populate config from plan
	filterConfig.Enabled = plan.FilteringEnabled.ValueBool()
	filterConfig.Interval = uint(plan.FilteringUpdateInterval.ValueInt64())

	// create new filter config using plan
	_, err := r.adg.ConfigureFiltering(filterConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Config",
			"Could not create config, unexpected error: "+err.Error(),
		)
		return
	}

	// there can be only one entry Config, so hardcode the ID as 1
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
func (r *configResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// get current state
	var state configResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get refreshed config value from AdGuard Home
	filterConfig, err := r.adg.GetAllFilters()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading AdGuard Home config",
			"Could not read AdGuard Home config ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// overwrite config with refreshed state
	state.FilteringEnabled = types.BoolValue(filterConfig.Enabled)
	state.FilteringUpdateInterval = types.Int64Value(int64(filterConfig.Interval))

	// set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success
func (r *configResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// retrieve values from plan
	var plan configResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// generate API request body from plan
	var filterConfig adguard.FilterConfig
	filterConfig.Enabled = plan.FilteringEnabled.ValueBool()
	filterConfig.Interval = uint(plan.FilteringUpdateInterval.ValueInt64())

	// update existing config
	_, err := r.adg.ConfigureFiltering(filterConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating AdGuard Home Config",
			"Could not update config, unexpected error: "+err.Error(),
		)
		return
	}

	// update resource state with updated items and timestamp
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// update state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success
func (r *configResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// there is no "real" delete for the configuration, so this means "restore defaults"

	var filterConfig adguard.FilterConfig

	// populate config with default values
	filterConfig.Enabled = true
	filterConfig.Interval = 24

	// set default values for the configuration
	_, err := r.adg.ConfigureFiltering(filterConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting AdGuard Home Config",
			"Could not delete config, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *configResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
