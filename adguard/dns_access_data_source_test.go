package adguard

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDnsAccessDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "adguard_dns_access" "test" { }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.adguard_dns_access.test", "allowed_clients.#", "0"),
					resource.TestCheckResourceAttr("data.adguard_dns_access.test", "disallowed_clients.#", "2"),
					resource.TestCheckResourceAttr("data.adguard_dns_access.test", "disallowed_clients.1", "10.0.0.0/8"),
					resource.TestCheckResourceAttr("data.adguard_dns_access.test", "blocked_hosts.#", "3"),
					resource.TestCheckResourceAttr("data.adguard_dns_access.test", "blocked_hosts.1", "id.server"),

					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.adguard_dns_access.test", "id", "placeholder"),
				),
			},
		},
	})
}
