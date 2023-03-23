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
	_ datasource.DataSource              = &userRulesDataSource{}
	_ datasource.DataSourceWithConfigure = &userRulesDataSource{}
)

// userRulesDataSource is the data source implementation
type userRulesDataSource struct {
	adg *adguard.ADG
}

// userRulesDataModel maps filter schema data
type userRulesDataModel struct {
	ID    types.String `tfsdk:"id"`
	Rules types.List   `tfsdk:"rules"`
}

// NewUserRulesDataSource is a helper function to simplify the provider implementation
func NewUserRulesDataSource() datasource.DataSource {
	return &userRulesDataSource{}
}

// Metadata returns the data source type name
func (d *userRulesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_rules"
}

// Schema defines the schema for the data source
func (d *userRulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder identifier attribute",
				Computed:    true,
			},
			"rules": schema.ListAttribute{
				Description: "List of user rules",
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data
func (d *userRulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// read Terraform configuration data into the model
	var state userRulesDataModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	// retrieve list filter info
	userRules, err := d.adg.GetUserRules()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home List Filter",
			err.Error(),
		)
		return
	}

	// map response body to model
	state.Rules, diags = types.ListValueFrom(ctx, types.StringType, userRules)
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
func (d *userRulesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.adg = req.ProviderData.(*adguard.ADG)
}
