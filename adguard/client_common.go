package adguard

import (
	"context"

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
	SafesearchEnabled            types.Bool   `tfsdk:"safesearch_enabled"` // deprecated
	SafeSearch                   types.Object `tfsdk:"safesearch"`
	UseGlobalBlockedServices     types.Bool   `tfsdk:"use_global_blocked_services"`
	BlockedServicesPauseSchedule types.Object `tfsdk:"blocked_services_pause_schedule"`
	BlockedServices              types.Set    `tfsdk:"blocked_services"`
	Upstreams                    types.List   `tfsdk:"upstreams"`
	Tags                         types.Set    `tfsdk:"tags"`
	IgnoreQuerylog               types.Bool   `tfsdk:"ignore_querylog"`
	IgnoreStatistics             types.Bool   `tfsdk:"ignore_statistics"`
}

// common `Read` function for both data source and resource
func (o *clientCommonModel) Read(ctx context.Context, adg adguard.ADG, currState *clientCommonModel, diags *diag.Diagnostics, rtype string) {
	// retrieve client info
	client, err := adg.GetClient(currState.Name.ValueString())
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
			"No client with name `"+currState.Name.ValueString()+"` exists in AdGuard Home.",
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
	// deprecated, copy from other value
	o.SafesearchEnabled = stateSafeSearchClient.Enabled
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
	stateBlockedServicesPauseScheduleClient := mapBlockedServicesScheduleDays(ctx, &client.BlockedServicesSchedule)

	// need special handling for timezone in resource due to inconsistent API response for `Local`
	if rtype == "resource" {
		// last updated will exist on create operation, null on import operation
		if !currState.LastUpdated.IsNull() {
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

	// if we got here, all went fine
}
