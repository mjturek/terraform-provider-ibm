// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package ibm

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccIBMCmCatalogDataSource(t *testing.T) {
	ResourceGroupID := os.Getenv("CATMGMT_RESOURCE_GROUP_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckIBMCmCatalogDataSourceConfig(ResourceGroupID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ibm_cm_catalog.cm_catalog_data", "label"),
					resource.TestCheckResourceAttrSet("data.ibm_cm_catalog.cm_catalog_data", "crn"),
					resource.TestCheckResourceAttrSet("data.ibm_cm_catalog.cm_catalog_data", "kind"),
					resource.TestCheckResourceAttrSet("ibm_cm_catalog.cm_catalog", "resource_group_id"),
				),
			},
			resource.TestStep{
				Config: testAccCheckIBMCmCatalogDataSourceConfigDefault(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ibm_cm_catalog.cm_catalog_data", "label"),
					resource.TestCheckResourceAttrSet("data.ibm_cm_catalog.cm_catalog_data", "crn"),
					resource.TestCheckResourceAttrSet("data.ibm_cm_catalog.cm_catalog_data", "kind"),
					resource.TestCheckResourceAttrSet("ibm_cm_catalog.cm_catalog", "resource_group_id"),
				),
			},
		},
	})
}

func testAccCheckIBMCmCatalogDataSourceConfig(resourceGroupID string) string {
	return fmt.Sprintf(`

		resource "ibm_cm_catalog" "cm_catalog" {
			label = "tf_test_datasource_catalog"
			short_description = "testing terraform provider with catalog"
			resource_group_id = "%s"
		}
		
		data "ibm_cm_catalog" "cm_catalog_data" {
			catalog_identifier = ibm_cm_catalog.cm_catalog.id
		}
		`, resourceGroupID)
}

func testAccCheckIBMCmCatalogDataSourceConfigDefault() string {
	return fmt.Sprintf(`

		resource "ibm_cm_catalog" "cm_catalog" {
			label = "tf_test_datasource_catalog"
			short_description = "testing terraform provider with catalog"
		}
		
		data "ibm_cm_catalog" "cm_catalog_data" {
			catalog_identifier = ibm_cm_catalog.cm_catalog.id
		}
		`)
}
