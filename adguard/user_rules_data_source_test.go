package adguard

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserRulesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `
data "adguard_user_rules" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.adguard_user_rules.test", "rules.#", "7"),
					resource.TestCheckResourceAttr("data.adguard_user_rules.test", "rules.1", "||blocked.org^"),
					resource.TestCheckResourceAttrSet("data.adguard_user_rules.test", "id"),
				),
			},
		},
	})
}
