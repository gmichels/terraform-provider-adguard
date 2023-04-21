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
					resource.TestCheckResourceAttr("data.adguard_config.test", "filtering.enabled", "true"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "filtering.update_interval", "24"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "safebrowsing.enabled", "false"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "parental.enabled", "false"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "safesearch.enabled", "false"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "safesearch.services.#", "6"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "safesearch.services.0", "bing"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "safesearch.services.4", "yandex"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "querylog.enabled", "true"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "querylog.anonymize_client_ip", "false"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "querylog.interval", "4"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "querylog.ignored.#", "3"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "querylog.ignored.0", "abc.com"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "querylog.ignored.1", "example.com"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "querylog.ignored.2", "test.com"),
					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.adguard_config.test", "id", "placeholder"),
				),
			},
		},
	})
}
