// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package ibm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccIBMPIDhcpDataSourceBasic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPIDhcpDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ibm_pi_dhcp.dhcp", "id"),
				),
			},
		},
	})
}

func testAccCheckIBMPIDhcpDataSourceConfig() string {
	return fmt.Sprintf(`
		data "ibm_pi_dhcp" "dhcp" {
			pi_cloud_instance_id = "%s"
			pi_dhcp_id = "%s"
		}`, pi_cloud_instance_id, pi_dhcp_id)
}