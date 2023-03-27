package adguard

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccClientResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "adguard_client" "test" {
  name = "Test Client"
  ids  = ["192.168.100.15", "test-client"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_client.test", "name", "Test Client"),
					resource.TestCheckResourceAttr("adguard_client.test", "ids.#", "2"),
					resource.TestCheckResourceAttr("adguard_client.test", "ids.1", "test-client"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("adguard_client.test", "id"),
					resource.TestCheckResourceAttrSet("adguard_client.test", "last_updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "adguard_client.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in AdGuard Home,
				// therefore there is no value for it during import
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "adguard_client" "test" {
  name = "Test Client"
  ids  = ["192.168.100.15", "test-client", "another-test-client"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_client.test", "ids.#", "3"),
					resource.TestCheckResourceAttr("adguard_client.test", "ids.2", "another-test-client"),
				),
			},
			// Update client name testing (requires recreate)
			{
				Config: providerConfig + `
resource "adguard_client" "test" {
  name = "Test Client Name Updated"
  ids  = ["192.168.100.15", "test-client", "another-test-client"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_client.test", "name", "Test Client Name Updated"),
					resource.TestCheckResourceAttr("adguard_client.test", "ids.#", "3"),
					resource.TestCheckResourceAttr("adguard_client.test", "ids.2", "another-test-client"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
