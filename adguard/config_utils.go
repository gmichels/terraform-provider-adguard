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

var DEFAULT_BOOTSTRAP_DNS = []string{"9.9.9.10", "149.112.112.10", "2620:fe::10", "2620:fe::fe:10"}
var DEFAULT_UPSTREAM_DNS = []string{"https://dns10.quad9.net/dns-query"}
var DEFAULT_SAFESEARCH_SERVICES = []string{"bing", "duckduckgo", "google", "pixabay", "yandex", "youtube"}
var BLOCKED_SERVICES_ALL = []string{"9gag", "amazon", "bilibili", "cloudflare", "crunchyroll", "dailymotion", "deezer",
	"discord", "disneyplus", "douban", "ebay", "epic_games", "facebook", "gog", "hbomax", "hulu", "icloud_private_relay", "imgur",
	"instagram", "iqiyi", "kakaotalk", "lazada", "leagueoflegends", "line", "mail_ru", "mastodon", "minecraft", "netflix", "ok",
	"onlyfans", "origin", "pinterest", "playstation", "qq", "rakuten_viki", "reddit", "riot_games", "roblox", "shopee", "skype", "snapchat",
	"soundcloud", "spotify", "steam", "telegram", "tiktok", "tinder", "twitch", "twitter", "valorant", "viber", "vimeo", "vk", "voot", "wechat",
	"weibo", "whatsapp", "xboxlive", "youtube", "zhihu"}

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

type configModel struct {
	Filtering       filteringModel
	SafeBrowsing    enabledModel
	ParentalControl enabledModel
	SafeSearch      safeSearchModel
	QueryLog        queryLogConfigModel
	Stats           statsConfigModel
	BlockedServices types.Set
	Dns             dnsConfigModel
}

// nested objects

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
		"enabled":         types.BoolValue(true),
		"update_interval": types.Int64Value(24),
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
	for _, service := range DEFAULT_SAFESEARCH_SERVICES {
		services = append(services, types.StringValue(service))
	}

	return map[string]attr.Value{
		"enabled":  types.BoolValue(false),
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
		"enabled":             types.BoolValue(true),
		"interval":            types.Int64Value(90 * 24),
		"anonymize_client_ip": types.BoolValue(false),
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
		"enabled":  types.BoolValue(true),
		"interval": types.Int64Value(24),
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
	// DefaultLocalPtrUpstreams types.List   `tfsdk:"default_local_ptr_upstreams"`
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
		// "default_local_ptr_upstreams": types.ListType{ElemType: types.StringType},
	}
}

// defaultObject - return default object for this model
func (o dnsConfigModel) defaultObject() map[string]attr.Value {
	bootstrap_dns := []attr.Value{}
	for _, entry := range DEFAULT_BOOTSTRAP_DNS {
		bootstrap_dns = append(bootstrap_dns, types.StringValue(entry))
	}

	upstream_dns := []attr.Value{}
	for _, entry := range DEFAULT_UPSTREAM_DNS {
		upstream_dns = append(upstream_dns, types.StringValue(entry))
	}

	return map[string]attr.Value{
		"bootstrap_dns":             basetypes.NewListValueMust(types.StringType, bootstrap_dns),
		"upstream_dns":              basetypes.NewListValueMust(types.StringType, upstream_dns),
		"rate_limit":                types.Int64Value(20),
		"blocking_mode":             types.StringValue("default"),
		"blocking_ipv4":             types.StringValue(""),
		"blocking_ipv6":             types.StringValue(""),
		"edns_cs_enabled":           types.BoolValue(false),
		"disable_ipv6":              types.BoolValue(false),
		"dnssec_enabled":            types.BoolValue(false),
		"cache_size":                types.Int64Value(4194304),
		"cache_ttl_min":             types.Int64Value(0),
		"cache_ttl_max":             types.Int64Value(0),
		"cache_optimistic":          types.BoolValue(false),
		"upstream_mode":             types.StringValue(""),
		"use_private_ptr_resolvers": types.BoolValue(true),
		"resolve_clients":           types.BoolValue(true),
		"local_ptr_upstreams":       types.SetValueMust(types.StringType, []attr.Value{}),
		// TODO ?
		// "default_local_ptr_upstreams": basetypes.NewListNull(types.StringType),
	}
}

func (d *configResource) ReadConfig(ctx context.Context) (configModel, diag.Diagnostics, error) {
	// initialize variables
	var config configModel
	var diags diag.Diagnostics

	// FILTERING CONFIG
	// get refreshed filtering config value from AdGuard Home
	filteringConfig, err := d.adg.GetAllFilters()
	if err != nil {
		return config, nil, err
	}
	// map filter config to state
	var stateFilteringConfig filteringModel
	stateFilteringConfig.Enabled = types.BoolValue(filteringConfig.Enabled)
	stateFilteringConfig.UpdateInterval = types.Int64Value(int64(filteringConfig.Interval))
	// add to config model
	config.Filtering = stateFilteringConfig

	// SAFE BROWSING
	// get refreshed safe browsing status from AdGuard Home
	safeBrowsingStatus, err := d.adg.GetSafeBrowsingStatus()
	if err != nil {
		return config, nil, err
	}
	// map safe browsing config to state
	var stateSafeBrowsingStatus enabledModel
	stateSafeBrowsingStatus.Enabled = types.BoolValue(*safeBrowsingStatus)
	// add to config model
	config.SafeBrowsing = stateSafeBrowsingStatus

	// PARENTAL CONTROL
	// get refreshed parental control status from AdGuard Home
	parentalStatus, err := d.adg.GetParentalStatus()
	if err != nil {
		return config, nil, err
	}
	// map parental control config to state
	var stateParentalStatus enabledModel
	stateParentalStatus.Enabled = types.BoolValue(*parentalStatus)
	// add to config model
	config.ParentalControl = stateParentalStatus

	// SAFE SEARCH
	// retrieve safe search info
	safeSearchConfig, err := d.adg.GetSafeSearchConfig()
	if err != nil {
		return config, nil, err
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
	if diags.HasError() {
		return config, diags, nil
	}
	// add to config model
	config.SafeSearch = stateSafeSearchConfig

	// QUERY LOG
	// retrieve query log config info
	queryLogConfig, err := d.adg.GetQueryLogConfig()
	if err != nil {
		return config, nil, err
	}
	var stateQueryLogConfig queryLogConfigModel
	stateQueryLogConfig.Enabled = types.BoolValue(queryLogConfig.Enabled)
	stateQueryLogConfig.Interval = types.Int64Value(int64(queryLogConfig.Interval / 1000 / 3600))
	stateQueryLogConfig.AnonymizeClientIp = types.BoolValue(queryLogConfig.AnonymizeClientIp)
	stateQueryLogConfig.Ignored, diags = types.SetValueFrom(ctx, types.StringType, queryLogConfig.Ignored)
	if diags.HasError() {
		return config, diags, nil
	}
	// add to config model
	config.QueryLog = stateQueryLogConfig

	// STATS
	// retrieve server statistics config info
	statsConfig, err := d.adg.GetStatsConfig()
	if err != nil {
		return config, nil, err
	}
	var stateStatsConfig statsConfigModel
	stateStatsConfig.Enabled = types.BoolValue(statsConfig.Enabled)
	stateStatsConfig.Interval = types.Int64Value(int64(statsConfig.Interval / 3600 / 1000))
	stateStatsConfig.Ignored, diags = types.SetValueFrom(ctx, types.StringType, statsConfig.Ignored)
	if diags.HasError() {
		return config, diags, nil
	}

	// add to config model
	config.Stats = stateStatsConfig

	// BLOCKED SERVICES
	// get refreshed blocked services from AdGuard Home
	blockedServices, err := d.adg.GetBlockedServices()
	if err != nil {
		return config, nil, err
	}
	// add to config model
	config.BlockedServices, diags = types.SetValueFrom(ctx, types.StringType, blockedServices)
	if diags.HasError() {
		return config, diags, nil
	}

	// DNS CONFIG
	// retrieve dns config info
	dnsConfig, err := d.adg.GetDnsInfo()
	if err != nil {
		return config, nil, err
	}
	var stateDnsConfig dnsConfigModel
	stateDnsConfig.BootstrapDns, diags = types.ListValueFrom(ctx, types.StringType, dnsConfig.BootstrapDns)
	if diags.HasError() {
		return config, diags, nil
	}
	stateDnsConfig.UpstreamDns, diags = types.ListValueFrom(ctx, types.StringType, dnsConfig.UpstreamDns)
	if diags.HasError() {
		return config, diags, nil
	}
	stateDnsConfig.RateLimit = types.Int64Value(int64(dnsConfig.RateLimit))
	stateDnsConfig.BlockingMode = types.StringValue(dnsConfig.BlockingMode)
	// upstream API does not unset blocking_ipv4 and blocking_ipv6 when previously set
	// and blocking mode changes, so force state to empty values here
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
	stateDnsConfig.LocalPtrUpstreams, diags = types.SetValueFrom(ctx, types.StringType, dnsConfig.LocalPtrUpstreams)
	if diags.HasError() {
		return config, diags, nil
	}

	// stateDnsConfig.DefaultLocalPtrUpstreams, diags = types.ListValueFrom(ctx, types.StringType, dnsConfig.DefaultLocalPtrUpstreams)
	// if diags.HasError() {
	// 	return config, diags, nil
	// }
	// add to config model
	config.Dns = stateDnsConfig

	// return the config
	return config, nil, nil
}

// CreateOrUpdateConfigResource - common function to create or update a config resource
func (r *configResource) CreateOrUpdateConfigResource(ctx context.Context, plan configResourceModel) (diag.Diagnostics, error) {
	// FILTERING CONFIG
	// unpack nested attributes from plan
	var planFiltering filteringModel
	diags := plan.Filtering.As(ctx, &planFiltering, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return diags, nil
	}
	// instantiate empty object for storing plan data
	var filteringConfig adguard.FilterConfig
	// populate filtering config from plan
	filteringConfig.Enabled = planFiltering.Enabled.ValueBool()
	filteringConfig.Interval = uint(planFiltering.UpdateInterval.ValueInt64())

	// set filtering config using plan
	_, err := r.adg.ConfigureFiltering(filteringConfig)
	if err != nil {
		return nil, err
	}

	// SAFE BROWSING
	// unpack nested attributes from plan
	var planSafeBrowsing enabledModel
	diags = plan.SafeBrowsing.As(ctx, &planSafeBrowsing, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return diags, nil
	}
	// set safe browsing status using plan
	err = r.adg.SetSafeBrowsingStatus(planSafeBrowsing.Enabled.ValueBool())
	if err != nil {
		return nil, err
	}

	// PARENTAL CONTROL
	// unpack nested attributes from plan
	var planParentalControl enabledModel
	diags = plan.ParentalControl.As(ctx, &planParentalControl, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return diags, nil
	}
	// set parental control status using plan
	err = r.adg.SetParentalStatus(planParentalControl.Enabled.ValueBool())
	if err != nil {
		return nil, err
	}

	// SAFE SEARCH
	// unpack nested attributes from plan
	var planSafeSearch safeSearchModel
	diags = plan.SafeSearch.As(ctx, &planSafeSearch, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return diags, nil
	}
	// instantiate empty object for storing plan data
	var safeSearchConfig adguard.SafeSearchConfig
	// populate safe search config using plan
	safeSearchConfig.Enabled = planSafeSearch.Enabled.ValueBool()
	if len(planSafeSearch.Services.Elements()) > 0 {
		var safeSearchServicesEnabled []string
		diags = planSafeSearch.Services.ElementsAs(ctx, &safeSearchServicesEnabled, false)
		if diags.HasError() {
			return diags, nil
		}
		// use reflection to set each safeSearchConfig service value dynamically
		v := reflect.ValueOf(&safeSearchConfig).Elem()
		t := v.Type()
		setSafeSearchConfigServices(v, t, safeSearchServicesEnabled)
	}
	// set safe search config using plan
	_, err = r.adg.SetSafeSearchConfig(safeSearchConfig)
	if err != nil {
		return nil, err
	}

	// QUERY LOG
	// unpack nested attributes from plan
	var planQueryLogConfig queryLogConfigModel
	diags = plan.QueryLog.As(ctx, &planQueryLogConfig, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return diags, nil
	}
	// instantiate empty object for storing plan data
	var queryLogConfig adguard.GetQueryLogConfigResponse
	// populate query log config from plan
	queryLogConfig.Enabled = planQueryLogConfig.Enabled.ValueBool()
	queryLogConfig.Interval = uint64(planQueryLogConfig.Interval.ValueInt64() * 3600 * 1000)
	queryLogConfig.AnonymizeClientIp = planQueryLogConfig.AnonymizeClientIp.ValueBool()
	if len(planQueryLogConfig.Ignored.Elements()) > 0 {
		diags = planQueryLogConfig.Ignored.ElementsAs(ctx, &queryLogConfig.Ignored, false)
		if diags.HasError() {
			return diags, nil
		}
	}
	// set query log config using plan
	_, err = r.adg.SetQueryLogConfig(queryLogConfig)
	if err != nil {
		return nil, err
	}

	// STATS
	// unpack nested attributes from plan
	var planStatsConfig statsConfigModel
	diags = plan.Stats.As(ctx, &planStatsConfig, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return diags, nil
	}
	// instantiate empty object for storing plan data
	var statsConfig adguard.GetStatsConfigResponse
	// populate stats from plan
	statsConfig.Enabled = planStatsConfig.Enabled.ValueBool()
	statsConfig.Interval = uint64(planStatsConfig.Interval.ValueInt64() * 3600 * 1000)
	if len(planStatsConfig.Ignored.Elements()) > 0 {
		diags = planStatsConfig.Ignored.ElementsAs(ctx, &statsConfig.Ignored, false)
		if diags.HasError() {
			return diags, nil
		}
	}
	// set stats config using plan
	_, err = r.adg.SetStatsConfig(statsConfig)
	if err != nil {
		return nil, err
	}

	// BLOCKED SERVICES
	// instantiate empty object for storing plan data
	var blockedServices []string
	// populate blocked services from plan
	if len(plan.BlockedServices.Elements()) > 0 {
		diags = plan.BlockedServices.ElementsAs(ctx, &blockedServices, false)
		if diags.HasError() {
			return diags, nil
		}
	}
	// set blocked services using plan
	_, err = r.adg.SetBlockedServices(blockedServices)
	if err != nil {
		return nil, err
	}

	// DNS CONFIG
	// unpack nested attributes from plan
	var planDnsConfig dnsConfigModel
	diags = plan.Dns.As(ctx, &planDnsConfig, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return diags, nil
	}
	// instantiate empty object for storing plan data
	var dnsConfig adguard.DNSConfig
	// populate DNS config from plan
	if len(planDnsConfig.BootstrapDns.Elements()) > 0 {
		diags = planDnsConfig.BootstrapDns.ElementsAs(ctx, &dnsConfig.BootstrapDns, false)
		if diags.HasError() {
			return diags, nil
		}
	}
	if len(planDnsConfig.UpstreamDns.Elements()) > 0 {
		diags = planDnsConfig.UpstreamDns.ElementsAs(ctx, &dnsConfig.UpstreamDns, false)
		if diags.HasError() {
			return diags, nil
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
		diags = planDnsConfig.LocalPtrUpstreams.ElementsAs(ctx, &dnsConfig.LocalPtrUpstreams, false)
		if diags.HasError() {
			return diags, nil
		}
	} else {
		dnsConfig.LocalPtrUpstreams = []string{}
	}
	// set DNS config using plan
	_, err = r.adg.SetDnsConfig(dnsConfig)
	if err != nil {
		return nil, err
	}

	// no errors to return
	return nil, nil
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
