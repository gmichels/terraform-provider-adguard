package adguard

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccUserRulesResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "adguard_user_rules" "test" {
  rules = [
    "! line 1 bang comment",
		"# line 2 respond with 127.0.0.1 for localhost.org (but not for its subdomains)",
		"127.0.0.1 localhost.org",
		"# line 4 unblock access to unblocked.org and all its subdomains",
		"@@||unblocked.org^",
		"# line 6 block access to blocked.org and all its subdomains",
		"||blocked.org^"
	]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_user_rules.test", "rules.#", "7"),
					resource.TestCheckResourceAttrSet("adguard_user_rules.test", "last_updated"),
					resource.TestCheckResourceAttr("adguard_user_rules.test", "id", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "adguard_user_rules.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in AdGuard Home,
				// therefore there is no value for it during import
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "adguard_user_rules" "test" {
  rules = [
		"# line 1 unblock access to unblocked.org and all its subdomains",
		"@@||unblocked.org^",
		"# line 3 block access to blocked.org and all its subdomains",
		"||blocked.org^"
	]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_user_rules.test", "rules.#", "4"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
