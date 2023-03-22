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
	_ datasource.DataSource              = &listFilterDataSource{}
	_ datasource.DataSourceWithConfigure = &listFilterDataSource{}
)

// listFilterDataSource is the data source implementation
type listFilterDataSource struct {
	adg *adguard.ADG
}

// listFilterDataModel maps filter schema data
type listFilterDataModel struct {
	Url         types.String `tfsdk:"url"`
	Name        types.String `tfsdk:"name"`
	LastUpdated types.String `tfsdk:"last_updated"`
	Id          types.Int64  `tfsdk:"id"`
	RulesCount  types.Int64  `tfsdk:"rules_count"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Whitelist   types.Bool   `tfsdk:"whitelist"`
}

// NewListFilterDataSource is a helper function to simplify the provider implementation
func NewListFilterDataSource() datasource.DataSource {
	return &listFilterDataSource{}
}

// Metadata returns the data source type name
func (d *listFilterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_list_filter"
}

// Schema defines the schema for the data source
func (d *listFilterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the list filter",
				Required:    true,
			},
			"whitelist": schema.BoolAttribute{
				Description: "Then `true`, will consider this list filter of type whitelist",
				Optional:    true,
			},
			"url": schema.StringAttribute{
				Description: "Url of the list filter",
				Computed:    true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of last synchronization",
				Computed:    true,
			},
			"id": schema.Int64Attribute{
				Description: "Identifier attribute",
				Computed:    true,
			},
			"rules_count": schema.Int64Attribute{
				Description: "Number of rules in the list filter",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether this list filter is enabled",
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data
func (d *listFilterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// read Terraform configuration data into the model
	var state listFilterDataModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	// default to a blacklist filter type
	filterType := "blacklist"
	whitelist := false
	if state.Whitelist.ValueBool() {
		filterType = "whitelist"
		whitelist = true
	}

	// retrieve list filter info
	listFilter, err := d.adg.GetListFilterByName(state.Name.ValueString(), filterType)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home List Filter",
			err.Error(),
		)
		return
	}
	if listFilter == nil {
		resp.Diagnostics.AddError(
			"Unable to Locate AdGuard Home List Filter",
			"No list filter with name `"+state.Name.ValueString()+"` exists in AdGuard Home.",
		)
		return
	}

	// map response body to model
	state.Url = types.StringValue(listFilter.Url)
	state.LastUpdated = types.StringValue(listFilter.LastUpdated)
	state.Id = types.Int64Value(listFilter.Id)
	state.RulesCount = types.Int64Value(int64(listFilter.RulesCount))
	state.Enabled = types.BoolValue(listFilter.Enabled)
	state.Whitelist = types.BoolValue(whitelist)

	// set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source
func (d *listFilterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.adg = req.ProviderData.(*adguard.ADG)
}
