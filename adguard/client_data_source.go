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

// NewClientDataSource is a helper function to simplify the provider implementation
func NewClientDataSource() datasource.DataSource {
	return &clientDataSource{}
}

// clientDataSource is the data source implementation
type clientDataSource struct {
	adg *adguard.ADG
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
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"ids": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
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

// clientDataModel maps client schema data
type clientDataModel struct {
	ID                  types.String   `tfsdk:"id"`
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

// Read refreshes the Terraform state with the latest data
func (d *clientDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state clientDataModel

	// read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	// retrieve client info
	client, err := d.adg.GetClient(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Adguard Home Client",
			err.Error(),
		)
		return
	}

	// map response body to model
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

	state.ID = types.StringValue("placeholder")

	// set state
	diags := resp.State.Set(ctx, &state)
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
