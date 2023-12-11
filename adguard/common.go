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
