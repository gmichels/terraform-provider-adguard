package adguard

import (
	"context"

	"github.com/gmichels/adguard-client-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ensure the implementation satisfies the expected interfaces
var (
	_ datasource.DataSource              = &clientDataSource{}
	_ datasource.DataSourceWithConfigure = &clientDataSource{}
)

// clientDataSource is the data source implementation
type clientDataSource struct {
	adg *adguard.ADG
}

// clientDataModel maps client schema data
type clientDataModel struct {
	ID                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Ids                      types.List   `tfsdk:"ids"`
	UseGlobalSettings        types.Bool   `tfsdk:"use_global_settings"`
	FilteringEnabled         types.Bool   `tfsdk:"filtering_enabled"`
	ParentalEnabled          types.Bool   `tfsdk:"parental_enabled"`
	SafebrowsingEnabled      types.Bool   `tfsdk:"safebrowsing_enabled"`
	SafesearchEnabled        types.Bool   `tfsdk:"safesearch_enabled"`
	UseGlobalBlockedServices types.Bool   `tfsdk:"use_global_blocked_services"`
	BlockedServices          types.Set    `tfsdk:"blocked_services"`
	Upstreams                types.List   `tfsdk:"upstreams"`
	Tags                     types.Set    `tfsdk:"tags"`
}

// NewClientDataSource is a helper function to simplify the provider implementation
func NewClientDataSource() datasource.DataSource {
	return &clientDataSource{}
}

// Metadata returns the data source type name
func (d *clientDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client"
}

// Schema defines the schema for the data source
func (d *clientDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder identifier attribute",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the client",
				Required:    true,
			},
			"ids": schema.ListAttribute{
				Description: "List of identifiers for this client (IP, CIDR, MAC, or ClientID)",
				ElementType: types.StringType,
				Optional:    true,
			},
			"use_global_settings": schema.BoolAttribute{
				Description: "Whether to use global settings on this client",
				Optional:    true,
			},
			"filtering_enabled": schema.BoolAttribute{
				Description: "Whether to have filtering enabled on this client",
				Optional:    true,
			},
			"parental_enabled": schema.BoolAttribute{
				Description: "Whether to have AdGuard parental controls enabled on this client",
				Optional:    true,
			},
			"safebrowsing_enabled": schema.BoolAttribute{
				Description: "Whether to have AdGuard browsing security enabled on this client",
				Optional:    true,
			},
			"safesearch_enabled": schema.BoolAttribute{
				Description: "Whether to enforce safe search on this client",
				Optional:    true,
			},
			"use_global_blocked_services": schema.BoolAttribute{
				Description: "Whether to use global settings for blocked services",
				Optional:    true,
			},
			"blocked_services": schema.SetAttribute{
				Description: "Set of blocked services for this client",
				ElementType: types.StringType,
				Optional:    true,
			},
			"upstreams": schema.ListAttribute{
				Description: "List of upstream DNS server for this client",
				ElementType: types.StringType,
				Optional:    true,
			},
			"tags": schema.SetAttribute{
				Description: "Set of tags for this client",
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data
func (d *clientDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// read Terraform configuration data into the model
	var state clientDataModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	// retrieve client info
	client, err := d.adg.GetClient(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home Client",
			err.Error(),
		)
		return
	}
	if client == nil {
		resp.Diagnostics.AddError(
			"Unable to Locate AdGuard Home Client",
			"No client with name `"+state.Name.ValueString()+"` exists in AdGuard Home.",
		)
		return
	}

	// map response body to model
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
	state.BlockedServices, diags = types.SetValueFrom(ctx, types.StringType, client.BlockedServices)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Upstreams, diags = types.ListValueFrom(ctx, types.StringType, client.Upstreams)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Tags, diags = types.SetValueFrom(ctx, types.StringType, client.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// set ID placeholder for testing
	state.ID = types.StringValue("placeholder")

	// set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source
func (d *clientDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.adg = req.ProviderData.(*adguard.ADG)
}
