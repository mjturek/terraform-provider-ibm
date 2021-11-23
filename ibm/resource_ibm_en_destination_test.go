// Copyright IBM Corp. 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package ibm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	en "github.com/IBM/event-notifications-go-admin-sdk/eventnotificationsv1"
)

func TestAccIBMEnDestinationAllArgs(t *testing.T) {
	var config en.Destination
	name := fmt.Sprintf("tf_name_%d", acctest.RandIntRange(10, 100))
	instanceName := fmt.Sprintf("tf_name_%d", acctest.RandIntRange(10, 100))
	description := fmt.Sprintf("tf_description_%d", acctest.RandIntRange(10, 100))
	newName := fmt.Sprintf("tf_name_%d", acctest.RandIntRange(10, 100))
	newDescription := fmt.Sprintf("tf_description_%d", acctest.RandIntRange(10, 100))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIBMEnDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMEnDestinationConfig(instanceName, name, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIBMEnDestinationExists("ibm_en_destination.en_destination_resource_1", config),
					resource.TestCheckResourceAttr("ibm_en_destination.en_destination_resource_1", "name", name),
					resource.TestCheckResourceAttr("ibm_en_destination.en_destination_resource_1", "type", "webhook"),
					resource.TestCheckResourceAttr("ibm_en_destination.en_destination_resource_1", "description", description),
				),
			},
			{
				Config: testAccCheckIBMEnDestinationConfig(instanceName, newName, newDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ibm_en_destination.en_destination_resource_1", "name", newName),
					resource.TestCheckResourceAttr("ibm_en_destination.en_destination_resource_1", "type", "webhook"),
					resource.TestCheckResourceAttr("ibm_en_destination.en_destination_resource_1", "description", newDescription),
				),
			},
			{
				ResourceName:      "ibm_en_destination.en_destination_resource_1",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckIBMEnDestinationConfig(instanceName, name, description string) string {
	return fmt.Sprintf(`
	resource "ibm_resource_instance" "en_destination_resource" {
		name     = "%s"
		location = "us-south"
		plan     = "standard"
		service  = "event-notifications"
	}
	
	resource "ibm_en_destination" "en_destination_resource_1" {
		instance_guid = ibm_resource_instance.en_destination_resource.guid
		name        = "%s"
		type        = "webhook"
		description = "%s"
		config {
			params {
				verb = "POST"
				url  = "https://demo.webhook.com"
			}
		}
	}
	`, instanceName, name, description)
}

func testAccCheckIBMEnDestinationExists(n string, obj en.Destination) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		enClient, err := testAccProvider.Meta().(ClientSession).EventNotificationsApiV1()
		if err != nil {
			return err
		}

		options := &en.GetDestinationOptions{}

		parts, err := sepIdParts(rs.Primary.ID, "/")
		if err != nil {
			return err
		}

		options.SetInstanceID(parts[0])
		options.SetID(parts[1])

		result, _, err := enClient.GetDestination(options)
		if err != nil {
			return err
		}

		obj = *result
		return nil
	}
}

func testAccCheckIBMEnDestinationDestroy(s *terraform.State) error {
	enClient, err := testAccProvider.Meta().(ClientSession).EventNotificationsApiV1()
	if err != nil {
		return err
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "en_destination_resource_1" {
			continue
		}

		options := &en.GetDestinationOptions{}

		parts, err := sepIdParts(rs.Primary.ID, "/")
		if err != nil {
			return err
		}

		options.SetInstanceID(parts[0])
		options.SetID(parts[1])

		// Try to find the key
		_, response, err := enClient.GetDestination(options)

		if err == nil {
			return fmt.Errorf("en_destination still exists: %s", rs.Primary.ID)
		} else if response.StatusCode != 404 {
			return fmt.Errorf("Error checking for en_destination (%s) has been destroyed: %s", rs.Primary.ID, err)
		}
	}

	return nil
}
