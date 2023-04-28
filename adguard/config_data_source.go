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
	_ datasource.DataSource              = &configDataSource{}
	_ datasource.DataSourceWithConfigure = &configDataSource{}
)

// configDataSource is the data source implementation
type configDataSource struct {
	adg *adguard.ADG
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
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform refresh",
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
			"parental_control": schema.SingleNestedAttribute{
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
					"local_ptr_upstreams": schema.SetAttribute{
						Description: "List of private reverse DNS servers",
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
	var state configCommonModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	// initialize object to store downstream API responses
	var apiResponse configApiResponseModel

	// get refreshed filtering config value from AdGuard Home
	filteringConfig, err := d.adg.GetAllFilters()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// add to response object
	apiResponse.Filtering = *filteringConfig

	// get refreshed safe browsing status from AdGuard Home
	safeBrowsingStatus, err := d.adg.GetSafeBrowsingStatus()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// add to response object
	apiResponse.SafeBrowsing = *safeBrowsingStatus

	// get refreshed safe parental control status from AdGuard Home
	parentalStatus, err := d.adg.GetParentalStatus()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// add to response object
	apiResponse.ParentalControl = *parentalStatus

	// retrieve safe search info
	safeSearchConfig, err := d.adg.GetSafeSearchConfig()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// add to response object
	apiResponse.SafeSearch = safeSearchConfig

	// retrieve query log config info
	queryLogConfig, err := d.adg.GetQueryLogConfig()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// add to response object
	apiResponse.QueryLog = *queryLogConfig

	// retrieve server statistics config info
	statsConfig, err := d.adg.GetStatsConfig()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// add to response object
	apiResponse.Stats = *statsConfig

	// get refreshed blocked services from AdGuard Home
	blockedServices, err := d.adg.GetBlockedServices()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// add to response object
	apiResponse.BlockedServices = *blockedServices

	// retrieve dns config info
	dnsConfig, err := d.adg.GetDnsInfo()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// add to response object
	apiResponse.DnsConfig = *dnsConfig.DNSConfig

	// process API responses into a state-like object
	newState, diags, err := ProcessConfigApiReadResponse(ctx, apiResponse)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	} else if diags.HasError() {
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

// Configure adds the provider configured config to the data source
func (d *configDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.adg = req.ProviderData.(*adguard.ADG)
}
