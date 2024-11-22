package adguard

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// validateSafeSearchServices takes a safeSearch object from a plan and confirms all service entries are accepted by AdGuard Home
func validateSafeSearchServices(ctx context.Context, safeSearch basetypes.ObjectValue, resp *resource.ModifyPlanResponse) safeSearchModel {
	// initialize output
	var planSafeSearch safeSearchModel

	// retrieve all safe search services from cache
	allSafeSearchServices := getFromCache("safesearch")
	if len(allSafeSearchServices) == 0 {
		resp.Diagnostics.AddError(
			"Error Fetching Valid Values for Safe Search Config",
			"Could not fetch valid values from AdGuard Home",
		)
		return planSafeSearch
	}

	// unpack nested attributes from plan
	diags := safeSearch.As(ctx, &planSafeSearch, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return planSafeSearch
	}

	// check if any services were provided
	if planSafeSearch.Services.IsNull() || planSafeSearch.Services.IsUnknown() {
		// nothing was provided, set default in plan
		planSafeSearch.Services, diags = types.SetValueFrom(ctx, types.StringType, allSafeSearchServices)
		diags.Append(diags...)
		if diags.HasError() {
			return planSafeSearch
		}
	} else {
		// some services were provided, validate them
		// convert the safe search services in the plan to a list
		var planSafeSearchServices []string
		diags = planSafeSearch.Services.ElementsAs(ctx, &planSafeSearchServices, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return planSafeSearch
		}

		// go through the entries in the plan and validate them
		for _, v := range planSafeSearchServices {
			if !contains(allSafeSearchServices, v) {
				resp.Diagnostics.AddAttributeError(
					path.Root("safesearch"),
					"Invalid Attribute Value Match",
					fmt.Sprintf("Attribute `safesearch.services` with value '%s' is not valid. Valid values are: %v", v, allSafeSearchServices),
				)
				return planSafeSearch
			}
		}
	}

	return planSafeSearch
}
