package adguard

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
				ImportStateVerifyIgnore: []string{
					"last_updated",
					// time zone implementation by the AdGuard Home provides inconsistent results,
					// which render verifying its import complicated
					"blocked_services_pause_schedule.time_zone",
				},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "adguard_client" "test" {
	name = "Test Client"
	ids  = ["192.168.100.15", "test-client", "another-test-client"]
	safesearch = {
		enabled  = true
		services = ["bing"]
	}
	blocked_services = ["reddit", "9gag"]
	blocked_services_pause_schedule = {
		time_zone = "America/New_York"
		sat = {
			start = "11:37"
			end   = "13:15"
		}
		sun = {
			start = "10:00"
			end   = "19:41"
		}
	}
	ignore_querylog = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_client.test", "ids.#", "3"),
					resource.TestCheckResourceAttr("adguard_client.test", "ids.2", "another-test-client"),
					resource.TestCheckResourceAttr("adguard_client.test", "safesearch.enabled", "true"),
					resource.TestCheckResourceAttr("adguard_client.test", "safesearch.services.#", "1"),
					resource.TestCheckResourceAttr("adguard_client.test", "safesearch.services.0", "bing"),
					resource.TestCheckResourceAttr("adguard_client.test", "blocked_services.#", "2"),
					resource.TestCheckResourceAttr("adguard_client.test", "blocked_services.0", "9gag"),
					resource.TestCheckResourceAttr("adguard_client.test", "blocked_services_pause_schedule.time_zone", "America/New_York"),
					resource.TestCheckResourceAttr("adguard_client.test", "blocked_services_pause_schedule.sat.start", "11:37"),
					resource.TestCheckResourceAttr("adguard_client.test", "blocked_services_pause_schedule.sat.end", "13:15"),
					resource.TestCheckResourceAttr("adguard_client.test", "blocked_services_pause_schedule.sun.start", "10:00"),
					resource.TestCheckResourceAttr("adguard_client.test", "blocked_services_pause_schedule.sun.end", "19:41"),
					resource.TestCheckResourceAttr("adguard_client.test", "ignore_querylog", "true"),
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
