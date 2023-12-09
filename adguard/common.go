package adguard

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// dayRangeModel maps day ranges to schema data
type dayRangeModel struct {
	Start types.Int64 `tfsdk:"start"`
	End   types.Int64 `tfsdk:"end"`
}

// attrTypes - return attribute types for this model
func (o dayRangeModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"start":  types.Int64Type,
		"end": types.Int64Type,
	}
}

// defaultObject - return default object for this model
func (o dayRangeModel) defaultObject() map[string]attr.Value {
	return map[string]attr.Value{
		"start":       types.Int64Value(BLOCKED_SERVICES_SCHEDULE_START_END),
		"end":       types.Int64Value(BLOCKED_SERVICES_SCHEDULE_START_END),
	}
}

// provides day range schema for datasources
func dayRangeDatasourceSchema(day string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: fmt.Sprintf("Schedule interval for `%s`", day),
		Computed:    true,
		Attributes: map[string]schema.Attribute{
			"start": schema.Int64Attribute{
				Description: "Number of minutes from start of day",
				Computed:    true,
			},
			"end": schema.Int64Attribute{
				Description: "Number of minutes from start of day",
				Computed:    true,
			},
		},
	}
}

// provides day range schema for resources
func dayRangeResourceSchema(day string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: fmt.Sprintf("Schedule interval for `%s`", day),
		Computed: true,
		Optional: true,
		Default: objectdefault.StaticValue(types.ObjectValueMust(
			dayRangeModel{}.attrTypes(), dayRangeModel{}.defaultObject()),
		),
		Attributes: map[string]schema.Attribute{
			"start": schema.Int64Attribute{
				Description: "Number of minutes from start of day",
				Computed:    true,
				Optional:    true,
				Default:     int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.Between(0, 1439),
				},
			},
			"end": schema.Int64Attribute{
				Description: "Number of minutes from start of day",
				Computed:    true,
				Optional:    true,
				Default:     int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.Between(0, 1440),
				},
			},
		},
	}
}
