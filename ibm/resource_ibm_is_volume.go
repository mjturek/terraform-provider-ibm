// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package ibm

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/IBM/vpc-go-sdk/vpcv1"
	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	isVolumeName                 = "name"
	isVolumeProfileName          = "profile"
	isVolumeZone                 = "zone"
	isVolumeEncryptionKey        = "encryption_key"
	isVolumeEncryptionType       = "encryption_type"
	isVolumeCapacity             = "capacity"
	isVolumeIops                 = "iops"
	isVolumeCrn                  = "crn"
	isVolumeTags                 = "tags"
	isVolumeStatus               = "status"
	isVolumeStatusReasons        = "status_reasons"
	isVolumeStatusReasonsCode    = "code"
	isVolumeStatusReasonsMessage = "message"
	isVolumeDeleting             = "deleting"
	isVolumeDeleted              = "done"
	isVolumeProvisioning         = "provisioning"
	isVolumeProvisioningDone     = "done"
	isVolumeResourceGroup        = "resource_group"
	isVolumeSourceSnapshot       = "source_snapshot"
	isVolumeDeleteAllSnapshots   = "delete_all_snapshots"
	isVolumeBandwidth            = "bandwidth"
)

func resourceIBMISVolume() *schema.Resource {
	return &schema.Resource{
		Create:   resourceIBMISVolumeCreate,
		Read:     resourceIBMISVolumeRead,
		Update:   resourceIBMISVolumeUpdate,
		Delete:   resourceIBMISVolumeDelete,
		Exists:   resourceIBMISVolumeExists,
		Importer: &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		CustomizeDiff: customdiff.All(
			customdiff.Sequence(
				func(diff *schema.ResourceDiff, v interface{}) error {
					return resourceTagsCustomizeDiff(diff)
				},
			),
			customdiff.Sequence(
				func(diff *schema.ResourceDiff, v interface{}) error {
					return resourceVolumeValidate(diff)
				}),
		),

		Schema: map[string]*schema.Schema{

			isVolumeName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: InvokeValidator("ibm_is_volume", isVolumeName),
				Description:  "Volume name",
			},

			isVolumeProfileName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: InvokeValidator("ibm_is_volume", isVolumeProfileName),
				Description:  "Volume profile name",
			},

			isVolumeZone: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Zone name",
			},

			isVolumeEncryptionKey: {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Volume encryption key info",
			},

			isVolumeEncryptionType: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Volume encryption type info",
			},

			isVolumeCapacity: {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      100,
				ForceNew:     false,
				ValidateFunc: InvokeValidator("ibm_is_volume", isVolumeCapacity),
				Description:  "Volume capacity value",
			},
			isVolumeResourceGroup: {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Resource group name",
			},
			isVolumeIops: {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: InvokeValidator("ibm_is_volume", isVolumeIops),
				Description:  "IOPS value for the Volume",
			},
			isVolumeCrn: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "CRN value for the volume instance",
			},
			isVolumeStatus: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Volume status",
			},

			isVolumeStatusReasons: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						isVolumeStatusReasonsCode: {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "A snake case string succinctly identifying the status reason",
						},

						isVolumeStatusReasonsMessage: {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "An explanation of the status reason",
						},
					},
				},
			},

			isVolumeSourceSnapshot: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Identifier of the snapshot from which this volume was cloned",
			},
			isVolumeDeleteAllSnapshots: {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Deletes all snapshots created from this volume",
			},
			isVolumeTags: {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString, ValidateFunc: InvokeValidator("ibm_is_volume", "tag")},
				Set:         resourceIBMVPCHash,
				Description: "Tags for the volume instance",
			},

			ResourceControllerURL: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL of the IBM Cloud dashboard that can be used to explore and view details about this instance",
			},

			ResourceName: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the resource",
			},

			ResourceCRN: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The crn of the resource",
			},

			ResourceStatus: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the resource",
			},

			ResourceGroupName: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource group name in which resource is provisioned",
			},

			isVolumeBandwidth: {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The maximum bandwidth (in megabits per second) for the volume",
			},
		},
	}
}

func resourceIBMISVolumeValidator() *ResourceValidator {

	validateSchema := make([]ValidateSchema, 0)
	validateSchema = append(validateSchema,
		ValidateSchema{
			Identifier:                 isVolumeName,
			ValidateFunctionIdentifier: ValidateRegexpLen,
			Type:                       TypeString,
			Required:                   true,
			Regexp:                     `^([a-z]|[a-z][-a-z0-9]*[a-z0-9])$`,
			MinValueLength:             1,
			MaxValueLength:             63})

	validateSchema = append(validateSchema,
		ValidateSchema{
			Identifier:                 "tag",
			ValidateFunctionIdentifier: ValidateRegexpLen,
			Type:                       TypeString,
			Optional:                   true,
			Regexp:                     `^[A-Za-z0-9:_ .-]+$`,
			MinValueLength:             1,
			MaxValueLength:             128})
	validateSchema = append(validateSchema,
		ValidateSchema{
			Identifier:                 isVolumeProfileName,
			ValidateFunctionIdentifier: ValidateAllowedStringValue,
			Type:                       TypeString,
			Optional:                   true,
			AllowedValues:              "general-purpose, 5iops-tier, 10iops-tier, custom",
		})
	validateSchema = append(validateSchema,
		ValidateSchema{
			Identifier:                 isVolumeCapacity,
			ValidateFunctionIdentifier: IntBetween,
			Type:                       TypeInt,
			MinValue:                   "10",
			MaxValue:                   "16000"})
	validateSchema = append(validateSchema,
		ValidateSchema{
			Identifier:                 isVolumeIops,
			ValidateFunctionIdentifier: IntBetween,
			Type:                       TypeInt,
			MinValue:                   "100",
			MaxValue:                   "48000"})

	ibmISVolumeResourceValidator := ResourceValidator{ResourceName: "ibm_is_volume", Schema: validateSchema}
	return &ibmISVolumeResourceValidator
}

func resourceIBMISVolumeCreate(d *schema.ResourceData, meta interface{}) error {

	volName := d.Get(isVolumeName).(string)
	profile := d.Get(isVolumeProfileName).(string)
	zone := d.Get(isVolumeZone).(string)
	var volCapacity int64
	if capacity, ok := d.GetOk(isVolumeCapacity); ok {
		volCapacity = int64(capacity.(int))
	} else {
		volCapacity = 100
	}

	err := volCreate(d, meta, volName, profile, zone, volCapacity)
	if err != nil {
		return err
	}

	return resourceIBMISVolumeRead(d, meta)
}

func volCreate(d *schema.ResourceData, meta interface{}, volName, profile, zone string, volCapacity int64) error {
	sess, err := vpcClient(meta)
	if err != nil {
		return err
	}
	options := &vpcv1.CreateVolumeOptions{
		VolumePrototype: &vpcv1.VolumePrototype{
			Name:     &volName,
			Capacity: &volCapacity,
			Zone: &vpcv1.ZoneIdentity{
				Name: &zone,
			},
			Profile: &vpcv1.VolumeProfileIdentity{
				Name: &profile,
			},
		},
	}
	volTemplate := options.VolumePrototype.(*vpcv1.VolumePrototype)

	if key, ok := d.GetOk(isVolumeEncryptionKey); ok {
		encryptionKey := key.(string)
		volTemplate.EncryptionKey = &vpcv1.EncryptionKeyIdentity{
			CRN: &encryptionKey,
		}
	}

	if rgrp, ok := d.GetOk(isVolumeResourceGroup); ok {
		rg := rgrp.(string)
		volTemplate.ResourceGroup = &vpcv1.ResourceGroupIdentity{
			ID: &rg,
		}
	}

	if i, ok := d.GetOk(isVolumeIops); ok {
		iops := int64(i.(int))
		volTemplate.Iops = &iops
	}

	vol, response, err := sess.CreateVolume(options)
	if err != nil {
		return fmt.Errorf("[DEBUG] Create volume err %s\n%s", err, response)
	}
	d.SetId(*vol.ID)
	log.Printf("[INFO] Volume : %s", *vol.ID)
	_, err = isWaitForVolumeAvailable(sess, d.Id(), d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return err
	}
	v := os.Getenv("IC_ENV_TAGS")
	if _, ok := d.GetOk(isVolumeTags); ok || v != "" {
		oldList, newList := d.GetChange(isVolumeTags)
		err = UpdateTagsUsingCRN(oldList, newList, meta, *vol.CRN)
		if err != nil {
			log.Printf(
				"Error on create of resource Volume (%s) tags: %s", d.Id(), err)
		}
	}
	return nil
}

func resourceIBMISVolumeRead(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()
	err := volGet(d, meta, id)
	if err != nil {
		return err
	}
	return nil
}

func volGet(d *schema.ResourceData, meta interface{}, id string) error {
	sess, err := vpcClient(meta)
	if err != nil {
		return err
	}
	options := &vpcv1.GetVolumeOptions{
		ID: &id,
	}
	vol, response, err := sess.GetVolume(options)
	if err != nil {
		if response != nil && response.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error getting Volume (%s): %s\n%s", id, err, response)
	}
	d.SetId(*vol.ID)
	d.Set(isVolumeName, *vol.Name)
	d.Set(isVolumeProfileName, *vol.Profile.Name)
	d.Set(isVolumeZone, *vol.Zone.Name)
	if vol.EncryptionKey != nil {
		d.Set(isVolumeEncryptionKey, vol.EncryptionKey.CRN)
	}
	if vol.Encryption != nil {
		d.Set(isVolumeEncryptionType, vol.Encryption)
	}
	d.Set(isVolumeIops, *vol.Iops)
	d.Set(isVolumeCapacity, *vol.Capacity)
	d.Set(isVolumeCrn, *vol.CRN)
	if vol.SourceSnapshot != nil {
		d.Set(isVolumeSourceSnapshot, *vol.SourceSnapshot.ID)
	}
	d.Set(isVolumeStatus, *vol.Status)
	d.Set(isVolumeBandwidth, int(*vol.Bandwidth))
	//set the status reasons
	if vol.StatusReasons != nil {
		statusReasonsList := make([]map[string]interface{}, 0)
		for _, sr := range vol.StatusReasons {
			currentSR := map[string]interface{}{}
			if sr.Code != nil && sr.Message != nil {
				currentSR[isVolumeStatusReasonsCode] = *sr.Code
				currentSR[isVolumeStatusReasonsMessage] = *sr.Message
				statusReasonsList = append(statusReasonsList, currentSR)
			}
		}
		d.Set(isVolumeStatusReasons, statusReasonsList)
	}
	tags, err := GetTagsUsingCRN(meta, *vol.CRN)
	if err != nil {
		log.Printf(
			"Error on get of resource vpc volume (%s) tags: %s", d.Id(), err)
	}
	d.Set(isVolumeTags, tags)
	controller, err := getBaseController(meta)
	if err != nil {
		return err
	}
	d.Set(ResourceControllerURL, controller+"/vpc-ext/storage/storageVolumes")
	d.Set(ResourceName, *vol.Name)
	d.Set(ResourceCRN, *vol.CRN)
	d.Set(ResourceStatus, *vol.Status)
	if vol.ResourceGroup != nil {
		d.Set(ResourceGroupName, vol.ResourceGroup.Name)
		d.Set(isVolumeResourceGroup, *vol.ResourceGroup.ID)
	}
	return nil
}

func resourceIBMISVolumeUpdate(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()
	name := ""
	hasNameChanged := false
	delete := false

	if delete_all_snapshots, ok := d.GetOk(isVolumeDeleteAllSnapshots); ok && delete_all_snapshots.(bool) {
		delete = true
	}

	if d.HasChange(isVolumeName) {
		name = d.Get(isVolumeName).(string)
		hasNameChanged = true
	}

	err := volUpdate(d, meta, id, name, hasNameChanged, delete)
	if err != nil {
		return err
	}
	return resourceIBMISVolumeRead(d, meta)
}

func volUpdate(d *schema.ResourceData, meta interface{}, id, name string, hasNameChanged, delete bool) error {
	sess, err := vpcClient(meta)
	if err != nil {
		return err
	}
	var capacity int64
	if delete {
		deleteAllSnapshots(sess, id)
	}

	// tags update
	if d.HasChange(isVolumeTags) {
		options := &vpcv1.GetVolumeOptions{
			ID: &id,
		}
		vol, response, err := sess.GetVolume(options)
		if err != nil {
			return fmt.Errorf("Error getting Volume : %s\n%s", err, response)
		}
		oldList, newList := d.GetChange(isVolumeTags)
		err = UpdateTagsUsingCRN(oldList, newList, meta, *vol.CRN)
		if err != nil {
			log.Printf(
				"Error on update of resource vpc volume (%s) tags: %s", id, err)
		}
	}

	options := &vpcv1.UpdateVolumeOptions{
		ID: &id,
	}

	//name update
	volumeNamePatchModel := &vpcv1.VolumePatch{}
	if hasNameChanged {
		volumeNamePatchModel.Name = &name
		volumeNamePatch, err := volumeNamePatchModel.AsPatch()
		if err != nil {
			return fmt.Errorf("Error calling asPatch for volumeNamePatch: %s", err)
		}
		options.VolumePatch = volumeNamePatch
		_, _, err = sess.UpdateVolume(options)
		_, err = isWaitForVolumeAvailable(sess, d.Id(), d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return err
		}
	}

	// profile/ iops update
	if d.HasChange(isVolumeProfileName) || d.HasChange(isVolumeIops) {
		volumeProfilePatchModel := &vpcv1.VolumePatch{}
		volId := d.Id()
		getvoloptions := &vpcv1.GetVolumeOptions{
			ID: &volId,
		}
		vol, response, err := sess.GetVolume(getvoloptions)
		if err != nil || vol == nil {
			return fmt.Errorf("Error retrieving Volume (%s) details: %s\n%s", volId, err, response)
		}
		if vol.VolumeAttachments == nil || len(vol.VolumeAttachments) < 1 {
			return fmt.Errorf("Error updating Volume profile/iops because the specified volume %s is not attached to a virtual server instance ", volId)
		}
		volAtt := &vol.VolumeAttachments[0]
		insId := *volAtt.Instance.ID
		getinsOptions := &vpcv1.GetInstanceOptions{
			ID: &insId,
		}
		instance, response, err := sess.GetInstance(getinsOptions)
		if err != nil || instance == nil {
			return fmt.Errorf("Error retrieving Instance (%s) to which the volume (%s) is attached : %s\n%s", insId, volId, err, response)
		}
		if instance != nil && *instance.Status != "running" {
			actiontype := "start"
			createinsactoptions := &vpcv1.CreateInstanceActionOptions{
				InstanceID: &insId,
				Type:       &actiontype,
			}
			_, response, err = sess.CreateInstanceAction(createinsactoptions)
			if err != nil {
				return fmt.Errorf("Error starting Instance (%s) to which the volume (%s) is attached  : %s\n%s", insId, volId, err, response)
			}
			_, err = isWaitForInstanceAvailable(sess, insId, d.Timeout(schema.TimeoutCreate), d)
			if err != nil {
				return err
			}
		}
		if d.HasChange(isVolumeProfileName) {
			profile := d.Get(isVolumeProfileName).(string)
			volumeProfilePatchModel.Profile = &vpcv1.VolumeProfileIdentity{
				Name: &profile,
			}
		} else if d.HasChange(isVolumeIops) {
			profile := d.Get(isVolumeProfileName).(string)
			volumeProfilePatchModel.Profile = &vpcv1.VolumeProfileIdentity{
				Name: &profile,
			}
			iops := int64(d.Get(isVolumeIops).(int))
			volumeProfilePatchModel.Iops = &iops
		}

		volumeProfilePatch, err := volumeProfilePatchModel.AsPatch()
		if err != nil {
			return fmt.Errorf("Error calling asPatch for VolumeProfilePatch: %s", err)
		}
		options.VolumePatch = volumeProfilePatch
		_, response, err = sess.UpdateVolume(options)
		_, err = isWaitForVolumeAvailable(sess, d.Id(), d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return err
		}
	}

	// capacity update
	if d.HasChange(isVolumeCapacity) {
		id := d.Id()
		getvolumeoptions := &vpcv1.GetVolumeOptions{
			ID: &id,
		}
		vol, response, err := sess.GetVolume(getvolumeoptions)
		if err != nil {
			if response != nil && response.StatusCode == 404 {
				d.SetId("")
				return nil
			}
			return fmt.Errorf("Error Getting Volume (%s): %s\n%s", id, err, response)
		}
		if vol.VolumeAttachments == nil || len(vol.VolumeAttachments) == 0 || *vol.VolumeAttachments[0].ID == "" {
			return fmt.Errorf("Error volume capacity can't be updated since volume %s is not attached to any instance for VolumePatch", id)
		}
		insId := vol.VolumeAttachments[0].Instance.ID
		getinsOptions := &vpcv1.GetInstanceOptions{
			ID: insId,
		}
		instance, response, err := sess.GetInstance(getinsOptions)
		if err != nil || instance == nil {
			return fmt.Errorf("Error retrieving Instance (%s) : %s\n%s", *insId, err, response)
		}
		if instance != nil && *instance.Status != "running" {
			actiontype := "start"
			createinsactoptions := &vpcv1.CreateInstanceActionOptions{
				InstanceID: insId,
				Type:       &actiontype,
			}
			_, response, err = sess.CreateInstanceAction(createinsactoptions)
			if err != nil {
				return fmt.Errorf("Error starting Instance (%s) : %s\n%s", *insId, err, response)
			}
			_, err = isWaitForInstanceAvailable(sess, *insId, d.Timeout(schema.TimeoutCreate), d)
			if err != nil {
				return err
			}
		}
		capacity = int64(d.Get(isVolumeCapacity).(int))
		volumeCapacityPatchModel := &vpcv1.VolumePatch{}
		volumeCapacityPatchModel.Capacity = &capacity

		volumeCapacityPatch, err := volumeCapacityPatchModel.AsPatch()
		if err != nil {
			return fmt.Errorf("Error calling asPatch for volumeCapacityPatch: %s", err)
		}
		options.VolumePatch = volumeCapacityPatch
		_, response, err = sess.UpdateVolume(options)
		if err != nil {
			return fmt.Errorf("Error updating vpc volume: %s\n%s", err, response)
		}
		_, err = isWaitForVolumeAvailable(sess, d.Id(), d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceIBMISVolumeDelete(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	err := volDelete(d, meta, id)
	if err != nil {
		return err
	}
	return nil
}

func volDelete(d *schema.ResourceData, meta interface{}, id string) error {
	sess, err := vpcClient(meta)
	if err != nil {
		return err
	}

	getvoloptions := &vpcv1.GetVolumeOptions{
		ID: &id,
	}
	volDetails, response, err := sess.GetVolume(getvoloptions)
	if err != nil {
		if response != nil && response.StatusCode == 404 {
			return nil
		}
		return fmt.Errorf("Error getting Volume (%s): %s\n%s", id, err, response)
	}

	if volDetails.VolumeAttachments != nil {
		for _, volAtt := range volDetails.VolumeAttachments {
			deleteVolumeAttachment := &vpcv1.DeleteInstanceVolumeAttachmentOptions{
				InstanceID: volAtt.Instance.ID,
				ID:         volAtt.ID,
			}
			_, err := sess.DeleteInstanceVolumeAttachment(deleteVolumeAttachment)
			if err != nil {
				return fmt.Errorf("Error while removing volume attachment %q for instance %s: %q", *volAtt.ID, *volAtt.Instance.ID, err)
			}
			_, err = isWaitForInstanceVolumeDetached(sess, d, d.Id(), *volAtt.ID)
			if err != nil {
				return err
			}

		}
	}

	options := &vpcv1.DeleteVolumeOptions{
		ID: &id,
	}
	response, err = sess.DeleteVolume(options)
	if err != nil {
		return fmt.Errorf("Error deleting Volume : %s\n%s", err, response)
	}
	_, err = isWaitForVolumeDeleted(sess, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func isWaitForVolumeDeleted(vol *vpcv1.VpcV1, id string, timeout time.Duration) (interface{}, error) {
	log.Printf("Waiting for  (%s) to be deleted.", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"retry", isVolumeDeleting},
		Target:     []string{"done", ""},
		Refresh:    isVolumeDeleteRefreshFunc(vol, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	return stateConf.WaitForState()
}

func isVolumeDeleteRefreshFunc(vol *vpcv1.VpcV1, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		volgetoptions := &vpcv1.GetVolumeOptions{
			ID: &id,
		}
		vol, response, err := vol.GetVolume(volgetoptions)
		if err != nil {
			if response != nil && response.StatusCode == 404 {
				return vol, isVolumeDeleted, nil
			}
			return vol, "", fmt.Errorf("Error getting Volume: %s\n%s", err, response)
		}
		return vol, isVolumeDeleting, err
	}
}

func resourceIBMISVolumeExists(d *schema.ResourceData, meta interface{}) (bool, error) {

	id := d.Id()

	exists, err := volExists(d, meta, id)
	return exists, err
}

func volExists(d *schema.ResourceData, meta interface{}, id string) (bool, error) {
	sess, err := vpcClient(meta)
	if err != nil {
		return false, err
	}
	options := &vpcv1.GetVolumeOptions{
		ID: &id,
	}
	_, response, err := sess.GetVolume(options)
	if err != nil {
		if response != nil && response.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("Error getting Volume: %s\n%s", err, response)
	}
	return true, nil
}

func isWaitForVolumeAvailable(client *vpcv1.VpcV1, id string, timeout time.Duration) (interface{}, error) {
	log.Printf("Waiting for Volume (%s) to be available.", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"retry", isVolumeProvisioning},
		Target:     []string{isVolumeProvisioningDone, ""},
		Refresh:    isVolumeRefreshFunc(client, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	return stateConf.WaitForState()
}

func isVolumeRefreshFunc(client *vpcv1.VpcV1, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		volgetoptions := &vpcv1.GetVolumeOptions{
			ID: &id,
		}
		vol, response, err := client.GetVolume(volgetoptions)
		if err != nil {
			return nil, "", fmt.Errorf("Error getting volume: %s\n%s", err, response)
		}

		if *vol.Status == "available" {
			return vol, isVolumeProvisioningDone, nil
		}

		return vol, isVolumeProvisioning, nil
	}
}

func deleteAllSnapshots(sess *vpcv1.VpcV1, id string) error {
	delete_all_snapshots := new(vpcv1.DeleteSnapshotsOptions)
	delete_all_snapshots.SourceVolumeID = &id
	response, err := sess.DeleteSnapshots(delete_all_snapshots)
	if err != nil {
		return fmt.Errorf("Error deleting snapshots from volume %s\n%s", err, response)
	}
	return nil
}
