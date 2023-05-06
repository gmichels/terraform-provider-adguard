package adguard

import (
	"context"
	"fmt"
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
						Description: fmt.Sprintf("Whether DNS filtering is enabled. Defaults to `%t`", CONFIG_FILTERING_ENABLED),
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(CONFIG_FILTERING_ENABLED),
					},
					"update_interval": schema.Int64Attribute{
						Description: fmt.Sprintf("Update interval for all list-based filters, in hours. Defaults to `%d`", CONFIG_FILTERING_UPDATE_INTERVAL),
						Computed:    true,
						Optional:    true,
						Default:     int64default.StaticInt64(int64(CONFIG_FILTERING_UPDATE_INTERVAL)),
						Validators: []validator.Int64{
							int64validator.OneOf([]int64{1, 12, 24, 72, 168}...),
						},
					},
				},
			},
			"safebrowsing": schema.BoolAttribute{
				Description: fmt.Sprintf("Whether Safe Browsing is enabled. Defaults to `%t`", CONFIG_SAFEBROWSING_ENABLED),
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(CONFIG_SAFEBROWSING_ENABLED),
			},
			"parental_control": schema.BoolAttribute{
				Description: fmt.Sprintf("Whether Parental Control is enabled. Defaults to `%t`", CONFIG_PARENTAL_CONTROL_ENABLED),
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(CONFIG_PARENTAL_CONTROL_ENABLED),
			},
			"safesearch": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					safeSearchModel{}.attrTypes(), safeSearchModel{}.defaultObject()),
				),
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: fmt.Sprintf("Whether Safe Search is enabled. Defaults to `%t`", CONFIG_SAFE_SEARCH_ENABLED),
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(CONFIG_SAFE_SEARCH_ENABLED),
					},
					"services": schema.SetAttribute{
						Description: "Services which SafeSearch is enabled.",
						ElementType: types.StringType,
						Computed:    true,
						Optional:    true,
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
							setvalidator.ValueStringsAre(
								stringvalidator.OneOf(CONFIG_SAFE_SEARCH_SERVICES_OPTIONS...),
							),
						},
						Default: setdefault.StaticValue(
							types.SetValueMust(types.StringType, convertToAttr(CONFIG_SAFE_SEARCH_SERVICES_OPTIONS)),
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
						Description: fmt.Sprintf("Whether the query log is enabled. Defaults to `%t`", CONFIG_QUERYLOG_ENABLED),
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(CONFIG_QUERYLOG_ENABLED),
					},
					"interval": schema.Int64Attribute{
						Description: fmt.Sprintf("Time period for query log rotation, in hours. Defaults to `%d` (%d days)", CONFIG_QUERYLOG_INTERVAL, CONFIG_QUERYLOG_INTERVAL/24),
						Computed:    true,
						Optional:    true,
						Default:     int64default.StaticInt64(int64(CONFIG_QUERYLOG_INTERVAL)),
					},
					"anonymize_client_ip": schema.BoolAttribute{
						Description: fmt.Sprintf("Whether anonymizing clients' IP addresses is enabled. Defaults to `%t`", CONFIG_QUERYLOG_ANONYMIZE_CLIENT_IP),
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(CONFIG_QUERYLOG_ANONYMIZE_CLIENT_IP),
					},
					"ignored": schema.SetAttribute{
						Description: "Set of host names which should not be written to log",
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
						Default: setdefault.StaticValue(types.SetNull(types.StringType)),
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
						Description: fmt.Sprintf("Whether server statistics are enabled. Defaults to `%t`", CONFIG_STATS_ENABLED),
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(CONFIG_STATS_ENABLED),
					},
					"interval": schema.Int64Attribute{
						Description: fmt.Sprintf("Time period for server statistics rotation, in hours. Defaults to `%d` (%d day)", CONFIG_STATS_INTERVAL, CONFIG_STATS_INTERVAL/24),
						Computed:    true,
						Optional:    true,
						Default:     int64default.StaticInt64(CONFIG_STATS_INTERVAL),
					},
					"ignored": schema.SetAttribute{
						Description: "Set of host names which should not be counted in the server statistics",
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
						Default: setdefault.StaticValue(types.SetNull(types.StringType)),
					},
				},
			},
			"blocked_services": schema.SetAttribute{
				Description: "Set of services to be blocked globally",
				ElementType: types.StringType,
				Computed:    true,
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(stringvalidator.OneOf(CONFIG_GLOBAL_BLOCKED_SERVICES_OPTIONS...)),
				},
				Default: setdefault.StaticValue(types.SetNull(types.StringType)),
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
							types.ListValueMust(types.StringType, convertToAttr(CONFIG_DNS_BOOTSTRAP)),
						),
					},
					"upstream_dns": schema.ListAttribute{
						Description: "Upstream DNS servers. Defaults to the ones supplied by the default AdGuard Home configuration",
						ElementType: types.StringType,
						Computed:    true,
						Optional:    true,
						Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
						Default: listdefault.StaticValue(
							types.ListValueMust(types.StringType, convertToAttr(CONFIG_DNS_UPSTREAM)),
						),
					},
					"rate_limit": schema.Int64Attribute{
						Description: fmt.Sprintf("The number of requests per second allowed per client. Defaults to `%d`", CONFIG_DNS_RATE_LIMIT),
						Computed:    true,
						Optional:    true,
						Default:     int64default.StaticInt64(CONFIG_DNS_RATE_LIMIT),
					},
					"blocking_mode": schema.StringAttribute{
						Description: "DNS response sent when request is blocked. Valid values are `default` (the default), `refused`, `nxdomain`, `null_ip` or `custom_ip`",
						Computed:    true,
						Optional:    true,
						Default:     stringdefault.StaticString(CONFIG_DNS_BLOCKING_MODE),
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
						Description: fmt.Sprintf("Whether EDNS Client Subnet (ECS) is enabled. Defaults to `%t`", CONFIG_DNS_EDNS_CS_ENABLED),
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(CONFIG_DNS_EDNS_CS_ENABLED),
					},
					"disable_ipv6": schema.BoolAttribute{
						Description: fmt.Sprintf("Whether dropping of all IPv6 DNS queries is enabled. Defaults to `%t`", CONFIG_DNS_DISABLE_IPV6),
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(CONFIG_DNS_DISABLE_IPV6),
					},
					"dnssec_enabled": schema.BoolAttribute{
						Description: fmt.Sprintf("Whether outgoing DNSSEC is enabled. Defaults to `%t`", CONFIG_DNS_DNSSEC_ENABLED),
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(CONFIG_DNS_DNSSEC_ENABLED),
					},
					"cache_size": schema.Int64Attribute{
						Description: fmt.Sprintf("DNS cache size (in bytes). Defaults to `%d`", CONFIG_DNS_CACHE_SIZE),
						Computed:    true,
						Optional:    true,
						Default:     int64default.StaticInt64(CONFIG_DNS_CACHE_SIZE),
					},
					"cache_ttl_min": schema.Int64Attribute{
						Description: fmt.Sprintf("Overridden minimum TTL (in seconds) received from upstream DNS servers. Defaults to `%d`", CONFIG_DNS_CACHE_TTL_MIN),
						Computed:    true,
						Optional:    true,
						Default:     int64default.StaticInt64(CONFIG_DNS_CACHE_TTL_MIN),
					},
					"cache_ttl_max": schema.Int64Attribute{
						Description: fmt.Sprintf("Overridden maximum TTL (in seconds) received from upstream DNS servers. Defaults to `%d`", CONFIG_DNS_CACHE_TTL_MAX),
						Computed:    true,
						Optional:    true,
						Default:     int64default.StaticInt64(CONFIG_DNS_CACHE_TTL_MAX),
					},
					"cache_optimistic": schema.BoolAttribute{
						Description: fmt.Sprintf("Whether optimistic DNS caching is enabled. Defaults to `%t`", CONFIG_DNS_CACHE_OPTIMISTIC),
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(CONFIG_DNS_CACHE_OPTIMISTIC),
					},
					"upstream_mode": schema.StringAttribute{
						Description: fmt.Sprintf("Upstream DNS resolvers usage strategy. Valid values are `%s` (default), `parallel` and `fastest_addr`", CONFIG_DNS_UPSTREAM_MODE),
						Computed:    true,
						Optional:    true,
						Default:     stringdefault.StaticString(CONFIG_DNS_UPSTREAM_MODE),
						Validators: []validator.String{
							stringvalidator.OneOf("load_balance", "parallel", "fastest_addr"),
						},
					},
					"use_private_ptr_resolvers": schema.BoolAttribute{
						Description: fmt.Sprintf("Whether to use private reverse DNS resolvers. Defaults to `%t`", CONFIG_DNS_USE_PRIVATE_PTR_RESOLVERS),
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(CONFIG_DNS_USE_PRIVATE_PTR_RESOLVERS),
					},
					"resolve_clients": schema.BoolAttribute{
						Description: fmt.Sprintf("Whether reverse DNS resolution of clients' IP addresses is enabled. Defaults to `%t`", CONFIG_DNS_RESOLVE_CLIENTS),
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(CONFIG_DNS_RESOLVE_CLIENTS),
					},
					"local_ptr_upstreams": schema.SetAttribute{
						Description: "Set of private reverse DNS servers",
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
			"dhcp": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					dhcpConfigModel{}.attrTypes(), dhcpConfigModel{}.defaultObject()),
				),
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: fmt.Sprintf("Whether the DHCP server is enabled. Defaults to `%t`", CONFIG_DHCP_ENABLED),
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(CONFIG_DHCP_ENABLED),
					},
					"interface": schema.StringAttribute{
						Description: "The interface to use for the DHCP server",
						Required:    true,
					},
					"ipv4_settings": schema.SingleNestedAttribute{
						Computed: true,
						Optional: true,
						Default: objectdefault.StaticValue(types.ObjectValueMust(
							dhcpIpv4Model{}.attrTypes(), dhcpIpv4Model{}.defaultObject()),
						),
						Attributes: map[string]schema.Attribute{
							"gateway_ip": schema.StringAttribute{
								Description: "The gateway IP for the DHCP server scope",
								Required:    true,
								Validators: []validator.String{
									stringvalidator.RegexMatches(
										regexp.MustCompile(`\b((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.|$)){4}\b`),
										"must be a valid IPv4 address",
									),
								},
							},
							"subnet_mask": schema.StringAttribute{
								Description: "The subnet mask for the DHCP server scope",
								Required:    true,
								Validators: []validator.String{
									stringvalidator.RegexMatches(
										regexp.MustCompile(`\b((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.|$)){4}\b`),
										"must be a valid IPv4 address",
									),
								},
							},
							"range_start": schema.StringAttribute{
								Description: "The start range for the DHCP server scope",
								Required:    true,
								Validators: []validator.String{
									stringvalidator.RegexMatches(
										regexp.MustCompile(`\b((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.|$)){4}\b`),
										"must be a valid IPv4 address",
									),
								},
							},
							"range_end": schema.StringAttribute{
								Description: "The start range for the DHCP server scope",
								Required:    true,
								Validators: []validator.String{
									stringvalidator.RegexMatches(
										regexp.MustCompile(`\b((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.|$)){4}\b`),
										"must be a valid IPv4 address",
									),
								},
							},
							"lease_duration": schema.Int64Attribute{
								Description: fmt.Sprintf("The lease duration for the DHCP server scope, in seconds. Defaults to `%d`", CONFIG_DHCP_LEASE_DURATION),
								Computed:    true,
								Optional:    true,
								Default:     int64default.StaticInt64(CONFIG_DHCP_LEASE_DURATION),
								Validators: []validator.Int64{
									int64validator.AtLeast(1),
								},
							},
						},
					},
					"ipv6_settings": schema.SingleNestedAttribute{
						Computed: true,
						Optional: true,
						Default: objectdefault.StaticValue(types.ObjectValueMust(
							dhcpIpv6Model{}.attrTypes(), dhcpIpv6Model{}.defaultObject()),
						),
						Attributes: map[string]schema.Attribute{
							"range_start": schema.StringAttribute{
								Description: "The start range for the DHCP server scope",
								Required:    true,
								Validators: []validator.String{
									stringvalidator.RegexMatches(
										regexp.MustCompile(`(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))`),
										"must be a valid IPv6 address",
									),
								},
							},
							"lease_duration": schema.Int64Attribute{
								Description: fmt.Sprintf("The lease duration for the DHCP server scope, in seconds. Defaults to `%d`", CONFIG_DHCP_LEASE_DURATION),
								Computed:    true,
								Optional:    true,
								Default:     int64default.StaticInt64(CONFIG_DHCP_LEASE_DURATION),
								Validators: []validator.Int64{
									int64validator.AtLeast(1),
								},
							},
						},
					},
					"static_leases": schema.SetNestedAttribute{
						Description: "Static leases for the DHCP server",
						Computed:    true,
						Optional:    true,
						Validators: []validator.Set{
							setvalidator.Any(
								setvalidator.AlsoRequires(path.Expressions{
									path.MatchRelative().AtParent().AtName("ipv4_settings"),
								}...),
								setvalidator.AlsoRequires(path.Expressions{
									path.MatchRelative().AtParent().AtName("ipv6_settings"),
								}...),
							),
						},
						Default: setdefault.StaticValue(types.SetNull(types.ObjectType{AttrTypes: dhcpStaticLeasesModel{}.attrTypes()})),
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"mac": schema.StringAttribute{
									Description: "MAC address associated with the static lease",
									Required:    true,
									Validators: []validator.String{
										stringvalidator.RegexMatches(
											regexp.MustCompile(`^[a-f0-9:]+$`),
											"must be a valid MAC address",
										),
									},
								},
								"ip": schema.StringAttribute{
									Description: "IP address associated with the static lease",
									Required:    true,
									Validators: []validator.String{
										stringvalidator.RegexMatches(
											regexp.MustCompile(`\b((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.|$)){4}\b`),
											"must be a valid IPv4 address",
										),
									},
								},
								"hostname": schema.StringAttribute{
									Description: "Hostname associated with the static lease",
									Required:    true,
									Validators: []validator.String{
										stringvalidator.RegexMatches(
											regexp.MustCompile(`^[a-z0-9-]+$`),
											"must be a valid hostname",
										),
									},
								},
							},
						},
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

	// empty state as it's a create operation
	var state configCommonModel

	// defer to common function to create or update the resource
	r.CreateOrUpdate(ctx, &plan, &state, &resp.Diagnostics)
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
	newState.Read(ctx, *r.adg, &resp.Diagnostics, "resource")
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

	// retrieve values from state
	var state configCommonModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// defer to common function to create or update the resource
	r.CreateOrUpdate(ctx, &plan, &state, &resp.Diagnostics)
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
	filterConfig.Enabled = CONFIG_FILTERING_ENABLED
	filterConfig.Interval = CONFIG_FILTERING_UPDATE_INTERVAL

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
	err = r.adg.SetSafeBrowsingStatus(CONFIG_SAFEBROWSING_ENABLED)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting AdGuard Home Config",
			"Could not delete config, unexpected error: "+err.Error(),
		)
		return
	}

	// set parental to default
	err = r.adg.SetParentalStatus(CONFIG_PARENTAL_CONTROL_ENABLED)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting AdGuard Home Config",
			"Could not delete config, unexpected error: "+err.Error(),
		)
		return
	}

	// populate safe search with default values
	var safeSearchConfig adguard.SafeSearchConfig
	safeSearchConfig.Enabled = CONFIG_SAFE_SEARCH_ENABLED
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
	queryLogConfig.Enabled = CONFIG_QUERYLOG_ENABLED
	queryLogConfig.Interval = CONFIG_QUERYLOG_INTERVAL * 3600 * 1000
	queryLogConfig.AnonymizeClientIp = CONFIG_QUERYLOG_ANONYMIZE_CLIENT_IP
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
	statsConfig.Enabled = CONFIG_STATS_ENABLED
	statsConfig.Interval = CONFIG_STATS_INTERVAL * 3600 * 1000
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
	dnsConfig.BootstrapDns = CONFIG_DNS_BOOTSTRAP
	dnsConfig.UpstreamDns = CONFIG_DNS_UPSTREAM
	dnsConfig.UpstreamDnsFile = ""
	dnsConfig.RateLimit = CONFIG_DNS_RATE_LIMIT
	dnsConfig.BlockingMode = CONFIG_DNS_BLOCKING_MODE
	dnsConfig.BlockingIpv4 = ""
	dnsConfig.BlockingIpv6 = ""
	dnsConfig.EDnsCsEnabled = CONFIG_DNS_EDNS_CS_ENABLED
	dnsConfig.DisableIpv6 = CONFIG_DNS_DISABLE_IPV6
	dnsConfig.DnsSecEnabled = CONFIG_DNS_DNSSEC_ENABLED
	dnsConfig.CacheSize = CONFIG_DNS_CACHE_SIZE
	dnsConfig.CacheTtlMin = CONFIG_DNS_CACHE_TTL_MIN
	dnsConfig.CacheTtlMax = CONFIG_DNS_CACHE_TTL_MAX
	dnsConfig.CacheOptimistic = CONFIG_DNS_CACHE_OPTIMISTIC
	dnsConfig.UpstreamMode = ""
	dnsConfig.UsePrivatePtrResolvers = CONFIG_DNS_USE_PRIVATE_PTR_RESOLVERS
	dnsConfig.ResolveClients = CONFIG_DNS_RESOLVE_CLIENTS
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

	// set dhcp config to defaults
	err = r.adg.ResetDhcpConfig()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting AdGuard Home Config",
			"Could not delete config, unexpected error: "+err.Error(),
		)
		return
	}

	// remove all dhcp static leases
	err = r.adg.ResetDhcpStaticLeases()
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
