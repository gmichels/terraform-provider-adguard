package adguard

import (
	"context"
	// "encoding/json"
	// "fmt"
	// "net/http"
	// "strconv"
	// "time"

	adg "github.com/gmichels/adguard-client-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceClient() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceClientRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"ids": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"use_global_settings": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"filtering_enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"parental_enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"safebrowsing_enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"blocked_services": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"upstreams": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"tags": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceClientRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// time.Sleep(30 * time.Second)

	c := meta.(*adg.ADG)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	clientName := d.Get("name").(string)
	client, err := c.GetClient(clientName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(client.Name)
	d.Set("ids", client.Ids)
	d.Set("use_global_settings", client.UseGlobalSettings)
	d.Set("filtering_enabled", client.FilteringEnabled)
	d.Set("parental_enabled", client.ParentalEnabled)
	d.Set("safebrowsing_enabled", client.SafebrowsingEnabled)
	d.Set("blocked_services", client.BlockedServices)
	d.Set("upstreams", client.Upstreams)
	d.Set("tags", client.Tags)

	return diags
}
