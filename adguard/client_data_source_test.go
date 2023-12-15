package adguard

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
					resource.TestCheckResourceAttr("data.adguard_client.test", "use_global_settings", "false"),
					resource.TestCheckResourceAttr("data.adguard_client.test", "safesearch.enabled", "true"),
					resource.TestCheckResourceAttr("data.adguard_client.test", "safesearch.services.#", "2"),
					resource.TestCheckResourceAttr("data.adguard_client.test", "safesearch.services.1", "youtube"),
					resource.TestCheckResourceAttr("data.adguard_client.test", "use_global_blocked_services", "false"),
					resource.TestCheckResourceAttr("data.adguard_client.test", "blocked_services.#", "3"),
					resource.TestCheckResourceAttr("data.adguard_client.test", "blocked_services.1", "instagram"),
					resource.TestCheckResourceAttr("data.adguard_client.test", "blocked_services_pause_schedule.time_zone", "America/Los_Angeles"),
					resource.TestCheckResourceAttr("data.adguard_client.test", "blocked_services_pause_schedule.mon.start", "08:00"),
					resource.TestCheckResourceAttr("data.adguard_client.test", "blocked_services_pause_schedule.mon.end", "17:45"),
					resource.TestCheckResourceAttr("data.adguard_client.test", "blocked_services_pause_schedule.thu.start", "08:00"),
					resource.TestCheckResourceAttr("data.adguard_client.test", "blocked_services_pause_schedule.thu.end", "17:45"),
					resource.TestCheckResourceAttr("data.adguard_client.test", "ignore_querylog", "false"),
					resource.TestCheckResourceAttr("data.adguard_client.test", "ignore_statistics", "true"),
					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.adguard_client.test", "id", "placeholder"),
				),
			},
		},
	})
}
