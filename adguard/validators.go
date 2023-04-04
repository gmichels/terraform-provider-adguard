package adguard

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Blocking Mode validator confirms the DNS Config blocking mode is correctly configured
var _ validator.String = checkBlockingModeValidator{}

type checkBlockingModeValidator struct {
	mode string
}

func (v checkBlockingModeValidator) Description(_ context.Context) string {
	return fmt.Sprintf("blocking_mode must be set to %s when specifying blocking_ipv4 and blocking_ipv6", v.mode)
}

func (v checkBlockingModeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v checkBlockingModeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		// If code block does not exist, config is valid.
		return
	}

	modePath := req.Path.ParentPath().AtName("blocking_mode")

	var m types.String

	diags := req.Config.GetAttribute(ctx, modePath, &m)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if m.IsNull() || m.IsUnknown() {
		// Only validate if mode value is known.
		return
	}

	if m.ValueString() != v.mode {
		resp.Diagnostics.AddAttributeError(
			modePath,
			"DNS Config Blocking Mode Value Invalid",
			v.Description(ctx),
		)
	}
}

func checkBlockingMode(m string) validator.String {
	return checkBlockingModeValidator{
		mode: m,
	}
}
