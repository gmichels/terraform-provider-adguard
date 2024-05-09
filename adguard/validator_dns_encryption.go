package adguard

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// validator confirms the DNS Config local PTR upstreams are set when the private PTR resolvers are enabled
var _ validator.Bool = checkDnsEncryptionValidator{}

type checkDnsEncryptionValidator struct {
}

func (v checkDnsEncryptionValidator) Description(_ context.Context) string {
	return "\"tls.serve_plain_dns\" must be `true` when \"tls.enabled\" is set to `false`"
}

func (v checkDnsEncryptionValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v checkDnsEncryptionValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	if req.ConfigValue.ValueBool() || req.ConfigValue.IsNull() {
		// if set to true or null, config is valid
		return
	}

	tlsEnabledPath := req.Path.ParentPath().AtName("enabled")

	var tlsEnabled types.Bool

	diags := req.Config.GetAttribute(ctx, tlsEnabledPath, &tlsEnabled)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !tlsEnabled.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			tlsEnabledPath,
			"DNS Encryption Config Invalid",
			v.Description(ctx),
		)
	}
}

func checkDnsEncryption() validator.Bool {
	return checkDnsEncryptionValidator{}
}
