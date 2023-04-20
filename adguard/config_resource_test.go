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
  filtering_enabled         = false
  filtering_update_interval = 1
	safebrowsing_enabled      = true
	parental_enabled          = true
	safesearch_enabled        = true
	safesearch_services       = ["bing", "youtube", "google"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_config.test", "filtering_enabled", "false"),
					resource.TestCheckResourceAttr("adguard_config.test", "filtering_update_interval", "1"),
					resource.TestCheckResourceAttr("adguard_config.test", "safebrowsing_enabled", "true"),
					resource.TestCheckResourceAttr("adguard_config.test", "parental_enabled", "true"),
					resource.TestCheckResourceAttr("adguard_config.test", "safesearch_enabled", "true"),
					resource.TestCheckResourceAttr("adguard_config.test", "safesearch_services.#", "3"),
					resource.TestCheckResourceAttr("adguard_config.test", "safesearch_services.1", "google"),
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
	filtering_update_interval = 72
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_config.test", "filtering_enabled", "true"),
					resource.TestCheckResourceAttr("adguard_config.test", "filtering_update_interval", "72"),
					resource.TestCheckResourceAttr("adguard_config.test", "safebrowsing_enabled", "false"),
					resource.TestCheckResourceAttr("adguard_config.test", "parental_enabled", "false"),
					resource.TestCheckResourceAttr("adguard_config.test", "safesearch_enabled", "false"),
					resource.TestCheckResourceAttr("adguard_config.test", "safesearch_services.#", "6"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
