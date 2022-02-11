// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package ibm

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	st "github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/errors"
	"github.com/IBM-Cloud/power-go-client/helpers"
	"github.com/IBM-Cloud/power-go-client/power/client/p_cloud_service_d_h_c_p"
)

const (
	PIDhcpStatusBuilding = "Building"
	PIDhcpStatusActive   = "ACTIVE"
	PIDhcpDeleting       = "Deleting"
	PIDhcpDeleted        = "Deleted"
	PIDhcpId             = "dhcp_id"
	PIDhcpStatus         = "status"
	PIDhcpNetwork        = "network"
	PIDhcpLeases         = "leases"
	PIDhcpInstanceIp     = "instance_ip"
	PIDhcpInstanceMac    = "instance_mac"
)

func resourceIBMPIDhcp() *schema.Resource {
	return &schema.Resource{
		Create:   resourceIBMPIDhcpCreate,
		Read:     resourceIBMPIDhcpRead,
		Delete:   resourceIBMPIDhcpDelete,
		Importer: &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			// Required Attributes
			helpers.PICloudInstanceId: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "PI cloud instance ID",
				ForceNew:    true,
			},
			helpers.PICloudConnectionId: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The cloud connection uuid to connect with DHCP private network",
				ForceNew:    true,
			},
			//Computed Attributes
			PIDhcpId: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the DHCP Server",
			},
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

func resourceIBMPIDhcpCreate(d *schema.ResourceData, meta interface{}) error {
	sess, err := meta.(ClientSession).IBMPISession()
	if err != nil {
		return err
	}

	cloudInstanceID := d.Get(helpers.PICloudInstanceId).(string)

	body := &models.DHCPServerCreate{}
	if cloudConnectionID, ok := d.GetOk(helpers.PICloudConnectionId); ok {
		body.CloudConnectionID = cloudConnectionID.(string)
	}

	client := st.NewIBMPIDhcpClient(sess, cloudInstanceID)
	dhcpServer, err := client.Create(body)
	if err != nil {
		log.Printf("[DEBUG] create DHCP failed %v", err)
		return fmt.Errorf(errors.CreateDchpOperationFailed, cloudInstanceID, err)
	}

	d.SetId(fmt.Sprintf("%s/%s", cloudInstanceID, *dhcpServer.ID))

	_, err = waitForIBMPIDhcpStatus(context.TODO(), client, *dhcpServer.ID, cloudInstanceID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return fmt.Errorf(errors.CreateDchpOperationFailed, cloudInstanceID, err)
	}

	return resourceIBMPIDhcpRead(d, meta)
}

func resourceIBMPIDhcpRead(d *schema.ResourceData, meta interface{}) error {
	sess, err := meta.(ClientSession).IBMPISession()
	if err != nil {
		return err
	}

	parts, err := idParts(d.Id())
	if err != nil {
		return err
	}

	cloudInstanceID := parts[0]
	dhcpID := parts[1]

	client := st.NewIBMPIDhcpClient(sess, cloudInstanceID)
	dhcpServer, err := client.GetWithContext(context.TODO(), dhcpID, cloudInstanceID)
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

	d.Set(PIDhcpId, *dhcpServer.ID)
	d.Set(PIDhcpStatus, *dhcpServer.Status)

	if dhcpServer.Network != nil {
		d.Set(PIDhcpNetwork, dhcpServer.Network.ID)
	}
	if dhcpServer.Leases != nil {
		leaseList := make([]map[string]string, len(dhcpServer.Leases))
		for i, lease := range dhcpServer.Leases {
			leaseList[i] = map[string]string{
				PIDhcpInstanceIp:  *lease.InstanceIP,
				PIDhcpInstanceMac: *lease.InstanceMacAddress,
			}
		}
		d.Set(PIDhcpLeases, leaseList)
	}

	return nil
}
func resourceIBMPIDhcpDelete(d *schema.ResourceData, meta interface{}) error {
	sess, err := meta.(ClientSession).IBMPISession()
	if err != nil {
		return err
	}

	parts, err := idParts(d.Id())
	if err != nil {
		return err
	}

	cloudInstanceID := parts[0]
	dhcpID := parts[1]

	client := st.NewIBMPIDhcpClient(sess, cloudInstanceID)
	_, err = client.DeleteWithContext(context.TODO(), dhcpID, cloudInstanceID)
	if err != nil {
		switch err.(type) {
		case *p_cloud_service_d_h_c_p.PcloudDhcpDeleteNotFound:
			log.Printf("[DEBUG] dhcp does not exist %v", err)
			d.SetId("")
			return nil
		}
		log.Printf("[DEBUG] delete DHCP failed %v", err)
		return fmt.Errorf(errors.DeleteDhcpOperationFailed, dhcpID, err)
	}
	_, err = waitForIBMPIDhcpDeleted(context.TODO(), client, dhcpID, cloudInstanceID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

func waitForIBMPIDhcpStatus(ctx context.Context, client *st.IBMPIDhcpClient, dhcpID, cloudInstanceID string, timeout time.Duration) (interface{}, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{PIDhcpStatusBuilding},
		Target:  []string{PIDhcpStatusActive},
		Refresh: func() (interface{}, string, error) {
			dhcpServer, err := client.GetWithContext(ctx, dhcpID, cloudInstanceID)
			if err != nil {
				log.Printf("[DEBUG] get DHCP failed %v", err)
				return nil, "", fmt.Errorf(errors.GetDhcpOperationFailed, dhcpID, err)
			}
			if *dhcpServer.Status != PIDhcpStatusActive {
				return dhcpServer, PIDhcpStatusBuilding, nil
			}
			return dhcpServer, *dhcpServer.Status, nil
		},
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}
	return stateConf.WaitForState()
}

func waitForIBMPIDhcpDeleted(ctx context.Context, client *st.IBMPIDhcpClient, dhcpID, cloudInstanceID string, timeout time.Duration) (interface{}, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{PIDhcpDeleting},
		Target:  []string{PIDhcpDeleted},
		Refresh: func() (interface{}, string, error) {
			dhcpServer, err := client.GetWithContext(ctx, dhcpID, cloudInstanceID)
			if err != nil {
				log.Printf("[DEBUG] dhcp does not exist %v", err)
				return dhcpServer, PIDhcpDeleted, nil
			}
			return dhcpServer, PIDhcpDeleting, nil
		},
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}
	return stateConf.WaitForState()
}
