package adguard

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRewriteResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "adguard_rewrite" "test" {
  domain = "example.com"
  answer = "4.3.2.1"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_rewrite.test", "domain", "example.com"),
					resource.TestCheckResourceAttr("adguard_rewrite.test", "answer", "4.3.2.1"),
					resource.TestCheckResourceAttr("adguard_rewrite.test", "id", "example.com||4.3.2.1"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("adguard_rewrite.test", "last_updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "adguard_rewrite.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in AdGuard Home,
				// therefore there is no value for it during import
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "adguard_rewrite" "test" {
  domain = "example.com"
  answer = "2400:cb00:2049:1::a29f:1804"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_rewrite.test", "domain", "example.com"),
					resource.TestCheckResourceAttr("adguard_rewrite.test", "answer", "2400:cb00:2049:1::a29f:1804"),
					resource.TestCheckResourceAttr("adguard_rewrite.test", "id", "example.com||2400:cb00:2049:1::a29f:1804"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
