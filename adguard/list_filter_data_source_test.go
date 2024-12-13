package adguard

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccListFilterDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `
data "adguard_list_filter" "test_blacklist" {
  name = "Test Blocklist Datasource"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.adguard_list_filter.test_blacklist", "url", "https://raw.githubusercontent.com/gmichels/terraform-provider-adguard/refs/heads/main/assets/list_filter_1.txt"),
					resource.TestCheckResourceAttr("data.adguard_list_filter.test_blacklist", "enabled", "true"),
					resource.TestCheckResourceAttr("data.adguard_list_filter.test_blacklist", "whitelist", "false"),
					resource.TestCheckResourceAttr("data.adguard_list_filter.test_blacklist", "rules_count", "13"),
					resource.TestCheckResourceAttr("data.adguard_list_filter.test_blacklist", "id", "1"),
				),
			},
			{
				Config: providerConfig + `
data "adguard_list_filter" "test_whitelist" {
  name      = "Test Whitelist Datasource"
  whitelist = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.adguard_list_filter.test_whitelist", "url", "https://raw.githubusercontent.com/gmichels/terraform-provider-adguard/refs/heads/main/assets/list_filter_2.txt"),
					resource.TestCheckResourceAttr("data.adguard_list_filter.test_whitelist", "enabled", "false"),
					resource.TestCheckResourceAttr("data.adguard_list_filter.test_whitelist", "whitelist", "true"),
					resource.TestCheckResourceAttr("data.adguard_list_filter.test_whitelist", "rules_count", "0"),
					resource.TestCheckResourceAttr("data.adguard_list_filter.test_whitelist", "id", "3"),
				),
			},
		},
	})
}
