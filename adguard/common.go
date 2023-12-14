package adguard

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
