package adguard

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// validator confirms the DNS Config local PTR upstreams are set when the private PTR resolvers are enabled
var _ validator.Bool = checkLocalPtrUpstreamsValidator{}

type checkLocalPtrUpstreamsValidator struct {
}

func (v checkLocalPtrUpstreamsValidator) Description(_ context.Context) string {
	return "\"dns.local_ptr_upstreams\" must contain values when \"dns.use_private_ptr_resolvers\" is set to `true`"
}

func (v checkLocalPtrUpstreamsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v checkLocalPtrUpstreamsValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	if !req.ConfigValue.ValueBool() || req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		// if set to false or not set, config is valid
		return
	}

	ptrUpstreamsPath := req.Path.ParentPath().AtName("local_ptr_upstreams")

	var ptrUpstreams types.Set

	diags := req.Config.GetAttribute(ctx, ptrUpstreamsPath, &ptrUpstreams)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if ptrUpstreams.IsNull() || ptrUpstreams.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			ptrUpstreamsPath,
			"DNS Config Use Private PTR Resolvers Invalid",
			v.Description(ctx),
		)
	}
}

func checkLocalPtrUpstreams() validator.Bool {
	return checkLocalPtrUpstreamsValidator{}
}
