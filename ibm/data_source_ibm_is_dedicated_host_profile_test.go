// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package ibm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccIbmIsDedicatedHostProfileDataSourceBasic(t *testing.T) {

	resName := "data.ibm_is_dedicated_host_profile.dhprofile"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckIbmIsDedicatedHostProfileDataSourceConfigBasic(dedicatedHostProfileName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "name", dedicatedHostProfileName),
					resource.TestCheckResourceAttrSet(resName, "class"),
					resource.TestCheckResourceAttrSet(resName, "family"),
				),
			},
		},
	})
}

func testAccCheckIbmIsDedicatedHostProfileDataSourceConfigBasic(profile string) string {
	return fmt.Sprintf(`
	 
	 data "ibm_is_dedicated_host_profile" "dhprofile" {
		 name = "%s"
	 }
	 `, profile)
}
