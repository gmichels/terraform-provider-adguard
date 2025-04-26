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
			"safebrowsing": schema.BoolAttribute{
				Description: "Whether Safe Browsing is enabled",
				Computed:    true,
			},
			"parental_control": schema.BoolAttribute{
				Description: "Whether Parental Control is enabled",
				Computed:    true,
			},
			"safesearch": safeSearchDatasourceSchema(),
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
						Description: "Set of host names which should not be written to log",
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
						Description: "Set of host names which should not be counted in the server statistics",
						ElementType: types.StringType,
						Computed:    true,
					},
				},
			},
			"blocked_services": schema.SetAttribute{
				Description: "Set of services that are blocked globally",
				ElementType: types.StringType,
				Computed:    true,
			},
			"blocked_services_pause_schedule": scheduleDatasourceSchema(),
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
					"fallback_dns": schema.ListAttribute{
						Description: "Fallback DNS servers",
						ElementType: types.StringType,
						Computed:    true,
					},
					"protection_enabled": schema.BoolAttribute{
						Description: "Whether protection is enabled",
						Computed:    true,
					},
					"rate_limit": schema.Int64Attribute{
						Description: "The number of requests per second allowed per client",
						Computed:    true,
					},
					"rate_limit_subnet_len_ipv4": schema.Int64Attribute{
						Description: "Subnet prefix length for IPv4 addresses used for rate limiting",
						Computed:    true,
					},
					"rate_limit_subnet_len_ipv6": schema.Int64Attribute{
						Description: "Subnet prefix length for IPv6 addresses used for rate limiting",
						Computed:    true,
					},
					"rate_limit_whitelist": schema.ListAttribute{
						Description: "IP addresses excluded from rate limiting",
						ElementType: types.StringType,
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
					"blocked_response_ttl": schema.Int64Attribute{
						Description: "How many seconds the clients should cache a filtered response",
						Computed:    true,
					},
					"edns_cs_enabled": schema.BoolAttribute{
						Description: "Whether EDNS Client Subnet (ECS) is enabled",
						Computed:    true,
					},
					"edns_cs_use_custom": schema.BoolAttribute{
						Description: "Whether EDNS Client Subnet (ECS) is using a custom IP",
						Computed:    true,
					},
					"edns_cs_custom_ip": schema.StringAttribute{
						Description: "The custom IP being used for EDNS Client Subnet (ECS)",
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
						Description: "Set of private reverse DNS servers",
						ElementType: types.StringType,
						Computed:    true,
					},
					"upstream_timeout": schema.Int64Attribute{
						Description: "The number of seconds to wait for a response from the upstream server",
						Computed:    true,
					},
					"allowed_clients": schema.SetAttribute{
						Description: "The allowlist of clients: IP addresses, CIDRs, or ClientIDs",
						ElementType: types.StringType,
						Computed:    true,
					},
					"disallowed_clients": schema.SetAttribute{
						Description: "The blocklist of clients: IP addresses, CIDRs, or ClientIDs",
						ElementType: types.StringType,
						Computed:    true,
					},
					"blocked_hosts": schema.SetAttribute{
						Description: "Disallowed domains",
						ElementType: types.StringType,
						Computed:    true,
					},
				},
			},
			"dhcp": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Whether the DHCP server is enabled",
						Computed:    true,
					},
					"interface": schema.StringAttribute{
						Description: "The interface to use for the DHCP server",
						Computed:    true,
					},
					"ipv4_settings": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"gateway_ip": schema.StringAttribute{
								Description: "The gateway IP for the DHCP server scope",
								Computed:    true,
							},
							"subnet_mask": schema.StringAttribute{
								Description: "The subnet mask for the DHCP server scope",
								Computed:    true,
							},
							"range_start": schema.StringAttribute{
								Description: "The start range for the DHCP server scope",
								Computed:    true,
							},
							"range_end": schema.StringAttribute{
								Description: "The start range for the DHCP server scope",
								Computed:    true,
							},
							"lease_duration": schema.Int64Attribute{
								Description: "The lease duration for the DHCP server scope, in seconds",
								Computed:    true,
							},
						},
					},
					"ipv6_settings": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"range_start": schema.StringAttribute{
								Description: "The start range for the DHCP server scope",
								Computed:    true,
							},
							"lease_duration": schema.Int64Attribute{
								Description: "The lease duration for the DHCP server scope",
								Computed:    true,
							},
						},
					},
					"leases": schema.ListNestedAttribute{
						Description: "Current leases in the DHCP server",
						Computed:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"mac": schema.StringAttribute{
									Description: "MAC address associated with the lease",
									Computed:    true,
								},
								"ip": schema.StringAttribute{
									Description: "IP address associated with the lease",
									Computed:    true,
								},
								"hostname": schema.StringAttribute{
									Description: "Hostname associated with the lease",
									Computed:    true,
								},
								"expires": schema.StringAttribute{
									Description: "Expiration timestamp for the lease",
									Computed:    true,
								},
							},
						},
					},
					"static_leases": schema.SetNestedAttribute{
						Description: "Current static leases in the DHCP server",
						Computed:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"mac": schema.StringAttribute{
									Description: "MAC address associated with the static lease",
									Computed:    true,
								},
								"ip": schema.StringAttribute{
									Description: "IP address associated with the static lease",
									Computed:    true,
								},
								"hostname": schema.StringAttribute{
									Description: "Hostname associated with the static lease",
									Computed:    true,
								},
							},
						},
					},
				},
			},
			"tls": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Whether encryption (DoT/DoH/HTTPS) is enabled",
						Computed:    true,
					},
					"server_name": schema.StringAttribute{
						Description: "The hostname of the TLS/HTTPS server",
						Computed:    true,
					},
					"certificate_chain": schema.StringAttribute{
						Description: "The certificates chain, either the path to a file or a base64 encoded string of the certificates chain in PEM format",
						Computed:    true,
					},
					"private_key": schema.StringAttribute{
						Description: "The private key, either the path to a file or a base64 encoded string of the private key in PEM format",
						Computed:    true,
					},
					"force_https": schema.BoolAttribute{
						Description: "When `true`, forces HTTP-to-HTTPS redirect",
						Computed:    true,
					},
					"port_https": schema.Int64Attribute{
						Description: "The HTTPS port",
						Computed:    true,
					},
					"port_dns_over_tls": schema.Int64Attribute{
						Description: "The DNS-over-TLS (DoT) port",
						Computed:    true,
					},
					"port_dns_over_quic": schema.Int64Attribute{
						Description: "The DNS-over-Quic (DoQ) port",
						Computed:    true,
					},
					"private_key_saved": schema.BoolAttribute{
						Description: "Whether the user has previously saved a private key",
						Computed:    true,
					},
					"valid_cert": schema.BoolAttribute{
						Description: "Whether the specified certificates chain is a valid chain of X.509 certificates",
						Computed:    true,
					},
					"valid_chain": schema.BoolAttribute{
						Description: "Whether the specified certificates chain is verified and issued by a known CA",
						Computed:    true,
					},
					"valid_key": schema.BoolAttribute{
						Description: "Whether the private key is valid",
						Computed:    true,
					},
					"valid_pair": schema.BoolAttribute{
						Description: "Whether both certificate and private key are correct",
						Computed:    true,
					},
					"key_type": schema.StringAttribute{
						Description: "The private key type, either `RSA` or `ECDSA`",
						Computed:    true,
					},
					"subject": schema.StringAttribute{
						Description: "The subject of the first certificate in the chain",
						Computed:    true,
					},
					"issuer": schema.StringAttribute{
						Description: "The issuer of the first certificate in the chain",
						Computed:    true,
					},
					"not_before": schema.StringAttribute{
						Description: "The NotBefore field of the first certificate in the chain",
						Computed:    true,
					},
					"not_after": schema.StringAttribute{
						Description: "The NotAfter field of the first certificate in the chain",
						Computed:    true,
					},
					"dns_names": schema.ListAttribute{
						Description: "The value of SubjectAltNames field of the first certificate in the chain",
						ElementType: types.StringType,
						Computed:    true,
					},
					"warning_validation": schema.StringAttribute{
						Description: "The validation warning message with the issue description",
						Computed:    true,
					},
					"serve_plain_dns": schema.BoolAttribute{
						Description: "When `true`, plain DNS is allowed for incoming requests",
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

	// use common model for state
	var newState configCommonModel
	// use common Read function
	newState.Read(ctx, *d.adg, &state, &resp.Diagnostics, "datasource")
	if resp.Diagnostics.HasError() {
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
