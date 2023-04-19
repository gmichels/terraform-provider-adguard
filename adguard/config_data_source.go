package adguard

import (
	"context"
	"reflect"
	"strings"

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
	SafeBrowsingEnabled     types.Bool   `tfsdk:"safebrowsing_enabled"`
	ParentalEnabled         types.Bool   `tfsdk:"parental_enabled"`
	SafeSearchEnabled       types.Bool   `tfsdk:"safesearch_enabled"`
	SafeSearchServices      types.Set    `tfsdk:"safesearch_services"`
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
			"safebrowsing_enabled": schema.BoolAttribute{
				Description: "Whether Safe Browsing is enabled",
				Computed:    true,
			},
			"parental_enabled": schema.BoolAttribute{
				Description: "Whether Parental Control is enabled",
				Computed:    true,
			},
			"safesearch_enabled": schema.BoolAttribute{
				Description: "Whether Safe Search is enabled",
				Computed:    true,
			},
			"safesearch_services": schema.SetAttribute{
				Description: "Services which SafeSearch is enabled",
				ElementType: types.StringType,
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

	// retrieve filter config info
	filterConfig, err := d.adg.GetAllFilters()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}

	// retrieve safe browsing info
	safeBrowsingEnabled, err := d.adg.GetSafeBrowsingStatus()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}

	// retrieve parental control info
	parentalEnabled, err := d.adg.GetParentalStatus()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}

	// retrieve safe search info
	safeSearchConfig, err := d.adg.GetSafeSearchConfig()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// initialize list for holding the enabled safe search services
	var enabledSafeSearchServices []string

	// logic to map enabled safe search services to a list
	// perform reflection of safeSearchConfig object
	reflected := reflect.ValueOf(safeSearchConfig).Elem()
	// grab the type of the reflected object
	reflectedType := reflected.Type()

	// loop over all safeSearchConfig fields
	for i := 0; i < reflected.NumField(); i++ {
		// skip the Enabled field
		if reflectedType.Field(i).Name != "Enabled" {
			// add service to list if its value is true
			if reflected.Field(i).Interface().(bool) {
				enabledSafeSearchServices = append(enabledSafeSearchServices, strings.ToLower(reflectedType.Field(i).Name))
			}
		}
	}

	// map response body to model
	state.FilteringEnabled = types.BoolValue(filterConfig.Enabled)
	state.FilteringUpdateInterval = types.Int64Value(int64(filterConfig.Interval))
	state.SafeBrowsingEnabled = types.BoolValue(*safeBrowsingEnabled)
	state.ParentalEnabled = types.BoolValue(*parentalEnabled)
	state.SafeSearchEnabled = types.BoolValue(safeSearchConfig.Enabled)
	state.SafeSearchServices, diags = types.SetValueFrom(ctx, types.StringType, enabledSafeSearchServices)
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

// Configure adds the provider configured config to the data source
func (d *configDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.adg = req.ProviderData.(*adguard.ADG)
}
