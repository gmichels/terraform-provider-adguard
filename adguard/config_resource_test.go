package adguard

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccConfigResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "adguard_config" "test" {
	filtering = {
		update_interval = 1
	}
	safesearch = {
		enabled  = true
		services = ["bing", "youtube", "google"]
	}
	querylog = {
		enabled             = false
		interval            = 8
		anonymize_client_ip = true
		ignored             = ["test2.com", "example2.com"]
	}
	stats = {
		enabled  = false
		interval = 2
		ignored  = ["test3.net", "example4.com"]
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_config.test", "filtering.enabled", "true"),
					resource.TestCheckResourceAttr("adguard_config.test", "filtering.update_interval", "1"),
					resource.TestCheckResourceAttr("adguard_config.test", "safebrowsing.enabled", "false"),
					resource.TestCheckResourceAttr("adguard_config.test", "parental_control.enabled", "false"),
					resource.TestCheckResourceAttr("adguard_config.test", "safesearch.enabled", "true"),
					resource.TestCheckResourceAttr("adguard_config.test", "safesearch.services.#", "3"),
					resource.TestCheckResourceAttr("adguard_config.test", "safesearch.services.1", "google"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.enabled", "false"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.interval", "8"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.anonymize_client_ip", "true"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.ignored.#", "2"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.ignored.0", "example2.com"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.ignored.1", "test2.com"),
					resource.TestCheckResourceAttr("adguard_config.test", "stats.enabled", "false"),
					resource.TestCheckResourceAttr("adguard_config.test", "stats.interval", "2"),
					resource.TestCheckResourceAttr("adguard_config.test", "stats.ignored.#", "2"),
					resource.TestCheckResourceAttr("adguard_config.test", "stats.ignored.0", "example4.com"),
					resource.TestCheckResourceAttr("adguard_config.test", "stats.ignored.1", "test3.net"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("adguard_config.test", "id"),
					resource.TestCheckResourceAttrSet("adguard_config.test", "last_updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "adguard_config.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in AdGuard Home,
				// therefore there is no value for it during import
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "adguard_config" "test" {
	filtering = {
		update_interval = 72
	}
	querylog = {
		ignored = ["test2.com", "example2.com", "abc2.com"]
	}
	stats = {
		ignored  = ["test9.com", "example15.com", "abc5.com"]
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_config.test", "filtering.enabled", "true"),
					resource.TestCheckResourceAttr("adguard_config.test", "filtering.update_interval", "72"),
					resource.TestCheckResourceAttr("adguard_config.test", "safebrowsing.enabled", "false"),
					resource.TestCheckResourceAttr("adguard_config.test", "parental_control.enabled", "false"),
					resource.TestCheckResourceAttr("adguard_config.test", "safesearch.enabled", "false"),
					resource.TestCheckResourceAttr("adguard_config.test", "safesearch.services.#", "6"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.enabled", "true"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.interval", "24"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.anonymize_client_ip", "false"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.ignored.#", "3"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.ignored.0", "abc2.com"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.ignored.1", "example2.com"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.ignored.2", "test2.com"),
					resource.TestCheckResourceAttr("adguard_config.test", "stats.enabled", "true"),
					resource.TestCheckResourceAttr("adguard_config.test", "stats.interval", "24"),
					resource.TestCheckResourceAttr("adguard_config.test", "stats.ignored.#", "3"),
					resource.TestCheckResourceAttr("adguard_config.test", "stats.ignored.0", "abc5.com"),
					resource.TestCheckResourceAttr("adguard_config.test", "stats.ignored.1", "example15.com"),
					resource.TestCheckResourceAttr("adguard_config.test", "stats.ignored.2", "test9.com"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
