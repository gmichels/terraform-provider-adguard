package adguard

import (
	"context"
	"regexp"
	"time"

	"github.com/gmichels/adguard-client-go"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource                = &dnsAccessResource{}
	_ resource.ResourceWithConfigure   = &dnsAccessResource{}
	_ resource.ResourceWithImportState = &dnsAccessResource{}
)

// dnsAccessResource is the resource implementation
type dnsAccessResource struct {
	adg *adguard.ADG
}

// dnsAccessResourceModel maps DNS Access List schema data
type dnsAccessResourceModel struct {
	ID                types.String `tfsdk:"id"`
	LastUpdated       types.String `tfsdk:"last_updated"`
	AllowedClients    types.List   `tfsdk:"allowed_clients"`
	DisallowedClients types.List   `tfsdk:"disallowed_clients"`
	BlockedHosts      types.List   `tfsdk:"blocked_hosts"`
}

// NewDnsAccessResource is a helper function to simplify the provider implementation
func NewDnsAccessResource() resource.Resource {
	return &dnsAccessResource{}
}

// Metadata returns the resource type name
func (r *dnsAccessResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_access"
}

// Schema defines the schema for the resource
func (r *dnsAccessResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internal identifier for this dnsAccess",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the dnsAccess",
				Computed:    true,
			},
			"allowed_clients": schema.ListAttribute{
				Description: "The allowlist of clients: IP addresses, CIDRs, or ClientIDs",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(
							regexp.MustCompile(`^[a-z0-9/.:-]+$`),
							"must be an IP address/CIDR or only contain numbers, lowercase letters, and hyphens",
						),
					),
				},
			},
			"disallowed_clients": schema.ListAttribute{
				Description: "The blocklist of clients: IP addresses, CIDRs, or ClientIDs",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
			},
			"blocked_hosts": schema.ListAttribute{
				Description: "Disallowed domains",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
				Default: listdefault.StaticValue(
					types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("version.bind"),
							types.StringValue("id.server"),
							types.StringValue("hostname.bind"),
						},
					),
				),
			},
		},
	}
}

// Configure adds the provider configured DNS Access List to the resource
func (r *dnsAccessResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.adg = req.ProviderData.(*adguard.ADG)
}

// Create creates the resource and sets the initial Terraform state
func (r *dnsAccessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// retrieve values from plan
	var plan dnsAccessResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// instantiate empty DNS Access List for storing plan data
	var dnsAccess adguard.AccessList

	// populate DNS Access List from plan
	if len(plan.AllowedClients.Elements()) > 0 {
		diags = plan.AllowedClients.ElementsAs(ctx, &dnsAccess.AllowedClients, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		plan.AllowedClients, diags = types.ListValueFrom(ctx, types.StringType, []string{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if len(plan.DisallowedClients.Elements()) > 0 {
		diags = plan.DisallowedClients.ElementsAs(ctx, &dnsAccess.DisallowedClients, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		plan.DisallowedClients, diags = types.ListValueFrom(ctx, types.StringType, []string{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if len(plan.BlockedHosts.Elements()) > 0 {
		diags = plan.BlockedHosts.ElementsAs(ctx, &dnsAccess.BlockedHosts, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		plan.BlockedHosts, diags = types.ListValueFrom(ctx, types.StringType, []string{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// set DNS Access List using plan
	_, err := r.adg.SetAccess(dnsAccess)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating DNS Access List",
			"Could not create DNS access list, unexpected error: "+err.Error(),
		)
		return
	}

	// response sent by AdGuard Home is the same as the sent payload,
	// just add missing attributes for state
	// there can be only one entry DNS Access List, so hardcode the ID as 1
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
func (r *dnsAccessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// get current state
	var state dnsAccessResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get refreshed DNS Access List value from AdGuard Home
	dnsAccess, err := r.adg.GetAccess()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading AdGuard Home DNS Access List",
			"Could not read AdGuard Home DNS access list with ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// overwrite DNS DNS Access List with refreshed state
	state.AllowedClients, diags = types.ListValueFrom(ctx, types.StringType, dnsAccess.AllowedClients)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.DisallowedClients, diags = types.ListValueFrom(ctx, types.StringType, dnsAccess.DisallowedClients)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.BlockedHosts, diags = types.ListValueFrom(ctx, types.StringType, dnsAccess.BlockedHosts)
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
func (r *dnsAccessResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// retrieve values from plan
	var plan dnsAccessResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// instantiate empty DNS Access List for storing plan data
	var dnsAccess adguard.AccessList

	// populate DNS Access List from plan
	if len(plan.AllowedClients.Elements()) > 0 {
		diags = plan.AllowedClients.ElementsAs(ctx, &dnsAccess.AllowedClients, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		dnsAccess.AllowedClients = []string{}
		plan.AllowedClients = types.ListNull(types.StringType)
		plan.AllowedClients, diags = types.ListValueFrom(ctx, types.StringType, []string{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if len(plan.DisallowedClients.Elements()) > 0 {
		diags = plan.DisallowedClients.ElementsAs(ctx, &dnsAccess.DisallowedClients, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		dnsAccess.DisallowedClients = []string{}
		plan.DisallowedClients, diags = types.ListValueFrom(ctx, types.StringType, []string{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if len(plan.BlockedHosts.Elements()) > 0 {
		diags = plan.BlockedHosts.ElementsAs(ctx, &dnsAccess.BlockedHosts, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		dnsAccess.BlockedHosts = []string{}
		plan.BlockedHosts, diags = types.ListValueFrom(ctx, types.StringType, []string{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// set DNS Config using plan
	_, err := r.adg.SetAccess(dnsAccess)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating DNS Access List",
			"Could not create DNS access list, unexpected error: "+err.Error(),
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
func (r *dnsAccessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// there is no "real" delete for DNS Access List, so this means "restore defaults"

	// instantiate empty DNS Access List for storing default values
	var dnsAccess adguard.AccessList

	// populate DNS Access List with default values
	dnsAccess.AllowedClients = []string{}
	dnsAccess.DisallowedClients = []string{}
	dnsAccess.BlockedHosts = []string{"version.bind", "id.server", "hostname.bind"}

	// set default values in DNS Access List
	_, err := r.adg.SetAccess(dnsAccess)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting DNS Access List",
			"Could not delete DNS access list, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *dnsAccessResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
