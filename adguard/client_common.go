package adguard

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/gmichels/adguard-client-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// common client model to be used for working with both resource and data source
type clientCommonModel struct {
	ID                           types.String `tfsdk:"id"`
	LastUpdated                  types.String `tfsdk:"last_updated"`
	Name                         types.String `tfsdk:"name"`
	Ids                          types.Set    `tfsdk:"ids"` // technically upstream accepts a list with duplicate values, but it doesn't make sense
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
	// initialize empty diags variable
	var d diag.Diagnostics

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
	// convert to JSON for response logging
	clientJson, err := json.Marshal(client)
	if err != nil {
		diags.AddError(
			"Unable to Parse AdGuard Home Client",
			err.Error(),
		)
		return
	}
	// log response body
	tflog.Debug(ctx, "ADG API response", map[string]interface{}{
		"object": "client",
		"body":   string(clientJson),
	})
	if client == nil {
		diags.AddError(
			"Unable to Locate AdGuard Home Client",
			"No client with name `"+clientName+"` exists in AdGuard Home.",
		)
		return
	}

	// map response body to model
	o.Name = types.StringValue(client.Name)
	o.Ids, d = types.SetValueFrom(ctx, types.StringType, client.Ids)
	diags.Append(d...)
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
	stateSafeSearchClient.Services, d = types.SetValueFrom(ctx, types.StringType, enabledSafeSearchServices)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	// add to config model
	o.SafeSearch, d = types.ObjectValueFrom(ctx, safeSearchModel{}.attrTypes(), &stateSafeSearchClient)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	o.UseGlobalBlockedServices = types.BoolValue(client.UseGlobalBlockedServices)

	// use common function to map blocked services pause schedules for each day
	stateBlockedServicesPauseScheduleClient := mapAdgScheduleToBlockedServicesPauseSchedule(ctx, &client.BlockedServicesSchedule, diags)
	if diags.HasError() {
		return
	}

	// need special handling for timezone in resource due to inconsistent API response for `Local`
	if rtype == "resource" {
		// last updated will exist on create operation, null on import operation
		if !currState.LastUpdated.IsNull() && !currState.BlockedServicesPauseSchedule.IsNull() {
			// unpack current state
			var currStateBlockedServicesPauseScheduleClient scheduleModel
			d = currState.BlockedServicesPauseSchedule.As(ctx, &currStateBlockedServicesPauseScheduleClient, basetypes.ObjectAsOptions{})
			diags.Append(d...)
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

	o.BlockedServices, d = types.SetValueFrom(ctx, types.StringType, client.BlockedServices)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	o.BlockedServicesPauseSchedule, d = types.ObjectValueFrom(ctx, scheduleModel{}.attrTypes(), &stateBlockedServicesPauseScheduleClient)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	o.Upstreams, d = types.ListValueFrom(ctx, types.StringType, client.Upstreams)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	o.Tags, d = types.SetValueFrom(ctx, types.StringType, client.Tags)
	diags.Append(d...)
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
	// initialize empty diags variable
	var d diag.Diagnostics

	// instantiate empty client for storing plan data
	var client adguard.Client

	// populate client from plan
	client.Name = plan.Name.ValueString()
	d = plan.Ids.ElementsAs(ctx, &client.Ids, false)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	client.UseGlobalSettings = plan.UseGlobalSettings.ValueBool()
	client.FilteringEnabled = plan.FilteringEnabled.ValueBool()
	client.ParentalEnabled = plan.ParentalEnabled.ValueBool()
	client.SafebrowsingEnabled = plan.SafebrowsingEnabled.ValueBool()

	// unpack nested attributes from plan
	var planSafeSearch safeSearchModel
	d = plan.SafeSearch.As(ctx, &planSafeSearch, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	client.SafeSearch.Enabled = planSafeSearch.Enabled.ValueBool()

	if len(planSafeSearch.Services.Elements()) > 0 {
		var safeSearchServicesEnabled []string
		d = planSafeSearch.Services.ElementsAs(ctx, &safeSearchServicesEnabled, false)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
		// use reflection to set each safe search services value dynamically
		v := reflect.ValueOf(&client.SafeSearch).Elem()
		t := v.Type()
		setSafeSearchServices(v, t, safeSearchServicesEnabled)
	}

	client.UseGlobalBlockedServices = plan.UseGlobalBlockedServices.ValueBool()

	// unpack nested attributes from plan
	var planBlockedServicesPauseScheduleClient scheduleModel
	d = plan.BlockedServicesPauseSchedule.As(ctx, &planBlockedServicesPauseScheduleClient, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	// defer to common function to populate schedule
	client.BlockedServicesSchedule = mapBlockedServicesPauseScheduleToAdgSchedule(ctx, planBlockedServicesPauseScheduleClient, diags)
	if diags.HasError() {
		return
	}

	if len(plan.BlockedServices.Elements()) > 0 {
		d = plan.BlockedServices.ElementsAs(ctx, &client.BlockedServices, false)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	}
	if len(plan.Upstreams.Elements()) > 0 {
		d = plan.Upstreams.ElementsAs(ctx, &client.Upstreams, false)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	}
	if len(plan.Tags.Elements()) > 0 {
		d = plan.Tags.ElementsAs(ctx, &client.Tags, false)
		diags.Append(d...)
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
