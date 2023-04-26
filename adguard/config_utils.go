package adguard

import (
	"context"
	"reflect"
	"strings"

	"github.com/gmichels/adguard-client-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var DEFAULT_SAFESEARCH_SERVICES = []string{"bing", "duckduckgo", "google", "pixabay", "yandex", "youtube"}
var BLOCKED_SERVICES_ALL = []string{"9gag", "amazon", "bilibili", "cloudflare", "crunchyroll", "dailymotion", "deezer",
	"discord", "disneyplus", "douban", "ebay", "epic_games", "facebook", "gog", "hbomax", "hulu", "icloud_private_relay", "imgur",
	"instagram", "iqiyi", "kakaotalk", "lazada", "leagueoflegends", "line", "mail_ru", "mastodon", "minecraft", "netflix", "ok",
	"onlyfans", "origin", "pinterest", "playstation", "qq", "rakuten_viki", "reddit", "riot_games", "roblox", "shopee", "skype", "snapchat",
	"soundcloud", "spotify", "steam", "telegram", "tiktok", "tinder", "twitch", "twitter", "valorant", "viber", "vimeo", "vk", "voot", "wechat",
	"weibo", "whatsapp", "xboxlive", "youtube", "zhihu"}

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
		"ignored":             basetypes.NewSetNull(types.StringType),
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
		"ignored":  basetypes.NewSetNull(types.StringType),
	}
}

func (r *configResource) CreateOrUpdateConfigResource(ctx context.Context, plan configResourceModel) (configResourceModel, error) {
	// unpack nested attributes from plan
	var planFiltering filteringModel
	var planSafeBrowsing enabledModel
	var planParentalControl enabledModel
	var planSafeSearch safeSearchModel
	var planQueryLogConfig queryLogConfigModel
	var planStatsConfig statsConfigModel
	_ = plan.Filtering.As(ctx, &planFiltering, basetypes.ObjectAsOptions{})
	_ = plan.SafeBrowsing.As(ctx, &planSafeBrowsing, basetypes.ObjectAsOptions{})
	_ = plan.ParentalControl.As(ctx, &planParentalControl, basetypes.ObjectAsOptions{})
	_ = plan.SafeSearch.As(ctx, &planSafeSearch, basetypes.ObjectAsOptions{})
	_ = plan.QueryLog.As(ctx, &planQueryLogConfig, basetypes.ObjectAsOptions{})
	_ = plan.Stats.As(ctx, &planStatsConfig, basetypes.ObjectAsOptions{})

	// instantiate empty object for storing plan data
	var filteringConfig adguard.FilterConfig
	// populate filtering config from plan
	filteringConfig.Enabled = planFiltering.Enabled.ValueBool()
	filteringConfig.Interval = uint(planFiltering.UpdateInterval.ValueInt64())

	// set filtering config using plan
	_, err := r.adg.ConfigureFiltering(filteringConfig)
	if err != nil {
		return plan, err
	}

	// set safe browsing status using plan
	err = r.adg.SetSafeBrowsingStatus(planSafeBrowsing.Enabled.ValueBool())
	if err != nil {
		return plan, err
	}

	// set parental control status using plan
	err = r.adg.SetParentalStatus(planParentalControl.Enabled.ValueBool())
	if err != nil {
		return plan, err
	}

	// instantiate empty object for storing plan data
	var safeSearchConfig adguard.SafeSearchConfig
	// populate safe search config using plan
	safeSearchConfig.Enabled = planSafeSearch.Enabled.ValueBool()
	if len(planSafeSearch.Services.Elements()) > 0 {
		var safeSearchServicesEnabled []string
		_ = planSafeSearch.Services.ElementsAs(ctx, &safeSearchServicesEnabled, false)

		// use reflection to set each safeSearchConfig service value dynamically
		v := reflect.ValueOf(&safeSearchConfig).Elem()
		t := v.Type()
		setSafeSearchConfigServices(v, t, safeSearchServicesEnabled)
	}

	// set safe search config using plan
	_, err = r.adg.SetSafeSearchConfig(safeSearchConfig)
	if err != nil {
		return plan, err
	}

	// instantiate empty object for storing plan data
	var queryLogConfig adguard.GetQueryLogConfigResponse
	// populate query log config from plan
	queryLogConfig.Enabled = planQueryLogConfig.Enabled.ValueBool()
	queryLogConfig.Interval = uint64(planQueryLogConfig.Interval.ValueInt64() * 3600 * 1000)
	queryLogConfig.AnonymizeClientIp = planQueryLogConfig.AnonymizeClientIp.ValueBool()

	if len(planQueryLogConfig.Ignored.Elements()) > 0 {
		_ = planQueryLogConfig.Ignored.ElementsAs(ctx, &queryLogConfig.Ignored, false)
	}

	// set query log config using plan
	_, err = r.adg.SetQueryLogConfig(queryLogConfig)
	if err != nil {
		return plan, err
	}

	// instantiate empty object for storing plan data
	var statsConfig adguard.GetStatsConfigResponse
	// populate stats from plan
	statsConfig.Enabled = planStatsConfig.Enabled.ValueBool()
	statsConfig.Interval = uint64(planStatsConfig.Interval.ValueInt64() * 3600 * 1000)

	if len(planStatsConfig.Ignored.Elements()) > 0 {
		_ = planStatsConfig.Ignored.ElementsAs(ctx, &statsConfig.Ignored, false)
	}

	// set stats config using plan
	_, err = r.adg.SetStatsConfig(statsConfig)
	if err != nil {
		return plan, err
	}

	// instantiate empty object for storing plan data
	var blockedServices []string
	// populate blocked services from plan
	if len(plan.BlockedServices.Elements()) > 0 {
		_ = plan.BlockedServices.ElementsAs(ctx, &blockedServices, false)
	}

	// set blocked services using plan
	_, err = r.adg.SetBlockedServices(blockedServices)
	if err != nil {
		return plan, err
	}

	// return the modified plan
	return plan, nil
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
