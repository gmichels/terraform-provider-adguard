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
  name = "AdGuard DNS filter"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.adguard_list_filter.test_blacklist", "url", "https://adguardteam.github.io/HostlistsRegistry/assets/filter_1.txt"),
					resource.TestCheckResourceAttr("data.adguard_list_filter.test_blacklist", "enabled", "true"),
					resource.TestCheckResourceAttr("data.adguard_list_filter.test_blacklist", "whitelist", "false"),
					resource.TestCheckResourceAttrSet("data.adguard_list_filter.test_blacklist", "rules_count"),
					resource.TestCheckResourceAttr("data.adguard_list_filter.test_blacklist", "id", "1"),
				),
			},
			{
				Config: providerConfig + `
data "adguard_list_filter" "test_whitelist" {
  name      = "Test Allow List"
  whitelist = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.adguard_list_filter.test_whitelist", "url", "https://adguardteam.github.io/HostlistsRegistry/assets/filter_3.txt"),
					resource.TestCheckResourceAttr("data.adguard_list_filter.test_whitelist", "enabled", "false"),
					resource.TestCheckResourceAttr("data.adguard_list_filter.test_whitelist", "whitelist", "true"),
					resource.TestCheckResourceAttrSet("data.adguard_list_filter.test_whitelist", "rules_count"),
					resource.TestCheckResourceAttr("data.adguard_list_filter.test_whitelist", "id", "3"),
				),
			},
		},
	})
}
