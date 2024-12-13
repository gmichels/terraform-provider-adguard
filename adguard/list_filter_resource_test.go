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
  name = "Test Blacklist Filter Resource"
  url  = "https://raw.githubusercontent.com/gmichels/terraform-provider-adguard/refs/heads/main/assets/list_filter_3.txt"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_list_filter.test_blacklist", "name", "Test Blacklist Filter Resource"),
					resource.TestCheckResourceAttr("adguard_list_filter.test_blacklist", "url", "https://raw.githubusercontent.com/gmichels/terraform-provider-adguard/refs/heads/main/assets/list_filter_3.txt"),
					resource.TestCheckResourceAttrSet("adguard_list_filter.test_blacklist", "last_updated"),
					resource.TestCheckResourceAttrSet("adguard_list_filter.test_blacklist", "id"),
					resource.TestCheckResourceAttr("adguard_list_filter.test_blacklist", "rules_count", "5"),
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
  name = "Test Blacklist Filter Resource Updated"
  url  = "https://raw.githubusercontent.com/gmichels/terraform-provider-adguard/refs/heads/main/assets/list_filter_4.txt"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_list_filter.test_blacklist", "name", "Test Blacklist Filter Resource Updated"),
					resource.TestCheckResourceAttr("adguard_list_filter.test_blacklist", "url", "https://raw.githubusercontent.com/gmichels/terraform-provider-adguard/refs/heads/main/assets/list_filter_4.txt"),
					resource.TestCheckResourceAttr("adguard_list_filter.test_blacklist", "rules_count", "8"),
				),
			},
			// Delete testing automatically occurs in TestCase

			// Whitelist
			// Create and Read testing
			{
				Config: providerConfig + `
resource "adguard_list_filter" "test_whitelist" {
  name      = "Test Whitelist Filter Resource"
  url       = "https://raw.githubusercontent.com/gmichels/terraform-provider-adguard/refs/heads/main/assets/list_filter_5.txt"
  whitelist = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_list_filter.test_whitelist", "name", "Test Whitelist Filter Resource"),
					resource.TestCheckResourceAttr("adguard_list_filter.test_whitelist", "url", "https://raw.githubusercontent.com/gmichels/terraform-provider-adguard/refs/heads/main/assets/list_filter_5.txt"),
					resource.TestCheckResourceAttrSet("adguard_list_filter.test_whitelist", "last_updated"),
					resource.TestCheckResourceAttrSet("adguard_list_filter.test_whitelist", "id"),
					resource.TestCheckResourceAttr("adguard_list_filter.test_whitelist", "rules_count", "9"),
					resource.TestCheckResourceAttr("adguard_list_filter.test_whitelist", "enabled", "true"),
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
  name      = "Test Whitelist Filter Resource Updated"
  url       = "https://raw.githubusercontent.com/gmichels/terraform-provider-adguard/refs/heads/main/assets/list_filter_6.txt"
  enabled   = false
  whitelist = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_list_filter.test_whitelist", "name", "Test Whitelist Filter Resource Updated"),
					resource.TestCheckResourceAttr("adguard_list_filter.test_whitelist", "url", "https://raw.githubusercontent.com/gmichels/terraform-provider-adguard/refs/heads/main/assets/list_filter_6.txt"),
					resource.TestCheckResourceAttr("adguard_list_filter.test_whitelist", "rules_count", "0"),
					resource.TestCheckResourceAttr("adguard_list_filter.test_whitelist", "enabled", "false"),
					resource.TestCheckResourceAttr("adguard_list_filter.test_whitelist", "whitelist", "true"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
