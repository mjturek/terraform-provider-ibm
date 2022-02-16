// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package ibm

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	st "github.com/IBM-Cloud/power-go-client/clients/instance"
)

func TestAccIBMPICloudConnectionbasic(t *testing.T) {
	name := fmt.Sprintf("tf-cloudconnection-%d", acctest.RandIntRange(10, 100))
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIBMPICloudConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPICloudConnectionConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIBMPICloudConnectionExists("ibm_pi_cloud_connection.cloud_connection"),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cloud_connection",
						"pi_cloud_connection_name", name),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cloud_connection",
						"pi_cloud_connection_speed", "50"),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cloud_connection",
						"pi_cloud_connection_global_routing", "false"),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cloud_connection",
						"pi_cloud_connection_metered", "false"),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cloud_connection",
						"pi_cloud_connection_networks.#", "0"),
				),
			},
			{
				Config: testAccCheckIBMPICloudConnectionUpdateConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIBMPICloudConnectionExists("ibm_pi_cloud_connection.cloud_connection"),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cloud_connection",
						"pi_cloud_connection_name", name),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cloud_connection",
						"pi_cloud_connection_speed", "100"),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cloud_connection",
						"pi_cloud_connection_global_routing", "true"),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cloud_connection",
						"pi_cloud_connection_metered", "true"),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cloud_connection",
						"pi_cloud_connection_networks.#", "1"),
				),
			},
		},
	})
}
func testAccCheckIBMPICloudConnectionDestroy(s *terraform.State) error {
	sess, err := testAccProvider.Meta().(ClientSession).IBMPISession()
	if err != nil {
		return err
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ibm_pi_cloud_connection" {
			continue
		}
		parts, err := idParts(rs.Primary.ID)
		if err != nil {
			return err
		}
		cloudInstanceID := parts[0]
		cloudConnectionID := parts[1]
		client := st.NewIBMPICloudConnectionClient(sess, cloudInstanceID)
		_, err = client.Get(cloudConnectionID, cloudInstanceID)
		if err == nil {
			return fmt.Errorf("Cloud Connection still exists: %s", rs.Primary.ID)
		}
	}
	return nil
}
func testAccCheckIBMPICloudConnectionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return errors.New("No Record ID is set")
		}

		sess, err := testAccProvider.Meta().(ClientSession).IBMPISession()
		if err != nil {
			return err
		}
		parts, err := idParts(rs.Primary.ID)
		if err != nil {
			return err
		}
		cloudInstanceID := parts[0]
		cloudConnectionID := parts[1]
		client := st.NewIBMPICloudConnectionClient(sess, cloudInstanceID)

		_, err = client.Get(cloudConnectionID, cloudInstanceID)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckIBMPICloudConnectionConfig(name string) string {
	return fmt.Sprintf(`
	  resource "ibm_pi_cloud_connection" "cloud_connection" {
		pi_cloud_instance_id		= "%[1]s"
		pi_cloud_connection_name	= "%[2]s"
		pi_cloud_connection_speed	= 50

		//pi_cloud_connection_status			=
		//pi_cloud_connection_user_ip_address	= ""
		//pi_cloud_connection_ibm_ip_address	= ""
		//pi_cloud_connection_port			= ""
		//pi_cloud_connection_id				= ""
		//pi_cloud_connection_classic			= ""
		//enabled   = ""
		//gre_source_address = ""
		//gre_destination_address   = ""
		//pi_cloud_connection_vpc              = ""
		//enabled       = ""
		//vpc_name          = ""
		//vpc_id            = ""
	  }
	`, pi_cloud_instance_id, name)
}

func testAccCheckIBMPICloudConnectionUpdateConfig(name string) string {
	return fmt.Sprintf(`
	  resource "ibm_pi_cloud_connection" "cloud_connection" {
		pi_cloud_instance_id				= "%[1]s"
		pi_cloud_connection_name			= "%[2]s"
		pi_cloud_connection_speed			= 100
		pi_cloud_connection_metered			= true
		pi_cloud_connection_global_routing	= true
		pi_cloud_connection_networks		= ["25cdcee0-4c39-47ab-b55e-50b8389ace1a"]
	  }
	//   resource "ibm_pi_network" "network1" {
	// 	pi_cloud_instance_id	= "%[1]s"
	// 	pi_network_name			= "%[2]s"
	// 	pi_network_type         = "vlan"
	// 	pi_cidr         		= "192.112.111.0/24"
	//   }
	`, pi_cloud_instance_id, name)
}

func TestAccIBMPICloudConnectionNetworks(t *testing.T) {
	name := fmt.Sprintf("tf-cloudconnection-%d", acctest.RandIntRange(10, 100))
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIBMPICloudConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPICloudConnectionNetworkConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIBMPICloudConnectionExists("ibm_pi_cloud_connection.cc_network"),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cc_network",
						"pi_cloud_connection_name", name),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cc_network",
						"pi_cloud_connection_networks.#", "1"),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cc_network",
						"pi_cloud_connection_networks.0", "25cdcee0-4c39-47ab-b55e-50b8389ace1a"),
				),
			},
			{
				Config: testAccCheckIBMPICloudConnectionNetworkUpdateConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIBMPICloudConnectionExists("ibm_pi_cloud_connection.cc_network"),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cc_network",
						"pi_cloud_connection_name", name),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cc_network",
						"pi_cloud_connection_networks.#", "1"),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cc_network",
						"pi_cloud_connection_networks.0", "b144138d-f418-4013-b87a-fbf1d10cded7"),
				),
			},
		},
	})
}

func testAccCheckIBMPICloudConnectionNetworkConfig(name string) string {
	return fmt.Sprintf(`
	  resource "ibm_pi_cloud_connection" "cc_network" {
		pi_cloud_instance_id		= "%[1]s"
		pi_cloud_connection_name	= "%[2]s"
		pi_cloud_connection_speed	= 1000
		pi_cloud_connection_networks	= ["25cdcee0-4c39-47ab-b55e-50b8389ace1a"]
	  }
	//   resource "ibm_pi_network" "network1" {
	// 	pi_cloud_instance_id	= "%[1]s"
	// 	pi_network_name			= "%[2]s_net1"
	// 	pi_network_type         = "vlan"
	// 	pi_cidr         		= "192.112.112.0/24"
	//   }
	`, pi_cloud_instance_id, name)
}

func testAccCheckIBMPICloudConnectionNetworkUpdateConfig(name string) string {
	return fmt.Sprintf(`
	  resource "ibm_pi_cloud_connection" "cc_network" {
		pi_cloud_instance_id				= "%[1]s"
		pi_cloud_connection_name			= "%[2]s"
		pi_cloud_connection_speed			= 1000
		pi_cloud_connection_networks		= ["b144138d-f418-4013-b87a-fbf1d10cded7"]
	  }
	//   resource "ibm_pi_network" "network2" {
	// 	pi_cloud_instance_id	= "%[1]s"
	// 	pi_network_name			= "%[2]s_net2"
	// 	pi_network_type         = "vlan"
	// 	pi_cidr         		= "192.112.113.0/24"
	//   }
	`, pi_cloud_instance_id, name)
}

func TestAccIBMPICloudConnectionClassicAndVPC(t *testing.T) {
	name := fmt.Sprintf("tf-cloudconnection-%d", acctest.RandIntRange(10, 100))
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIBMPICloudConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPICloudConnectionVPCConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIBMPICloudConnectionExists("ibm_pi_cloud_connection.cc_network"),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cc_network",
						"pi_cloud_connection_name", name),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cc_network",
						"pi_cloud_connection_networks.#", "0"),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cc_network",
						"pi_cloud_connection_classic_enabled", "false"),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cc_network",
						"pi_cloud_connection_vpc_enabled", "true"),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cc_network",
						"pi_cloud_connection_vpc_crns.#", "1"),
				),
			},
			{
				Config: testAccCheckIBMPICloudConnectionClassicConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIBMPICloudConnectionExists("ibm_pi_cloud_connection.cc_network"),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cc_network",
						"pi_cloud_connection_name", name),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cc_network",
						"pi_cloud_connection_networks.#", "0"),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cc_network",
						"pi_cloud_connection_classic_enabled", "true"),
					resource.TestCheckResourceAttr("ibm_pi_cloud_connection.cc_network",
						"pi_cloud_connection_vpc_enabled", "false"),
				),
			},
		},
	})
}

func testAccCheckIBMPICloudConnectionClassicConfig(name string) string {
	return fmt.Sprintf(`
	  resource "ibm_pi_cloud_connection" "cc_network" {
		pi_cloud_instance_id		= "%[1]s"
		pi_cloud_connection_name	= "%[2]s"
		pi_cloud_connection_speed	= 50
		pi_cloud_connection_classic_enabled	= true
	  }
	`, pi_cloud_instance_id, name)
}

func testAccCheckIBMPICloudConnectionVPCConfig(name string) string {
	return fmt.Sprintf(`
	  resource "ibm_pi_cloud_connection" "cc_network" {
		pi_cloud_instance_id		= "%[1]s"
		pi_cloud_connection_name	= "%[2]s"
		pi_cloud_connection_speed	= 50
		pi_cloud_connection_vpc_enabled	= true
		pi_cloud_connection_vpc_crns = ["crn:v1:bluemix:public:is:us-south:a/d9cec80d0adc400ead8e2076afe26698::vpc:r006-6486cf73-451d-4d44-b90d-83dff504cbed"]
	  }
	`, pi_cloud_instance_id, name)
}