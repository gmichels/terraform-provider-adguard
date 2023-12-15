package adguard

import (
	"context"
	"time"

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
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform refresh",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the client",
				Required:    true,
			},
			"ids": schema.ListAttribute{
				Description: "List of identifiers for this client (IP, CIDR, MAC, or ClientID)",
				ElementType: types.StringType,
				Computed:    true,
			},
			"use_global_settings": schema.BoolAttribute{
				Description: "Whether to use global settings on this client",
				Computed:    true,
			},
			"filtering_enabled": schema.BoolAttribute{
				Description: "Whether to have filtering enabled on this client",
				Computed:    true,
			},
			"parental_enabled": schema.BoolAttribute{
				Description: "Whether to have AdGuard parental controls enabled on this client",
				Computed:    true,
			},
			"safebrowsing_enabled": schema.BoolAttribute{
				Description: "Whether to have AdGuard browsing security enabled on this client",
				Computed:    true,
			},
			"safesearch": safeSearchDatasourceSchema(),
			"use_global_blocked_services": schema.BoolAttribute{
				Description: "Whether to use global settings for blocked services",
				Computed:    true,
			},
			"blocked_services": schema.SetAttribute{
				Description: "Set of blocked services for this client",
				ElementType: types.StringType,
				Computed:    true,
			},
			"blocked_services_pause_schedule": scheduleDatasourceSchema(),
			"upstreams": schema.ListAttribute{
				Description: "List of upstream DNS server for this client",
				ElementType: types.StringType,
				Computed:    true,
			},
			"tags": schema.SetAttribute{
				Description: "Set of tags for this client",
				ElementType: types.StringType,
				Computed:    true,
			},
			"ignore_querylog": schema.BoolAttribute{
				Description: "Whether to this client writes to the query log",
				Computed:    true,
			},
			"ignore_statistics": schema.BoolAttribute{
				Description: "Whether to this client is included in the statistics",
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data
func (d *clientDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// read Terraform configuration data into the model
	var state clientCommonModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	// use common model for state
	var newState clientCommonModel
	// use common Read function
	newState.Read(ctx, *d.adg, &state, &resp.Diagnostics, "datasource")
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// set ID placeholder for testing
	newState.ID = types.StringValue("placeholder")
	// set last updated just because we need it in the resource as well
	newState.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// set state
	diags = resp.State.Set(ctx, &newState)
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
