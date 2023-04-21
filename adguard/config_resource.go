package adguard

import (
	"context"
	"reflect"
	"time"

	"github.com/gmichels/adguard-client-go"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource                = &configResource{}
	_ resource.ResourceWithConfigure   = &configResource{}
	_ resource.ResourceWithImportState = &configResource{}
)

// configResource is the resource implementation
type configResource struct {
	adg *adguard.ADG
}

// configResourceModel maps config schema data
type configResourceModel struct {
	ID              types.String `tfsdk:"id"`
	LastUpdated     types.String `tfsdk:"last_updated"`
	Filtering       types.Object `tfsdk:"filtering"`
	SafeBrowsing    types.Object `tfsdk:"safebrowsing"`
	ParentalControl types.Object `tfsdk:"parental_control"`
	SafeSearch      types.Object `tfsdk:"safesearch"`
}

// NewConfigResource is a helper function to simplify the provider implementation
func NewConfigResource() resource.Resource {
	return &configResource{}
}

// Metadata returns the resource type name
func (r *configResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config"
}

// Schema defines the schema for the resource
func (r *configResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internal identifier for this config",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the config",
				Computed:    true,
			},
			"filtering": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					filteringModel{}.attrTypes(), filteringModel{}.defaultObject()),
				),
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Whether DNS filtering is enabled. Defaults to `true`",
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(true),
					},
					"update_interval": schema.Int64Attribute{
						Description: "Update interval for all list-based filters, in hours. Defaults to `24`",
						Computed:    true,
						Optional:    true,
						Default:     int64default.StaticInt64(24),
						Validators: []validator.Int64{
							int64validator.OneOf([]int64{1, 12, 24, 72, 168}...),
						},
					},
				},
			},
			"safebrowsing": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					enabledModel{}.attrTypes(), enabledModel{}.defaultObject()),
				),
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Whether Safe Browsing is enabled. Defaults to `false`",
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(false),
					},
				},
			},
			"parental_control": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					enabledModel{}.attrTypes(), enabledModel{}.defaultObject()),
				),
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Whether Parental Control is enabled. Defaults to `false`",
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(false),
					},
				},
			},
			"safesearch": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					safeSearchModel{}.attrTypes(), safeSearchModel{}.defaultObject()),
				),
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Whether Safe Search is enabled. Defaults to `false`",
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(false),
					},
					"services": schema.SetAttribute{
						Description: "Services which SafeSearch is enabled. Defaults to Bing, DuckDuckGo, Google, Pixabay, Yandex and YouTube",
						ElementType: types.StringType,
						Computed:    true,
						Optional:    true,
						Validators: []validator.Set{
							setvalidator.AlsoRequires(path.Expressions{
								path.MatchRelative().AtParent().AtName("enabled"),
							}...),
							setvalidator.SizeAtLeast(1),
							setvalidator.ValueStringsAre(
								stringvalidator.OneOf(
									"bing", "duckduckgo", "google", "pixabay", "yandex", "youtube",
								),
							),
						},
						Default: setdefault.StaticValue(
							types.SetValueMust(
								types.StringType,
								[]attr.Value{
									types.StringValue("bing"),
									types.StringValue("duckduckgo"),
									types.StringValue("google"),
									types.StringValue("pixabay"),
									types.StringValue("yandex"),
									types.StringValue("youtube"),
								},
							),
						),
					},
				},
			},
		},
	}
}

// Configure adds the provider configured config to the resource
func (r *configResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.adg = req.ProviderData.(*adguard.ADG)
}

// Create creates the resource and sets the initial Terraform state
func (r *configResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// retrieve values from plan
	var plan configResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// unpack nested attributes from plan
	var planFiltering filteringModel
	var planSafeBrowsing enabledModel
	var planParentalControl enabledModel
	var planSafeSearch safeSearchModel
	_ = plan.Filtering.As(ctx, &planFiltering, basetypes.ObjectAsOptions{})
	_ = plan.SafeBrowsing.As(ctx, &planSafeBrowsing, basetypes.ObjectAsOptions{})
	_ = plan.ParentalControl.As(ctx, &planParentalControl, basetypes.ObjectAsOptions{})
	_ = plan.SafeSearch.As(ctx, &planSafeSearch, basetypes.ObjectAsOptions{})

	// instantiate empty objects for storing plan data
	var filteringConfig adguard.FilterConfig
	var safeSearchConfig adguard.SafeSearchConfig

	// populate filtering config from plan
	filteringConfig.Enabled = planFiltering.Enabled.ValueBool()
	filteringConfig.Interval = uint(planFiltering.UpdateInterval.ValueInt64())

	// set filtering config using plan
	_, err := r.adg.ConfigureFiltering(filteringConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Config",
			"Could not create config, unexpected error: "+err.Error(),
		)
		return
	}

	// set safe browsing status using plan
	err = r.adg.SetSafeBrowsingStatus(planSafeBrowsing.Enabled.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Config",
			"Could not create config, unexpected error: "+err.Error(),
		)
		return
	}

	// set parental control status using plan
	err = r.adg.SetParentalStatus(planParentalControl.Enabled.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Config",
			"Could not create config, unexpected error: "+err.Error(),
		)
		return
	}

	// populate search config using plan
	safeSearchConfig.Enabled = planSafeSearch.Enabled.ValueBool()
	if len(planSafeSearch.Services.Elements()) > 0 {
		var safeSearchServicesEnabled []string
		diags = planSafeSearch.Services.ElementsAs(ctx, &safeSearchServicesEnabled, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// use reflection to set each safeSearchConfig service value dynamically
		v := reflect.ValueOf(&safeSearchConfig).Elem()
		t := v.Type()
		setSafeSearchConfigServices(v, t, safeSearchServicesEnabled)
	}

	// set safe search config using plan
	_, err = r.adg.SetSafeSearchConfig(safeSearchConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Config",
			"Could not create config, unexpected error: "+err.Error(),
		)
		return
	}

	// there can be only one entry Config, so hardcode the ID as 1
	plan.ID = types.StringValue("1")
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
func (r *configResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// get current state
	var state configResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get refreshed filtering config value from AdGuard Home
	filteringConfig, err := r.adg.GetAllFilters()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading AdGuard Home Config",
			"Could not read AdGuard Home config ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	// map filter config to state
	var stateFilteringConfig filteringModel
	stateFilteringConfig.Enabled = types.BoolValue(filteringConfig.Enabled)
	stateFilteringConfig.UpdateInterval = types.Int64Value(int64(filteringConfig.Interval))

	// get refreshed safe browsing status from AdGuard Home
	safeBrowsingStatus, err := r.adg.GetSafeBrowsingStatus()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading AdGuard Home Config",
			"Could not read AdGuard Home config ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	// map safe browsing config to state
	var stateSafeBrowsingStatus enabledModel
	stateSafeBrowsingStatus.Enabled = types.BoolValue(*safeBrowsingStatus)

	// get refreshed safe parental control status from AdGuard Home
	parentalStatus, err := r.adg.GetParentalStatus()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading AdGuard Home Config",
			"Could not read AdGuard Home config ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	// map parental control config to state
	var stateParentalStatus enabledModel
	stateParentalStatus.Enabled = types.BoolValue(*parentalStatus)

	// retrieve safe search info
	safeSearchConfig, err := r.adg.GetSafeSearchConfig()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading AdGuard Home Config",
			"Could not read AdGuard Home config ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	// perform reflection of safeSearchConfig object
	v := reflect.ValueOf(safeSearchConfig).Elem()
	// grab the type of the reflected object
	t := v.Type()
	// map the reflected object to a list of enabled services
	enabledSafeSearchServices := mapSafeSearchConfigServices(v, t)

	// map safe search to state
	var stateSafeSearchConfig safeSearchModel
	stateSafeSearchConfig.Enabled = types.BoolValue(safeSearchConfig.Enabled)
	stateSafeSearchConfig.Services, diags = types.SetValueFrom(ctx, types.StringType, enabledSafeSearchServices)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// overwrite config with refreshed state
	state.Filtering, _ = types.ObjectValueFrom(ctx, filteringModel{}.attrTypes(), &stateFilteringConfig)
	state.SafeBrowsing, _ = types.ObjectValueFrom(ctx, enabledModel{}.attrTypes(), &stateSafeBrowsingStatus)
	state.ParentalControl, _ = types.ObjectValueFrom(ctx, enabledModel{}.attrTypes(), &stateParentalStatus)
	state.SafeSearch, _ = types.ObjectValueFrom(ctx, safeSearchModel{}.attrTypes(), &stateSafeSearchConfig)

	// set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success
func (r *configResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// retrieve values from plan
	var plan configResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// unpack nested attributes from plan
	var planFiltering filteringModel
	var planSafeBrowsing enabledModel
	var planParentalControl enabledModel
	var planSafeSearch safeSearchModel
	_ = plan.Filtering.As(ctx, &planFiltering, basetypes.ObjectAsOptions{})
	_ = plan.SafeBrowsing.As(ctx, &planSafeBrowsing, basetypes.ObjectAsOptions{})
	_ = plan.ParentalControl.As(ctx, &planParentalControl, basetypes.ObjectAsOptions{})
	_ = plan.SafeSearch.As(ctx, &planSafeSearch, basetypes.ObjectAsOptions{})

	// generate API request body from plan
	var filteringConfig adguard.FilterConfig
	var safeSearchConfig adguard.SafeSearchConfig

	filteringConfig.Enabled = planFiltering.Enabled.ValueBool()
	filteringConfig.Interval = uint(planFiltering.UpdateInterval.ValueInt64())

	// update existing filtering config
	_, err := r.adg.ConfigureFiltering(filteringConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating AdGuard Home Config",
			"Could not update config, unexpected error: "+err.Error(),
		)
		return
	}

	// update safe browsing status
	err = r.adg.SetSafeBrowsingStatus(planSafeBrowsing.Enabled.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating AdGuard Home Config",
			"Could not update config, unexpected error: "+err.Error(),
		)
		return
	}

	// update parental status
	err = r.adg.SetParentalStatus(planParentalControl.Enabled.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating AdGuard Home Config",
			"Could not update config, unexpected error: "+err.Error(),
		)
		return
	}

	// populate search config using plan
	safeSearchConfig.Enabled = planSafeSearch.Enabled.ValueBool()
	if len(planSafeSearch.Services.Elements()) > 0 {
		var safeSearchServicesEnabled []string
		diags = planSafeSearch.Services.ElementsAs(ctx, &safeSearchServicesEnabled, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// use reflection to set each safeSearchConfig service value dynamically
		v := reflect.ValueOf(&safeSearchConfig).Elem()
		t := v.Type()
		setSafeSearchConfigServices(v, t, safeSearchServicesEnabled)
	}

	// set safe search config using plan
	_, err = r.adg.SetSafeSearchConfig(safeSearchConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Config",
			"Could not create config, unexpected error: "+err.Error(),
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
func (r *configResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// there is no "real" delete for the configuration, so this means "restore defaults"

	var filterConfig adguard.FilterConfig
	var safeSearchConfig adguard.SafeSearchConfig

	// populate filtering config with default values
	filterConfig.Enabled = true
	filterConfig.Interval = 24

	// set filtering config to default
	_, err := r.adg.ConfigureFiltering(filterConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting AdGuard Home Config",
			"Could not delete config, unexpected error: "+err.Error(),
		)
		return
	}

	// set safebrowsing to default
	err = r.adg.SetSafeBrowsingStatus(false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting AdGuard Home Config",
			"Could not update config, unexpected error: "+err.Error(),
		)
		return
	}

	// set parental to default
	err = r.adg.SetParentalStatus(false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting AdGuard Home Config",
			"Could not update config, unexpected error: "+err.Error(),
		)
		return
	}

	// set safe search to defaults
	safeSearchConfig.Enabled = false
	safeSearchConfig.Bing = true
	safeSearchConfig.Duckduckgo = true
	safeSearchConfig.Google = true
	safeSearchConfig.Pixabay = true
	safeSearchConfig.Yandex = true
	safeSearchConfig.Youtube = true

	_, err = r.adg.SetSafeSearchConfig(safeSearchConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting AdGuard Home Config",
			"Could not update config, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *configResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
