package adguard

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccListFilterResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Blacklist
			// Create and Read testing
			{
				Config: providerConfig + `
resource "adguard_list_filter" "test_blacklist" {
  name = "Test Blacklist Filter"
  url  = "https://adguardteam.github.io/HostlistsRegistry/assets/filter_4.txt"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_list_filter.test_blacklist", "name", "Test Blacklist Filter"),
					resource.TestCheckResourceAttr("adguard_list_filter.test_blacklist", "url", "https://adguardteam.github.io/HostlistsRegistry/assets/filter_4.txt"),
					resource.TestCheckResourceAttrSet("adguard_list_filter.test_blacklist", "last_updated"),
					resource.TestCheckResourceAttrSet("adguard_list_filter.test_blacklist", "id"),
					resource.TestCheckResourceAttrSet("adguard_list_filter.test_blacklist", "rules_count"),
					resource.TestCheckResourceAttr("adguard_list_filter.test_blacklist", "enabled", "true"),
					resource.TestCheckResourceAttr("adguard_list_filter.test_blacklist", "whitelist", "false"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "adguard_list_filter.test_blacklist",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "adguard_list_filter" "test_blacklist" {
  name = "Test Blacklist Filter Updated"
  url  = "https://adguardteam.github.io/HostlistsRegistry/assets/filter_5.txt"
  enabled   = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_list_filter.test_blacklist", "name", "Test Blacklist Filter Updated"),
					resource.TestCheckResourceAttr("adguard_list_filter.test_blacklist", "url", "https://adguardteam.github.io/HostlistsRegistry/assets/filter_5.txt"),
					resource.TestCheckResourceAttr("adguard_list_filter.test_blacklist", "enabled", "false"),
				),
			},
			// Delete testing automatically occurs in TestCase

			// Whitelist
			// Create and Read testing
			{
				Config: providerConfig + `
resource "adguard_list_filter" "test_whitelist" {
  name      = "Test Whitelist Filter"
  url       = "https://adguardteam.github.io/HostlistsRegistry/assets/filter_6.txt"
  enabled   = false
  whitelist = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_list_filter.test_whitelist", "name", "Test Whitelist Filter"),
					resource.TestCheckResourceAttr("adguard_list_filter.test_whitelist", "url", "https://adguardteam.github.io/HostlistsRegistry/assets/filter_6.txt"),
					resource.TestCheckResourceAttrSet("adguard_list_filter.test_whitelist", "last_updated"),
					resource.TestCheckResourceAttrSet("adguard_list_filter.test_whitelist", "id"),
					resource.TestCheckResourceAttrSet("adguard_list_filter.test_whitelist", "rules_count"),
					resource.TestCheckResourceAttr("adguard_list_filter.test_whitelist", "enabled", "false"),
					resource.TestCheckResourceAttr("adguard_list_filter.test_whitelist", "whitelist", "true"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "adguard_list_filter.test_whitelist",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "adguard_list_filter" "test_whitelist" {
  name      = "Test Whitelist Filter Updated"
  url       = "https://adguardteam.github.io/HostlistsRegistry/assets/filter_7.txt"
  enabled   = false
  whitelist = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_list_filter.test_whitelist", "name", "Test Whitelist Filter Updated"),
					resource.TestCheckResourceAttr("adguard_list_filter.test_whitelist", "url", "https://adguardteam.github.io/HostlistsRegistry/assets/filter_7.txt"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
