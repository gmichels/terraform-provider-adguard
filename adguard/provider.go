package adguard

import (
	"context"

	"github.com/gmichels/adguard-client-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ADGUARD_HOST", nil),
			},
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ADGUARD_USERNAME", nil),
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ADGUARD_PASSWORD", nil),
			},
			"scheme": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "https",
				DefaultFunc: schema.EnvDefaultFunc("ADGUARD_SCHEME", nil),
			},
			"timeout": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     10,
				DefaultFunc: schema.EnvDefaultFunc("ADGUARD_TIMEOUT", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{},
		DataSourcesMap: map[string]*schema.Resource{
			"adguard_client": dataSourceClient(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	host := d.Get("host").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	scheme := d.Get("scheme").(string)
	timeout := d.Get("timeout").(int)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c, err := adguard.NewClient(&host, &username, &password, &scheme, &timeout)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return c, diags
}
