package adguard

import (
	"context"
	"reflect"

	"github.com/gmichels/adguard-client-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// common client model to be used for working with both resource and data source
type clientCommonModel struct {
	ID                           types.String `tfsdk:"id"`
	LastUpdated                  types.String `tfsdk:"last_updated"`
	Name                         types.String `tfsdk:"name"`
	Ids                          types.List   `tfsdk:"ids"`
	UseGlobalSettings            types.Bool   `tfsdk:"use_global_settings"`
	FilteringEnabled             types.Bool   `tfsdk:"filtering_enabled"`
	ParentalEnabled              types.Bool   `tfsdk:"parental_enabled"`
	SafebrowsingEnabled          types.Bool   `tfsdk:"safebrowsing_enabled"`
	SafeSearch                   types.Object `tfsdk:"safesearch"`
	UseGlobalBlockedServices     types.Bool   `tfsdk:"use_global_blocked_services"`
	BlockedServicesPauseSchedule types.Object `tfsdk:"blocked_services_pause_schedule"`
	BlockedServices              types.Set    `tfsdk:"blocked_services"`
	Upstreams                    types.List   `tfsdk:"upstreams"`
	Tags                         types.Set    `tfsdk:"tags"`
	IgnoreQuerylog               types.Bool   `tfsdk:"ignore_querylog"`
	IgnoreStatistics             types.Bool   `tfsdk:"ignore_statistics"`
	UpstreamsCacheEnabled        types.Bool   `tfsdk:"upstreams_cache_enabled"`
	UpstreamsCacheSize           types.Int64  `tfsdk:"upstreams_cache_size"`
}

// common `Read` function for both data source and resource
func (o *clientCommonModel) Read(ctx context.Context, adg adguard.ADG, currState *clientCommonModel, diags *diag.Diagnostics, rtype string) {
	// need to define client name based whether it's an import operation
	var clientName string
	if !currState.Name.IsNull() {
		// name exists in plan
		clientName = currState.Name.ValueString()
	} else {
		// this is an import operation, use the ID
		clientName = currState.ID.ValueString()
	}
	// retrieve client info
	client, err := adg.GetClient(clientName)
	if err != nil {
		diags.AddError(
			"Unable to Read AdGuard Home Client",
			err.Error(),
		)
		return
	}
	if client == nil {
		diags.AddError(
			"Unable to Locate AdGuard Home Client",
			"No client with name `"+clientName+"` exists in AdGuard Home.",
		)
		return
	}

	// map response body to model
	o.Name = types.StringValue(client.Name)
	o.Ids, *diags = types.ListValueFrom(ctx, types.StringType, client.Ids)
	if diags.HasError() {
		return
	}
	o.UseGlobalSettings = types.BoolValue(client.UseGlobalSettings)
	o.FilteringEnabled = types.BoolValue(client.FilteringEnabled)
	o.ParentalEnabled = types.BoolValue(client.ParentalEnabled)
	o.SafebrowsingEnabled = types.BoolValue(client.SafebrowsingEnabled)

	var stateSafeSearchClient safeSearchModel
	stateSafeSearchClient.Enabled = types.BoolValue(client.SafeSearch.Enabled)
	// map safe search config object to a list of enabled services
	enabledSafeSearchServices := mapSafeSearchServices(&client.SafeSearch)
	stateSafeSearchClient.Services, *diags = types.SetValueFrom(ctx, types.StringType, enabledSafeSearchServices)
	if diags.HasError() {
		return
	}
	// add to config model
	o.SafeSearch, _ = types.ObjectValueFrom(ctx, safeSearchModel{}.attrTypes(), &stateSafeSearchClient)

	o.UseGlobalBlockedServices = types.BoolValue(client.UseGlobalBlockedServices)

	// use common function to map blocked services pause schedules for each day
	stateBlockedServicesPauseScheduleClient := mapAdgScheduleToBlockedServicesPauseSchedule(ctx, &client.BlockedServicesSchedule)

	// need special handling for timezone in resource due to inconsistent API response for `Local`
	if rtype == "resource" {
		// last updated will exist on create operation, null on import operation
		if !currState.LastUpdated.IsNull() && !currState.BlockedServicesPauseSchedule.IsNull() {
			// unpack current state
			var currStateBlockedServicesPauseScheduleClient scheduleModel
			*diags = currState.BlockedServicesPauseSchedule.As(ctx, &currStateBlockedServicesPauseScheduleClient, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				return
			}
			// if timezone in state is null, it means it was never defined, so we should ignore the inconsistent response from ADG
			if !currStateBlockedServicesPauseScheduleClient.TimeZone.IsNull() {
				// map timezone from response
				stateBlockedServicesPauseScheduleClient.TimeZone = types.StringValue(client.BlockedServicesSchedule.TimeZone)
			}
			// ID exists in both create and import operations, but if we got here, it's an import
			// still, imports for this attribute are finicky and error-prone, therefore ignored in tests
		} else if !currState.ID.IsNull() {
			// map timezone from response
			stateBlockedServicesPauseScheduleClient.TimeZone = types.StringValue(client.BlockedServicesSchedule.TimeZone)
		}
	} else {
		// used for datasource
		stateBlockedServicesPauseScheduleClient.TimeZone = types.StringValue(client.BlockedServicesSchedule.TimeZone)
	}

	o.BlockedServices, *diags = types.SetValueFrom(ctx, types.StringType, client.BlockedServices)
	if diags.HasError() {
		return
	}
	o.BlockedServicesPauseSchedule, *diags = types.ObjectValueFrom(ctx, scheduleModel{}.attrTypes(), &stateBlockedServicesPauseScheduleClient)
	if diags.HasError() {
		return
	}
	o.Upstreams, *diags = types.ListValueFrom(ctx, types.StringType, client.Upstreams)
	if diags.HasError() {
		return
	}
	o.Tags, *diags = types.SetValueFrom(ctx, types.StringType, client.Tags)
	if diags.HasError() {
		return
	}
	o.IgnoreQuerylog = types.BoolValue(client.IgnoreQuerylog)
	o.IgnoreStatistics = types.BoolValue(client.IgnoreStatistics)
	o.UpstreamsCacheEnabled = types.BoolValue(client.UpstreamsCacheEnabled)
	o.UpstreamsCacheSize = types.Int64Value(int64(client.UpstreamsCacheSize))

	// if we got here, all went fine
}

// common `Create` and `Update` function for the resource
func (r *clientResource) CreateOrUpdate(ctx context.Context, plan *clientCommonModel, diags *diag.Diagnostics, create_operation bool) {
	// instantiate empty client for storing plan data
	var client adguard.Client

	// populate client from plan
	client.Name = plan.Name.ValueString()
	*diags = plan.Ids.ElementsAs(ctx, &client.Ids, false)
	if diags.HasError() {
		return
	}
	client.UseGlobalSettings = plan.UseGlobalSettings.ValueBool()
	client.FilteringEnabled = plan.FilteringEnabled.ValueBool()
	client.ParentalEnabled = plan.ParentalEnabled.ValueBool()
	client.SafebrowsingEnabled = plan.SafebrowsingEnabled.ValueBool()

	// unpack nested attributes from plan
	var planSafeSearch safeSearchModel
	*diags = plan.SafeSearch.As(ctx, &planSafeSearch, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return
	}

	client.SafeSearch.Enabled = planSafeSearch.Enabled.ValueBool()

	if len(planSafeSearch.Services.Elements()) > 0 {
		var safeSearchServicesEnabled []string
		*diags = planSafeSearch.Services.ElementsAs(ctx, &safeSearchServicesEnabled, false)
		if diags.HasError() {
			return
		} // use reflection to set each safe search services value dynamically
		v := reflect.ValueOf(&client.SafeSearch).Elem()
		t := v.Type()
		setSafeSearchServices(v, t, safeSearchServicesEnabled)
	}

	client.UseGlobalBlockedServices = plan.UseGlobalBlockedServices.ValueBool()

	// unpack nested attributes from plan
	var planBlockedServicesPauseScheduleClient scheduleModel
	*diags = plan.BlockedServicesPauseSchedule.As(ctx, &planBlockedServicesPauseScheduleClient, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return
	} // defer to common function to populate schedule
	client.BlockedServicesSchedule = mapBlockedServicesPauseScheduleToAdgSchedule(ctx, planBlockedServicesPauseScheduleClient)

	if len(plan.BlockedServices.Elements()) > 0 {
		*diags = plan.BlockedServices.ElementsAs(ctx, &client.BlockedServices, false)
		if diags.HasError() {
			return
		}
	}
	if len(plan.Upstreams.Elements()) > 0 {
		*diags = plan.Upstreams.ElementsAs(ctx, &client.Upstreams, false)
		if diags.HasError() {
			return
		}
	}
	if len(plan.Tags.Elements()) > 0 {
		*diags = plan.Tags.ElementsAs(ctx, &client.Tags, false)
		if diags.HasError() {
			return
		}
	}
	client.IgnoreQuerylog = plan.IgnoreQuerylog.ValueBool()
	client.IgnoreStatistics = plan.IgnoreStatistics.ValueBool()
	client.UpstreamsCacheEnabled = plan.UpstreamsCacheEnabled.ValueBool()
	client.UpstreamsCacheSize = uint(plan.UpstreamsCacheSize.ValueInt64())

	if create_operation {
		// create new client using plan
		_, err := r.adg.CreateClient(client)
		if err != nil {
			diags.AddError(
				"Error Creating Client",
				"Could not create client, unexpected error: "+err.Error(),
			)
			return
		}
	} else {
		// instantiate specific empty update client for storing plan data
		var updateClient adguard.ClientUpdate
		updateClient.Name = plan.ID.ValueString()
		// grab our client and place in object
		updateClient.Data = client
		// update existing client
		_, err := r.adg.UpdateClient(updateClient)
		if err != nil {
			diags.AddError(
				"Error Updating AdGuard Home Client",
				"Could not update client, unexpected error: "+err.Error(),
			)
			return
		}
	}
}
