// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package ibm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccIBMISRegionsDataSource_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckIBMISRegionsDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ibm_is_regions.testacc_ds_regions", "regions.0.name"),
					resource.TestCheckResourceAttrSet("data.ibm_is_regions.testacc_ds_regions", "regions.0.status"),
					resource.TestCheckResourceAttrSet("data.ibm_is_regions.testacc_ds_regions", "regions.0.endpoint"),
					resource.TestCheckResourceAttrSet("data.ibm_is_regions.testacc_ds_regions", "regions.0.href"),
				),
			},
		},
	})
}

func testAccCheckIBMISRegionsDataSourceConfig() string {
	return fmt.Sprintf(`

		data "ibm_is_regions" "testacc_ds_regions" {
		}`)

}
