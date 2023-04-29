package adguard

import (
	"context"
	"regexp"
	"time"

	"github.com/gmichels/adguard-client-go"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource                = &configResource{}
	_ resource.ResourceWithConfigure   = &configResource{}
	_ resource.ResourceWithImportState = &configResource{}
)

// configResource is the resource implementation
type configResource struct {
	adg *adguard.ADG
}

// NewConfigResource is a helper function to simplify the provider implementation
func NewConfigResource() resource.Resource {
	return &configResource{}
}

// Metadata returns the resource type name
func (r *configResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config"
}

// Schema defines the schema for the resource
func (r *configResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internal identifier for this config",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the config",
				Computed:    true,
			},
			"filtering": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					filteringModel{}.attrTypes(), filteringModel{}.defaultObject()),
				),
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Whether DNS filtering is enabled. Defaults to `true`",
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(true),
					},
					"update_interval": schema.Int64Attribute{
						Description: "Update interval for all list-based filters, in hours. Defaults to `24`",
						Computed:    true,
						Optional:    true,
						Default:     int64default.StaticInt64(24),
						Validators: []validator.Int64{
							int64validator.OneOf([]int64{1, 12, 24, 72, 168}...),
						},
					},
				},
			},
			"safebrowsing": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					enabledModel{}.attrTypes(), enabledModel{}.defaultObject()),
				),
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Whether Safe Browsing is enabled. Defaults to `false`",
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(false),
					},
				},
			},
			"parental_control": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					enabledModel{}.attrTypes(), enabledModel{}.defaultObject()),
				),
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Whether Parental Control is enabled. Defaults to `false`",
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(false),
					},
				},
			},
			"safesearch": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					safeSearchModel{}.attrTypes(), safeSearchModel{}.defaultObject()),
				),
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Whether Safe Search is enabled. Defaults to `false`",
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(false),
					},
					"services": schema.SetAttribute{
						Description: "Services which SafeSearch is enabled. Defaults to Bing, DuckDuckGo, Google, Pixabay, Yandex and YouTube",
						ElementType: types.StringType,
						Computed:    true,
						Optional:    true,
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
							setvalidator.ValueStringsAre(
								stringvalidator.OneOf(DEFAULT_SAFESEARCH_SERVICES...),
							),
						},
						Default: setdefault.StaticValue(
							types.SetValueMust(
								types.StringType,
								[]attr.Value{
									types.StringValue("bing"),
									types.StringValue("duckduckgo"),
									types.StringValue("google"),
									types.StringValue("pixabay"),
									types.StringValue("yandex"),
									types.StringValue("youtube"),
								},
							),
						),
					},
				},
			},
			"querylog": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					queryLogConfigModel{}.attrTypes(), queryLogConfigModel{}.defaultObject()),
				),
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Whether the query log is enabled. Defaults to `true`",
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(true),
					},
					"interval": schema.Int64Attribute{
						Description: "Time period for query log rotation, in hours. Defaults to `2160` (90 days)",
						Computed:    true,
						Optional:    true,
						Default:     int64default.StaticInt64(24),
					},
					"anonymize_client_ip": schema.BoolAttribute{
						Description: "Whether anonymizing clients' IP addresses is enabled. Defaults to `false`",
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(false),
					},
					"ignored": schema.SetAttribute{
						Description: "List of host names which should not be written to log",
						ElementType: types.StringType,
						Computed:    true,
						Optional:    true,
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
							setvalidator.ValueStringsAre(
								stringvalidator.RegexMatches(
									regexp.MustCompile(`^[a-z0-9.-_]+$`),
									"must be a valid domain name",
								),
							),
						},
						Default: setdefault.StaticValue(
							types.SetNull(types.StringType),
						),
					},
				},
			},
			"stats": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					statsConfigModel{}.attrTypes(), statsConfigModel{}.defaultObject()),
				),
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Whether server statistics are enabled. Defaults to `true`",
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(true),
					},
					"interval": schema.Int64Attribute{
						Description: "Time period for server statistics rotation, in hours. Defaults to `24` (1 day)",
						Computed:    true,
						Optional:    true,
						Default:     int64default.StaticInt64(24),
					},
					"ignored": schema.SetAttribute{
						Description: "List of host names which should not be counted in the server statistics",
						ElementType: types.StringType,
						Computed:    true,
						Optional:    true,
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
							setvalidator.ValueStringsAre(
								stringvalidator.RegexMatches(
									regexp.MustCompile(`^[a-z0-9.-_]+$`),
									"must be a valid domain name",
								),
							),
						},
						Default: setdefault.StaticValue(
							types.SetNull(types.StringType),
						),
					},
				},
			},
			"blocked_services": schema.SetAttribute{
				Description: "List of services to be blocked globally",
				ElementType: types.StringType,
				Computed:    true,
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(BLOCKED_SERVICES_ALL...),
					),
				},
				Default: setdefault.StaticValue(
					types.SetNull(types.StringType),
				),
			},
			"dns": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					dnsConfigModel{}.attrTypes(), dnsConfigModel{}.defaultObject()),
				),
				Attributes: map[string]schema.Attribute{
					"bootstrap_dns": schema.ListAttribute{
						Description: "Booststrap DNS servers. Defaults to the ones supplied by the default AdGuard Home configuration",
						ElementType: types.StringType,
						Computed:    true,
						Optional:    true,
						Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
						Default: listdefault.StaticValue(
							types.ListValueMust(
								types.StringType,
								[]attr.Value{
									types.StringValue("9.9.9.10"),
									types.StringValue("149.112.112.10"),
									types.StringValue("2620:fe::10"),
									types.StringValue("2620:fe::fe:10"),
								},
							),
						),
					},
					"upstream_dns": schema.ListAttribute{
						Description: "Upstream DNS servers. Defaults to the ones supplied by the default AdGuard Home configuration",
						ElementType: types.StringType,
						Computed:    true,
						Optional:    true,
						Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
						Default: listdefault.StaticValue(
							types.ListValueMust(
								types.StringType,
								[]attr.Value{
									types.StringValue("https://dns10.quad9.net/dns-query"),
								},
							),
						),
					},
					"rate_limit": schema.Int64Attribute{
						Description: "The number of requests per second allowed per client. Defaults to `20`",
						Computed:    true,
						Optional:    true,
						Default:     int64default.StaticInt64(20),
					},
					"blocking_mode": schema.StringAttribute{
						Description: "DNS response sent when request is blocked. Valid values are `default` (the default), `refused`, `nxdomain`, `null_ip` or `custom_ip`",
						Computed:    true,
						Optional:    true,
						Default:     stringdefault.StaticString("default"),
						Validators: []validator.String{
							stringvalidator.OneOf("default", "refused", "nxdomain", "null_ip", "custom_ip"),
						},
					},
					"blocking_ipv4": schema.StringAttribute{
						Description: "When `blocking_mode` is set to `custom_ip`, the IPv4 address to be returned for a blocked A request",
						Computed:    true,
						Optional:    true,
						Default:     stringdefault.StaticString(""),
						Validators: []validator.String{
							stringvalidator.All(
								stringvalidator.RegexMatches(
									regexp.MustCompile(`\b((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.|$)){4}\b`),
									"must be a valid IPv4 address",
								),
								checkBlockingMode("custom_ip"),
								stringvalidator.AlsoRequires(path.Expressions{
									path.MatchRelative().AtParent().AtName("blocking_mode"),
									path.MatchRelative().AtParent().AtName("blocking_ipv6"),
								}...),
							),
						},
					},
					"blocking_ipv6": schema.StringAttribute{
						Description: "When `blocking_mode` is set to `custom_ip`, the IPv6 address to be returned for a blocked A request",
						Computed:    true,
						Optional:    true,
						Default:     stringdefault.StaticString(""),
						Validators: []validator.String{
							stringvalidator.All(
								stringvalidator.RegexMatches(
									regexp.MustCompile(`(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))`),
									"must be a valid IPv6 address",
								),
								checkBlockingMode("custom_ip"),
								stringvalidator.AlsoRequires(path.Expressions{
									path.MatchRelative().AtParent().AtName("blocking_mode"),
									path.MatchRelative().AtParent().AtName("blocking_ipv4"),
								}...),
							),
						},
					},
					"edns_cs_enabled": schema.BoolAttribute{
						Description: "Whether EDNS Client Subnet (ECS) is enabled. Defaults to `false`",
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(false),
					},
					"disable_ipv6": schema.BoolAttribute{
						Description: "Whether dropping of all IPv6 DNS queries is enabled. Defaults to `false`",
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(false),
					},
					"dnssec_enabled": schema.BoolAttribute{
						Description: "Whether outgoing DNSSEC is enabled. Defaults to `false`",
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(false),
					},
					"cache_size": schema.Int64Attribute{
						Description: "DNS cache size (in bytes). Defaults to `4194304`",
						Computed:    true,
						Optional:    true,
						Default:     int64default.StaticInt64(4194304),
					},
					"cache_ttl_min": schema.Int64Attribute{
						Description: "Overridden minimum TTL (in seconds) received from upstream DNS servers. Defaults to `0`",
						Computed:    true,
						Optional:    true,
						Default:     int64default.StaticInt64(0),
					},
					"cache_ttl_max": schema.Int64Attribute{
						Description: "Overridden maximum TTL (in seconds) received from upstream DNS servers. Defaults to `0`",
						Computed:    true,
						Optional:    true,
						Default:     int64default.StaticInt64(0),
					},
					"cache_optimistic": schema.BoolAttribute{
						Description: "Whether optimistic DNS caching is enabled. Defaults to `false`",
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(false),
					},
					"upstream_mode": schema.StringAttribute{
						Description: "Upstream DNS resolvers usage strategy. Valid values are `load_balance` (default), `parallel` and `fastest_addr`",
						Computed:    true,
						Optional:    true,
						Default:     stringdefault.StaticString("load_balance"),
						Validators: []validator.String{
							stringvalidator.OneOf("load_balance", "parallel", "fastest_addr"),
						},
					},
					"use_private_ptr_resolvers": schema.BoolAttribute{
						Description: "Whether to use private reverse DNS resolvers. Defaults to `true`",
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(true),
					},
					"resolve_clients": schema.BoolAttribute{
						Description: "Whether reverse DNS resolution of clients' IP addresses is enabled. Defaults to `true`",
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(true),
					},
					"local_ptr_upstreams": schema.SetAttribute{
						Description: "List of private reverse DNS servers",
						ElementType: types.StringType,
						Computed:    true,
						Optional:    true,
						Validators:  []validator.Set{setvalidator.SizeAtLeast(1)},
						Default: setdefault.StaticValue(
							types.SetValueMust(types.StringType, []attr.Value{}),
						),
					},
					"allowed_clients": schema.SetAttribute{
						Description: "The allowlist of clients: IP addresses, CIDRs, or ClientIDs",
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
						Default: setdefault.StaticValue(
							types.SetNull(types.StringType),
						),
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
							setvalidator.ValueStringsAre(
								stringvalidator.RegexMatches(
									regexp.MustCompile(`^[a-z0-9/.:-]+$`),
									"must be an IP address/CIDR or only contain numbers, lowercase letters, and hyphens",
								),
							),
						},
					},
					"disallowed_clients": schema.SetAttribute{
						Description: "The blocklist of clients: IP addresses, CIDRs, or ClientIDs",
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
						Default: setdefault.StaticValue(
							types.SetNull(types.StringType),
						),
						Validators: []validator.Set{
							setvalidator.All(
								setvalidator.SizeAtLeast(1),
								setvalidator.ConflictsWith(path.Expressions{
									path.MatchRelative().AtParent().AtName("allowed_clients"),
								}...),
							),
							setvalidator.ValueStringsAre(
								stringvalidator.RegexMatches(
									regexp.MustCompile(`^[a-z0-9/.:-]+$`),
									"must be an IP address/CIDR or only contain numbers, lowercase letters, and hyphens",
								),
							),
						},
					},
					"blocked_hosts": schema.SetAttribute{
						Description: "Disallowed domains. Defaults to the ones supplied by the default AdGuard Home configuration",
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
						Validators:  []validator.Set{setvalidator.SizeAtLeast(1)},
						Default: setdefault.StaticValue(
							types.SetValueMust(
								types.StringType,
								[]attr.Value{
									types.StringValue("version.bind"),
									types.StringValue("id.server"),
									types.StringValue("hostname.bind"),
								},
							),
						),
					},
				},
			},
		},
	}
}

// Configure adds the provider configured config to the resource
func (r *configResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.adg = req.ProviderData.(*adguard.ADG)
}

// Create creates the resource and sets the initial Terraform state
func (r *configResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// retrieve values from plan
	var plan configCommonModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// defer to common function to create or update the resource
	r.CreateOrUpdate(ctx, plan, &resp.Diagnostics)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// there can be only one entry Config, so hardcode the ID as 1
	plan.ID = types.StringValue("1")
	// add the last updated attribute
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data
func (r *configResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// get current state
	var state configCommonModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// use common model for state
	var newState configCommonModel
	// use common Read function
	newState.Read(ctx, *r.adg, &resp.Diagnostics)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// populate internal fields into new state
	newState.ID = state.ID
	newState.LastUpdated = state.LastUpdated

	// set refreshed state
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success
func (r *configResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// retrieve values from plan
	var plan configCommonModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// defer to common function to create or update the resource
	r.CreateOrUpdate(ctx, plan, &resp.Diagnostics)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// update resource state with updated items and timestamp
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// update state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success
func (r *configResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// there is no "real" delete for the configuration, so this means "restore defaults"

	// populate filtering config with default values
	var filterConfig adguard.FilterConfig
	filterConfig.Enabled = true
	filterConfig.Interval = 24

	// set filtering config to default
	_, err := r.adg.ConfigureFiltering(filterConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting AdGuard Home Config",
			"Could not delete config, unexpected error: "+err.Error(),
		)
		return
	}

	// set safebrowsing to default
	err = r.adg.SetSafeBrowsingStatus(false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting AdGuard Home Config",
			"Could not delete config, unexpected error: "+err.Error(),
		)
		return
	}

	// set parental to default
	err = r.adg.SetParentalStatus(false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting AdGuard Home Config",
			"Could not delete config, unexpected error: "+err.Error(),
		)
		return
	}

	// populate safe search with default values
	var safeSearchConfig adguard.SafeSearchConfig
	safeSearchConfig.Enabled = false
	safeSearchConfig.Bing = true
	safeSearchConfig.Duckduckgo = true
	safeSearchConfig.Google = true
	safeSearchConfig.Pixabay = true
	safeSearchConfig.Yandex = true
	safeSearchConfig.Youtube = true

	// set safe search to defaults
	_, err = r.adg.SetSafeSearchConfig(safeSearchConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting AdGuard Home Config",
			"Could not delete config, unexpected error: "+err.Error(),
		)
		return
	}

	// populate query log config with default values
	var queryLogConfig adguard.GetQueryLogConfigResponse
	queryLogConfig.Enabled = true
	queryLogConfig.Interval = 90 * 86400 * 1000
	queryLogConfig.AnonymizeClientIp = false
	queryLogConfig.Ignored = []string{}

	// set query log config to defaults
	_, err = r.adg.SetQueryLogConfig(queryLogConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting AdGuard Home Config",
			"Could not delete config, unexpected error: "+err.Error(),
		)
		return
	}

	// populate server statistics config with default values
	var statsConfig adguard.GetStatsConfigResponse
	statsConfig.Enabled = true
	statsConfig.Interval = 1 * 86400 * 1000
	statsConfig.Ignored = []string{}

	// set server statistics to defaults
	_, err = r.adg.SetStatsConfig(statsConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting AdGuard Home Config",
			"Could not delete config, unexpected error: "+err.Error(),
		)
		return
	}

	// set blocked services to defaults
	_, err = r.adg.SetBlockedServices(make([]string, 0))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting AdGuard Home Config",
			"Could not delete config, unexpected error: "+err.Error(),
		)
		return
	}

	// instantiate empty DNS config for storing default values
	var dnsConfig adguard.DNSConfig

	// populate DNS config with default values
	dnsConfig.BootstrapDns = []string{"9.9.9.10", "149.112.112.10", "2620:fe::10", "2620:fe::fe:10"}
	dnsConfig.UpstreamDns = []string{"https://dns10.quad9.net/dns-query"}
	dnsConfig.UpstreamDnsFile = ""
	dnsConfig.RateLimit = 20
	dnsConfig.BlockingMode = "default"
	dnsConfig.BlockingIpv4 = ""
	dnsConfig.BlockingIpv6 = ""
	dnsConfig.EDnsCsEnabled = false
	dnsConfig.DisableIpv6 = false
	dnsConfig.DnsSecEnabled = false
	dnsConfig.CacheSize = 4194304
	dnsConfig.CacheTtlMin = 0
	dnsConfig.CacheTtlMax = 0
	dnsConfig.CacheOptimistic = false
	dnsConfig.UpstreamMode = ""
	dnsConfig.UsePrivatePtrResolvers = true
	dnsConfig.ResolveClients = true
	dnsConfig.LocalPtrUpstreams = []string{}

	// set dns config to defaults
	_, err = r.adg.SetDnsConfig(dnsConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting AdGuard Home Config",
			"Could not delete config, unexpected error: "+err.Error(),
		)
		return
	}

	// instantiate empty dns access list for storing default values
	var dnsAccess adguard.AccessList

	// populate dns access list with default values
	dnsAccess.AllowedClients = []string{}
	dnsAccess.DisallowedClients = []string{}
	dnsAccess.BlockedHosts = []string{"version.bind", "id.server", "hostname.bind"}

	// set dns access list to defaults
	_, err = r.adg.SetAccess(dnsAccess)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting AdGuard Home Config",
			"Could not delete config, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *configResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
