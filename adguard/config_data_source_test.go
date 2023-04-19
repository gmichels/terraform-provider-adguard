package adguard

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccConfigDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "adguard_config" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.adguard_config.test", "filtering_enabled", "true"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "filtering_update_interval", "12"),

					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.adguard_config.test", "id", "placeholder"),
				),
			},
		},
	})
}
