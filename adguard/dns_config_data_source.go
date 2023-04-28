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
	_ datasource.DataSource              = &dnsConfigDataSource{}
	_ datasource.DataSourceWithConfigure = &dnsConfigDataSource{}
)

// dnsConfigDataSource is the data source implementation
type dnsConfigDataSource struct {
	adg *adguard.ADG
}

// dnsConfigDataModel maps DNS Config schema data
type dnsConfigDataModel struct {
	ID                       types.String `tfsdk:"id"`
	BootstrapDns             types.List   `tfsdk:"bootstrap_dns"`
	UpstreamDns              types.List   `tfsdk:"upstream_dns"`
	RateLimit                types.Int64  `tfsdk:"rate_limit"`
	BlockingMode             types.String `tfsdk:"blocking_mode"`
	BlockingIpv4             types.String `tfsdk:"blocking_ipv4"`
	BlockingIpv6             types.String `tfsdk:"blocking_ipv6"`
	EDnsCsEnabled            types.Bool   `tfsdk:"edns_cs_enabled"`
	DisableIpv6              types.Bool   `tfsdk:"disable_ipv6"`
	DnsSecEnabled            types.Bool   `tfsdk:"dnssec_enabled"`
	CacheSize                types.Int64  `tfsdk:"cache_size"`
	CacheTtlMin              types.Int64  `tfsdk:"cache_ttl_min"`
	CacheTtlMax              types.Int64  `tfsdk:"cache_ttl_max"`
	CacheOptimistic          types.Bool   `tfsdk:"cache_optimistic"`
	UpstreamMode             types.String `tfsdk:"upstream_mode"`
	UsePrivatePtrResolvers   types.Bool   `tfsdk:"use_private_ptr_resolvers"`
	ResolveClients           types.Bool   `tfsdk:"resolve_clients"`
	LocalPtrUpstreams        types.List   `tfsdk:"local_ptr_upstreams"`
	DefaultLocalPtrUpstreams types.List   `tfsdk:"default_local_ptr_upstreams"`
}

// NewDnsConfigDataSource is a helper function to simplify the provider implementation
func NewDnsConfigDataSource() datasource.DataSource {
	return &dnsConfigDataSource{}
}

// Metadata returns the data source type name
func (d *dnsConfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_config"
}

// Schema defines the schema for the data source
func (d *dnsConfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		DeprecationMessage: "The `adguard_dns_config` data source is deprecated and will be removed in a future release. Use the `dns` block in the `adguard_config` data source instead.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier attribute",
				Computed:    true,
			},
			"bootstrap_dns": schema.ListAttribute{
				Description: "Booststrap DNS servers",
				ElementType: types.StringType,
				Computed:    true,
			},
			"upstream_dns": schema.ListAttribute{
				Description: "Upstream DNS servers",
				ElementType: types.StringType,
				Computed:    true,
			},
			"rate_limit": schema.Int64Attribute{
				Description: "The number of requests per second allowed per client",
				Computed:    true,
			},
			"blocking_mode": schema.StringAttribute{
				Description: "DNS response sent when request is blocked",
				Computed:    true,
			},
			"blocking_ipv4": schema.StringAttribute{
				Description: "When `blocking_mode` is set to `custom_ip`, the IPv4 address to be returned for a blocked A request",
				Computed:    true,
			},
			"blocking_ipv6": schema.StringAttribute{
				Description: "When `blocking_mode` is set to `custom_ip`, the IPv6 address to be returned for a blocked A request",
				Computed:    true,
			},
			"edns_cs_enabled": schema.BoolAttribute{
				Description: "Whether EDNS Client Subnet (ECS) is enabled",
				Computed:    true,
			},
			"disable_ipv6": schema.BoolAttribute{
				Description: "Whether dropping of all IPv6 DNS queries is enabled",
				Computed:    true,
			},
			"dnssec_enabled": schema.BoolAttribute{
				Description: "Whether outgoing DNSSEC is enabled",
				Computed:    true,
			},
			"cache_size": schema.Int64Attribute{
				Description: "DNS cache size (in bytes)",
				Computed:    true,
			},
			"cache_ttl_min": schema.Int64Attribute{
				Description: "Overridden minimum TTL received from upstream DNS servers",
				Computed:    true,
			},
			"cache_ttl_max": schema.Int64Attribute{
				Description: "Overridden maximum TTL received from upstream DNS servers",
				Computed:    true,
			},
			"cache_optimistic": schema.BoolAttribute{
				Description: "Whether optimistic DNS caching is enabled",
				Computed:    true,
			},
			"upstream_mode": schema.StringAttribute{
				Description: "Upstream DNS resolvers usage strategy",
				Computed:    true,
			},
			"use_private_ptr_resolvers": schema.BoolAttribute{
				Description: "Whether to use private reverse DNS resolvers",
				Computed:    true,
			},
			"resolve_clients": schema.BoolAttribute{
				Description: "Whether reverse DNS resolution of clients' IP addresses is enabled",
				Computed:    true,
			},
			"local_ptr_upstreams": schema.ListAttribute{
				Description: "List of private reverse DNS servers",
				ElementType: types.StringType,
				Computed:    true,
			},
			"default_local_ptr_upstreams": schema.ListAttribute{
				Description: "List of discovered private reverse DNS servers",
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data
func (d *dnsConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// read Terraform configuration data into the model
	var state dnsConfigDataModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	// retrieve dnsConfig info
	dnsConfig, err := d.adg.GetDnsInfo()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home DNS Config",
			err.Error(),
		)
		return
	}

	// map response body to model
	state.BootstrapDns, diags = types.ListValueFrom(ctx, types.StringType, dnsConfig.BootstrapDns)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.UpstreamDns, diags = types.ListValueFrom(ctx, types.StringType, dnsConfig.UpstreamDns)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.RateLimit = types.Int64Value(int64(dnsConfig.RateLimit))
	state.BlockingMode = types.StringValue(dnsConfig.BlockingMode)
	state.BlockingIpv4 = types.StringValue(dnsConfig.BlockingIpv4)
	state.BlockingIpv6 = types.StringValue(dnsConfig.BlockingIpv6)
	state.EDnsCsEnabled = types.BoolValue(dnsConfig.EDnsCsEnabled)
	state.DisableIpv6 = types.BoolValue(dnsConfig.DisableIpv6)
	state.DnsSecEnabled = types.BoolValue(dnsConfig.DnsSecEnabled)
	state.CacheSize = types.Int64Value(int64(dnsConfig.CacheSize))
	state.CacheTtlMin = types.Int64Value(int64(dnsConfig.CacheTtlMin))
	state.CacheTtlMax = types.Int64Value(int64(dnsConfig.CacheTtlMax))
	state.CacheOptimistic = types.BoolValue(dnsConfig.CacheOptimistic)
	state.UpstreamMode = types.StringValue(dnsConfig.UpstreamMode)
	state.UsePrivatePtrResolvers = types.BoolValue(dnsConfig.UsePrivatePtrResolvers)
	state.ResolveClients = types.BoolValue(dnsConfig.ResolveClients)
	state.LocalPtrUpstreams, diags = types.ListValueFrom(ctx, types.StringType, dnsConfig.LocalPtrUpstreams)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.DefaultLocalPtrUpstreams, diags = types.ListValueFrom(ctx, types.StringType, dnsConfig.DefaultLocalPtrUpstreams)
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

// Configure adds the provider configured dnsConfig to the data source
func (d *dnsConfigDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.adg = req.ProviderData.(*adguard.ADG)
}
