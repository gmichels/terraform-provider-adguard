package adguard

import (
	"context"
	"os"
	"regexp"
	"strconv"

	"github.com/gmichels/adguard-client-go"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// define a max Adguard Home client timeout
const MAX_TIMEOUT int = 60

// ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &adguardProvider{}
)

// New is a helper function to simplify provider server and testing implementation
func New() provider.Provider {
	return &adguardProvider{}
}

// adguardProvider is the provider implementation
type adguardProvider struct{}

// Metadata returns the provider type name
func (p *adguardProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "adguard"
}

// Schema defines the provider-level schema for configuration data
func (p *adguardProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "The hostname of the Adguard Home instance. Include the port if not on a standard HTTP/HTTPS port",
				Optional:    true,
			},
			"username": schema.StringAttribute{
				Description: "The username of the Adguard Home instance",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "The password of the Adguard Home instance",
				Optional:    true,
				Sensitive:   true,
			},
			"scheme": schema.StringAttribute{
				Description: "The HTTP scheme of the Adguard Home instance. Can be either `http` or `https` (default)",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(4, 5),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^https{0,1}$`),
						"must be either http or https",
					),
				},
			},
			"timeout": schema.Int64Attribute{
				Description: "The timeout (in seconds) for making requests to Adguard Home. Defaults to 10",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, int64(MAX_TIMEOUT)),
				},
			},
		},
	}
}

// adguardProviderModel maps provider schema data to a Go type
type adguardProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Scheme   types.String `tfsdk:"scheme"`
	Timeout  types.Int64  `tfsdk:"timeout"`
}

// Configure prepares an Adguard API client for data sources and resources
func (p *adguardProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config adguardProviderModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// if provided a configuration value for any of the attributes, it must be a known value
	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Adguard Home Host",
			"The provider cannot create the Adguard Home client as there is an unknown configuration value for the Adguard Home host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ADGUARD_HOST environment variable.",
		)
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown Adguard Home Username",
			"The provider cannot create the Adguard Home client as there is an unknown configuration value for the Adguard Home username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ADGUARD_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown Adguard Home Password",
			"The provider cannot create the Adguard Home client as there is an unknown configuration value for the Adguard Home password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ADGUARD_PASSWORD environment variable.",
		)
	}

	if config.Scheme.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("scheme"),
			"Unknown Adguard Home Scheme",
			"The provider cannot create the Adguard Home client as there is an unknown configuration value for the Adguard Home scheme. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ADGUARD_SCHEME environment variable.",
		)
	}

	if config.Timeout.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("timeout"),
			"Unknown Adguard Home Timeout",
			"The provider cannot create the Adguard Home client as there is an unknown configuration value for the Adguard Home timeout. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ADGUARD_TIMEOUT environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// default values to environment variables, but override with Terraform configuration value if set
	host := os.Getenv("ADGUARD_HOST")
	username := os.Getenv("ADGUARD_USERNAME")
	password := os.Getenv("ADGUARD_PASSWORD")
	scheme := os.Getenv("ADGUARD_SCHEME")
	// sanity check for scheme when provided via env variable
	if scheme != "" && scheme != "http" && scheme != "https" {
		resp.Diagnostics.AddAttributeError(
			path.Root("scheme"),
			"Unable to parse Adguard Home Scheme value",
			"The provider cannot create the Adguard Home client as the provided value for ADGUARD_SCHEME needs to be either `http` or `https`.")
		return

	}
	timeout_env := os.Getenv("ADGUARD_TIMEOUT")
	// sanity check for timeout when provided via env variable
	var timeout int
	if timeout_env != "" {
		var err error
		timeout, err = strconv.Atoi(timeout_env)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("timeout"),
				"Unable to parse Adguard Home Timeout value",
				"The provider cannot create the Adguard Home client as it was unable to parse the provided value for ADGUARD_TIMEOUT.")
			return
		} else if timeout <= 0 || timeout > MAX_TIMEOUT {
			resp.Diagnostics.AddAttributeError(
				path.Root("timeout"),
				"Unable to parse Adguard Home Timeout value",
				"The provider cannot create the Adguard Home client as the provided value for ADGUARD_TIMEOUT was outside the acceptable range (1, "+strconv.Itoa(MAX_TIMEOUT)+").")
			return
		}
	}

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	if !config.Scheme.IsNull() {
		scheme = config.Scheme.ValueString()
	}

	if !config.Timeout.IsNull() {
		timeout = int(config.Timeout.ValueInt64())
	}

	// if any of the expected configurations are missing, return errors with provider-specific guidance
	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Adguard Home Host",
			"The provider cannot create the Adguard Home client as there is a missing or empty value for the Adguard Home host. "+
				"Set the host value in the configuration or use the ADGUARD_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing Adguard Home Username",
			"The provider cannot create the Adguard Home client as there is a missing or empty value for the Adguard Home username. "+
				"Set the username value in the configuration or use the ADGUARD_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing Adguard Home Password",
			"The provider cannot create the Adguard Home client as there is a missing or empty value for the Adguard Home password. "+
				"Set the password value in the configuration or use the ADGUARD_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if scheme == "" {
		// default to https
		scheme = "https"
	}

	if timeout == 0 {
		// default to 10 seconds
		timeout = 10
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// create a new Adguard Home client using the configuration values
	client, err := adguard.NewClient(&host, &username, &password, &scheme, &timeout)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Adguard Home Client",
			"An unexpected error occurred when creating the Adguard Home client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Adguard Home Client Error: "+err.Error(),
		)
		return
	}

	// make the Adguard Home client available during DataSource and Resource type Configure methods
	resp.DataSourceData = client
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider
func (p *adguardProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewClientDataSource,
	}
}

// Resources defines the resources implemented in the provider
func (p *adguardProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewClientResource,
	}
}
