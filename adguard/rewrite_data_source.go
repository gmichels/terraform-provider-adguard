package adguard

import (
	"context"
	"encoding/json"

	"github.com/gmichels/adguard-client-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ensure the implementation satisfies the expected interfaces
var (
	_ datasource.DataSource              = &rewriteDataSource{}
	_ datasource.DataSourceWithConfigure = &rewriteDataSource{}
)

// rewriteDataSource is the data source implementation
type rewriteDataSource struct {
	adg *adguard.ADG
}

// rewriteDataModel maps rewrite schema data
type rewriteDataModel struct {
	ID     types.String `tfsdk:"id"`
	Domain types.String `tfsdk:"domain"`
	Answer types.String `tfsdk:"answer"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

// NewRewriteDataSource is a helper function to simplify the provider implementation
func NewRewriteDataSource() datasource.DataSource {
	return &rewriteDataSource{}
}

// Metadata returns the data source type name
func (d *rewriteDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rewrite"
}

// Schema defines the schema for the data source
func (d *rewriteDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier attribute",
				Computed:    true,
			},
			"domain": schema.StringAttribute{
				Description: "Domain name",
				Required:    true,
			},
			"answer": schema.StringAttribute{
				Description: "Value of A, AAAA or CNAME DNS record",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the rewrite rule is enabled",
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data
func (d *rewriteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// read Terraform configuration data into the model
	var state rewriteDataModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	// retrieve rewrite info
	rewrite, err := GetRewrite(d.adg, state.Domain.ValueString(), state.Answer.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home Rewrite Rule",
			err.Error(),
		)
		return
	}
	// convert to JSON for response logging
	rewriteJson, err := json.Marshal(rewrite)
	if err != nil {
		diags.AddError(
			"Unable to Parse AdGuard Home Rewrite Rule",
			err.Error(),
		)
		return
	}
	// log response body
	tflog.Debug(ctx, "ADG API response", map[string]interface{}{
		"object": "rewrite",
		"body":   string(rewriteJson),
	})
	if rewrite == nil {
		resp.Diagnostics.AddError(
			"Unable to Locate AdGuard Home Rewrite Rule",
			"No rewrite rule with domain `"+state.Domain.ValueString()+" and answer `"+state.Answer.ValueString()+"` exists in AdGuard Home.",
		)
		return
	}

	// map response body to model
	state.Domain = types.StringValue(rewrite.Domain)
	state.Answer = types.StringValue(rewrite.Answer)
	state.Enabled = types.BoolValue(rewrite.Enabled)

	// set ID placeholder for testing
	state.ID = types.StringValue("placeholder")

	// set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured rewrite to the data source
func (d *rewriteDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.adg = req.ProviderData.(*adguard.ADG)
}
