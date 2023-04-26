package adguard

import (
	"context"
	"reflect"

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
	ID              types.String `tfsdk:"id"`
	Filtering       types.Object `tfsdk:"filtering"`
	SafeBrowsing    types.Object `tfsdk:"safebrowsing"`
	ParentalControl types.Object `tfsdk:"parental"`
	SafeSearch      types.Object `tfsdk:"safesearch"`
	QueryLog        types.Object `tfsdk:"querylog"`
	Stats           types.Object `tfsdk:"stats"`
	BlockedServices types.Set    `tfsdk:"blocked_services"`
	Dns             types.Object `tfsdk:"dns"`
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
			"filtering": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Whether DNS filtering is enabled",
						Computed:    true,
					},
					"update_interval": schema.Int64Attribute{
						Description: "Update interval for all list-based filters, in hours",
						Computed:    true,
					},
				},
			},
			"safebrowsing": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Whether Safe Browsing is enabled",
						Computed:    true,
					},
				},
			},
			"parental": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Whether Parental Control is enabled",
						Computed:    true,
					},
				},
			},
			"safesearch": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Whether Safe Search is enabled",
						Computed:    true,
					},
					"services": schema.SetAttribute{
						Description: "Services which SafeSearch is enabled",
						ElementType: types.StringType,
						Computed:    true,
					},
				},
			},
			"querylog": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Whether the query log is enabled",
						Computed:    true,
					},
					"interval": schema.Int64Attribute{
						Description: "Time period for query log rotation, in hours",
						Computed:    true,
					},
					"anonymize_client_ip": schema.BoolAttribute{
						Description: "Whether anonymizing clients' IP addresses is enabled",
						Computed:    true,
					},
					"ignored": schema.SetAttribute{
						Description: "List of host names which should not be written to log",
						ElementType: types.StringType,
						Computed:    true,
					},
				},
			},
			"stats": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Whether server statistics enabled",
						Computed:    true,
					},
					"interval": schema.Int64Attribute{
						Description: "Time period for the server statistics rotation, in hours",
						Computed:    true,
					},
					"ignored": schema.SetAttribute{
						Description: "List of host names which should not be counted in the server statistics",
						ElementType: types.StringType,
						Computed:    true,
					},
				},
			},
			"blocked_services": schema.SetAttribute{
				Description: "List of services that are blocked globally",
				ElementType: types.StringType,
				Computed:    true,
			},
			"dns": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
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

	// get refreshed filtering config value from AdGuard Home
	filteringConfig, err := d.adg.GetAllFilters()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// map filter config to state
	var stateFilteringConfig filteringModel
	stateFilteringConfig.Enabled = types.BoolValue(filteringConfig.Enabled)
	stateFilteringConfig.UpdateInterval = types.Int64Value(int64(filteringConfig.Interval))

	// get refreshed safe browsing status from AdGuard Home
	safeBrowsingStatus, err := d.adg.GetSafeBrowsingStatus()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// map safe browsing config to state
	var stateSafeBrowsingStatus enabledModel
	stateSafeBrowsingStatus.Enabled = types.BoolValue(*safeBrowsingStatus)

	// get refreshed safe parental control status from AdGuard Home
	parentalStatus, err := d.adg.GetParentalStatus()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// map parental control config to state
	var stateParentalStatus enabledModel
	stateParentalStatus.Enabled = types.BoolValue(*parentalStatus)

	// retrieve safe search info
	safeSearchConfig, err := d.adg.GetSafeSearchConfig()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}

	// perform reflection of safeSearchConfig object
	v := reflect.ValueOf(safeSearchConfig).Elem()
	// grab the type of the reflected object
	t := v.Type()
	// map the reflected object to a list
	enabledSafeSearchServices := mapSafeSearchConfigServices(v, t)

	// map safe search to state
	var stateSafeSearchConfig safeSearchModel
	stateSafeSearchConfig.Enabled = types.BoolValue(safeSearchConfig.Enabled)
	stateSafeSearchConfig.Services, diags = types.SetValueFrom(ctx, types.StringType, enabledSafeSearchServices)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// retrieve query log config info
	queryLogConfig, err := d.adg.GetQueryLogConfig()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	var stateQueryLogConfig queryLogConfigModel
	stateQueryLogConfig.Enabled = types.BoolValue(queryLogConfig.Enabled)
	stateQueryLogConfig.Interval = types.Int64Value(int64(queryLogConfig.Interval / 1000 / 3600))
	stateQueryLogConfig.AnonymizeClientIp = types.BoolValue(queryLogConfig.AnonymizeClientIp)
	stateQueryLogConfig.Ignored, diags = types.SetValueFrom(ctx, types.StringType, queryLogConfig.Ignored)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// retrieve server statistics config info
	statsConfig, err := d.adg.GetStatsConfig()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	var stateStatsConfig statsConfigModel
	stateStatsConfig.Enabled = types.BoolValue(statsConfig.Enabled)
	stateStatsConfig.Interval = types.Int64Value(int64(statsConfig.Interval / 3600 / 1000))
	stateStatsConfig.Ignored, diags = types.SetValueFrom(ctx, types.StringType, statsConfig.Ignored)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get refreshed blocked services from AdGuard Home
	blockedServices, err := d.adg.GetBlockedServices()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}

	// retrieve dns config info
	dnsConfig, err := d.adg.GetDnsInfo()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	var stateDnsConfig dnsConfigModel
	stateDnsConfig.BootstrapDns, diags = types.ListValueFrom(ctx, types.StringType, dnsConfig.BootstrapDns)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	stateDnsConfig.UpstreamDns, diags = types.ListValueFrom(ctx, types.StringType, dnsConfig.UpstreamDns)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	stateDnsConfig.RateLimit = types.Int64Value(int64(dnsConfig.RateLimit))
	stateDnsConfig.BlockingMode = types.StringValue(dnsConfig.BlockingMode)
	stateDnsConfig.BlockingIpv4 = types.StringValue(dnsConfig.BlockingIpv4)
	stateDnsConfig.BlockingIpv6 = types.StringValue(dnsConfig.BlockingIpv6)
	stateDnsConfig.EDnsCsEnabled = types.BoolValue(dnsConfig.EDnsCsEnabled)
	stateDnsConfig.DisableIpv6 = types.BoolValue(dnsConfig.DisableIpv6)
	stateDnsConfig.DnsSecEnabled = types.BoolValue(dnsConfig.DnsSecEnabled)
	stateDnsConfig.CacheSize = types.Int64Value(int64(dnsConfig.CacheSize))
	stateDnsConfig.CacheTtlMin = types.Int64Value(int64(dnsConfig.CacheTtlMin))
	stateDnsConfig.CacheTtlMax = types.Int64Value(int64(dnsConfig.CacheTtlMax))
	stateDnsConfig.CacheOptimistic = types.BoolValue(dnsConfig.CacheOptimistic)
	stateDnsConfig.UpstreamMode = types.StringValue(dnsConfig.UpstreamMode)
	stateDnsConfig.UsePrivatePtrResolvers = types.BoolValue(dnsConfig.UsePrivatePtrResolvers)
	stateDnsConfig.ResolveClients = types.BoolValue(dnsConfig.ResolveClients)
	stateDnsConfig.LocalPtrUpstreams, diags = types.ListValueFrom(ctx, types.StringType, dnsConfig.LocalPtrUpstreams)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	stateDnsConfig.DefaultLocalPtrUpstreams, diags = types.ListValueFrom(ctx, types.StringType, dnsConfig.DefaultLocalPtrUpstreams)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// map response body to model
	state.Filtering, _ = types.ObjectValueFrom(ctx, filteringModel{}.attrTypes(), &stateFilteringConfig)
	state.SafeBrowsing, _ = types.ObjectValueFrom(ctx, enabledModel{}.attrTypes(), &stateSafeBrowsingStatus)
	state.ParentalControl, _ = types.ObjectValueFrom(ctx, enabledModel{}.attrTypes(), &stateParentalStatus)
	state.SafeSearch, _ = types.ObjectValueFrom(ctx, safeSearchModel{}.attrTypes(), &stateSafeSearchConfig)
	state.QueryLog, _ = types.ObjectValueFrom(ctx, queryLogConfigModel{}.attrTypes(), &stateQueryLogConfig)
	state.Stats, _ = types.ObjectValueFrom(ctx, statsConfigModel{}.attrTypes(), &stateStatsConfig)
	state.BlockedServices, diags = types.SetValueFrom(ctx, types.StringType, blockedServices)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Dns, _ = types.ObjectValueFrom(ctx, dnsConfigModel{}.attrTypes(), &stateDnsConfig)

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
