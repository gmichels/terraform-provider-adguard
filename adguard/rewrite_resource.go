package adguard

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/gmichels/adguard-client-go"
	adgmodels "github.com/gmichels/adguard-client-go/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource                = &rewriteResource{}
	_ resource.ResourceWithConfigure   = &rewriteResource{}
	_ resource.ResourceWithImportState = &rewriteResource{}
)

// rewriteResource is the resource implementation
type rewriteResource struct {
	adg *adguard.ADG
}

// rewriteResourceModel maps DNS rewrite rule schema data
type rewriteResourceModel struct {
	ID          types.String `tfsdk:"id"`
	LastUpdated types.String `tfsdk:"last_updated"`
	Domain      types.String `tfsdk:"domain"`
	Answer      types.String `tfsdk:"answer"`
}

// NewRewriteResource is a helper function to simplify the provider implementation
func NewRewriteResource() resource.Resource {
	return &rewriteResource{}
}

// Metadata returns the resource type name
func (r *rewriteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rewrite"
}

// Schema defines the schema for the resource
func (r *rewriteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internal identifier for this rewrite",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the rewrite",
				Computed:    true,
			},
			"domain": schema.StringAttribute{
				Description: "Domain name",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"answer": schema.StringAttribute{
				Description: "Value of A, AAAA or CNAME DNS record",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[A-Za-z0-9/.:-]+$`),
						"must be an IP address/CIDR, MAC address, or only contain numbers, lowercase letters, and hyphens",
					),
				},
			},
		},
	}
}

// Configure adds the provider configured DNS rewrite rule to the resource
func (r *rewriteResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.adg = req.ProviderData.(*adguard.ADG)
}

// Create creates the resource and sets the initial Terraform state
func (r *rewriteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// retrieve values from plan
	var plan rewriteResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// instantiate empty DNS rewrite rule for storing plan data
	var rewrite adgmodels.RewriteEntry

	// populate DNS rewrite rule from plan
	rewrite.Domain = plan.Domain.ValueString()
	rewrite.Answer = plan.Answer.ValueString()

	// create new DNS rewrite rule using plan
	err := r.adg.RewriteAdd(rewrite)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating DNS Rewrite Rule",
			"Could not create DNS rewrite rule, unexpected error: "+err.Error(),
		)
		return
	}

	// add missing attributes for state
	plan.ID = types.StringValue(rewrite.Domain + "||" + rewrite.Answer)
	// add the last updated attribute
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data
func (r *rewriteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// get current state
	var state rewriteResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// detect old ID format
	if !strings.Contains(state.ID.ValueString(), "||") {
		// set the ID to the new format
		newID := state.Domain.ValueString() + "||" + state.Answer.ValueString()
		state.ID = types.StringValue(newID)
	}

	// split the ID to separate domain and answer
	idSplit := strings.Split(state.ID.ValueString(), "||")

	// get refreshed DNS rewrite rule value from AdGuard Home
	rewrite, err := GetRewrite(r.adg, idSplit[0], idSplit[1])
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading AdGuard Home DNS Rewrite Rule",
			"Could not read AdGuard Home DNS rewrite rule with ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	// convert to JSON for response logging
	rewriteJson, err := json.Marshal(rewrite)
	if err != nil {
		diags.AddError(
			"Unable to Parse AdGuard Home Rewrite Rule",
			err.Error(),
		)
		return
	}
	// log response body
	tflog.Debug(ctx, "ADG API response", map[string]interface{}{
		"object": "rewrite",
		"body":   string(rewriteJson),
	})
	if rewrite == nil {
		resp.Diagnostics.AddWarning(
			"AdGuard Home DNS Rewrite Rule was deleted outside of Terraform",
			"No such rewrite rule with id "+state.ID.ValueString(),
		)
		// remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// overwrite DNS rewrite rule with refreshed state
	state.Domain = types.StringValue(rewrite.Domain)
	state.Answer = types.StringValue(rewrite.Answer)

	// set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success
func (r *rewriteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// retrieve values from plan
	var plan rewriteResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// retrieve state as we need the current info
	var state rewriteResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// generate API request body from plan and state
	var updateRewriteTarget adgmodels.RewriteEntry
	updateRewriteTarget.Domain = state.Domain.ValueString()
	updateRewriteTarget.Answer = state.Answer.ValueString()

	var updateRewriteUpdate adgmodels.RewriteEntry
	updateRewriteUpdate.Domain = plan.Domain.ValueString()
	updateRewriteUpdate.Answer = plan.Answer.ValueString()

	var updateRewrite adgmodels.RewriteUpdate
	updateRewrite.Target = updateRewriteTarget
	updateRewrite.Update = updateRewriteUpdate

	// update existing DNS rewrite rule
	err := r.adg.RewriteUpdate(updateRewrite)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating AdGuard Home DNS Rewrite Rule",
			"Could not update DNS rewrite rule, unexpected error: "+err.Error(),
		)
		return
	}

	// update resource state with updated items and timestamp
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// update state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success
func (r *rewriteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// retrieve values from state
	var state rewriteResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// generate API request body from state
	var deleteRewrite adgmodels.RewriteEntry
	deleteRewrite.Domain = state.Domain.ValueString()
	deleteRewrite.Answer = state.Answer.ValueString()

	// delete existing DNS rewrite rule
	err := r.adg.RewriteDelete(deleteRewrite)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting AdGuard Home DNS Rewrite Rule",
			"Could not delete DNS rewrite rule, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *rewriteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
