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
				Config: providerConfig + `data "adguard_client" "test" { name = "Amcrest-Left" }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.adguard_client.test", "name", "Amcrest-Left"),
					resource.TestCheckResourceAttr("data.adguard_client.test", "ids.#", "1"),
					resource.TestCheckResourceAttr("data.adguard_client.test", "ids.0", "172.16.128.53"),

					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.adguard_client.test", "id", "placeholder"),
				),
			},
		},
	})
}
