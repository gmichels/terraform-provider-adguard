package adguard

import (
	"context"
	"regexp"
	"time"

	"github.com/gmichels/adguard-client-go"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource                = &dnsConfigResource{}
	_ resource.ResourceWithConfigure   = &dnsConfigResource{}
	_ resource.ResourceWithImportState = &dnsConfigResource{}
)

// dnsConfigResource is the resource implementation
type dnsConfigResource struct {
	adg *adguard.ADG
}

// dnsConfigResourceModel maps DNS Config schema data
type dnsConfigResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	LastUpdated            types.String `tfsdk:"last_updated"`
	BootstrapDns           types.List   `tfsdk:"bootstrap_dns"`
	UpstreamDns            types.List   `tfsdk:"upstream_dns"`
	UpstreamDnsFile        types.String `tfsdk:"upstream_dns_file"`
	RateLimit              types.Int64  `tfsdk:"rate_limit"`
	BlockingMode           types.String `tfsdk:"blocking_mode"`
	BlockingIpv4           types.String `tfsdk:"blocking_ipv4"`
	BlockingIpv6           types.String `tfsdk:"blocking_ipv6"`
	EDnsCsEnabled          types.Bool   `tfsdk:"edns_cs_enabled"`
	DisableIpv6            types.Bool   `tfsdk:"disable_ipv6"`
	DnsSecEnabled          types.Bool   `tfsdk:"dnssec_enabled"`
	CacheSize              types.Int64  `tfsdk:"cache_size"`
	CacheTtlMin            types.Int64  `tfsdk:"cache_ttl_min"`
	CacheTtlMax            types.Int64  `tfsdk:"cache_ttl_max"`
	CacheOptimistic        types.Bool   `tfsdk:"cache_optimistic"`
	UpstreamMode           types.String `tfsdk:"upstream_mode"`
	UsePrivatePtrResolvers types.Bool   `tfsdk:"use_private_ptr_resolvers"`
	ResolveClients         types.Bool   `tfsdk:"resolve_clients"`
	LocalPtrUpstreams      types.List   `tfsdk:"local_ptr_upstreams"`
}

// NewDnsConfigResource is a helper function to simplify the provider implementation
func NewDnsConfigResource() resource.Resource {
	return &dnsConfigResource{}
}

// Metadata returns the resource type name
func (r *dnsConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_config"
}

// Schema defines the schema for the resource
func (r *dnsConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internal identifier for this dnsConfig",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the dnsConfig",
				Computed:    true,
			},
			"bootstrap_dns": schema.ListAttribute{
				Description: "Booststrap DNS servers",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
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
				Description: "Upstream DNS servers",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
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
			"upstream_dns_file": schema.StringAttribute{
				Description: "File with upstream DNS servers",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"rate_limit": schema.Int64Attribute{
				Description: "The number of requests per second allowed per client",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(20),
			},
			"blocking_mode": schema.StringAttribute{
				Description: "DNS response sent when request is blocked",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("default"),
				Validators: []validator.String{
					stringvalidator.All(
						stringvalidator.OneOf("default", "refused", "nxdomain", "null_ip", "custom_ip"),
					),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRoot("blocking_ipv4"),
						path.MatchRoot("blocking_ipv6"),
					}...),
				},
			},
			"blocking_ipv4": schema.StringAttribute{
				Description: "When `blocking_mode` is set to `custom_ip`, the IPv4 address to be returned for a blocked A request",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.All(
						stringvalidator.RegexMatches(
							regexp.MustCompile(`\b((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.|$)){4}\b`),
							"must be a valid IPv4 address",
						),
						checkBlockingMode("custom_ip"),
						stringvalidator.AlsoRequires(path.Expressions{
							path.MatchRoot("blocking_mode"),
							path.MatchRoot("blocking_ipv6"),
						}...),
					),
				},
			},
			"blocking_ipv6": schema.StringAttribute{
				Description: "When `blocking_mode` is set to `custom_ip`, the IPv6 address to be returned for a blocked A request",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.All(
						stringvalidator.RegexMatches(
							regexp.MustCompile(`(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))`),
							"must be a valid IPv6 address",
						),
						checkBlockingMode("custom_ip"),
						stringvalidator.AlsoRequires(path.Expressions{
							path.MatchRoot("blocking_mode"),
							path.MatchRoot("blocking_ipv4"),
						}...),
					),
				},
			},
			"edns_cs_enabled": schema.BoolAttribute{
				Description: "Whether EDNS Client Subnet (ECS) is enabled",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"disable_ipv6": schema.BoolAttribute{
				Description: "Whether dropping of all IPv6 DNS queries is enabled",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"dnssec_enabled": schema.BoolAttribute{
				Description: "Whether outgoing DNSSEC is enabled",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"cache_size": schema.Int64Attribute{
				Description: "DNS cache size (in bytes)",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(4194304),
			},
			"cache_ttl_min": schema.Int64Attribute{
				Description: "Overridden minimum TTL received from upstream DNS servers",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
			"cache_ttl_max": schema.Int64Attribute{
				Description: "Overridden maximum TTL received from upstream DNS servers",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
			"cache_optimistic": schema.BoolAttribute{
				Description: "Whether optimistic DNS caching is enabled",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"upstream_mode": schema.StringAttribute{
				Description: "Upstream DNS resolvers usage strategy",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("load_balance"),
				Validators: []validator.String{
					stringvalidator.OneOf("load_balance", "parallel", "fastest_addr"),
				},
			},
			"use_private_ptr_resolvers": schema.BoolAttribute{
				Description: "Whether to use private reverse DNS resolvers",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"resolve_clients": schema.BoolAttribute{
				Description: "Whether reverse DNS resolution of clients' IP addresses is enabled",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"local_ptr_upstreams": schema.ListAttribute{
				Description: "List of private reverse DNS servers",
				ElementType: types.StringType,
				Optional:    true,
				Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
			},
		},
	}
}

// Configure adds the provider configured DNS Config to the resource
func (r *dnsConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.adg = req.ProviderData.(*adguard.ADG)
}

// Create creates the resource and sets the initial Terraform state
func (r *dnsConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// retrieve values from plan
	var plan dnsConfigResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// instantiate empty DNS Config for storing plan data
	var dnsConfig adguard.DNSConfig

	// populate DNS Config from plan
	if len(plan.BootstrapDns.Elements()) > 0 {
		diags = plan.BootstrapDns.ElementsAs(ctx, &dnsConfig.BootstrapDns, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if len(plan.UpstreamDns.Elements()) > 0 {
		diags = plan.UpstreamDns.ElementsAs(ctx, &dnsConfig.UpstreamDns, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	dnsConfig.UpstreamDnsFile = plan.UpstreamDnsFile.ValueString()
	dnsConfig.RateLimit = uint(plan.RateLimit.ValueInt64())
	dnsConfig.BlockingMode = plan.BlockingMode.ValueString()
	dnsConfig.BlockingIpv4 = plan.BlockingIpv4.ValueString()
	dnsConfig.BlockingIpv6 = plan.BlockingIpv6.ValueString()
	dnsConfig.EDnsCsEnabled = plan.EDnsCsEnabled.ValueBool()
	dnsConfig.DisableIpv6 = plan.DisableIpv6.ValueBool()
	dnsConfig.DnsSecEnabled = plan.DnsSecEnabled.ValueBool()
	dnsConfig.CacheSize = uint(plan.CacheSize.ValueInt64())
	dnsConfig.CacheTtlMin = uint(plan.CacheTtlMin.ValueInt64())
	dnsConfig.CacheTtlMax = uint(plan.CacheTtlMax.ValueInt64())
	dnsConfig.CacheOptimistic = plan.CacheOptimistic.ValueBool()
	if plan.UpstreamMode.ValueString() == "load_balance" {
		dnsConfig.UpstreamMode = ""
	} else {
		dnsConfig.UpstreamMode = plan.UpstreamMode.ValueString()
	}
	dnsConfig.UsePrivatePtrResolvers = plan.UsePrivatePtrResolvers.ValueBool()
	dnsConfig.ResolveClients = plan.ResolveClients.ValueBool()
	if len(plan.LocalPtrUpstreams.Elements()) > 0 {
		diags = plan.LocalPtrUpstreams.ElementsAs(ctx, &dnsConfig.LocalPtrUpstreams, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// set DNS Config using plan
	_, err := r.adg.SetDnsConfig(dnsConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating DNS Config",
			"Could not create DNS Config, unexpected error: "+err.Error(),
		)
		return
	}

	// response sent by AdGuard Home is the same as the sent payload,
	// just add missing attributes for state
	// there can be only one entry DNS Config, so hardcode the ID as 1
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
func (r *dnsConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// get current state
	var state dnsConfigResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get refreshed DNS dnsConfig rule value from AdGuard Home
	dnsConfig, err := r.adg.GetDnsInfo()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading AdGuard Home DNS Config",
			"Could not read AdGuard Home DNS Config with ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// overwrite DNS dnsConfig rule with refreshed state
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
	state.UpstreamDnsFile = types.StringValue(dnsConfig.UpstreamDnsFile)
	state.RateLimit = types.Int64Value(int64(dnsConfig.RateLimit))
	state.BlockingMode = types.StringValue(dnsConfig.BlockingMode)
	// upstream API does not unset blocking_ipv4 and blocking_ipv6 when previously set
	// and blocking mode changes, so force state to empty values here
	if dnsConfig.BlockingMode != "custom_ip" {
		state.BlockingIpv4 = types.StringValue("")
		state.BlockingIpv6 = types.StringValue("")
	} else {
		state.BlockingIpv4 = types.StringValue(dnsConfig.BlockingIpv4)
		state.BlockingIpv6 = types.StringValue(dnsConfig.BlockingIpv6)
	}
	state.EDnsCsEnabled = types.BoolValue(dnsConfig.EDnsCsEnabled)
	state.DisableIpv6 = types.BoolValue(dnsConfig.DisableIpv6)
	state.DnsSecEnabled = types.BoolValue(dnsConfig.DnsSecEnabled)
	state.CacheSize = types.Int64Value(int64(dnsConfig.CacheSize))
	state.CacheTtlMin = types.Int64Value(int64(dnsConfig.CacheTtlMin))
	state.CacheTtlMax = types.Int64Value(int64(dnsConfig.CacheTtlMax))
	state.CacheOptimistic = types.BoolValue(dnsConfig.CacheOptimistic)
	if dnsConfig.UpstreamMode != "" {
		state.UpstreamMode = types.StringValue(dnsConfig.UpstreamMode)
	} else {
		state.UpstreamMode = types.StringValue("load_balance")
	}
	state.UsePrivatePtrResolvers = types.BoolValue(dnsConfig.UsePrivatePtrResolvers)
	state.ResolveClients = types.BoolValue(dnsConfig.ResolveClients)
	if len(dnsConfig.LocalPtrUpstreams) > 0 {
		state.LocalPtrUpstreams, diags = types.ListValueFrom(ctx, types.StringType, dnsConfig.LocalPtrUpstreams)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	// set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success
func (r *dnsConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// updating is exactly the same as creating and unfortunately I don't know
	// of a way to reuse the code, hence the duplication

	// retrieve values from plan
	var plan dnsConfigResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// instantiate empty DNS Config for storing plan data
	var dnsConfig adguard.DNSConfig

	// populate DNS Config from plan
	if len(plan.BootstrapDns.Elements()) > 0 {
		diags = plan.BootstrapDns.ElementsAs(ctx, &dnsConfig.BootstrapDns, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if len(plan.UpstreamDns.Elements()) > 0 {
		diags = plan.UpstreamDns.ElementsAs(ctx, &dnsConfig.UpstreamDns, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	dnsConfig.UpstreamDnsFile = plan.UpstreamDnsFile.ValueString()
	dnsConfig.RateLimit = uint(plan.RateLimit.ValueInt64())
	dnsConfig.BlockingMode = plan.BlockingMode.ValueString()
	dnsConfig.BlockingIpv4 = plan.BlockingIpv4.ValueString()
	dnsConfig.BlockingIpv6 = plan.BlockingIpv6.ValueString()
	dnsConfig.EDnsCsEnabled = plan.EDnsCsEnabled.ValueBool()
	dnsConfig.DisableIpv6 = plan.DisableIpv6.ValueBool()
	dnsConfig.DnsSecEnabled = plan.DnsSecEnabled.ValueBool()
	dnsConfig.CacheSize = uint(plan.CacheSize.ValueInt64())
	dnsConfig.CacheTtlMin = uint(plan.CacheTtlMin.ValueInt64())
	dnsConfig.CacheTtlMax = uint(plan.CacheTtlMax.ValueInt64())
	dnsConfig.CacheOptimistic = plan.CacheOptimistic.ValueBool()
	if plan.UpstreamMode.ValueString() == "load_balance" {
		dnsConfig.UpstreamMode = ""
	} else {
		dnsConfig.UpstreamMode = plan.UpstreamMode.ValueString()
	}
	dnsConfig.UsePrivatePtrResolvers = plan.UsePrivatePtrResolvers.ValueBool()
	dnsConfig.ResolveClients = plan.ResolveClients.ValueBool()
	if len(plan.LocalPtrUpstreams.Elements()) > 0 {
		diags = plan.LocalPtrUpstreams.ElementsAs(ctx, &dnsConfig.LocalPtrUpstreams, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		dnsConfig.LocalPtrUpstreams = []string{}
	}

	// set DNS Config using plan
	_, err := r.adg.SetDnsConfig(dnsConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating DNS Config",
			"Could not create DNS Config, unexpected error: "+err.Error(),
		)
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
func (r *dnsConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// there is no "real" delete for DNS Configs, so this means "restore defaults"

	// instantiate empty DNS Config for storing default values
	var dnsConfig adguard.DNSConfig

	// populate DNS Config with default values
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

	// set default values in DNS Config
	_, err := r.adg.SetDnsConfig(dnsConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting DNS Config",
			"Could not delete DNS Config, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *dnsConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
