package adguard

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gmichels/adguard-client-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource                = &listFilterResource{}
	_ resource.ResourceWithConfigure   = &listFilterResource{}
	_ resource.ResourceWithImportState = &listFilterResource{}
)

// listFilterResource is the resource implementation
type listFilterResource struct {
	adg *adguard.ADG
}

// listFilterResourceModel maps list filter schema data
type listFilterResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Url         types.String `tfsdk:"url"`
	Name        types.String `tfsdk:"name"`
	LastUpdated types.String `tfsdk:"last_updated"`
	RulesCount  types.Int64  `tfsdk:"rules_count"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Whitelist   types.Bool   `tfsdk:"whitelist"`
}

// NewlistFilterResource is a helper function to simplify the provider implementation
func NewListFilterResource() resource.Resource {
	return &listFilterResource{}
}

// Metadata returns the resource type name
func (r *listFilterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_list_filter"
}

// Schema defines the schema for the resource
func (r *listFilterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the list filter",
				Required:    true,
			},
			"url": schema.StringAttribute{
				Description: "Url of the list filter",
				Required:    true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of last synchronization",
				Computed:    true,
			},
			"id": schema.StringAttribute{
				Description: "Identifier attribute",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"rules_count": schema.Int64Attribute{
				Description: "Number of rules in the list filter",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: fmt.Sprintf("Whether this list filter is enabled. Defaults to `%t`", LIST_FILTER_ENABLED),
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(LIST_FILTER_ENABLED),
			},
			"whitelist": schema.BoolAttribute{
				Description: fmt.Sprintf("When `true`, will consider this list filter of type whitelist. Defaults to `%t`", LIST_FILTER_WHITELIST),
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(LIST_FILTER_WHITELIST),
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *listFilterResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.adg = req.ProviderData.(*adguard.ADG)
}

// Create creates the resource and sets the initial Terraform state
func (r *listFilterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// retrieve values from plan
	var plan listFilterResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// instantiate empty client for storing plan data
	var listFilter adguard.AddUrlRequest

	// populate list filter from plan
	listFilter.Name = plan.Name.ValueString()
	listFilter.Url = plan.Url.ValueString()
	listFilter.Whitelist = plan.Whitelist.ValueBool()

	// create new list filter using plan
	newListFilter, _, err := r.adg.CreateListFilter(listFilter)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating AdGuard Home List Filter",
			"Could not create list filter, unexpected error: "+err.Error(),
		)
		return
	}

	// update plan with computed attributes
	plan.ID = types.StringValue(strconv.FormatInt(newListFilter.Id, 10))
	plan.LastUpdated = types.StringValue(newListFilter.LastUpdated)
	plan.RulesCount = types.Int64Value(int64(newListFilter.RulesCount))

	// if list filter is expected to be disabled, need to update it after creation
	// as the create endpoint does not have control over it
	if !plan.Enabled.ValueBool() {

		// generate API request body from plan
		var updateListFilterData adguard.FilterSetUrlData
		updateListFilterData.Enabled = plan.Enabled.ValueBool()
		updateListFilterData.Name = plan.Name.ValueString()
		updateListFilterData.Url = plan.Url.ValueString()

		var updateListFilter adguard.FilterSetUrl
		updateListFilter.Url = listFilter.Url
		updateListFilter.Whitelist = plan.Whitelist.ValueBool()
		updateListFilter.Data = updateListFilterData

		// update existing list filter
		_, _, err := r.adg.UpdateListFilter(updateListFilter)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating AdGuard Home List Filter",
				"Could not update list filter, unexpected error: "+err.Error(),
			)
			return
		}

		// rules count gets set to 0 when the list filter is disabled
		plan.RulesCount = types.Int64Value(0)
	}

	// set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data
func (r *listFilterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// get current state
	var state listFilterResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// convert id to int64
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading AdGuard Home List Filter",
			"Could not read list filter with id "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// get refreshed list filter from AdGuard Home
	listFilter, whitelist, err := r.adg.GetListFilterById(id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading AdGuard Home List Filter",
			"Could not read list filter with id "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	// convert to JSON for response logging
	listFilterJson, err := json.Marshal(listFilter)
	if err != nil {
		diags.AddError(
			"Unable to Parse AdGuard Home List Filter",
			err.Error(),
		)
		return
	}
	// log response body
	tflog.Debug(ctx, "ADG API response", map[string]interface{}{
		"object": "listFilter",
		"body":   string(listFilterJson),
	})
	if listFilter == nil {
		resp.Diagnostics.AddWarning(
			"AdGuard Home List Filter was deleted outside of Terraform",
			"No such list filter with id "+state.ID.ValueString(),
		)
		// remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// overwrite list filter with refreshed state
	state.Name = types.StringValue(listFilter.Name)
	state.Enabled = types.BoolValue(listFilter.Enabled)
	state.LastUpdated = types.StringValue(listFilter.LastUpdated)
	state.Url = types.StringValue(listFilter.Url)
	state.RulesCount = types.Int64Value(int64(listFilter.RulesCount))
	state.Whitelist = types.BoolValue(whitelist)

	// set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success
func (r *listFilterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// retrieve values from plan
	var plan listFilterResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// convert id to int64
	id, err := strconv.ParseInt(plan.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading AdGuard Home List Filter",
			"Could not read list filter with id "+plan.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	// retrieve current list filter as we need the current URL
	currentListFilter, _, err := r.adg.GetListFilterById(id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Retrieving AdGuard Home List Filter",
			"Could not retrieve list filter, unexpected error: "+err.Error(),
		)
		return
	}
	// convert to JSON for response logging
	currentListFilterJson, err := json.Marshal(currentListFilter)
	if err != nil {
		diags.AddError(
			"Unable to Parse AdGuard Home List Filter",
			err.Error(),
		)
		return
	}
	// log response body
	tflog.Debug(ctx, "ADG API response", map[string]interface{}{
		"object": "currentListFilter",
		"body":   string(currentListFilterJson),
	})

	// generate API request body from plan
	var updateListFilterData adguard.FilterSetUrlData
	updateListFilterData.Enabled = plan.Enabled.ValueBool()
	updateListFilterData.Name = plan.Name.ValueString()
	updateListFilterData.Url = plan.Url.ValueString()

	var updateListFilter adguard.FilterSetUrl
	updateListFilter.Url = currentListFilter.Url
	updateListFilter.Whitelist = plan.Whitelist.ValueBool()
	updateListFilter.Data = updateListFilterData

	// update existing list filter
	updatedlistFilter, _, err := r.adg.UpdateListFilter(updateListFilter)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating AdGuard Home List Filter",
			"Could not update list filter, unexpected error: "+err.Error(),
		)
		return
	}

	// update plan with computed attributes
	plan.LastUpdated = types.StringValue(updatedlistFilter.LastUpdated)
	plan.RulesCount = types.Int64Value(int64(updatedlistFilter.RulesCount))

	// update state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success
func (r *listFilterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// retrieve values from state
	var state listFilterResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var deleteListFilter adguard.RemoveUrlRequest
	deleteListFilter.Url = state.Url.ValueString()
	deleteListFilter.Whitelist = state.Whitelist.ValueBool()

	// delete existing list filter
	err := r.adg.DeleteListFilter(deleteListFilter)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting AdGuard Home List Filter",
			"Could not delete list filter, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *listFilterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
