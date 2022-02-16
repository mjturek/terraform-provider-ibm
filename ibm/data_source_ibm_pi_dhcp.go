// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package ibm

import (
	"fmt"

	"log"

	st "github.com/IBM-Cloud/power-go-client/clients/instance"

	"github.com/IBM-Cloud/power-go-client/errors"

	"github.com/IBM-Cloud/power-go-client/helpers"
	"github.com/IBM-Cloud/power-go-client/power/client/p_cloud_service_d_h_c_p"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceIBMPIDhcp() *schema.Resource {

	return &schema.Resource{
		Read: dataSourceIBMPIDhcpRead,
		Schema: map[string]*schema.Schema{
			helpers.PICloudInstanceId: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			PIDhcpId: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the DHCP Server",
			},
			// Computed Attributes
			PIDhcpStatus: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the DHCP Server",
			},
			PIDhcpNetwork: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The DHCP Server private network",
			},
			PIDhcpLeases: {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of DHCP Server PVM Instance leases",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						PIDhcpInstanceIp: {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The IP of the PVM Instance",
						},
						PIDhcpInstanceMac: {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The MAC Address of the PVM Instance",
						},
					},
				},
			},
		},
	}
}

func dataSourceIBMPIDhcpRead(d *schema.ResourceData, meta interface{}) error {
	sess, err := meta.(ClientSession).IBMPISession()
	if err != nil {
		return err
	}

	cloudInstanceID := d.Get(helpers.PICloudInstanceId).(string)
	dhcpID := d.Get(PIDhcpId).(string)

	client := st.NewIBMPIDhcpClient(sess, cloudInstanceID)
	dhcpServer, err := client.Get(dhcpID, cloudInstanceID)
	if err != nil {
		switch err.(type) {
		case *p_cloud_service_d_h_c_p.PcloudDhcpGetNotFound:
			log.Printf("[DEBUG] dhcp does not exist %v", err)
			d.SetId("")
			return nil
		}
		log.Printf("[DEBUG] get DHCP failed %v", err)
		return fmt.Errorf(errors.GetDhcpOperationFailed, dhcpID, err)
	}

	d.SetId(*dhcpServer.ID)
	d.Set(PIDhcpStatus, *dhcpServer.Status)
	dhcpNetwork := dhcpServer.Network
	if dhcpNetwork != nil {
		d.Set(PIDhcpNetwork, *dhcpNetwork.ID)
	}
	dhcpLeases := dhcpServer.Leases
	if dhcpLeases != nil {
		leaseList := make([]map[string]string, len(dhcpLeases))
		for i, lease := range dhcpLeases {
			leaseList[i] = map[string]string{
				PIDhcpInstanceIp:  *lease.InstanceIP,
				PIDhcpInstanceMac: *lease.InstanceMacAddress,
			}
		}
		d.Set(PIDhcpLeases, leaseList)
	}

	return nil
}