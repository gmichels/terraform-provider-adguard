package adguard

import (
	"context"
	"time"

	"github.com/gmichels/adguard-client-go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource              = &clientResource{}
	_ resource.ResourceWithConfigure = &clientResource{}
)

// NewClientResource is a helper function to simplify the provider implementation
func NewClientResource() resource.Resource {
	return &clientResource{}
}

// clientResource is the resource implementation
type clientResource struct {
	adg *adguard.ADG
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
				Computed: true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"ids": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			// default values are not yet an easy task using the plugin framework
			// see https://github.com/hashicorp/terraform-plugin-framework/issues/668
			"use_global_settings": schema.BoolAttribute{
				Computed: true,
				Optional: true,
			},
			"filtering_enabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
			},
			"parental_enabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
			},
			"safebrowsing_enabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
			},
			"blocked_services": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Optional:    true,
			},
			"upstreams": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Optional:    true,
			},
			"tags": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Optional:    true,
			},
		},
	}
}

// clientResourceModel maps client schema data
type clientResourceModel struct {
	ID                  types.String   `tfsdk:"id"`
	LastUpdated         types.String   `tfsdk:"last_updated"`
	Name                types.String   `tfsdk:"name"`
	Ids                 []types.String `tfsdk:"ids"`
	UseGlobalSettings   types.Bool     `tfsdk:"use_global_settings"`
	FilteringEnabled    types.Bool     `tfsdk:"filtering_enabled"`
	ParentalEnabled     types.Bool     `tfsdk:"parental_enabled"`
	SafebrowsingEnabled types.Bool     `tfsdk:"safebrowsing_enabled"`
	BlockedServices     []types.String `tfsdk:"blocked_services"`
	Upstreams           []types.String `tfsdk:"upstreams"`
	Tags                []types.String `tfsdk:"tags"`
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
	for _, id := range plan.Ids {
		clientPlan.Ids = append(clientPlan.Ids, id.ValueString())
	}
	clientPlan.UseGlobalSettings = plan.UseGlobalSettings.ValueBool()
	clientPlan.FilteringEnabled = plan.FilteringEnabled.ValueBool()
	clientPlan.ParentalEnabled = plan.ParentalEnabled.ValueBool()
	clientPlan.SafebrowsingEnabled = plan.SafebrowsingEnabled.ValueBool()
	for _, blockedService := range plan.BlockedServices {
		clientPlan.BlockedServices = append(clientPlan.BlockedServices, blockedService.ValueString())
	}
	for _, upstream := range plan.Upstreams {
		clientPlan.Upstreams = append(clientPlan.Upstreams, upstream.ValueString())
	}
	for _, tag := range plan.Tags {
		clientPlan.Tags = append(clientPlan.Tags, tag.ValueString())
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

	// map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(clientState.Name)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	plan.Name = types.StringValue(clientState.Name)
	for _, id := range clientState.Ids {
		plan.Ids = append(plan.Ids, types.StringValue(id))
	}
	plan.UseGlobalSettings = types.BoolValue(clientState.UseGlobalSettings)
	plan.FilteringEnabled = types.BoolValue(clientState.FilteringEnabled)
	plan.ParentalEnabled = types.BoolValue(clientState.ParentalEnabled)
	plan.SafebrowsingEnabled = types.BoolValue(clientState.SafebrowsingEnabled)
	for _, blockedService := range clientState.BlockedServices {
		plan.BlockedServices = append(plan.BlockedServices, types.StringValue(blockedService))
	}
	for _, upstream := range clientState.Upstreams {
		plan.Upstreams = append(plan.Upstreams, types.StringValue(upstream))
	}
	for _, tag := range clientState.Tags {
		plan.Tags = append(plan.Tags, types.StringValue(tag))
	}

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

	// get refreshed order value from Adguard Home
	client, err := r.adg.GetClient(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Adguard Home Client",
			"Could not read Adguard Home client ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// overwrite items with refreshed state
	state.Name = types.StringValue(client.Name)
	for _, id := range client.Ids {
		state.Ids = append(state.Ids, types.StringValue(id))
	}
	state.UseGlobalSettings = types.BoolValue(client.UseGlobalSettings)
	state.FilteringEnabled = types.BoolValue(client.FilteringEnabled)
	state.ParentalEnabled = types.BoolValue(client.ParentalEnabled)
	state.SafebrowsingEnabled = types.BoolValue(client.SafebrowsingEnabled)
	for _, blockedService := range client.BlockedServices {
		state.BlockedServices = append(state.BlockedServices, types.StringValue(blockedService))
	}
	for _, upstream := range client.Upstreams {
		state.Upstreams = append(state.Upstreams, types.StringValue(upstream))
	}
	for _, tag := range client.Tags {
		state.Tags = append(state.Tags, types.StringValue(tag))
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
	for _, id := range plan.Ids {
		clientUpdate.Data.Ids = append(clientUpdate.Data.Ids, id.ValueString())
	}
	clientUpdate.Data.UseGlobalSettings = plan.UseGlobalSettings.ValueBool()
	clientUpdate.Data.FilteringEnabled = plan.FilteringEnabled.ValueBool()
	clientUpdate.Data.ParentalEnabled = plan.ParentalEnabled.ValueBool()
	clientUpdate.Data.SafebrowsingEnabled = plan.SafebrowsingEnabled.ValueBool()
	for _, blockedService := range plan.BlockedServices {
		clientUpdate.Data.BlockedServices = append(clientUpdate.Data.BlockedServices, blockedService.ValueString())
	}
	for _, upstream := range plan.Upstreams {
		clientUpdate.Data.Upstreams = append(clientUpdate.Data.Upstreams, upstream.ValueString())
	}
	for _, tag := range plan.Tags {
		clientUpdate.Data.Tags = append(clientUpdate.Data.Tags, tag.ValueString())
	}

	// update existing client
	_, err := r.adg.UpdateClient(clientUpdate)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Adguard Home Client",
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
			"Error Deleting Adguard Home Client",
			"Could not delete client, unexpected error: "+err.Error(),
		)
		return
	}
}
