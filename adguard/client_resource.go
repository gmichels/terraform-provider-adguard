package adguard

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/gmichels/adguard-client-go"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	_ resource.Resource                = &clientResource{}
	_ resource.ResourceWithConfigure   = &clientResource{}
	_ resource.ResourceWithImportState = &clientResource{}
)

// clientResource is the resource implementation
type clientResource struct {
	adg *adguard.ADG
}

// NewClientResource is a helper function to simplify the provider implementation
func NewClientResource() resource.Resource {
	return &clientResource{}
}

// Metadata returns the resource type name
func (r *clientResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client"
}

// Schema defines the schema for the resource
func (r *clientResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internal identifier for this client",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the client",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the client",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ids": schema.SetAttribute{
				Description: "Set of identifiers for this client (IP, CIDR, MAC, or ClientID)",
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(
							regexp.MustCompile(`^[a-z0-9/.:-]+$`),
							"must be an IP address/CIDR, MAC address, or only contain numbers, lowercase letters, and hyphens",
						),
					),
				},
			},
			"use_global_settings": schema.BoolAttribute{
				Description: fmt.Sprintf("Whether to use global settings on this client. Defaults to `%t`", CLIENT_USE_GLOBAL_SETTINGS),
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(CLIENT_USE_GLOBAL_SETTINGS),
			},
			"filtering_enabled": schema.BoolAttribute{
				Description: fmt.Sprintf("Whether to have filtering enabled on this client. Defaults to `%t`", CLIENT_FILTERING_ENABLED),
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(CLIENT_FILTERING_ENABLED),
			},
			"parental_enabled": schema.BoolAttribute{
				Description: fmt.Sprintf("Whether to have AdGuard parental controls enabled on this client. Defaults to `%t`", CLIENT_PARENTAL_CONTROL_ENABLED),
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(CLIENT_PARENTAL_CONTROL_ENABLED),
			},
			"safebrowsing_enabled": schema.BoolAttribute{
				Description: fmt.Sprintf("Whether to have AdGuard browsing security enabled on this client. Defaults to `%t`", CLIENT_SAFEBROWSING_ENABLED),
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(CLIENT_SAFEBROWSING_ENABLED),
			},
			"safesearch": safeSearchResourceSchema(),
			"use_global_blocked_services": schema.BoolAttribute{
				Description: fmt.Sprintf("Whether to use global settings for blocked services. Defaults to `%t`", CLIENT_USE_GLOBAL_BLOCKED_SERVICES),
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(CLIENT_USE_GLOBAL_BLOCKED_SERVICES),
			},
			"blocked_services_pause_schedule": scheduleResourceSchema(),
			"blocked_services": schema.SetAttribute{
				Description: "Set of blocked services for this client",
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(stringvalidator.OneOf(BLOCKED_SERVICES_OPTIONS...)),
				},
			},
			"upstreams": schema.ListAttribute{
				Description: "List of upstream DNS server for this client",
				ElementType: types.StringType,
				Optional:    true,
				Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
			},
			"tags": schema.SetAttribute{
				Description: "Set of tags for this client",
				ElementType: types.StringType,
				Optional:    true,
				Validators:  []validator.Set{setvalidator.SizeAtLeast(1)},
			},
			"ignore_querylog": schema.BoolAttribute{
				Description: fmt.Sprintf("Whether to write to the query log. Defaults to `%t`", CLIENT_IGNORE_QUERYLOG),
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(CLIENT_IGNORE_QUERYLOG),
			},
			"ignore_statistics": schema.BoolAttribute{
				Description: fmt.Sprintf("Whether to be included in the statistics. Defaults to `%t`", CLIENT_IGNORE_STATISTICS),
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(CLIENT_IGNORE_STATISTICS),
			},
			"upstreams_cache_enabled": schema.BoolAttribute{
				Description: fmt.Sprintf("Whether to enable DNS caching for this client's custom upstream configuration. Defaults to `%t`", CLIENT_UPSTREAMS_CACHE_ENABLED),
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(CLIENT_UPSTREAMS_CACHE_ENABLED),
			},
			"upstreams_cache_size": schema.Int64Attribute{
				Description: "The upstreams DNS cache size, in bytes",
				Computed:    true,
				Optional:    true,
				Default:     int64default.StaticInt64(CLIENT_UPSTREAMS_CACHE_SIZE),
				Validators: []validator.Int64{
					int64validator.Between(CLIENT_UPSTREAMS_CACHE_SIZE, CLIENT_UPSTREAMS_CACHE_SIZE_MAX),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *clientResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.adg = req.ProviderData.(*adguard.ADG)
}

// Create creates the resource and sets the initial Terraform state
func (r *clientResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// retrieve values from plan
	var plan clientCommonModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// defer to common function to create or update the resource
	r.CreateOrUpdate(ctx, &plan, &resp.Diagnostics, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// response sent by AdGuard Home is the same as the sent payload,
	// just add missing attributes for state
	plan.ID = types.StringValue(plan.Name.ValueString())
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
func (r *clientResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// get current state
	var state clientCommonModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// use common model for state
	var newState clientCommonModel
	// use common Read function
	newState.Read(ctx, *r.adg, &state, &resp.Diagnostics, "resource")
	if resp.Diagnostics.HasError() {
		return
	}

	// populate internal fields into new state
	newState.ID = state.ID
	newState.LastUpdated = state.LastUpdated

	// set refreshed state
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success
func (r *clientResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// retrieve values from plan
	var plan clientCommonModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// defer to common function to create or update the resource
	r.CreateOrUpdate(ctx, &plan, &resp.Diagnostics, false)
	if resp.Diagnostics.HasError() {
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
func (r *clientResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// retrieve values from state
	var state clientCommonModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var deleteClient adguard.ClientDelete
	deleteClient.Name = state.ID.ValueString()
	// delete existing client
	err := r.adg.DeleteClient(deleteClient)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting AdGuard Home Client",
			"Could not delete client, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *clientResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
