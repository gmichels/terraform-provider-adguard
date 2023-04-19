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
	_ datasource.DataSource              = &configDataSource{}
	_ datasource.DataSourceWithConfigure = &configDataSource{}
)

// configDataSource is the data source implementation
type configDataSource struct {
	adg *adguard.ADG
}

// configDataModel maps Config schema data
type configDataModel struct {
	ID                      types.String `tfsdk:"id"`
	FilteringEnabled        types.Bool   `tfsdk:"filtering_enabled"`
	FilteringUpdateInterval types.Int64  `tfsdk:"filtering_update_interval"`
}

// NewConfigDataSource is a helper function to simplify the provider implementation
func NewConfigDataSource() datasource.DataSource {
	return &configDataSource{}
}

// Metadata returns the data source type name
func (d *configDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config"
}

// Schema defines the schema for the data source
func (d *configDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier attribute",
				Computed:    true,
			},
			"filtering_enabled": schema.BoolAttribute{
				Description: "Whether DNS filtering is enabled",
				Computed:    true,
			},
			"filtering_update_interval": schema.Int64Attribute{
				Description: "Update interval for all list-based filters, in hours",
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data
func (d *configDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// read Terraform configuration data into the model
	var state configDataModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	// retrieve config info
	config, err := d.adg.GetAllFilters()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}

	// map response body to model
	state.FilteringEnabled = types.BoolValue(config.Enabled)
	state.FilteringUpdateInterval = types.Int64Value(int64(config.Interval))

	// set ID placeholder for testing
	state.ID = types.StringValue("placeholder")

	// set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured config to the data source
func (d *configDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.adg = req.ProviderData.(*adguard.ADG)
}
