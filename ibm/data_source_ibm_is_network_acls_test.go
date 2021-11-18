// Copyright IBM Corp. 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package ibm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccIBMIsNetworkAclsDataSourceBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMIsNetworkAclsDataSourceConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ibm_is_network_acls.is_network_acls", "id"),
					resource.TestCheckResourceAttrSet("data.ibm_is_network_acls.is_network_acls", "network_acls.#"),
				),
			},
		},
	})
}

func testAccCheckIBMIsNetworkAclsDataSourceConfigBasic() string {
	return fmt.Sprintf(`
		data "ibm_is_network_acls" "is_network_acls" {
		}
	`)
}
