package adguard

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// validateBlockedServices takes a BlockedServices SetValue from a plan and confirms all entries are accepted by AdGuard Home
func validateBlockedServices(ctx context.Context, blockedServices basetypes.SetValue, resp *resource.ModifyPlanResponse) {
	// only proceed if there are blocked services to deal with in the plan
	if blockedServices.IsNull() || blockedServices.IsUnknown() {
		return
	}

	// convert the blocked services in the plan to a list
	var planBlockedServices []string
	diags := blockedServices.ElementsAs(ctx, &planBlockedServices, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// retrieve all blocked services from cache
	allBlockedServices := getFromCache("blocked_services")
	if len(allBlockedServices) == 0 {
		resp.Diagnostics.AddError(
			"Error Fetching Valid Values for Blocked Services",
			"Could not fetch valid values from AdGuard Home",
		)
		return
	}

	// go through the entries in the plan and validate them
	for _, v := range planBlockedServices {
		if !contains(allBlockedServices, v) {
			resp.Diagnostics.AddAttributeError(
				path.Root("blocked_services"),
				"Invalid Attribute Value Match",
				fmt.Sprintf("Attribute `blocked_services` with value '%s' is not valid. Valid values are: %v", v, allBlockedServices),
			)
			return
		}
	}
}
