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
	_ datasource.DataSource              = &dnsAccessDataSource{}
	_ datasource.DataSourceWithConfigure = &dnsAccessDataSource{}
)

// dnsAccessDataSource is the data source implementation
type dnsAccessDataSource struct {
	adg *adguard.ADG
}

// dnsAccessDataModel maps DNS Access List schema data
type dnsAccessDataModel struct {
	ID                types.String `tfsdk:"id"`
	AllowedClients    types.List   `tfsdk:"allowed_clients"`
	DisallowedClients types.List   `tfsdk:"disallowed_clients"`
	BlockedHosts      types.List   `tfsdk:"blocked_hosts"`
}

// NewDnsAccessDataSource is a helper function to simplify the provider implementation
func NewDnsAccessDataSource() datasource.DataSource {
	return &dnsAccessDataSource{}
}

// Metadata returns the data source type name
func (d *dnsAccessDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_access"
}

// Schema defines the schema for the data source
func (d *dnsAccessDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier attribute",
				Computed:    true,
			},
			"allowed_clients": schema.ListAttribute{
				Description: "The allowlist of clients: IP addresses, CIDRs, or ClientIDs",
				ElementType: types.StringType,
				Computed:    true,
			},
			"disallowed_clients": schema.ListAttribute{
				Description: "The blocklist of clients: IP addresses, CIDRs, or ClientIDs",
				ElementType: types.StringType,
				Computed:    true,
			},
			"blocked_hosts": schema.ListAttribute{
				Description: "Disallowed domains",
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data
func (d *dnsAccessDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// read Terraform configuration data into the model
	var state dnsAccessDataModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	// retrieve DNS Access List info
	dnsAccess, err := d.adg.GetAccess()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home DNS Access List",
			err.Error(),
		)
		return
	}

	// map response body to model
	state.AllowedClients, diags = types.ListValueFrom(ctx, types.StringType, dnsAccess.AllowedClients)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.DisallowedClients, diags = types.ListValueFrom(ctx, types.StringType, dnsAccess.DisallowedClients)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.BlockedHosts, diags = types.ListValueFrom(ctx, types.StringType, dnsAccess.BlockedHosts)
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

// Configure adds the provider configured DNS Access List to the data source
func (d *dnsAccessDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.adg = req.ProviderData.(*adguard.ADG)
}
