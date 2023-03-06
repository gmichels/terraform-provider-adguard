package adguard

import (
	"context"
	"time"

	"github.com/gmichels/adguard-client-go"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-community-providers/terraform-plugin-framework-utils/modifiers"
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

// clientResourceModel maps client schema data
type clientResourceModel struct {
	ID                       types.String `tfsdk:"id"`
	LastUpdated              types.String `tfsdk:"last_updated"`
	Name                     types.String `tfsdk:"name"`
	Ids                      types.List   `tfsdk:"ids"`
	UseGlobalSettings        types.Bool   `tfsdk:"use_global_settings"`
	FilteringEnabled         types.Bool   `tfsdk:"filtering_enabled"`
	ParentalEnabled          types.Bool   `tfsdk:"parental_enabled"`
	SafebrowsingEnabled      types.Bool   `tfsdk:"safebrowsing_enabled"`
	SafesearchEnabled        types.Bool   `tfsdk:"safesearch_enabled"`
	UseGlobalBlockedServices types.Bool   `tfsdk:"use_global_blocked_services"`
	BlockedServices          types.List   `tfsdk:"blocked_services"`
	Upstreams                types.List   `tfsdk:"upstreams"`
	Tags                     types.List   `tfsdk:"tags"`
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
			},
			"ids": schema.ListAttribute{
				Description: "List of identifiers for this client (IP, CIDR, MAC, or ClientID)",
				ElementType: types.StringType,
				Required:    true,
			},
			// default values are not yet an easy task using the plugin framework
			// see https://github.com/hashicorp/terraform-plugin-framework/issues/668
			// using instead https://github.com/terraform-community-providers/terraform-plugin-framework-utils/modifiers
			"use_global_settings": schema.BoolAttribute{
				Description: "Whether to use global settings on this client",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					modifiers.DefaultBool(true),
				},
			},
			"filtering_enabled": schema.BoolAttribute{
				Description: "Whether to have filtering enabled on this client",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					modifiers.DefaultBool(false),
				},
			},
			"parental_enabled": schema.BoolAttribute{
				Description: "Whether to have AdGuard parental controls enabled on this client",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					modifiers.DefaultBool(false),
				},
			},
			"safebrowsing_enabled": schema.BoolAttribute{
				Description: "Whether to have AdGuard browsing security enabled on this client",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					modifiers.DefaultBool(false),
				},
			},
			"safesearch_enabled": schema.BoolAttribute{
				Description: "Whether to enforce safe search on this client",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					modifiers.DefaultBool(false),
				},
			},
			"use_global_blocked_services": schema.BoolAttribute{
				Description: "Whether to use global settings for blocked services",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					modifiers.DefaultBool(true),
				},
			},
			"blocked_services": schema.ListAttribute{
				Description: "List of blocked services for this client",
				ElementType: types.StringType,
				Optional:    true,
				Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
			},
			"upstreams": schema.ListAttribute{
				Description: "List of upstream DNS server for this client",
				ElementType: types.StringType,
				Optional:    true,
				Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
			},
			"tags": schema.ListAttribute{
				Description: "List of tags for this client",
				ElementType: types.StringType,
				Optional:    true,
				Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
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
	var plan clientResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// instantiate empty client for storing plan data
	var clientPlan adguard.Client

	// populate client from plan
	clientPlan.Name = plan.Name.ValueString()
	diags = plan.Ids.ElementsAs(ctx, &clientPlan.Ids, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	clientPlan.UseGlobalSettings = plan.UseGlobalSettings.ValueBool()
	clientPlan.FilteringEnabled = plan.FilteringEnabled.ValueBool()
	clientPlan.ParentalEnabled = plan.ParentalEnabled.ValueBool()
	clientPlan.SafebrowsingEnabled = plan.SafebrowsingEnabled.ValueBool()
	clientPlan.SafesearchEnabled = plan.SafesearchEnabled.ValueBool()
	clientPlan.UseGlobalBlockedServices = plan.UseGlobalBlockedServices.ValueBool()
	if len(plan.BlockedServices.Elements()) > 0 {
		diags = plan.BlockedServices.ElementsAs(ctx, &clientPlan.BlockedServices, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if len(plan.Upstreams.Elements()) > 0 {
		diags = plan.Upstreams.ElementsAs(ctx, &clientPlan.Upstreams, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if len(plan.Tags.Elements()) > 0 {
		diags = plan.Tags.ElementsAs(ctx, &clientPlan.Tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// create new clientState using plan
	clientState, err := r.adg.CreateClient(clientPlan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating client",
			"Could not create client, unexpected error: "+err.Error(),
		)
		return
	}

	// response sent by AdGuard Home is the same as the sent payload,
	// just add missing attributes for state
	plan.ID = types.StringValue(clientState.Name)
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
	var state clientResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get refreshed client value from AdGuard Home
	client, err := r.adg.GetClient(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading AdGuard Home Client",
			"Could not read AdGuard Home client ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	} else if client == nil {
		resp.Diagnostics.AddError(
			"Error Reading AdGuard Home Client",
			"No such AdGuard Home client with ID "+state.ID.ValueString(),
		)
		return
	}

	// overwrite client with refreshed state
	state.Name = types.StringValue(client.Name)
	state.Ids, diags = types.ListValueFrom(ctx, types.StringType, client.Ids)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.UseGlobalSettings = types.BoolValue(client.UseGlobalSettings)
	state.FilteringEnabled = types.BoolValue(client.FilteringEnabled)
	state.ParentalEnabled = types.BoolValue(client.ParentalEnabled)
	state.SafebrowsingEnabled = types.BoolValue(client.SafebrowsingEnabled)
	state.SafesearchEnabled = types.BoolValue(client.SafesearchEnabled)
	state.UseGlobalBlockedServices = types.BoolValue(client.UseGlobalBlockedServices)
	state.BlockedServices, diags = types.ListValueFrom(ctx, types.StringType, client.BlockedServices)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Upstreams, diags = types.ListValueFrom(ctx, types.StringType, client.Upstreams)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Tags, diags = types.ListValueFrom(ctx, types.StringType, client.Tags)
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
func (r *clientResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// retrieve values from plan
	var plan clientResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// generate API request body from plan
	var clientUpdate adguard.ClientUpdate
	clientUpdate.Name = plan.Name.ValueString()
	clientUpdate.Data.Name = plan.Name.ValueString()
	diags = plan.Ids.ElementsAs(ctx, &clientUpdate.Data.Ids, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	clientUpdate.Data.UseGlobalSettings = plan.UseGlobalSettings.ValueBool()
	clientUpdate.Data.FilteringEnabled = plan.FilteringEnabled.ValueBool()
	clientUpdate.Data.ParentalEnabled = plan.ParentalEnabled.ValueBool()
	clientUpdate.Data.SafebrowsingEnabled = plan.SafebrowsingEnabled.ValueBool()
	clientUpdate.Data.SafesearchEnabled = plan.SafesearchEnabled.ValueBool()
	clientUpdate.Data.UseGlobalBlockedServices = plan.UseGlobalBlockedServices.ValueBool()
	if len(plan.BlockedServices.Elements()) > 0 {
		diags = plan.BlockedServices.ElementsAs(ctx, &clientUpdate.Data.BlockedServices, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if len(plan.Upstreams.Elements()) > 0 {
		diags = plan.Upstreams.ElementsAs(ctx, &clientUpdate.Data.Upstreams, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if len(plan.Tags.Elements()) > 0 {
		diags = plan.Tags.ElementsAs(ctx, &clientUpdate.Data.Tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// update existing client
	_, err := r.adg.UpdateClient(clientUpdate)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating AdGuard Home Client",
			"Could not update client, unexpected error: "+err.Error(),
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
func (r *clientResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// retrieve values from state
	var state clientResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var clientDelete adguard.ClientDelete
	clientDelete.Name = state.ID.ValueString()
	// delete existing client
	err := r.adg.DeleteClient(clientDelete)
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
