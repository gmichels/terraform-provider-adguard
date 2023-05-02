package adguard

import (
	"context"
	"reflect"
	"strings"

	"github.com/gmichels/adguard-client-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// common config model to be used for working with both resource and data source
type configCommonModel struct {
	ID              types.String `tfsdk:"id"`
	LastUpdated     types.String `tfsdk:"last_updated"`
	Filtering       types.Object `tfsdk:"filtering"`
	SafeBrowsing    types.Object `tfsdk:"safebrowsing"`
	ParentalControl types.Object `tfsdk:"parental_control"`
	SafeSearch      types.Object `tfsdk:"safesearch"`
	QueryLog        types.Object `tfsdk:"querylog"`
	Stats           types.Object `tfsdk:"stats"`
	BlockedServices types.Set    `tfsdk:"blocked_services"`
	Dns             types.Object `tfsdk:"dns"`
}

// nested attributes objects

// filteringModel maps filtering schema data
type filteringModel struct {
	Enabled        types.Bool  `tfsdk:"enabled"`
	UpdateInterval types.Int64 `tfsdk:"update_interval"`
}

// attrTypes - return attribute types for this model
func (o filteringModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled":         types.BoolType,
		"update_interval": types.Int64Type,
	}
}

// defaultObject - return default object for this model
func (o filteringModel) defaultObject() map[string]attr.Value {
	return map[string]attr.Value{
		"enabled":         types.BoolValue(CONFIG_FILTERING_ENABLED),
		"update_interval": types.Int64Value(int64(CONFIG_FILTERING_UPDATE_INTERVAL)),
	}
}

// enabledModel maps both safe browsing and parental control schema data
type enabledModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

// attrTypes - return attribute types for this model
func (o enabledModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled": types.BoolType,
	}
}

// defaultObject - return default object for this model
func (o enabledModel) defaultObject() map[string]attr.Value {
	return map[string]attr.Value{
		"enabled": types.BoolValue(false),
	}
}

// safeSearchModel maps safe search schema data
type safeSearchModel struct {
	Enabled  types.Bool `tfsdk:"enabled"`
	Services types.Set  `tfsdk:"services"`
}

// attrTypes - return attribute types for this model
func (o safeSearchModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled":  types.BoolType,
		"services": types.SetType{ElemType: types.StringType},
	}
}

// defaultObject - return default object for this model
func (o safeSearchModel) defaultObject() map[string]attr.Value {
	services := []attr.Value{}
	for _, service := range CONFIG_SAFE_SEARCH_SERVICES_OPTIONS {
		services = append(services, types.StringValue(service))
	}

	return map[string]attr.Value{
		"enabled":  types.BoolValue(CONFIG_SAFE_SEARCH_ENABLED),
		"services": types.SetValueMust(types.StringType, services),
	}
}

// queryLogConfigModel maps query log configuration schema data
type queryLogConfigModel struct {
	Enabled           types.Bool  `tfsdk:"enabled"`
	Interval          types.Int64 `tfsdk:"interval"`
	AnonymizeClientIp types.Bool  `tfsdk:"anonymize_client_ip"`
	Ignored           types.Set   `tfsdk:"ignored"`
}

// attrTypes - return attribute types for this model
func (o queryLogConfigModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled":             types.BoolType,
		"interval":            types.Int64Type,
		"anonymize_client_ip": types.BoolType,
		"ignored":             types.SetType{ElemType: types.StringType},
	}
}

// defaultObject - return default object for this model
func (o queryLogConfigModel) defaultObject() map[string]attr.Value {
	return map[string]attr.Value{
		"enabled":             types.BoolValue(CONFIG_QUERYLOG_ENABLED),
		"interval":            types.Int64Value(int64(CONFIG_QUERYLOG_INTERVAL)),
		"anonymize_client_ip": types.BoolValue(CONFIG_QUERYLOG_ANONYMIZE_CLIENT_IP),
		"ignored":             types.SetValueMust(types.StringType, []attr.Value{}),
	}
}

// statsConfigModel maps stats configuration schema data
type statsConfigModel struct {
	Enabled  types.Bool  `tfsdk:"enabled"`
	Interval types.Int64 `tfsdk:"interval"`
	Ignored  types.Set   `tfsdk:"ignored"`
}

// attrTypes - return attribute types for this model
func (o statsConfigModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled":  types.BoolType,
		"interval": types.Int64Type,
		"ignored":  types.SetType{ElemType: types.StringType},
	}
}

// defaultObject - return default object for this model
func (o statsConfigModel) defaultObject() map[string]attr.Value {
	return map[string]attr.Value{
		"enabled":  types.BoolValue(CONFIG_STATS_ENABLED),
		"interval": types.Int64Value(CONFIG_STATS_INTERVAL),
		"ignored":  types.SetValueMust(types.StringType, []attr.Value{}),
	}
}

// dnsConfigModel maps DNS configuration schema data
type dnsConfigModel struct {
	BootstrapDns           types.List   `tfsdk:"bootstrap_dns"`
	UpstreamDns            types.List   `tfsdk:"upstream_dns"`
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
	LocalPtrUpstreams      types.Set    `tfsdk:"local_ptr_upstreams"`
	AllowedClients         types.Set    `tfsdk:"allowed_clients"`
	DisallowedClients      types.Set    `tfsdk:"disallowed_clients"`
	BlockedHosts           types.Set    `tfsdk:"blocked_hosts"`
}

// attrTypes - return attribute types for this model
func (o dnsConfigModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"bootstrap_dns":             types.ListType{ElemType: types.StringType},
		"upstream_dns":              types.ListType{ElemType: types.StringType},
		"rate_limit":                types.Int64Type,
		"blocking_mode":             types.StringType,
		"blocking_ipv4":             types.StringType,
		"blocking_ipv6":             types.StringType,
		"edns_cs_enabled":           types.BoolType,
		"disable_ipv6":              types.BoolType,
		"dnssec_enabled":            types.BoolType,
		"cache_size":                types.Int64Type,
		"cache_ttl_min":             types.Int64Type,
		"cache_ttl_max":             types.Int64Type,
		"cache_optimistic":          types.BoolType,
		"upstream_mode":             types.StringType,
		"use_private_ptr_resolvers": types.BoolType,
		"resolve_clients":           types.BoolType,
		"local_ptr_upstreams":       types.SetType{ElemType: types.StringType},
		"allowed_clients":           types.SetType{ElemType: types.StringType},
		"disallowed_clients":        types.SetType{ElemType: types.StringType},
		"blocked_hosts":             types.SetType{ElemType: types.StringType},
	}
}

// defaultObject - return default object for this model
func (o dnsConfigModel) defaultObject() map[string]attr.Value {
	bootstrap_dns := convertToAttr(CONFIG_DNS_BOOTSTRAP)
	upstream_dns := convertToAttr(CONFIG_DNS_UPSTREAM)

	return map[string]attr.Value{
		"bootstrap_dns":             types.ListValueMust(types.StringType, bootstrap_dns),
		"upstream_dns":              types.ListValueMust(types.StringType, upstream_dns),
		"rate_limit":                types.Int64Value(CONFIG_DNS_RATE_LIMIT),
		"blocking_mode":             types.StringValue(CONFIG_DNS_BLOCKING_MODE),
		"blocking_ipv4":             types.StringValue(""),
		"blocking_ipv6":             types.StringValue(""),
		"edns_cs_enabled":           types.BoolValue(CONFIG_DNS_EDNS_CS_ENABLED),
		"disable_ipv6":              types.BoolValue(CONFIG_DNS_DISABLE_IPV6),
		"dnssec_enabled":            types.BoolValue(CONFIG_DNS_DNSSEC_ENABLED),
		"cache_size":                types.Int64Value(CONFIG_DNS_CACHE_SIZE),
		"cache_ttl_min":             types.Int64Value(CONFIG_DNS_CACHE_TTL_MIN),
		"cache_ttl_max":             types.Int64Value(CONFIG_DNS_CACHE_TTL_MAX),
		"cache_optimistic":          types.BoolValue(CONFIG_DNS_CACHE_OPTIMISTIC),
		"upstream_mode":             types.StringValue(CONFIG_DNS_UPSTREAM_MODE),
		"use_private_ptr_resolvers": types.BoolValue(CONFIG_DNS_USE_PRIVATE_PTR_RESOLVERS),
		"resolve_clients":           types.BoolValue(CONFIG_DNS_RESOLVE_CLIENTS),
		"local_ptr_upstreams":       types.SetValueMust(types.StringType, []attr.Value{}),
		"allowed_clients":           types.SetValueMust(types.StringType, []attr.Value{}),
		"disallowed_clients":        types.SetValueMust(types.StringType, []attr.Value{}),
		"blocked_hosts":             types.SetValueMust(types.StringType, []attr.Value{}),
	}
}

// common `Read` function for both data source and resource
func (o *configCommonModel) Read(ctx context.Context, adg adguard.ADG, diags *diag.Diagnostics) {
	// FILTERING CONFIG
	// get refreshed filtering config value from AdGuard Home
	filteringConfig, err := adg.GetAllFilters()
	if err != nil {
		diags.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// map filter config to state
	var stateFilteringConfig filteringModel
	stateFilteringConfig.Enabled = types.BoolValue(filteringConfig.Enabled)
	stateFilteringConfig.UpdateInterval = types.Int64Value(int64(filteringConfig.Interval))
	// add to config model
	o.Filtering, _ = types.ObjectValueFrom(ctx, filteringModel{}.attrTypes(), &stateFilteringConfig)

	// SAFE BROWSING
	// get refreshed safe browsing status from AdGuard Home
	safeBrowsingStatus, err := adg.GetSafeBrowsingStatus()
	if err != nil {
		diags.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// map safe browsing config to state
	var stateSafeBrowsingStatus enabledModel
	stateSafeBrowsingStatus.Enabled = types.BoolValue(*safeBrowsingStatus)
	// add to config model
	o.SafeBrowsing, _ = types.ObjectValueFrom(ctx, enabledModel{}.attrTypes(), &stateSafeBrowsingStatus)

	// PARENTAL CONTROL
	// get refreshed safe parental control status from AdGuard Home
	parentalStatus, err := adg.GetParentalStatus()
	if err != nil {
		diags.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// map parental control config to state
	var stateParentalStatus enabledModel
	stateParentalStatus.Enabled = types.BoolValue(*parentalStatus)
	// add to config model
	o.ParentalControl, _ = types.ObjectValueFrom(ctx, enabledModel{}.attrTypes(), &stateParentalStatus)

	// SAFE SEARCH
	// retrieve safe search info
	safeSearchConfig, err := adg.GetSafeSearchConfig()
	if err != nil {
		diags.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// perform reflection of safe search object
	v := reflect.ValueOf(safeSearchConfig).Elem()
	// grab the type of the reflected object
	t := v.Type()
	// map the reflected object to a list
	enabledSafeSearchServices := mapSafeSearchConfigServices(v, t)
	// map safe search to state
	var stateSafeSearchConfig safeSearchModel
	stateSafeSearchConfig.Enabled = types.BoolValue(safeSearchConfig.Enabled)
	stateSafeSearchConfig.Services, *diags = types.SetValueFrom(ctx, types.StringType, enabledSafeSearchServices)
	if diags.HasError() {
		return
	}
	// add to config model
	o.SafeSearch, _ = types.ObjectValueFrom(ctx, safeSearchModel{}.attrTypes(), &stateSafeSearchConfig)

	// QUERY LOG
	// retrieve query log config info
	queryLogConfig, err := adg.GetQueryLogConfig()
	if err != nil {
		diags.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	var stateQueryLogConfig queryLogConfigModel
	stateQueryLogConfig.Enabled = types.BoolValue(queryLogConfig.Enabled)
	stateQueryLogConfig.Interval = types.Int64Value(int64(queryLogConfig.Interval / 1000 / 3600))
	stateQueryLogConfig.AnonymizeClientIp = types.BoolValue(queryLogConfig.AnonymizeClientIp)
	stateQueryLogConfig.Ignored, *diags = types.SetValueFrom(ctx, types.StringType, queryLogConfig.Ignored)
	if diags.HasError() {
		return
	}
	// add to config model
	o.QueryLog, _ = types.ObjectValueFrom(ctx, queryLogConfigModel{}.attrTypes(), &stateQueryLogConfig)

	// STATS
	// retrieve server statistics config info
	statsConfig, err := adg.GetStatsConfig()
	if err != nil {
		diags.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	var stateStatsConfig statsConfigModel
	stateStatsConfig.Enabled = types.BoolValue(statsConfig.Enabled)
	stateStatsConfig.Interval = types.Int64Value(int64(statsConfig.Interval / 3600 / 1000))
	stateStatsConfig.Ignored, *diags = types.SetValueFrom(ctx, types.StringType, statsConfig.Ignored)
	if diags.HasError() {
		return
	}
	// add to config model
	o.Stats, _ = types.ObjectValueFrom(ctx, statsConfigModel{}.attrTypes(), &stateStatsConfig)

	// BLOCKED SERVICES
	// get refreshed blocked services from AdGuard Home
	blockedServices, err := adg.GetBlockedServices()
	if err != nil {
		diags.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// add to config model
	o.BlockedServices, *diags = types.SetValueFrom(ctx, types.StringType, blockedServices)
	if diags.HasError() {
		return
	}

	// DNS CONFIG
	dnsConfig, err := adg.GetDnsInfo()
	if err != nil {
		diags.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// retrieve dns config info
	var stateDnsConfig dnsConfigModel
	stateDnsConfig.BootstrapDns, *diags = types.ListValueFrom(ctx, types.StringType, dnsConfig.BootstrapDns)
	if diags.HasError() {
		return
	}
	stateDnsConfig.UpstreamDns, *diags = types.ListValueFrom(ctx, types.StringType, dnsConfig.UpstreamDns)
	if diags.HasError() {
		return
	}
	stateDnsConfig.RateLimit = types.Int64Value(int64(dnsConfig.RateLimit))
	stateDnsConfig.BlockingMode = types.StringValue(dnsConfig.BlockingMode)
	// upstream API does not unset blocking_ipv4 and blocking_ipv6 when previously set and blocking mode changes,
	// so force state to empty values here
	if dnsConfig.BlockingMode != "custom_ip" {
		stateDnsConfig.BlockingIpv4 = types.StringValue("")
		stateDnsConfig.BlockingIpv6 = types.StringValue("")
	} else {
		stateDnsConfig.BlockingIpv4 = types.StringValue(dnsConfig.BlockingIpv4)
		stateDnsConfig.BlockingIpv6 = types.StringValue(dnsConfig.BlockingIpv6)
	}
	stateDnsConfig.EDnsCsEnabled = types.BoolValue(dnsConfig.EDnsCsEnabled)
	stateDnsConfig.DisableIpv6 = types.BoolValue(dnsConfig.DisableIpv6)
	stateDnsConfig.DnsSecEnabled = types.BoolValue(dnsConfig.DnsSecEnabled)
	stateDnsConfig.CacheSize = types.Int64Value(int64(dnsConfig.CacheSize))
	stateDnsConfig.CacheTtlMin = types.Int64Value(int64(dnsConfig.CacheTtlMin))
	stateDnsConfig.CacheTtlMax = types.Int64Value(int64(dnsConfig.CacheTtlMax))
	stateDnsConfig.CacheOptimistic = types.BoolValue(dnsConfig.CacheOptimistic)
	if dnsConfig.UpstreamMode != "" {
		stateDnsConfig.UpstreamMode = types.StringValue(dnsConfig.UpstreamMode)
	} else {
		stateDnsConfig.UpstreamMode = types.StringValue("load_balance")
	}
	stateDnsConfig.UsePrivatePtrResolvers = types.BoolValue(dnsConfig.UsePrivatePtrResolvers)
	stateDnsConfig.ResolveClients = types.BoolValue(dnsConfig.ResolveClients)
	stateDnsConfig.LocalPtrUpstreams, *diags = types.SetValueFrom(ctx, types.StringType, dnsConfig.LocalPtrUpstreams)
	if diags.HasError() {
		return
	}

	// DNS ACCESS
	// retrieve dns access info
	dnsAccess, err := adg.GetAccess()
	if err != nil {
		diags.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	stateDnsConfig.AllowedClients, *diags = types.SetValueFrom(ctx, types.StringType, dnsAccess.AllowedClients)
	if diags.HasError() {
		return
	}
	stateDnsConfig.DisallowedClients, *diags = types.SetValueFrom(ctx, types.StringType, dnsAccess.DisallowedClients)
	if diags.HasError() {
		return
	}
	stateDnsConfig.BlockedHosts, *diags = types.SetValueFrom(ctx, types.StringType, dnsAccess.BlockedHosts)
	if diags.HasError() {
		return
	}
	// add to config model
	o.Dns, _ = types.ObjectValueFrom(ctx, dnsConfigModel{}.attrTypes(), &stateDnsConfig)

	// if we got here, all went fine
}

// common `Create` and `Update` function for the resource
func (r *configResource) CreateOrUpdate(ctx context.Context, plan configCommonModel, diags *diag.Diagnostics) {
	// FILTERING CONFIG
	// unpack nested attributes from plan
	var planFiltering filteringModel
	*diags = plan.Filtering.As(ctx, &planFiltering, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return
	}
	// instantiate empty object for storing plan data
	var filteringConfig adguard.FilterConfig
	// populate filtering config from plan
	filteringConfig.Enabled = planFiltering.Enabled.ValueBool()
	filteringConfig.Interval = uint(planFiltering.UpdateInterval.ValueInt64())

	// set filtering config using plan
	_, err := r.adg.ConfigureFiltering(filteringConfig)
	if err != nil {
		diags.AddError(
			"Unable to Update AdGuard Home Config",
			err.Error(),
		)
		return
	}

	// SAFE BROWSING
	// unpack nested attributes from plan
	var planSafeBrowsing enabledModel
	*diags = plan.SafeBrowsing.As(ctx, &planSafeBrowsing, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return
	}
	// set safe browsing status using plan
	err = r.adg.SetSafeBrowsingStatus(planSafeBrowsing.Enabled.ValueBool())
	if err != nil {
		diags.AddError(
			"Unable to Update AdGuard Home Config",
			err.Error(),
		)
		return
	}

	// PARENTAL CONTROL
	// unpack nested attributes from plan
	var planParentalControl enabledModel
	*diags = plan.ParentalControl.As(ctx, &planParentalControl, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return
	}
	// set parental control status using plan
	if err != nil {
		diags.AddError(
			"Unable to Update AdGuard Home Config",
			err.Error(),
		)
		return
	}

	// SAFE SEARCH
	// unpack nested attributes from plan
	var planSafeSearch safeSearchModel
	*diags = plan.SafeSearch.As(ctx, &planSafeSearch, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return
	}
	// instantiate empty object for storing plan data
	var safeSearchConfig adguard.SafeSearchConfig
	// populate safe search config using plan
	safeSearchConfig.Enabled = planSafeSearch.Enabled.ValueBool()
	if len(planSafeSearch.Services.Elements()) > 0 {
		var safeSearchServicesEnabled []string
		*diags = planSafeSearch.Services.ElementsAs(ctx, &safeSearchServicesEnabled, false)
		if diags.HasError() {
			return
		}
		// use reflection to set each safeSearchConfig service value dynamically
		v := reflect.ValueOf(&safeSearchConfig).Elem()
		t := v.Type()
		setSafeSearchConfigServices(v, t, safeSearchServicesEnabled)
	}
	// set safe search config using plan
	_, err = r.adg.SetSafeSearchConfig(safeSearchConfig)
	if err != nil {
		diags.AddError(
			"Unable to Update AdGuard Home Config",
			err.Error(),
		)
		return
	}

	// QUERY LOG
	// unpack nested attributes from plan
	var planQueryLogConfig queryLogConfigModel
	*diags = plan.QueryLog.As(ctx, &planQueryLogConfig, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return
	}
	// instantiate empty object for storing plan data
	var queryLogConfig adguard.GetQueryLogConfigResponse
	// populate query log config from plan
	queryLogConfig.Enabled = planQueryLogConfig.Enabled.ValueBool()
	queryLogConfig.Interval = uint64(planQueryLogConfig.Interval.ValueInt64() * 3600 * 1000)
	queryLogConfig.AnonymizeClientIp = planQueryLogConfig.AnonymizeClientIp.ValueBool()
	if len(planQueryLogConfig.Ignored.Elements()) > 0 {
		*diags = planQueryLogConfig.Ignored.ElementsAs(ctx, &queryLogConfig.Ignored, false)
		if diags.HasError() {
			return
		}
	}
	// set query log config using plan
	_, err = r.adg.SetQueryLogConfig(queryLogConfig)
	if err != nil {
		diags.AddError(
			"Unable to Update AdGuard Home Config",
			err.Error(),
		)
		return
	}

	// STATS
	// unpack nested attributes from plan
	var planStatsConfig statsConfigModel
	*diags = plan.Stats.As(ctx, &planStatsConfig, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return
	}
	// instantiate empty object for storing plan data
	var statsConfig adguard.GetStatsConfigResponse
	// populate stats from plan
	statsConfig.Enabled = planStatsConfig.Enabled.ValueBool()
	statsConfig.Interval = uint64(planStatsConfig.Interval.ValueInt64() * 3600 * 1000)
	if len(planStatsConfig.Ignored.Elements()) > 0 {
		*diags = planStatsConfig.Ignored.ElementsAs(ctx, &statsConfig.Ignored, false)
		if diags.HasError() {
			return
		}
	}
	// set stats config using plan
	_, err = r.adg.SetStatsConfig(statsConfig)
	if err != nil {
		diags.AddError(
			"Unable to Update AdGuard Home Config",
			err.Error(),
		)
		return
	}

	// BLOCKED SERVICES
	// instantiate empty object for storing plan data
	var blockedServices []string
	// populate blocked services from plan
	if len(plan.BlockedServices.Elements()) > 0 {
		*diags = plan.BlockedServices.ElementsAs(ctx, &blockedServices, false)
		if diags.HasError() {
			return
		}
	}
	// set blocked services using plan
	_, err = r.adg.SetBlockedServices(blockedServices)
	if err != nil {
		diags.AddError(
			"Unable to Update AdGuard Home Config",
			err.Error(),
		)
		return
	}

	// DNS CONFIG
	// unpack nested attributes from plan
	var planDnsConfig dnsConfigModel
	*diags = plan.Dns.As(ctx, &planDnsConfig, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return
	}
	// instantiate empty object for storing plan data
	var dnsConfig adguard.DNSConfig
	// populate DNS config from plan
	if len(planDnsConfig.BootstrapDns.Elements()) > 0 {
		*diags = planDnsConfig.BootstrapDns.ElementsAs(ctx, &dnsConfig.BootstrapDns, false)
		if diags.HasError() {
			return
		}
	}
	if len(planDnsConfig.UpstreamDns.Elements()) > 0 {
		*diags = planDnsConfig.UpstreamDns.ElementsAs(ctx, &dnsConfig.UpstreamDns, false)
		if diags.HasError() {
			return
		}
	}
	dnsConfig.RateLimit = uint(planDnsConfig.RateLimit.ValueInt64())
	dnsConfig.BlockingMode = planDnsConfig.BlockingMode.ValueString()
	dnsConfig.BlockingIpv4 = planDnsConfig.BlockingIpv4.ValueString()
	dnsConfig.BlockingIpv6 = planDnsConfig.BlockingIpv6.ValueString()
	dnsConfig.EDnsCsEnabled = planDnsConfig.EDnsCsEnabled.ValueBool()
	dnsConfig.DisableIpv6 = planDnsConfig.DisableIpv6.ValueBool()
	dnsConfig.DnsSecEnabled = planDnsConfig.DnsSecEnabled.ValueBool()
	dnsConfig.CacheSize = uint(planDnsConfig.CacheSize.ValueInt64())
	dnsConfig.CacheTtlMin = uint(planDnsConfig.CacheTtlMin.ValueInt64())
	dnsConfig.CacheTtlMax = uint(planDnsConfig.CacheTtlMax.ValueInt64())
	dnsConfig.CacheOptimistic = planDnsConfig.CacheOptimistic.ValueBool()
	if planDnsConfig.UpstreamMode.ValueString() == "load_balance" {
		dnsConfig.UpstreamMode = ""
	} else {
		dnsConfig.UpstreamMode = planDnsConfig.UpstreamMode.ValueString()
	}
	dnsConfig.UsePrivatePtrResolvers = planDnsConfig.UsePrivatePtrResolvers.ValueBool()
	dnsConfig.ResolveClients = planDnsConfig.ResolveClients.ValueBool()
	if len(planDnsConfig.LocalPtrUpstreams.Elements()) > 0 {
		*diags = planDnsConfig.LocalPtrUpstreams.ElementsAs(ctx, &dnsConfig.LocalPtrUpstreams, false)
		if diags.HasError() {
			return
		}
	} else {
		dnsConfig.LocalPtrUpstreams = []string{}
	}
	// set DNS config using plan
	_, err = r.adg.SetDnsConfig(dnsConfig)
	if err != nil {
		diags.AddError(
			"Unable to Update AdGuard Home Config",
			err.Error(),
		)
		return
	}

	// instantiate empty dns access list for storing plan data
	var dnsAccess adguard.AccessList
	// populate dns access list from plan
	if len(planDnsConfig.AllowedClients.Elements()) > 0 {
		*diags = planDnsConfig.AllowedClients.ElementsAs(ctx, &dnsAccess.AllowedClients, false)
		if diags.HasError() {
			return
		}
	}
	if len(planDnsConfig.DisallowedClients.Elements()) > 0 {
		*diags = planDnsConfig.DisallowedClients.ElementsAs(ctx, &dnsAccess.DisallowedClients, false)
		if diags.HasError() {
			return
		}
	}
	if len(planDnsConfig.BlockedHosts.Elements()) > 0 {
		*diags = planDnsConfig.BlockedHosts.ElementsAs(ctx, &dnsAccess.BlockedHosts, false)
		if diags.HasError() {
			return
		}
	}
	// set DNS access list using plan
	_, err = r.adg.SetAccess(dnsAccess)
	if err != nil {
		diags.AddError(
			"Unable to Update AdGuard Home Config",
			err.Error(),
		)
		return
	}

	// if we got here, all went fine
}

// mapSafeSearchConfigFields - will return the list of safe search services that are enabled
func mapSafeSearchConfigServices(v reflect.Value, t reflect.Type) []string {
	// initalize output
	var services []string

	// loop over all safeSearchConfig fields
	for i := 0; i < v.NumField(); i++ {
		// skip the Enabled field
		if t.Field(i).Name != "Enabled" {
			// add service to list if its value is true
			if v.Field(i).Interface().(bool) {
				services = append(services, strings.ToLower(t.Field(i).Name))
			}
		}
	}
	return services
}

// setSafeSearchConfigServices - based on a list of enabled safe search services, will set the safeSearchConfig fields appropriately
func setSafeSearchConfigServices(v reflect.Value, t reflect.Type, services []string) {
	for i := 0; i < v.NumField(); i++ {
		fieldName := strings.ToLower(t.Field(i).Name)
		if contains(services, fieldName) {
			v.Field(i).Set(reflect.ValueOf(true))
		}
	}
}
