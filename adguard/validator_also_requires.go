package adguard

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// AllTypesValidator implements all validation interfaces offered by
// github.com/hashicorp/terraform-plugin-framework/schema/validator
type AllTypesValidator interface {
	validator.Bool
	validator.Float64
	validator.Int64
	validator.List
	validator.Map
	validator.Number
	validator.Object
	validator.Set
	validator.String
}

// This type of validator can be used with all types from
// github.com/hashicorp/terraform-plugin-framework/types
var _ AllTypesValidator = AlsoRequiresNOfValidator{}

// AlsoRequiresNOfValidator is the underlying struct implementing AlsoRequiresNOf.
type AlsoRequiresNOfValidator struct {
	N               int
	PathExpressions path.Expressions
}

type AlsoRequiresNOfValidatorRequest struct {
	Config         tfsdk.Config
	ConfigValue    attr.Value
	N              int
	Path           path.Path
	PathExpression path.Expression
}

type AlsoRequiresNOfValidatorResponse struct {
	Diagnostics diag.Diagnostics
}

func (o AlsoRequiresNOfValidator) Description(ctx context.Context) string {
	return o.MarkdownDescription(ctx)
}

func (o AlsoRequiresNOfValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Ensure that at least %d attribute(s) from this collection is set: %s", o.N, o.PathExpressions)
}

func (o AlsoRequiresNOfValidator) Validate(ctx context.Context, req AlsoRequiresNOfValidatorRequest, resp *AlsoRequiresNOfValidatorResponse) {
	expressions := req.PathExpression.MergeExpressions(o.PathExpressions...)

	found := 0
	for _, expression := range expressions {
		matchedPaths, diags := req.Config.PathMatches(ctx, expression)

		resp.Diagnostics.Append(diags...)

		// Collect all errors
		if diags.HasError() {
			continue
		}

		for _, mp := range matchedPaths {
			if mp.Equal(req.Path) {
				// ignore the attribute being validated when
				// counting other required attributes
				continue
			}

			var mpVal attr.Value
			diags = req.Config.GetAttribute(ctx, mp, &mpVal)
			resp.Diagnostics.Append(diags...)

			// Collect all errors
			if diags.HasError() {
				continue
			}

			// Delay validation until all involved attribute have a known value
			if mpVal.IsUnknown() {
				return
			}

			if !mpVal.IsNull() {
				found++
				if found == req.N {
					return
				}
			}
		}
	}

	resp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
		req.Path,
		fmt.Sprintf("At least %d attributes out of %s must be set",
			req.N, expressions),
	))
}

func (o AlsoRequiresNOfValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	validateReq := AlsoRequiresNOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		N:              o.N,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AlsoRequiresNOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o AlsoRequiresNOfValidator) ValidateFloat64(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	validateReq := AlsoRequiresNOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		N:              o.N,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AlsoRequiresNOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o AlsoRequiresNOfValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	validateReq := AlsoRequiresNOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		N:              o.N,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AlsoRequiresNOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o AlsoRequiresNOfValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	validateReq := AlsoRequiresNOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		N:              o.N,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AlsoRequiresNOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o AlsoRequiresNOfValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	validateReq := AlsoRequiresNOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		N:              o.N,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AlsoRequiresNOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o AlsoRequiresNOfValidator) ValidateNumber(ctx context.Context, req validator.NumberRequest, resp *validator.NumberResponse) {
	validateReq := AlsoRequiresNOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		N:              o.N,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AlsoRequiresNOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o AlsoRequiresNOfValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	validateReq := AlsoRequiresNOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		N:              o.N,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AlsoRequiresNOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o AlsoRequiresNOfValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	validateReq := AlsoRequiresNOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		N:              o.N,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AlsoRequiresNOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o AlsoRequiresNOfValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	validateReq := AlsoRequiresNOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		N:              o.N,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AlsoRequiresNOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

// AlsoRequiresNOf ensures that of a set of path.Expression,
// at least 'n' have a non-null value.
//
// Any relative path.Expression will be resolved using the attribute being
// validated.
func AlsoRequiresNOf(n int, expressions ...path.Expression) AllTypesValidator {
	return AlsoRequiresNOfValidator{
		N:               n,
		PathExpressions: expressions,
	}
}
