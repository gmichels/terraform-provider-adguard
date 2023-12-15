package adguard

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/gmichels/adguard-client-go"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

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
	for _, service := range SAFE_SEARCH_SERVICES_OPTIONS {
		services = append(services, types.StringValue(service))
	}

	return map[string]attr.Value{
		"enabled":  types.BoolValue(SAFE_SEARCH_ENABLED),
		"services": types.SetValueMust(types.StringType, services),
	}
}

// provides safe search schema for datasources
func safeSearchDatasourceSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
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
	}
}

// provides safe search schema for resources
func safeSearchResourceSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Computed: true,
		Optional: true,
		Default: objectdefault.StaticValue(types.ObjectValueMust(
			safeSearchModel{}.attrTypes(), safeSearchModel{}.defaultObject()),
		),
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: fmt.Sprintf("Whether Safe Search is enabled. Defaults to `%t`", SAFE_SEARCH_ENABLED),
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(SAFE_SEARCH_ENABLED),
			},
			"services": schema.SetAttribute{
				Description: "Services which SafeSearch is enabled.",
				ElementType: types.StringType,
				Computed:    true,
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(SAFE_SEARCH_SERVICES_OPTIONS...),
					),
				},
				Default: setdefault.StaticValue(
					types.SetValueMust(types.StringType, convertToAttr(SAFE_SEARCH_SERVICES_OPTIONS)),
				),
			},
		},
	}
}

// mapSafeSearchConfigFields - will return the list of safe search services that are enabled
func mapSafeSearchServices(adgSafeSearchConfig *adguard.SafeSearchConfig) []string {
	// perform reflection of safe search object
	v := reflect.ValueOf(adgSafeSearchConfig).Elem()
	// grab the type of the reflected object
	t := v.Type()

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

// setSafeSearchServices - based on a list of enabled safe search services, will set the safeSearchConfig fields appropriately
func setSafeSearchServices(v reflect.Value, t reflect.Type, services []string) {
	for i := 0; i < v.NumField(); i++ {
		fieldName := strings.ToLower(t.Field(i).Name)
		if contains(services, fieldName) {
			v.Field(i).Set(reflect.ValueOf(true))
		}
	}
}

// scheduleModel maps schedule configuration schema data
type scheduleModel struct {
	TimeZone  types.String `tfsdk:"time_zone"`
	Sunday    types.Object `tfsdk:"sun"`
	Monday    types.Object `tfsdk:"mon"`
	Tuesday   types.Object `tfsdk:"tue"`
	Wednesday types.Object `tfsdk:"wed"`
	Thursday  types.Object `tfsdk:"thu"`
	Friday    types.Object `tfsdk:"fri"`
	Saturday  types.Object `tfsdk:"sat"`
}

// attrTypes - return attribute types for this model
func (o scheduleModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"time_zone": types.StringType,
		"sun":       types.ObjectType{AttrTypes: dayRangeModel{}.attrTypes()},
		"mon":       types.ObjectType{AttrTypes: dayRangeModel{}.attrTypes()},
		"tue":       types.ObjectType{AttrTypes: dayRangeModel{}.attrTypes()},
		"wed":       types.ObjectType{AttrTypes: dayRangeModel{}.attrTypes()},
		"thu":       types.ObjectType{AttrTypes: dayRangeModel{}.attrTypes()},
		"fri":       types.ObjectType{AttrTypes: dayRangeModel{}.attrTypes()},
		"sat":       types.ObjectType{AttrTypes: dayRangeModel{}.attrTypes()},
	}
}

// defaultObject - return default object for this model
func (o scheduleModel) defaultObject() map[string]attr.Value {
	return map[string]attr.Value{
		"time_zone": types.StringNull(),
		"sun":       types.ObjectValueMust(dayRangeModel{}.attrTypes(), dayRangeModel{}.defaultObject()),
		"mon":       types.ObjectValueMust(dayRangeModel{}.attrTypes(), dayRangeModel{}.defaultObject()),
		"tue":       types.ObjectValueMust(dayRangeModel{}.attrTypes(), dayRangeModel{}.defaultObject()),
		"wed":       types.ObjectValueMust(dayRangeModel{}.attrTypes(), dayRangeModel{}.defaultObject()),
		"thu":       types.ObjectValueMust(dayRangeModel{}.attrTypes(), dayRangeModel{}.defaultObject()),
		"fri":       types.ObjectValueMust(dayRangeModel{}.attrTypes(), dayRangeModel{}.defaultObject()),
		"sat":       types.ObjectValueMust(dayRangeModel{}.attrTypes(), dayRangeModel{}.defaultObject()),
	}
}

// dayRangeModel maps day ranges to schema data
type dayRangeModel struct {
	Start types.String `tfsdk:"start"`
	End   types.String `tfsdk:"end"`
}

// attrTypes - return attribute types for this model
func (o dayRangeModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"start": types.StringType,
		"end":   types.StringType,
	}
}

// defaultObject - return default object for this model
func (o dayRangeModel) defaultObject() map[string]attr.Value {
	return map[string]attr.Value{
		"start": types.StringNull(),
		"end":   types.StringNull(),
	}
}

// provides schedule schema for datasources
func scheduleDatasourceSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Sets periods of inactivity for filtering blocked services. The schedule contains 7 days (Sunday to Saturday) and a time zone.",
		Computed:    true,
		Attributes: map[string]schema.Attribute{
			"time_zone": schema.StringAttribute{
				Description: "Time zone name according to IANA time zone database. For example `America/New_York`. `Local` represents the system's local time zone.",
				Computed:    true,
			},
			"sun": dayRangeDatasourceSchema("Sunday"),
			"mon": dayRangeDatasourceSchema("Monday"),
			"tue": dayRangeDatasourceSchema("Tueday"),
			"wed": dayRangeDatasourceSchema("Wednesday"),
			"thu": dayRangeDatasourceSchema("Thursday"),
			"fri": dayRangeDatasourceSchema("Friday"),
			"sat": dayRangeDatasourceSchema("Saturday"),
		},
	}
}

// provides schedule schema for resources
func scheduleResourceSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Sets periods of inactivity for filtering blocked services. The schedule contains 7 days (Sunday to Saturday) and a time zone.",
		Computed:    true,
		Optional:    true,
		Default: objectdefault.StaticValue(types.ObjectValueMust(
			scheduleModel{}.attrTypes(), scheduleModel{}.defaultObject()),
		),
		Attributes: map[string]schema.Attribute{
			"time_zone": schema.StringAttribute{
				Description: "Time zone name according to IANA time zone database. For example `America/New_York`. `Local` represents the system's local time zone.",
				Optional:    true,
				Validators: []validator.String{AlsoRequiresNOf(1,
					path.MatchRelative().AtParent().AtName("sun"),
					path.MatchRelative().AtParent().AtName("mon"),
					path.MatchRelative().AtParent().AtName("tue"),
					path.MatchRelative().AtParent().AtName("wed"),
					path.MatchRelative().AtParent().AtName("thu"),
					path.MatchRelative().AtParent().AtName("fri"),
					path.MatchRelative().AtParent().AtName("sat"),
				)},
			},
			"sun": dayRangeResourceSchema("Sunday"),
			"mon": dayRangeResourceSchema("Monday"),
			"tue": dayRangeResourceSchema("Tueday"),
			"wed": dayRangeResourceSchema("Wednesday"),
			"thu": dayRangeResourceSchema("Thursday"),
			"fri": dayRangeResourceSchema("Friday"),
			"sat": dayRangeResourceSchema("Saturday"),
		},
	}
}

// provides day range schema for datasources
func dayRangeDatasourceSchema(day string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: fmt.Sprintf("Paused service blocking interval for `%s`", day),
		Computed:    true,
		Attributes: map[string]schema.Attribute{
			"start": schema.StringAttribute{
				Description: "Start of paused service blocking schedule, in HH:MM format",
				Computed:    true,
			},
			"end": schema.StringAttribute{
				Description: "End of paused service blocking schedule, in HH:MM format",
				Computed:    true,
			},
		},
	}
}

// provides day range schema for resources
func dayRangeResourceSchema(day string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: fmt.Sprintf("Paused service blocking interval for `%s`", day),
		Computed:    true,
		Optional:    true,
		Default: objectdefault.StaticValue(types.ObjectValueMust(
			dayRangeModel{}.attrTypes(), dayRangeModel{}.defaultObject()),
		),
		Attributes: map[string]schema.Attribute{
			"start": schema.StringAttribute{
				Description: "Start of paused service blocking schedule, in HH:MM format",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.All(
						stringvalidator.RegexMatches(
							regexp.MustCompile(`^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$`),
							"must be in HH:MM format",
						),
						stringvalidator.AlsoRequires(path.Expressions{
							path.MatchRelative().AtParent().AtName("end"),
						}...),
					),
				},
			},
			"end": schema.StringAttribute{
				Description: "End of paused service blocking schedule, in HH:MM format",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.All(
						stringvalidator.RegexMatches(
							regexp.MustCompile(`^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$`),
							"must be in HH:MM format",
						),
						stringvalidator.AlsoRequires(path.Expressions{
							path.MatchRelative().AtParent().AtName("start"),
						}...),
					),
				},
			},
		},
	}
}

// mapAdgScheduleToBlockedServicesPauseSchedule takes an ADG Schedule object and maps the the days into a scheduleModel
func mapAdgScheduleToBlockedServicesPauseSchedule(ctx context.Context, adgBlockedServicesSchedule *adguard.Schedule) scheduleModel {
	// instantiate empty intermediate object
	var blockedServicesPauseSchedule scheduleModel

	// go over each day and map to intermediate object
	var sunDayRangeConfig dayRangeModel
	if adgBlockedServicesSchedule.Sunday.End > 0 {
		sunDayRangeConfig.Start = types.StringValue(convertMsToHourMinutes(int64(adgBlockedServicesSchedule.Sunday.Start)))
		sunDayRangeConfig.End = types.StringValue(convertMsToHourMinutes(int64(adgBlockedServicesSchedule.Sunday.End)))
	}
	blockedServicesPauseSchedule.Sunday, _ = types.ObjectValueFrom(ctx, dayRangeModel{}.attrTypes(), &sunDayRangeConfig)

	var monDayRangeConfig dayRangeModel
	if adgBlockedServicesSchedule.Monday.End > 0 {
		monDayRangeConfig.Start = types.StringValue(convertMsToHourMinutes(int64(adgBlockedServicesSchedule.Monday.Start)))
		monDayRangeConfig.End = types.StringValue(convertMsToHourMinutes(int64(adgBlockedServicesSchedule.Monday.End)))
	}
	blockedServicesPauseSchedule.Monday, _ = types.ObjectValueFrom(ctx, dayRangeModel{}.attrTypes(), &monDayRangeConfig)

	var tueDayRangeConfig dayRangeModel
	if adgBlockedServicesSchedule.Tuesday.End > 0 {
		tueDayRangeConfig.Start = types.StringValue(convertMsToHourMinutes(int64(adgBlockedServicesSchedule.Tuesday.Start)))
		tueDayRangeConfig.End = types.StringValue(convertMsToHourMinutes(int64(adgBlockedServicesSchedule.Tuesday.End)))
	}
	blockedServicesPauseSchedule.Tuesday, _ = types.ObjectValueFrom(ctx, dayRangeModel{}.attrTypes(), &tueDayRangeConfig)

	var wedDayRangeConfig dayRangeModel
	if adgBlockedServicesSchedule.Wednesday.End > 0 {
		wedDayRangeConfig.Start = types.StringValue(convertMsToHourMinutes(int64(adgBlockedServicesSchedule.Wednesday.Start)))
		wedDayRangeConfig.End = types.StringValue(convertMsToHourMinutes(int64(adgBlockedServicesSchedule.Wednesday.End)))
	}
	blockedServicesPauseSchedule.Wednesday, _ = types.ObjectValueFrom(ctx, dayRangeModel{}.attrTypes(), &wedDayRangeConfig)

	var thuDayRangeConfig dayRangeModel
	if adgBlockedServicesSchedule.Thursday.End > 0 {
		thuDayRangeConfig.Start = types.StringValue(convertMsToHourMinutes(int64(adgBlockedServicesSchedule.Thursday.Start)))
		thuDayRangeConfig.End = types.StringValue(convertMsToHourMinutes(int64(adgBlockedServicesSchedule.Thursday.End)))
	}
	blockedServicesPauseSchedule.Thursday, _ = types.ObjectValueFrom(ctx, dayRangeModel{}.attrTypes(), &thuDayRangeConfig)

	var friDayRangeConfig dayRangeModel
	if adgBlockedServicesSchedule.Friday.End > 0 {
		friDayRangeConfig.Start = types.StringValue(convertMsToHourMinutes(int64(adgBlockedServicesSchedule.Friday.Start)))
		friDayRangeConfig.End = types.StringValue(convertMsToHourMinutes(int64(adgBlockedServicesSchedule.Friday.End)))
	}
	blockedServicesPauseSchedule.Friday, _ = types.ObjectValueFrom(ctx, dayRangeModel{}.attrTypes(), &friDayRangeConfig)

	var satDayRangeConfig dayRangeModel
	if adgBlockedServicesSchedule.Saturday.End > 0 {
		satDayRangeConfig.Start = types.StringValue(convertMsToHourMinutes(int64(adgBlockedServicesSchedule.Saturday.Start)))
		satDayRangeConfig.End = types.StringValue(convertMsToHourMinutes(int64(adgBlockedServicesSchedule.Saturday.End)))
	}
	blockedServicesPauseSchedule.Saturday, _ = types.ObjectValueFrom(ctx, dayRangeModel{}.attrTypes(), &satDayRangeConfig)

	return blockedServicesPauseSchedule
}

// mapBlockedServicesPauseScheduleToAdgSchedule takes a scheduleModel from plan and maps into an ADG Schedule object
func mapBlockedServicesPauseScheduleToAdgSchedule(ctx context.Context, schedule scheduleModel) adguard.Schedule {
	// unpack nested attributes for each day from plan
	var planSunDayRangeConfig dayRangeModel
	_ = schedule.Sunday.As(ctx, &planSunDayRangeConfig, basetypes.ObjectAsOptions{})

	var planMonDayRangeConfig dayRangeModel
	_ = schedule.Monday.As(ctx, &planMonDayRangeConfig, basetypes.ObjectAsOptions{})

	var planTueDayRangeConfig dayRangeModel
	_ = schedule.Tuesday.As(ctx, &planTueDayRangeConfig, basetypes.ObjectAsOptions{})

	var planWedDayRangeConfig dayRangeModel
	_ = schedule.Wednesday.As(ctx, &planWedDayRangeConfig, basetypes.ObjectAsOptions{})

	var planThuDayRangeConfig dayRangeModel
	_ = schedule.Thursday.As(ctx, &planThuDayRangeConfig, basetypes.ObjectAsOptions{})

	var planFriDayRangeConfig dayRangeModel
	_ = schedule.Friday.As(ctx, &planFriDayRangeConfig, basetypes.ObjectAsOptions{})

	var planSatDayRangeConfig dayRangeModel
	_ = schedule.Saturday.As(ctx, &planSatDayRangeConfig, basetypes.ObjectAsOptions{})

	// instantiate empty object for storing plan data
	var blockedServicesSchedule adguard.Schedule
	// populate blocked services schedule from plan
	blockedServicesSchedule.TimeZone = schedule.TimeZone.ValueString()
	blockedServicesSchedule.Sunday.Start = uint(convertHoursMinutesToMs(planSunDayRangeConfig.Start.ValueString()))
	blockedServicesSchedule.Sunday.End = uint(convertHoursMinutesToMs(planSunDayRangeConfig.End.ValueString()))
	blockedServicesSchedule.Monday.Start = uint(convertHoursMinutesToMs(planMonDayRangeConfig.Start.ValueString()))
	blockedServicesSchedule.Monday.End = uint(convertHoursMinutesToMs(planMonDayRangeConfig.End.ValueString()))
	blockedServicesSchedule.Tuesday.Start = uint(convertHoursMinutesToMs(planTueDayRangeConfig.Start.ValueString()))
	blockedServicesSchedule.Tuesday.End = uint(convertHoursMinutesToMs(planTueDayRangeConfig.End.ValueString()))
	blockedServicesSchedule.Wednesday.Start = uint(convertHoursMinutesToMs(planWedDayRangeConfig.Start.ValueString()))
	blockedServicesSchedule.Wednesday.End = uint(convertHoursMinutesToMs(planWedDayRangeConfig.End.ValueString()))
	blockedServicesSchedule.Thursday.Start = uint(convertHoursMinutesToMs(planThuDayRangeConfig.Start.ValueString()))
	blockedServicesSchedule.Thursday.End = uint(convertHoursMinutesToMs(planThuDayRangeConfig.End.ValueString()))
	blockedServicesSchedule.Friday.Start = uint(convertHoursMinutesToMs(planFriDayRangeConfig.Start.ValueString()))
	blockedServicesSchedule.Friday.End = uint(convertHoursMinutesToMs(planFriDayRangeConfig.End.ValueString()))
	blockedServicesSchedule.Saturday.Start = uint(convertHoursMinutesToMs(planSatDayRangeConfig.Start.ValueString()))
	blockedServicesSchedule.Saturday.End = uint(convertHoursMinutesToMs(planSatDayRangeConfig.End.ValueString()))

	return blockedServicesSchedule
}
