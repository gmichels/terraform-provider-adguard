package adguard

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccClientDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "adguard_client" "test" { name = "Test Client Data Source" }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.adguard_client.test", "name", "Test Client Data Source"),
					resource.TestCheckResourceAttr("data.adguard_client.test", "ids.#", "1"),
					resource.TestCheckResourceAttr("data.adguard_client.test", "ids.0", "192.168.100.100"),
					resource.TestCheckResourceAttr("data.adguard_client.test", "tags.0", "device_other"),
					resource.TestCheckResourceAttr("data.adguard_client.test", "use_global_settings", "true"),
					resource.TestCheckResourceAttr("data.adguard_client.test", "use_global_blocked_services", "true"),

					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.adguard_client.test", "id", "placeholder"),
				),
			},
		},
	})
}
