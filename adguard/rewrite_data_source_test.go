package adguard

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccRewriteDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "adguard_rewrite" "test" { domain = "example.org" }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.adguard_rewrite.test", "domain", "example.org"),
					resource.TestCheckResourceAttr("data.adguard_rewrite.test", "answer", "1.2.3.4"),

					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.adguard_rewrite.test", "id", "placeholder"),
				),
			},
		},
	})
}
