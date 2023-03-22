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
			"url": schema.StringAttribute{
				Description: "Url of the list filter",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the list filter",
				Required:    true,
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
				Optional:    true,
			},
		},
	}
}

