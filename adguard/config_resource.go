package adguard

import (
	"context"
	"reflect"
	"regexp"
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
	QueryLog        types.Object `tfsdk:"querylog"`
	Stats           types.Object `tfsdk:"stats"`
	BlockedServices types.Set    `tfsdk:"blocked_services"`
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
							setvalidator.SizeAtLeast(1),
							setvalidator.ValueStringsAre(
								stringvalidator.OneOf(
									DEFAULT_SAFESEARCH_SERVICES...,
								),
							),
						},
						Default: setdefault.StaticValue(
							types.SetValueMust(
								types.StringType,
								[]attr.Value{
									// TODO
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
			"querylog": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					queryLogConfigModel{}.attrTypes(), queryLogConfigModel{}.defaultObject()),
				),
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Whether the query log is enabled. Defaults to `true`",
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(true),
					},
					"interval": schema.Int64Attribute{
						Description: "Time period for query log rotation, in hours. Defaults to `2160` (90 days)",
						Computed:    true,
						Optional:    true,
						Default:     int64default.StaticInt64(24),
					},
					"anonymize_client_ip": schema.BoolAttribute{
						Description: "Whether anonymizing clients' IP addresses is enabled",
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(false),
					},
					"ignored": schema.SetAttribute{
						Description: "List of host names which should not be written to log",
						ElementType: types.StringType,
						Computed:    true,
						Optional:    true,
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
							setvalidator.ValueStringsAre(
								stringvalidator.RegexMatches(
									regexp.MustCompile(`^[a-z0-9.-_]+$`),
									"must be a valid domain name",
								),
							),
						},
						Default: setdefault.StaticValue(
							types.SetNull(types.StringType),
						),
					},
				},
			},
			"stats": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					statsConfigModel{}.attrTypes(), statsConfigModel{}.defaultObject()),
				),
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Whether server statistics are enabled. Defaults to `true`",
						Computed:    true,
						Optional:    true,
						Default:     booldefault.StaticBool(true),
					},
					"interval": schema.Int64Attribute{
						Description: "Time period for server statistics rotation, in hours. Defaults to `24` (1 day)",
						Computed:    true,
						Optional:    true,
						Default:     int64default.StaticInt64(24),
					},
					"ignored": schema.SetAttribute{
						Description: "List of host names which should not be counted in the server statistics",
						ElementType: types.StringType,
						Computed:    true,
						Optional:    true,
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
							setvalidator.ValueStringsAre(
								stringvalidator.RegexMatches(
									regexp.MustCompile(`^[a-z0-9.-_]+$`),
									"must be a valid domain name",
								),
							),
						},
						Default: setdefault.StaticValue(
							types.SetNull(types.StringType),
						),
					},
				},
			},
			"blocked_services": schema.SetAttribute{
				Description: "List of services to be blocked globally",
				ElementType: types.StringType,
				Computed:    true,
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(
							r.DefaultBlockedServicesList()...,
						),
					),
				},
				Default: setdefault.StaticValue(
					types.SetNull(types.StringType),
				),
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

	// defer to common function to create or update the resource
	plan, err := r.CreateOrUpdateConfigResource(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating AdGuard Home Config",
			"Could not create AdGuard Home config: "+err.Error(),
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
	// map filtering config to state
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

	// get refreshed parental control status from AdGuard Home
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

	// retrieve query log config info
	queryLogConfig, err := r.adg.GetQueryLogConfig()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading AdGuard Home Config",
			"Could not read AdGuard Home config ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	var stateQueryLogConfig queryLogConfigModel
	stateQueryLogConfig.Enabled = types.BoolValue(queryLogConfig.Enabled)
	stateQueryLogConfig.Interval = types.Int64Value(int64(queryLogConfig.Interval / 1000 / 3600))
	stateQueryLogConfig.AnonymizeClientIp = types.BoolValue(queryLogConfig.AnonymizeClientIp)
	stateQueryLogConfig.Ignored, diags = types.SetValueFrom(ctx, types.StringType, queryLogConfig.Ignored)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// retrieve stats config info
	statsConfig, err := r.adg.GetStatsConfig()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading AdGuard Home Config",
			"Could not read AdGuard Home config ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	var stateStatsConfig statsConfigModel
	stateStatsConfig.Enabled = types.BoolValue(statsConfig.Enabled)
	stateStatsConfig.Interval = types.Int64Value(int64(statsConfig.Interval / 1000 / 3600))
	stateStatsConfig.Ignored, diags = types.SetValueFrom(ctx, types.StringType, statsConfig.Ignored)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get refreshed blocked services from AdGuard Home
	blockedServices, err := r.adg.GetBlockedServices()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading AdGuard Home Config",
			"Could not read AdGuard Home config ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// overwrite config with refreshed state
	state.Filtering, _ = types.ObjectValueFrom(ctx, filteringModel{}.attrTypes(), &stateFilteringConfig)
	state.SafeBrowsing, _ = types.ObjectValueFrom(ctx, enabledModel{}.attrTypes(), &stateSafeBrowsingStatus)
	state.ParentalControl, _ = types.ObjectValueFrom(ctx, enabledModel{}.attrTypes(), &stateParentalStatus)
	state.SafeSearch, _ = types.ObjectValueFrom(ctx, safeSearchModel{}.attrTypes(), &stateSafeSearchConfig)
	state.QueryLog, _ = types.ObjectValueFrom(ctx, queryLogConfigModel{}.attrTypes(), &stateQueryLogConfig)
	state.Stats, _ = types.ObjectValueFrom(ctx, statsConfigModel{}.attrTypes(), &stateStatsConfig)
	state.BlockedServices, diags = types.SetValueFrom(ctx, types.StringType, blockedServices)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	// defer to common function to create or update the resource
	plan, err := r.CreateOrUpdateConfigResource(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating AdGuard Home Config",
			"Could not create AdGuard Home config: "+err.Error(),
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

	// populate filtering config with default values
	var filterConfig adguard.FilterConfig
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

	// populate safe search with default values
	var safeSearchConfig adguard.SafeSearchConfig
	safeSearchConfig.Enabled = false
	safeSearchConfig.Bing = true
	safeSearchConfig.Duckduckgo = true
	safeSearchConfig.Google = true
	safeSearchConfig.Pixabay = true
	safeSearchConfig.Yandex = true
	safeSearchConfig.Youtube = true

	// set safe search to defaults
	_, err = r.adg.SetSafeSearchConfig(safeSearchConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting AdGuard Home Config",
			"Could not update config, unexpected error: "+err.Error(),
		)
		return
	}

	// populate query log config with default values
	var queryLogConfig adguard.GetQueryLogConfigResponse
	queryLogConfig.Enabled = true
	queryLogConfig.Interval = 90 * 86400 * 1000
	queryLogConfig.AnonymizeClientIp = false
	queryLogConfig.Ignored = []string{}

	// set query log config to defaults
	_, err = r.adg.SetQueryLogConfig(queryLogConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting AdGuard Home Config",
			"Could not update config, unexpected error: "+err.Error(),
		)
		return
	}

	// populate server statistics config with default values
	var statsConfig adguard.GetStatsConfigResponse
	statsConfig.Enabled = true
	statsConfig.Interval = 1 * 86400 * 1000
	statsConfig.Ignored = []string{}

	// set server statistics to defaults
	_, err = r.adg.SetStatsConfig(statsConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Server Statistics Config",
			"Could not delete server statistics configuration, unexpected error: "+err.Error(),
		)
		return
	}

	// set blocked services to defaults
	_, err = r.adg.SetBlockedServices(make([]string, 0))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Server Statistics Config",
			"Could not delete server statistics configuration, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *configResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
