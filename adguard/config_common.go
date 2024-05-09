package adguard

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"

	"github.com/gmichels/adguard-client-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// common config model to be used for working with both resource and data source
type configCommonModel struct {
	ID                           types.String `tfsdk:"id"`
	LastUpdated                  types.String `tfsdk:"last_updated"`
	Filtering                    types.Object `tfsdk:"filtering"`
	SafeBrowsing                 types.Bool   `tfsdk:"safebrowsing"`
	ParentalControl              types.Bool   `tfsdk:"parental_control"`
	SafeSearch                   types.Object `tfsdk:"safesearch"`
	QueryLog                     types.Object `tfsdk:"querylog"`
	Stats                        types.Object `tfsdk:"stats"`
	BlockedServices              types.Set    `tfsdk:"blocked_services"`
	BlockedServicesPauseSchedule types.Object `tfsdk:"blocked_services_pause_schedule"`
	Dns                          types.Object `tfsdk:"dns"`
	Dhcp                         types.Object `tfsdk:"dhcp"`
	Tls                          types.Object `tfsdk:"tls"`
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
	FallbackDns            types.List   `tfsdk:"fallback_dns"`
	ProtectionEnabled      types.Bool   `tfsdk:"protection_enabled"`
	RateLimit              types.Int64  `tfsdk:"rate_limit"`
	RateLimitSubnetLenIpv4 types.Int64  `tfsdk:"rate_limit_subnet_len_ipv4"`
	RateLimitSubnetLenIpv6 types.Int64  `tfsdk:"rate_limit_subnet_len_ipv6"`
	RateLimitWhitelist     types.List   `tfsdk:"rate_limit_whitelist"`
	BlockingMode           types.String `tfsdk:"blocking_mode"`
	BlockingIpv4           types.String `tfsdk:"blocking_ipv4"`
	BlockingIpv6           types.String `tfsdk:"blocking_ipv6"`
	BlockedResponseTtl     types.Int64  `tfsdk:"blocked_response_ttl"`
	EDnsCsEnabled          types.Bool   `tfsdk:"edns_cs_enabled"`
	EDnsCsUseCustom        types.Bool   `tfsdk:"edns_cs_use_custom"`
	EDnsCsCustomIp         types.String `tfsdk:"edns_cs_custom_ip"`
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
		"bootstrap_dns":              types.ListType{ElemType: types.StringType},
		"upstream_dns":               types.ListType{ElemType: types.StringType},
		"fallback_dns":               types.ListType{ElemType: types.StringType},
		"protection_enabled":         types.BoolType,
		"rate_limit":                 types.Int64Type,
		"rate_limit_subnet_len_ipv4": types.Int64Type,
		"rate_limit_subnet_len_ipv6": types.Int64Type,
		"rate_limit_whitelist":       types.ListType{ElemType: types.StringType},
		"blocking_mode":              types.StringType,
		"blocking_ipv4":              types.StringType,
		"blocking_ipv6":              types.StringType,
		"blocked_response_ttl":       types.Int64Type,
		"edns_cs_enabled":            types.BoolType,
		"edns_cs_use_custom":         types.BoolType,
		"edns_cs_custom_ip":          types.StringType,
		"disable_ipv6":               types.BoolType,
		"dnssec_enabled":             types.BoolType,
		"cache_size":                 types.Int64Type,
		"cache_ttl_min":              types.Int64Type,
		"cache_ttl_max":              types.Int64Type,
		"cache_optimistic":           types.BoolType,
		"upstream_mode":              types.StringType,
		"use_private_ptr_resolvers":  types.BoolType,
		"resolve_clients":            types.BoolType,
		"local_ptr_upstreams":        types.SetType{ElemType: types.StringType},
		"allowed_clients":            types.SetType{ElemType: types.StringType},
		"disallowed_clients":         types.SetType{ElemType: types.StringType},
		"blocked_hosts":              types.SetType{ElemType: types.StringType},
	}
}

// defaultObject - return default object for this model
func (o dnsConfigModel) defaultObject() map[string]attr.Value {
	bootstrap_dns := convertToAttr(CONFIG_DNS_BOOTSTRAP)
	upstream_dns := convertToAttr(CONFIG_DNS_UPSTREAM)
	blocked_hosts := convertToAttr(CONFIG_DNS_BLOCKED_HOSTS)

	return map[string]attr.Value{
		"bootstrap_dns":              types.ListValueMust(types.StringType, bootstrap_dns),
		"upstream_dns":               types.ListValueMust(types.StringType, upstream_dns),
		"fallback_dns":               types.ListNull(types.StringType),
		"protection_enabled":         types.BoolValue(CONFIG_DNS_PROTECTION_ENABLED),
		"rate_limit":                 types.Int64Value(CONFIG_DNS_RATE_LIMIT),
		"rate_limit_subnet_len_ipv4": types.Int64Value(CONFIG_DNS_RATE_LIMIT_SUBNET_LEN_IPV4),
		"rate_limit_subnet_len_ipv6": types.Int64Value(CONFIG_DNS_RATE_LIMIT_SUBNET_LEN_IPV6),
		"rate_limit_whitelist":       types.ListNull(types.StringType),
		"blocking_mode":              types.StringValue(CONFIG_DNS_BLOCKING_MODE),
		"blocking_ipv4":              types.StringValue(""),
		"blocking_ipv6":              types.StringValue(""),
		"blocked_response_ttl":       types.Int64Value(CONFIG_DNS_BLOCKED_RESPONSE_TTL),
		"edns_cs_enabled":            types.BoolValue(CONFIG_DNS_EDNS_CS_ENABLED),
		"edns_cs_use_custom":         types.BoolValue(CONFIG_DNS_EDNS_CS_USE_CUSTOM),
		"edns_cs_custom_ip":          types.StringValue(""),
		"disable_ipv6":               types.BoolValue(CONFIG_DNS_DISABLE_IPV6),
		"dnssec_enabled":             types.BoolValue(CONFIG_DNS_DNSSEC_ENABLED),
		"cache_size":                 types.Int64Value(CONFIG_DNS_CACHE_SIZE),
		"cache_ttl_min":              types.Int64Value(CONFIG_DNS_CACHE_TTL_MIN),
		"cache_ttl_max":              types.Int64Value(CONFIG_DNS_CACHE_TTL_MAX),
		"cache_optimistic":           types.BoolValue(CONFIG_DNS_CACHE_OPTIMISTIC),
		"upstream_mode":              types.StringValue(CONFIG_DNS_UPSTREAM_MODE),
		"use_private_ptr_resolvers":  types.BoolValue(CONFIG_DNS_USE_PRIVATE_PTR_RESOLVERS),
		"resolve_clients":            types.BoolValue(CONFIG_DNS_RESOLVE_CLIENTS),
		"local_ptr_upstreams":        types.SetValueMust(types.StringType, []attr.Value{}),
		"allowed_clients":            types.SetNull(types.StringType),
		"disallowed_clients":         types.SetNull(types.StringType),
		"blocked_hosts":              types.SetValueMust(types.StringType, blocked_hosts),
	}
}

// dhcpStatusModel maps DHCP schema data
type dhcpStatusModel struct {
	Enabled      types.Bool   `tfsdk:"enabled"`
	Interface    types.String `tfsdk:"interface"`
	Ipv4Settings types.Object `tfsdk:"ipv4_settings"`
	Ipv6Settings types.Object `tfsdk:"ipv6_settings"`
	Leases       types.List   `tfsdk:"leases"`
	StaticLeases types.Set    `tfsdk:"static_leases"`
}

// attrTypes - return attribute types for this model
func (o dhcpStatusModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled":       types.BoolType,
		"interface":     types.StringType,
		"ipv4_settings": types.ObjectType{AttrTypes: dhcpIpv4Model{}.attrTypes()},
		"ipv6_settings": types.ObjectType{AttrTypes: dhcpIpv6Model{}.attrTypes()},
		"leases":        types.ListType{ElemType: types.ObjectType{AttrTypes: dhcpLeasesModel{}.attrTypes()}},
		"static_leases": types.SetType{ElemType: types.ObjectType{AttrTypes: dhcpStaticLeasesModel{}.attrTypes()}},
	}
}

// dhcpConfigModel maps DHCP schema data
type dhcpConfigModel struct {
	Enabled      types.Bool   `tfsdk:"enabled"`
	Interface    types.String `tfsdk:"interface"`
	Ipv4Settings types.Object `tfsdk:"ipv4_settings"`
	Ipv6Settings types.Object `tfsdk:"ipv6_settings"`
	StaticLeases types.Set    `tfsdk:"static_leases"`
}

// attrTypes - return attribute types for this model
func (o dhcpConfigModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled":       types.BoolType,
		"interface":     types.StringType,
		"ipv4_settings": types.ObjectType{AttrTypes: dhcpIpv4Model{}.attrTypes()},
		"ipv6_settings": types.ObjectType{AttrTypes: dhcpIpv6Model{}.attrTypes()},
		"static_leases": types.SetType{ElemType: types.ObjectType{AttrTypes: dhcpStaticLeasesModel{}.attrTypes()}},
	}
}

// defaultObject - return default object for this model
func (o dhcpConfigModel) defaultObject() map[string]attr.Value {
	return map[string]attr.Value{
		"enabled":       types.BoolValue(CONFIG_DHCP_ENABLED),
		"interface":     types.StringValue(""),
		"ipv4_settings": types.ObjectValueMust(dhcpIpv4Model{}.attrTypes(), dhcpIpv4Model{}.defaultObject()),
		"ipv6_settings": types.ObjectValueMust(dhcpIpv6Model{}.attrTypes(), dhcpIpv6Model{}.defaultObject()),
		"static_leases": types.SetNull(types.ObjectType{AttrTypes: dhcpStaticLeasesModel{}.attrTypes()}),
	}
}

// dhcpIpv4Model maps DHCP IPv4 settings schema data
type dhcpIpv4Model struct {
	GatewayIp     types.String `tfsdk:"gateway_ip"`
	SubnetMask    types.String `tfsdk:"subnet_mask"`
	RangeStart    types.String `tfsdk:"range_start"`
	RangeEnd      types.String `tfsdk:"range_end"`
	LeaseDuration types.Int64  `tfsdk:"lease_duration"`
}

// attrTypes - return attribute types for this model
func (o dhcpIpv4Model) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"gateway_ip":     types.StringType,
		"subnet_mask":    types.StringType,
		"range_start":    types.StringType,
		"range_end":      types.StringType,
		"lease_duration": types.Int64Type,
	}
}

// defaultObject - return default object for this model
func (o dhcpIpv4Model) defaultObject() map[string]attr.Value {
	return map[string]attr.Value{
		"gateway_ip":     types.StringValue(""),
		"subnet_mask":    types.StringValue(""),
		"range_start":    types.StringValue(""),
		"range_end":      types.StringValue(""),
		"lease_duration": types.Int64Value(CONFIG_DHCP_V4_LEASE_DURATION),
	}
}

// dhcpIpv6Model maps DHCP IPv6 settings schema data
type dhcpIpv6Model struct {
	RangeStart    types.String `tfsdk:"range_start"`
	LeaseDuration types.Int64  `tfsdk:"lease_duration"`
}

// attrTypes - return attribute types for this model
func (o dhcpIpv6Model) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"range_start":    types.StringType,
		"lease_duration": types.Int64Type,
	}
}

// defaultObject - return default object for this model
func (o dhcpIpv6Model) defaultObject() map[string]attr.Value {
	return map[string]attr.Value{
		"range_start":    types.StringValue(""),
		"lease_duration": types.Int64Value(CONFIG_DHCP_V6_LEASE_DURATION),
	}
}

// dhcpLeasesModel maps DHCP leases schema data
type dhcpLeasesModel struct {
	Mac      types.String `tfsdk:"mac"`
	Ip       types.String `tfsdk:"ip"`
	Hostname types.String `tfsdk:"hostname"`
	Expires  types.String `tfsdk:"expires"`
}

// attrTypes - return attribute types for this model
func (o dhcpLeasesModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"mac":      types.StringType,
		"ip":       types.StringType,
		"hostname": types.StringType,
		"expires":  types.StringType,
	}
}

// dhcpStaticLeasesModel maps DHCP leases schema data
type dhcpStaticLeasesModel struct {
	Mac      types.String `tfsdk:"mac"`
	Ip       types.String `tfsdk:"ip"`
	Hostname types.String `tfsdk:"hostname"`
}

// attrTypes - return attribute types for this model
func (o dhcpStaticLeasesModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"mac":      types.StringType,
		"ip":       types.StringType,
		"hostname": types.StringType,
	}
}

// tlsConfigModel maps filtering schema data
type tlsConfigModel struct {
	Enabled           types.Bool   `tfsdk:"enabled"`
	ServerName        types.String `tfsdk:"server_name"`
	ForceHttps        types.Bool   `tfsdk:"force_https"`
	PortHttps         types.Int64  `tfsdk:"port_https"`
	PortDnsOverTls    types.Int64  `tfsdk:"port_dns_over_tls"`
	PortDnsOverQuic   types.Int64  `tfsdk:"port_dns_over_quic"`
	CertificateChain  types.String `tfsdk:"certificate_chain"`
	PrivateKey        types.String `tfsdk:"private_key"`
	PrivateKeySaved   types.Bool   `tfsdk:"private_key_saved"`
	ValidCert         types.Bool   `tfsdk:"valid_cert"`
	ValidChain        types.Bool   `tfsdk:"valid_chain"`
	Subject           types.String `tfsdk:"subject"`
	Issuer            types.String `tfsdk:"issuer"`
	NotBefore         types.String `tfsdk:"not_before"`
	NotAfter          types.String `tfsdk:"not_after"`
	DnsNames          types.List   `tfsdk:"dns_names"`
	ValidKey          types.Bool   `tfsdk:"valid_key"`
	KeyType           types.String `tfsdk:"key_type"`
	WarningValidation types.String `tfsdk:"warning_validation"`
	ValidPair         types.Bool   `tfsdk:"valid_pair"`
	ServePlainDns     types.Bool   `tfsdk:"serve_plain_dns"`
}

// attrTypes - return attribute types for this model
func (o tlsConfigModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled":            types.BoolType,
		"server_name":        types.StringType,
		"force_https":        types.BoolType,
		"port_https":         types.Int64Type,
		"port_dns_over_tls":  types.Int64Type,
		"port_dns_over_quic": types.Int64Type,
		"certificate_chain":  types.StringType,
		"private_key":        types.StringType,
		"private_key_saved":  types.BoolType,
		"valid_cert":         types.BoolType,
		"valid_chain":        types.BoolType,
		"subject":            types.StringType,
		"issuer":             types.StringType,
		"not_before":         types.StringType,
		"not_after":          types.StringType,
		"dns_names":          types.ListType{ElemType: types.StringType},
		"valid_key":          types.BoolType,
		"key_type":           types.StringType,
		"warning_validation": types.StringType,
		"valid_pair":         types.BoolType,
		"serve_plain_dns":    types.BoolType,
	}
}

// defaultObject - return default object for this model
func (o tlsConfigModel) defaultObject() map[string]attr.Value {
	return map[string]attr.Value{
		"enabled":            types.BoolValue(CONFIG_TLS_ENABLED),
		"server_name":        types.StringValue(""),
		"force_https":        types.BoolValue(CONFIG_TLS_FORCE_HTTPS),
		"port_https":         types.Int64Value(CONFIG_TLS_PORT_HTTPS),
		"port_dns_over_tls":  types.Int64Value(CONFIG_TLS_PORT_DNS_OVER_TLS),
		"port_dns_over_quic": types.Int64Value(CONFIG_TLS_PORT_DNS_OVER_QUIC),
		"certificate_chain":  types.StringValue(""),
		"private_key":        types.StringValue(""),
		"private_key_saved":  types.BoolValue(false),
		"valid_cert":         types.BoolValue(false),
		"valid_chain":        types.BoolValue(false),
		"valid_key":          types.BoolValue(false),
		"valid_pair":         types.BoolValue(false),
		"key_type":           types.StringValue(""),
		"subject":            types.StringValue(""),
		"issuer":             types.StringValue(""),
		"not_before":         types.StringValue(""),
		"not_after":          types.StringValue(""),
		"dns_names":          types.ListValueMust(types.StringType, []attr.Value{}),
		"warning_validation": types.StringValue(""),
		"serve_plain_dns":    types.BoolValue(CONFIG_TLS_SERVE_PLAIN_DNS),
	}
}

// common `Read` function for both data source and resource
func (o *configCommonModel) Read(ctx context.Context, adg adguard.ADG, currState *configCommonModel, diags *diag.Diagnostics, rtype string) {
	// initialize empty diags variable
	var d diag.Diagnostics

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
	// convert to JSON for response logging
	filteringConfigJson, err := json.Marshal(filteringConfig)
	if err != nil {
		diags.AddError(
			"Unable to Parse AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// log response body
	tflog.Debug(ctx, "ADG API response", map[string]interface{}{
		"object": "filteringConfig",
		"body":   string(filteringConfigJson),
	})
	// map filter config to state
	var stateFilteringConfig filteringModel
	stateFilteringConfig.Enabled = types.BoolValue(filteringConfig.Enabled)
	stateFilteringConfig.UpdateInterval = types.Int64Value(int64(filteringConfig.Interval))
	// add to config model
	o.Filtering, d = types.ObjectValueFrom(ctx, filteringModel{}.attrTypes(), &stateFilteringConfig)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

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
	// log response body
	tflog.Debug(ctx, "ADG API response", map[string]interface{}{
		"object": "safeBrowsingStatus",
		"body":   strconv.FormatBool(*safeBrowsingStatus),
	})
	// add to config model
	o.SafeBrowsing = types.BoolValue(*safeBrowsingStatus)

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
	tflog.Debug(ctx, "ADG API response", map[string]interface{}{
		"object": "parentalStatus",
		"body":   strconv.FormatBool(*parentalStatus),
	})
	// add to config model
	o.ParentalControl = types.BoolValue(*parentalStatus)

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
	// convert to JSON for response logging
	safeSearchConfigJson, err := json.Marshal(safeSearchConfig)
	if err != nil {
		diags.AddError(
			"Unable to Parse AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// log response body
	tflog.Debug(ctx, "ADG API response", map[string]interface{}{
		"object": "safeSearchConfig",
		"body":   string(safeSearchConfigJson),
	})
	// map safe search config object to a list of enabled services
	enabledSafeSearchServices := mapSafeSearchServices(safeSearchConfig)
	// map safe search to state
	var stateSafeSearchConfig safeSearchModel
	stateSafeSearchConfig.Enabled = types.BoolValue(safeSearchConfig.Enabled)
	stateSafeSearchConfig.Services, d = types.SetValueFrom(ctx, types.StringType, enabledSafeSearchServices)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	// add to config model
	o.SafeSearch, d = types.ObjectValueFrom(ctx, safeSearchModel{}.attrTypes(), &stateSafeSearchConfig)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

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
	// convert to JSON for response logging
	queryLogConfigJson, err := json.Marshal(queryLogConfig)
	if err != nil {
		diags.AddError(
			"Unable to Parse AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// log response body
	tflog.Debug(ctx, "ADG API response", map[string]interface{}{
		"object": "queryLogConfig",
		"body":   string(queryLogConfigJson),
	})
	var stateQueryLogConfig queryLogConfigModel
	stateQueryLogConfig.Enabled = types.BoolValue(queryLogConfig.Enabled)
	stateQueryLogConfig.Interval = types.Int64Value(int64(queryLogConfig.Interval / 1000 / 3600))
	stateQueryLogConfig.AnonymizeClientIp = types.BoolValue(queryLogConfig.AnonymizeClientIp)
	if len(queryLogConfig.Ignored) > 0 {
		stateQueryLogConfig.Ignored, d = types.SetValueFrom(ctx, types.StringType, queryLogConfig.Ignored)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	} else {
		stateQueryLogConfig.Ignored = types.SetValueMust(types.StringType, []attr.Value{})
	}
	// add to config model
	o.QueryLog, d = types.ObjectValueFrom(ctx, queryLogConfigModel{}.attrTypes(), &stateQueryLogConfig)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

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
	// convert to JSON for response logging
	statsConfigJson, err := json.Marshal(statsConfig)
	if err != nil {
		diags.AddError(
			"Unable to Parse AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// log response body
	tflog.Debug(ctx, "ADG API response", map[string]interface{}{
		"object": "statsConfig",
		"body":   string(statsConfigJson),
	})
	var stateStatsConfig statsConfigModel
	stateStatsConfig.Enabled = types.BoolValue(statsConfig.Enabled)
	stateStatsConfig.Interval = types.Int64Value(int64(statsConfig.Interval / 3600 / 1000))
	if len(statsConfig.Ignored) > 0 {
		stateStatsConfig.Ignored, d = types.SetValueFrom(ctx, types.StringType, statsConfig.Ignored)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	} else {
		stateStatsConfig.Ignored = types.SetValueMust(types.StringType, []attr.Value{})
	}
	// add to config model
	o.Stats, d = types.ObjectValueFrom(ctx, statsConfigModel{}.attrTypes(), &stateStatsConfig)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	// BLOCKED SERVICES
	// get refreshed blocked services from AdGuard Home
	blockedServicesPauseSchedule, err := adg.GetBlockedServices()
	if err != nil {
		diags.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// convert to JSON for response logging
	blockedServicesPauseScheduleJson, err := json.Marshal(blockedServicesPauseSchedule)
	if err != nil {
		diags.AddError(
			"Unable to Parse AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// log response body
	tflog.Debug(ctx, "ADG API response", map[string]interface{}{
		"object": "blockedServicesPauseSchedule",
		"body":   string(blockedServicesPauseScheduleJson),
	})

	// use common function to map blocked services pause schedules for each day
	stateBlockedServicesPauseScheduleConfig := mapAdgScheduleToBlockedServicesPauseSchedule(ctx, &blockedServicesPauseSchedule.Schedule, diags)
	if diags.HasError() {
		return
	}

	// need special handling for timezone in resource due to inconsistent API response for `Local`
	if rtype == "resource" && !currState.BlockedServicesPauseSchedule.IsNull() {
		// last updated will exist on create operation, null on import operation
		if !currState.LastUpdated.IsNull() {
			// unpack current state
			var currStateBlockedServicesPauseScheduleConfig scheduleModel
			d = currState.BlockedServicesPauseSchedule.As(ctx, &currStateBlockedServicesPauseScheduleConfig, basetypes.ObjectAsOptions{})
			diags.Append(d...)
			if diags.HasError() {
				return
			}
			// if timezone in state is null, it means it was never defined, so we should ignore the inconsistent response from ADG
			if !currStateBlockedServicesPauseScheduleConfig.TimeZone.IsNull() {
				// map timezone from response
				stateBlockedServicesPauseScheduleConfig.TimeZone = types.StringValue(blockedServicesPauseSchedule.Schedule.TimeZone)
			}
			// ID exists in both create and import operations, but if we got here, it's an import
			// still, imports for this attribute are finicky and error-prone, therefore ignored in tests
		} else if !currState.ID.IsNull() {
			// map timezone from response
			stateBlockedServicesPauseScheduleConfig.TimeZone = types.StringValue(blockedServicesPauseSchedule.Schedule.TimeZone)
		}
	} else {
		// used for datasource
		stateBlockedServicesPauseScheduleConfig.TimeZone = types.StringValue(blockedServicesPauseSchedule.Schedule.TimeZone)
	}

	// add to config model
	o.BlockedServices, d = types.SetValueFrom(ctx, types.StringType, blockedServicesPauseSchedule.Ids)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	o.BlockedServicesPauseSchedule, d = types.ObjectValueFrom(ctx, scheduleModel{}.attrTypes(), &stateBlockedServicesPauseScheduleConfig)
	diags.Append(d...)
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
	// convert to JSON for response logging
	dnsConfigJson, err := json.Marshal(dnsConfig)
	if err != nil {
		diags.AddError(
			"Unable to Parse AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// log response body
	tflog.Debug(ctx, "ADG API response", map[string]interface{}{
		"object": "dnsConfig",
		"body":   string(dnsConfigJson),
	})
	// retrieve dns config info
	var stateDnsConfig dnsConfigModel
	stateDnsConfig.BootstrapDns, d = types.ListValueFrom(ctx, types.StringType, dnsConfig.BootstrapDns)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	stateDnsConfig.UpstreamDns, d = types.ListValueFrom(ctx, types.StringType, dnsConfig.UpstreamDns)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	if len(dnsConfig.FallbackDns) == 0 && rtype == "resource" {
		stateDnsConfig.FallbackDns = types.ListNull(types.StringType)
	} else {
		stateDnsConfig.FallbackDns, d = types.ListValueFrom(ctx, types.StringType, dnsConfig.FallbackDns)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	}
	stateDnsConfig.ProtectionEnabled = types.BoolValue(dnsConfig.ProtectionEnabled)
	stateDnsConfig.RateLimit = types.Int64Value(int64(dnsConfig.RateLimit))
	stateDnsConfig.RateLimitSubnetLenIpv4 = types.Int64Value(int64(dnsConfig.RateLimitSubnetSubnetLenIpv4))
	stateDnsConfig.RateLimitSubnetLenIpv6 = types.Int64Value(int64(dnsConfig.RateLimitSubnetSubnetLenIpv6))
	if len(dnsConfig.RateLimitWhitelist) == 0 && rtype == "resource" {
		stateDnsConfig.RateLimitWhitelist = types.ListNull(types.StringType)
	} else {
		stateDnsConfig.RateLimitWhitelist, d = types.ListValueFrom(ctx, types.StringType, dnsConfig.RateLimitWhitelist)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	}
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
	stateDnsConfig.BlockedResponseTtl = types.Int64Value(int64(dnsConfig.BlockedResponseTtl))
	stateDnsConfig.EDnsCsEnabled = types.BoolValue(dnsConfig.EDnsCsEnabled)
	stateDnsConfig.EDnsCsUseCustom = types.BoolValue(dnsConfig.EDnsCsUseCustom)
	if !dnsConfig.EDnsCsUseCustom {
		// ignore whatever is in the API response for EDNS custom IP
		// as it doesn't get actually removed when not in use
		stateDnsConfig.EDnsCsCustomIp = types.StringValue("")
	} else {
		stateDnsConfig.EDnsCsCustomIp = types.StringValue(dnsConfig.EDnsCsCustomIp)
	}
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
	stateDnsConfig.LocalPtrUpstreams, d = types.SetValueFrom(ctx, types.StringType, dnsConfig.LocalPtrUpstreams)
	diags.Append(d...)
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
	// convert to JSON for response logging
	dnsAccessJson, err := json.Marshal(dnsAccess)
	if err != nil {
		diags.AddError(
			"Unable to Parse AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// log response body
	tflog.Debug(ctx, "ADG API response", map[string]interface{}{
		"object": "dnsAccess",
		"body":   string(dnsAccessJson),
	})
	stateDnsConfig.AllowedClients, d = types.SetValueFrom(ctx, types.StringType, dnsAccess.AllowedClients)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	stateDnsConfig.DisallowedClients, d = types.SetValueFrom(ctx, types.StringType, dnsAccess.DisallowedClients)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	stateDnsConfig.BlockedHosts, d = types.SetValueFrom(ctx, types.StringType, dnsAccess.BlockedHosts)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	// add to config model
	o.Dns, d = types.ObjectValueFrom(ctx, dnsConfigModel{}.attrTypes(), &stateDnsConfig)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	// DHCP
	// retrieve dhcp info
	dhcpStatus, err := adg.GetDhcpStatus()
	if err != nil {
		diags.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// convert to JSON for response logging
	dhcpStatusJson, err := json.Marshal(dhcpStatus)
	if err != nil {
		diags.AddError(
			"Unable to Parse AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// log response body
	tflog.Debug(ctx, "ADG API response", map[string]interface{}{
		"object": "dhcpStatus",
		"body":   string(dhcpStatusJson),
	})
	// parse double-nested attributes first
	var stateDhcpIpv4Config dhcpIpv4Model
	stateDhcpIpv4Config.GatewayIp = types.StringValue(dhcpStatus.V4.GatewayIp)
	stateDhcpIpv4Config.SubnetMask = types.StringValue(dhcpStatus.V4.SubnetMask)
	stateDhcpIpv4Config.RangeStart = types.StringValue(dhcpStatus.V4.RangeStart)
	stateDhcpIpv4Config.RangeEnd = types.StringValue(dhcpStatus.V4.RangeEnd)
	stateDhcpIpv4Config.LeaseDuration = types.Int64Value(int64(dhcpStatus.V4.LeaseDuration))

	var stateDhcpIpv6Config dhcpIpv6Model
	stateDhcpIpv6Config.RangeStart = types.StringValue(dhcpStatus.V6.RangeStart)
	stateDhcpIpv6Config.LeaseDuration = types.Int64Value(int64(dhcpStatus.V6.LeaseDuration))

	// now parse the top nested attribute
	var stateDhcpConfig dhcpConfigModel
	stateDhcpConfig.Enabled = types.BoolValue(dhcpStatus.Enabled)
	stateDhcpConfig.Interface = types.StringValue(dhcpStatus.InterfaceName)

	// add double-nested to top nested
	stateDhcpConfig.Ipv4Settings, d = types.ObjectValueFrom(ctx, dhcpIpv4Model{}.attrTypes(), &stateDhcpIpv4Config)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	stateDhcpConfig.Ipv6Settings, d = types.ObjectValueFrom(ctx, dhcpIpv6Model{}.attrTypes(), &stateDhcpIpv6Config)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	if len(dhcpStatus.StaticLeases) > 0 {
		// need to go through all entries to create a slice
		var dhcpStaticLeases []dhcpStaticLeasesModel
		var stateDhcpConfigStaticLease dhcpStaticLeasesModel
		for _, dhcpStaticLease := range dhcpStatus.StaticLeases {
			stateDhcpConfigStaticLease.Mac = types.StringValue(dhcpStaticLease.Mac)
			stateDhcpConfigStaticLease.Ip = types.StringValue(dhcpStaticLease.Ip)
			stateDhcpConfigStaticLease.Hostname = types.StringValue(dhcpStaticLease.Hostname)
			dhcpStaticLeases = append(dhcpStaticLeases, stateDhcpConfigStaticLease)
		}
		// convert to a set
		stateDhcpConfig.StaticLeases, d = types.SetValueFrom(ctx, types.ObjectType{AttrTypes: dhcpStaticLeasesModel{}.attrTypes()}, dhcpStaticLeases)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	} else {
		// use a null set
		stateDhcpConfig.StaticLeases = types.SetNull(types.ObjectType{AttrTypes: dhcpStaticLeasesModel{}.attrTypes()})
	}

	if rtype == "resource" {
		// no need to do anything else, just add to config model
		o.Dhcp, d = types.ObjectValueFrom(ctx, dhcpConfigModel{}.attrTypes(), &stateDhcpConfig)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	} else {
		// data source has a slightly different model, need to transfer over the attributes
		var stateDhcpStatus dhcpStatusModel
		stateDhcpStatus.Enabled = stateDhcpConfig.Enabled
		stateDhcpStatus.Interface = stateDhcpConfig.Interface
		stateDhcpStatus.Ipv4Settings = stateDhcpConfig.Ipv4Settings
		stateDhcpStatus.Ipv6Settings = stateDhcpConfig.Ipv6Settings
		stateDhcpStatus.StaticLeases = stateDhcpConfig.StaticLeases
		if len(dhcpStatus.Leases) > 0 {
			// need to go through all entries to create a slice
			var dhcpLeases []dhcpLeasesModel
			var stateDhcpConfigLease dhcpLeasesModel
			for _, dhcpLease := range dhcpStatus.Leases {
				stateDhcpConfigLease.Mac = types.StringValue(dhcpLease.Mac)
				stateDhcpConfigLease.Ip = types.StringValue(dhcpLease.Ip)
				stateDhcpConfigLease.Hostname = types.StringValue(dhcpLease.Hostname)
				stateDhcpConfigLease.Expires = types.StringValue(dhcpLease.Expires)
				dhcpLeases = append(dhcpLeases, stateDhcpConfigLease)
			}
			// convert to a set
			stateDhcpStatus.Leases, d = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: dhcpLeasesModel{}.attrTypes()}, dhcpLeases)
			diags.Append(d...)
			if diags.HasError() {
				return
			}
		} else {
			// use a null set
			stateDhcpStatus.Leases = types.ListNull(types.ObjectType{AttrTypes: dhcpLeasesModel{}.attrTypes()})
		}
		// add to config model
		o.Dhcp, d = types.ObjectValueFrom(ctx, dhcpStatusModel{}.attrTypes(), &stateDhcpStatus)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	}

	// TLS
	// get refreshed filtering config value from AdGuard Home
	tlsConfig, err := adg.GetTlsConfig()
	if err != nil {
		diags.AddError(
			"Unable to Read AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// convert to JSON for response logging
	tlsConfigJson, err := json.Marshal(tlsConfig)
	if err != nil {
		diags.AddError(
			"Unable to Parse AdGuard Home Config",
			err.Error(),
		)
		return
	}
	// log response body
	tflog.Debug(ctx, "ADG API response", map[string]interface{}{
		"object": "tlsConfig",
		"body":   string(tlsConfigJson),
	})
	// map filter config to state
	var stateTlsConfig tlsConfigModel
	stateTlsConfig.Enabled = types.BoolValue(tlsConfig.Enabled)
	stateTlsConfig.ServerName = types.StringValue(tlsConfig.ServerName)
	stateTlsConfig.ForceHttps = types.BoolValue(tlsConfig.ForceHttps)
	stateTlsConfig.PortHttps = types.Int64Value(int64(tlsConfig.PortHttps))
	stateTlsConfig.PortDnsOverTls = types.Int64Value(int64(tlsConfig.PortDnsOverTls))
	stateTlsConfig.PortDnsOverQuic = types.Int64Value(int64(tlsConfig.PortDnsOverQuic))
	// check if the certificate chain is provided directly or as a file path
	if tlsConfig.CertificateChain != "" {
		// it's the base64 PEM file
		stateTlsConfig.CertificateChain = types.StringValue(tlsConfig.CertificateChain)
	} else {
		// it's a file path
		stateTlsConfig.CertificateChain = types.StringValue(tlsConfig.CertificatePath)
	}
	// check if the private key is provided directly or as a file path
	if tlsConfig.PrivateKey != "" {
		// it's the base64 PEM file
		stateTlsConfig.PrivateKey = types.StringValue(tlsConfig.PrivateKey)
	} else {
		// it's a file path
		stateTlsConfig.PrivateKey = types.StringValue(tlsConfig.PrivateKeyPath)
	}
	stateTlsConfig.PrivateKeySaved = types.BoolValue(tlsConfig.PrivateKeySaved)
	stateTlsConfig.ValidCert = types.BoolValue(tlsConfig.ValidCert)
	stateTlsConfig.ValidChain = types.BoolValue(tlsConfig.ValidChain)
	stateTlsConfig.Subject = types.StringValue(tlsConfig.Subject)
	stateTlsConfig.Issuer = types.StringValue(tlsConfig.Issuer)
	// handle default timestamp from upstream
	if tlsConfig.NotBefore != "0001-01-01T00:00:00Z" {
		stateTlsConfig.NotBefore = types.StringValue(tlsConfig.NotBefore)
	} else {
		stateTlsConfig.NotBefore = types.StringValue("")
	}
	// handle default timestamp from upstream
	if tlsConfig.NotAfter != "0001-01-01T00:00:00Z" {
		stateTlsConfig.NotAfter = types.StringValue(tlsConfig.NotAfter)
	} else {
		stateTlsConfig.NotAfter = types.StringValue("")
	}
	if len(tlsConfig.DnsNames) > 0 {
		stateTlsConfig.DnsNames, d = types.ListValueFrom(ctx, types.StringType, tlsConfig.DnsNames)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	} else {
		stateTlsConfig.DnsNames = types.ListValueMust(types.StringType, []attr.Value{})
	}
	stateTlsConfig.ValidKey = types.BoolValue(tlsConfig.ValidKey)
	stateTlsConfig.KeyType = types.StringValue(tlsConfig.KeyType)
	stateTlsConfig.WarningValidation = types.StringValue(tlsConfig.WarningValidation)
	stateTlsConfig.ValidPair = types.BoolValue(tlsConfig.ValidPair)
	stateTlsConfig.ServePlainDns = types.BoolValue(tlsConfig.ServePlainDns)

	// add to config model
	o.Tls, d = types.ObjectValueFrom(ctx, tlsConfigModel{}.attrTypes(), &stateTlsConfig)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	// if we got here, all went fine
}

// common `Create` and `Update` function for the resource
func (r *configResource) CreateOrUpdate(ctx context.Context, plan *configCommonModel, state *configCommonModel, diags *diag.Diagnostics) {
	// initialize empty diags variable
	var d diag.Diagnostics

	// FILTERING CONFIG
	// unpack nested attributes from plan
	var planFiltering filteringModel
	d = plan.Filtering.As(ctx, &planFiltering, basetypes.ObjectAsOptions{})
	diags.Append(d...)
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
	// set safe browsing status using plan
	err = r.adg.SetSafeBrowsingStatus(plan.SafeBrowsing.ValueBool())
	if err != nil {
		diags.AddError(
			"Unable to Update AdGuard Home Config",
			err.Error(),
		)
		return
	}

	// PARENTAL CONTROL
	// set parental control status using plan
	err = r.adg.SetParentalStatus(plan.ParentalControl.ValueBool())
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
	d = plan.SafeSearch.As(ctx, &planSafeSearch, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	// instantiate empty object for storing plan data
	var safeSearchConfig adguard.SafeSearchConfig
	// populate safe search config using plan
	safeSearchConfig.Enabled = planSafeSearch.Enabled.ValueBool()
	if len(planSafeSearch.Services.Elements()) > 0 {
		var safeSearchServicesEnabled []string
		d = planSafeSearch.Services.ElementsAs(ctx, &safeSearchServicesEnabled, false)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
		// use reflection to set each safeSearchConfig service value dynamically
		v := reflect.ValueOf(&safeSearchConfig).Elem()
		t := v.Type()
		setSafeSearchServices(v, t, safeSearchServicesEnabled)
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
	d = plan.QueryLog.As(ctx, &planQueryLogConfig, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	// instantiate empty object for storing plan data
	var queryLogConfig adguard.GetQueryLogConfigResponse
	// populate query log config from plan
	queryLogConfig.Enabled = planQueryLogConfig.Enabled.ValueBool()
	queryLogConfig.Interval = uint64(planQueryLogConfig.Interval.ValueInt64() * 3600 * 1000)
	queryLogConfig.AnonymizeClientIp = planQueryLogConfig.AnonymizeClientIp.ValueBool()
	d = planQueryLogConfig.Ignored.ElementsAs(ctx, &queryLogConfig.Ignored, false)
	diags.Append(d...)
	if diags.HasError() {
		return
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
	d = plan.Stats.As(ctx, &planStatsConfig, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	// instantiate empty object for storing plan data
	var statsConfig adguard.GetStatsConfigResponse
	// populate stats from plan
	statsConfig.Enabled = planStatsConfig.Enabled.ValueBool()
	statsConfig.Interval = uint64(planStatsConfig.Interval.ValueInt64() * 3600 * 1000)
	d = planStatsConfig.Ignored.ElementsAs(ctx, &statsConfig.Ignored, false)
	diags.Append(d...)
	if diags.HasError() {
		return
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
		d = plan.BlockedServices.ElementsAs(ctx, &blockedServices, false)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	}
	// unpack nested attributes from plan
	var planBlockedServicesPauseScheduleConfig scheduleModel
	d = plan.BlockedServicesPauseSchedule.As(ctx, &planBlockedServicesPauseScheduleConfig, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	// instantiate empty object for storing plan data
	var blockedServicesPauseScheduleConfig adguard.BlockedServicesSchedule
	// populate blocked services schedule from plan
	blockedServicesPauseScheduleConfig.Ids = blockedServices
	// defer to common function to populate schedule
	blockedServicesPauseScheduleConfig.Schedule = mapBlockedServicesPauseScheduleToAdgSchedule(ctx, planBlockedServicesPauseScheduleConfig, diags)
	if diags.HasError() {
		return
	}

	// set blocked services and schedule using plan
	_, err = r.adg.SetBlockedServices(blockedServicesPauseScheduleConfig)
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
	d = plan.Dns.As(ctx, &planDnsConfig, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	// instantiate empty object for storing plan data
	var dnsConfig adguard.DNSConfig
	// populate DNS config from plan
	if len(planDnsConfig.BootstrapDns.Elements()) > 0 {
		d = planDnsConfig.BootstrapDns.ElementsAs(ctx, &dnsConfig.BootstrapDns, false)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	}
	if len(planDnsConfig.UpstreamDns.Elements()) > 0 {
		d = planDnsConfig.UpstreamDns.ElementsAs(ctx, &dnsConfig.UpstreamDns, false)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	}
	if len(planDnsConfig.FallbackDns.Elements()) > 0 {
		d = planDnsConfig.FallbackDns.ElementsAs(ctx, &dnsConfig.FallbackDns, false)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	} else {
		dnsConfig.FallbackDns = []string{}
	}
	dnsConfig.ProtectionEnabled = planDnsConfig.ProtectionEnabled.ValueBool()
	dnsConfig.RateLimit = uint(planDnsConfig.RateLimit.ValueInt64())
	dnsConfig.RateLimitSubnetSubnetLenIpv4 = uint(planDnsConfig.RateLimitSubnetLenIpv4.ValueInt64())
	dnsConfig.RateLimitSubnetSubnetLenIpv6 = uint(planDnsConfig.RateLimitSubnetLenIpv6.ValueInt64())
	if len(planDnsConfig.RateLimitWhitelist.Elements()) > 0 {
		d = planDnsConfig.RateLimitWhitelist.ElementsAs(ctx, &dnsConfig.RateLimitWhitelist, false)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	} else {
		dnsConfig.RateLimitWhitelist = []string{}
	}
	dnsConfig.BlockingMode = planDnsConfig.BlockingMode.ValueString()
	dnsConfig.BlockingIpv4 = planDnsConfig.BlockingIpv4.ValueString()
	dnsConfig.BlockingIpv6 = planDnsConfig.BlockingIpv6.ValueString()
	dnsConfig.BlockedResponseTtl = uint(planDnsConfig.BlockedResponseTtl.ValueInt64())
	dnsConfig.EDnsCsEnabled = planDnsConfig.EDnsCsEnabled.ValueBool()
	dnsConfig.EDnsCsUseCustom = planDnsConfig.EDnsCsUseCustom.ValueBool()
	dnsConfig.EDnsCsCustomIp = planDnsConfig.EDnsCsCustomIp.ValueString()
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
		d = planDnsConfig.LocalPtrUpstreams.ElementsAs(ctx, &dnsConfig.LocalPtrUpstreams, false)
		diags.Append(d...)
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
		d = planDnsConfig.AllowedClients.ElementsAs(ctx, &dnsAccess.AllowedClients, false)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	}
	if len(planDnsConfig.DisallowedClients.Elements()) > 0 {
		d = planDnsConfig.DisallowedClients.ElementsAs(ctx, &dnsAccess.DisallowedClients, false)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	}
	if len(planDnsConfig.BlockedHosts.Elements()) > 0 {
		d = planDnsConfig.BlockedHosts.ElementsAs(ctx, &dnsAccess.BlockedHosts, false)
		diags.Append(d...)
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

	// DHCP
	// unpack nested attributes from plan
	var planDhcpConfig dhcpConfigModel
	d = plan.Dhcp.As(ctx, &planDhcpConfig, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	var planDhcpIpv4Settings dhcpIpv4Model
	if !planDhcpConfig.Ipv4Settings.IsNull() {
		d = planDhcpConfig.Ipv4Settings.As(ctx, &planDhcpIpv4Settings, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	}
	var planDhcpIpv6Settings dhcpIpv6Model
	if !planDhcpConfig.Ipv6Settings.IsNull() {
		d = planDhcpConfig.Ipv6Settings.As(ctx, &planDhcpIpv6Settings, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	}
	var planDhcpStaticLeases []dhcpStaticLeasesModel
	if !planDhcpConfig.StaticLeases.IsNull() {
		d = planDhcpConfig.StaticLeases.ElementsAs(ctx, &planDhcpStaticLeases, false)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	}

	// instantiate empty object for storing plan data
	var dhcpConfig adguard.DhcpConfig
	// populate dhcp config from plan
	dhcpConfig.Enabled = planDhcpConfig.Enabled.ValueBool()
	dhcpConfig.InterfaceName = planDhcpConfig.Interface.ValueString()
	dhcpConfig.V4.GatewayIp = planDhcpIpv4Settings.GatewayIp.ValueString()
	dhcpConfig.V4.SubnetMask = planDhcpIpv4Settings.SubnetMask.ValueString()
	dhcpConfig.V4.RangeStart = planDhcpIpv4Settings.RangeStart.ValueString()
	dhcpConfig.V4.RangeEnd = planDhcpIpv4Settings.RangeEnd.ValueString()
	dhcpConfig.V4.LeaseDuration = uint64(planDhcpIpv4Settings.LeaseDuration.ValueInt64())
	dhcpConfig.V6.RangeStart = planDhcpIpv6Settings.RangeStart.ValueString()
	dhcpConfig.V6.LeaseDuration = uint64(planDhcpIpv6Settings.LeaseDuration.ValueInt64())

	// set dhcp config using plan
	_, err = r.adg.SetDhcpConfig(dhcpConfig)
	if err != nil {
		diags.AddError(
			"Unable to Update AdGuard Home Config",
			err.Error(),
		)
		return
	}

	// initialize variables related to state
	var allStateDhcpStaticLeases []string
	var stateDhcpConfig dhcpConfigModel
	var stateDhcpStaticLeases []dhcpStaticLeasesModel

	// check if we had dhcp config previously in state
	if !state.Dhcp.IsNull() {
		// check if the entire dhcp server has been turned off
		if dhcpConfig.InterfaceName == "" {
			// it was, set dhcp config to defaults
			err = r.adg.ResetDhcpConfig()
			if err != nil {
				diags.AddError(
					"Unable to Update AdGuard Home Config",
					err.Error(),
				)
				return
			}
		}

		d = state.Dhcp.As(ctx, &stateDhcpConfig, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return
		}
		if !stateDhcpConfig.StaticLeases.IsNull() {
			d = stateDhcpConfig.StaticLeases.ElementsAs(ctx, &stateDhcpStaticLeases, false)
			diags.Append(d...)
			if diags.HasError() {
				return
			}
		}

		// go through the dhcp static leases existing in state
		for _, stateDhcpStaticLease := range stateDhcpStaticLeases {
			// track dhcp static leases in state
			allStateDhcpStaticLeases = append(
				allStateDhcpStaticLeases, fmt.Sprintf(
					"%s_%s_%s", stateDhcpStaticLease.Hostname.ValueString(), stateDhcpStaticLease.Mac.ValueString(), stateDhcpStaticLease.Ip.ValueString(),
				),
			)
		}
	}

	// initialize slice for all dhcp static leases in the plan
	var allPlanDhcpStaticLeases []string

	// go through all dhcp static leases in plan
	for _, planDhcpStaticLease := range planDhcpStaticLeases {
		// create a unique key for this static lease
		dhcpStaticLease_key := fmt.Sprintf("%s_%s_%s", planDhcpStaticLease.Hostname.ValueString(), planDhcpStaticLease.Mac.ValueString(), planDhcpStaticLease.Ip.ValueString())
		// track dhcp static leases in plan
		allPlanDhcpStaticLeases = append(allPlanDhcpStaticLeases, dhcpStaticLease_key)
	}

	// only work on state dhcp leases if the dhcp server is configured appropriately
	if dhcpConfig.InterfaceName != "" {
		// go through the dhcp static leases existing in state
		for _, stateDhcpStaticLease := range stateDhcpStaticLeases {
			// instantiate empty object for storing state data
			var dhcpStaticLease adguard.DhcpStaticLease
			dhcpStaticLease.Mac = stateDhcpStaticLease.Mac.ValueString()
			dhcpStaticLease.Ip = stateDhcpStaticLease.Ip.ValueString()
			dhcpStaticLease.Hostname = stateDhcpStaticLease.Hostname.ValueString()
			// create a unique key for this static lease
			dhcpStaticLease_key := fmt.Sprintf("%s_%s_%s", dhcpStaticLease.Hostname, dhcpStaticLease.Mac, dhcpStaticLease.Ip)

			// check if this dhcp static lease is still in the plan
			if !contains(allPlanDhcpStaticLeases, dhcpStaticLease_key) {
				// not in plan, delete it
				_, err = r.adg.ManageDhcpStaticLease(false, dhcpStaticLease)
				if err != nil {
					diags.AddError(
						"Unable to Update AdGuard Home Config",
						err.Error(),
					)
					return
				}
			}
		}
	}

	// go through all dhcp static leases in plan
	for _, planDhcpStaticLease := range planDhcpStaticLeases {
		// instantiate empty object for storing plan data
		var dhcpStaticLease adguard.DhcpStaticLease
		dhcpStaticLease.Mac = planDhcpStaticLease.Mac.ValueString()
		dhcpStaticLease.Ip = planDhcpStaticLease.Ip.ValueString()
		dhcpStaticLease.Hostname = planDhcpStaticLease.Hostname.ValueString()

		// check if this dhcp static lease isn't already in state
		if !contains(allStateDhcpStaticLeases, fmt.Sprintf("%s_%s_%s", dhcpStaticLease.Hostname, dhcpStaticLease.Mac, dhcpStaticLease.Ip)) {
			// set this dhcp static lease using plan
			_, err = r.adg.ManageDhcpStaticLease(true, dhcpStaticLease)
			if err != nil {
				diags.AddError(
					"Unable to Update AdGuard Home Config",
					err.Error(),
				)
				return
			}
		}
	}

	// TLS CONFIG
	// unpack nested attributes from plan
	var planTlsConfig tlsConfigModel
	d = plan.Tls.As(ctx, &planTlsConfig, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	// instantiate empty object for storing plan data
	var tlsConfig adguard.TlsConfig
	// populate tls config from plan
	tlsConfig.Enabled = planTlsConfig.Enabled.ValueBool()
	tlsConfig.ServerName = planTlsConfig.ServerName.ValueString()
	tlsConfig.ForceHttps = planTlsConfig.ForceHttps.ValueBool()
	tlsConfig.PortHttps = uint16(planTlsConfig.PortHttps.ValueInt64())
	tlsConfig.PortDnsOverTls = uint16(planTlsConfig.PortDnsOverTls.ValueInt64())
	tlsConfig.PortDnsOverQuic = uint16(planTlsConfig.PortDnsOverQuic.ValueInt64())
	tlsConfig.ServePlainDns = planTlsConfig.ServePlainDns.ValueBool()

	// regex to match a file path
	var filePathIdentifier = regexp.MustCompile(`^/\w|\w:`)

	// check what is the certificate chain
	if len(planTlsConfig.CertificateChain.ValueString()) > 0 && filePathIdentifier.MatchString(planTlsConfig.CertificateChain.ValueString()[0:2]) {
		// it's a file path
		tlsConfig.CertificatePath = planTlsConfig.CertificateChain.ValueString()
	} else {
		// it's the base64 PEM file
		tlsConfig.CertificateChain = planTlsConfig.CertificateChain.ValueString()
	}

	// check what is the private key
	if len(planTlsConfig.PrivateKey.ValueString()) > 0 && filePathIdentifier.MatchString(planTlsConfig.PrivateKey.ValueString()[0:2]) {
		// it's a file path
		tlsConfig.PrivateKeyPath = planTlsConfig.PrivateKey.ValueString()
	} else {
		// it's the base64 PEM file
		tlsConfig.PrivateKey = planTlsConfig.PrivateKey.ValueString()
	}

	// set tls config using plan
	tlsConfigResponse, err := r.adg.SetTlsConfig(tlsConfig)
	if err != nil {
		diags.AddError(
			"Unable to Update AdGuard Home Config",
			err.Error(),
		)
		return
	}

	// populate computed attributes
	planTlsConfig.PrivateKeySaved = types.BoolValue(tlsConfigResponse.PrivateKeySaved)
	planTlsConfig.ValidCert = types.BoolValue(tlsConfigResponse.ValidCert)
	planTlsConfig.ValidChain = types.BoolValue(tlsConfigResponse.ValidChain)
	planTlsConfig.ValidKey = types.BoolValue(tlsConfigResponse.ValidKey)
	planTlsConfig.ValidPair = types.BoolValue(tlsConfigResponse.ValidPair)
	planTlsConfig.KeyType = types.StringValue(tlsConfigResponse.KeyType)
	planTlsConfig.Subject = types.StringValue(tlsConfigResponse.Subject)
	planTlsConfig.Issuer = types.StringValue(tlsConfigResponse.Issuer)
	planTlsConfig.NotBefore = types.StringValue(tlsConfigResponse.NotBefore)
	planTlsConfig.NotAfter = types.StringValue(tlsConfigResponse.NotAfter)
	planTlsConfig.DnsNames, d = types.ListValueFrom(ctx, types.StringType, tlsConfig.DnsNames)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	planTlsConfig.WarningValidation = types.StringValue(tlsConfigResponse.WarningValidation)

	// overwrite plan with computed values
	plan.Tls, d = types.ObjectValueFrom(ctx, tlsConfigModel{}.attrTypes(), &planTlsConfig)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	// if we got here, all went fine
}
