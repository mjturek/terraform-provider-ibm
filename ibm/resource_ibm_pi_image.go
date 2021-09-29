// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package ibm

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/IBM-Cloud/bluemix-go/bmxerror"
	st "github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/errors"
	"github.com/IBM-Cloud/power-go-client/helpers"
	"github.com/IBM-Cloud/power-go-client/power/client/p_cloud_images"
	"github.com/IBM-Cloud/power-go-client/power/models"
)

func resourceIBMPIImage() *schema.Resource {
	return &schema.Resource{
		Create:   resourceIBMPIImageCreate,
		Read:     resourceIBMPIImageRead,
		Update:   resourceIBMPIImageUpdate,
		Delete:   resourceIBMPIImageDelete,
		Exists:   resourceIBMPIImageExists,
		Importer: &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			helpers.PICloudInstanceId: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "PI cloud instance ID",
			},
			helpers.PIImageName: {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Image name",
				DiffSuppressFunc: applyOnce,
			},
			helpers.PIImageId: {
				Type:             schema.TypeString,
				Optional:         true,
				ExactlyOneOf:     []string{helpers.PIImageId, helpers.PIImageBucketName},
				Description:      "Image id",
				DiffSuppressFunc: applyOnce,
				ConflictsWith:    []string{helpers.PIImageBucketName},
				ForceNew:         true,
			},

			// COS import variables
			helpers.PIImageBucketName: {
				Type:          schema.TypeString,
				Optional:      true,
				ExactlyOneOf:  []string{helpers.PIImageId, helpers.PIImageBucketName},
				Description:   "Cloud Object Storage bucket name; bucket-name[/optional/folder]",
				ConflictsWith: []string{helpers.PIImageId},
				RequiredWith:  []string{helpers.PIImageBucketRegion, helpers.PIImageBucketFileName},
				ForceNew:      true,
			},
			helpers.PIImageBucketAccess: {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Indicates if the bucket has public or private access",
				Default:       "public",
				ValidateFunc:  validateAllowedStringValue([]string{"public", "private"}),
				ConflictsWith: []string{helpers.PIImageId},
				ForceNew:      true,
			},
			helpers.PIImageAccessKey: {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Cloud Object Storage access key; required for buckets with private access",
				ForceNew:     true,
				RequiredWith: []string{helpers.PIImageSecretKey},
			},
			helpers.PIImageSecretKey: {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Cloud Object Storage secret key; required for buckets with private access",
				ForceNew:     true,
				RequiredWith: []string{helpers.PIImageAccessKey},
			},
			helpers.PIImageBucketRegion: {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Cloud Object Storage region",
				ConflictsWith: []string{helpers.PIImageId},
				RequiredWith:  []string{helpers.PIImageBucketName},
				ForceNew:      true,
			},
			helpers.PIImageBucketFileName: {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Cloud Object Storage image filename",
				ConflictsWith: []string{helpers.PIImageId},
				RequiredWith:  []string{helpers.PIImageBucketName},
				ForceNew:      true,
			},
			helpers.PIImageStorageType: {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Type of storage",
				RequiredWith: []string{helpers.PIImageBucketName},
				ForceNew:     true,
			},

			// Computed Attribute
			"image_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Image ID",
			},
		},
	}
}

func resourceIBMPIImageCreate(d *schema.ResourceData, meta interface{}) error {
	sess, err := meta.(ClientSession).IBMPISession()
	if err != nil {
		log.Printf("Failed to get the session")
		return err
	}

	cloudInstanceID := d.Get(helpers.PICloudInstanceId).(string)
	imageName := d.Get(helpers.PIImageName).(string)

	client := st.NewIBMPIImageClient(sess, cloudInstanceID)
	// image copy
	if v, ok := d.GetOk(helpers.PIImageId); ok {
		imageid := v.(string)
		imageResponse, err := client.Create(imageName, imageid, cloudInstanceID)
		if err != nil {
			return err
		}

		IBMPIImageID := imageResponse.ImageID
		d.SetId(fmt.Sprintf("%s/%s", cloudInstanceID, *IBMPIImageID))

		_, err = isWaitForIBMPIImageAvailable(client, *IBMPIImageID, d.Timeout(schema.TimeoutCreate), cloudInstanceID)
		if err != nil {
			log.Printf("[DEBUG]  err %s", err)
			return err
		}
	}

	// COS image import
	if v, ok := d.GetOk(helpers.PIImageBucketName); ok {
		bucketName := v.(string)
		bucketImageFileName := d.Get(helpers.PIImageBucketFileName).(string)
		bucketRegion := d.Get(helpers.PIImageBucketRegion).(string)
		bucketAccess := d.Get(helpers.PIImageBucketAccess).(string)
		storageType := d.Get(helpers.PIImageStorageType).(string)

		body := &models.CreateCosImageImportJob{
			ImageName:     &imageName,
			BucketName:    &bucketName,
			BucketAccess:  &bucketAccess,
			ImageFilename: &bucketImageFileName,
			Region:        &bucketRegion,
			StorageType:   storageType,
		}

		if v, ok := d.GetOk(helpers.PIImageAccessKey); ok {
			body.AccessKey = v.(string)
		}
		if v, ok := d.GetOk(helpers.PIImageSecretKey); ok {
			body.SecretKey = v.(string)
		}

		imageResponse, err := client.CreateCosImage(body, cloudInstanceID)
		if err != nil {
			return err
		}

		jobClient := st.NewIBMPIJobClient(sess, cloudInstanceID)
		_, err = waitForIBMPIJobCompleted(jobClient, *imageResponse.ID, cloudInstanceID, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return fmt.Errorf(errors.CreateImageOperationFailed, cloudInstanceID, err)
		}

		// Once the job is completed find by name
		image, err := client.Get(imageName, cloudInstanceID)
		if err != nil {
			return err
		}
		d.SetId(fmt.Sprintf("%s/%s", cloudInstanceID, *image.ImageID))
	}

	return resourceIBMPIImageRead(d, meta)
}

func resourceIBMPIImageRead(d *schema.ResourceData, meta interface{}) error {
	sess, err := meta.(ClientSession).IBMPISession()
	if err != nil {
		return err
	}

	parts, err := idParts(d.Id())
	if err != nil {
		return err
	}

	cloudInstanceID := parts[0]
	imageID := parts[1]

	imageC := st.NewIBMPIImageClient(sess, cloudInstanceID)
	imagedata, err := imageC.Get(imageID, cloudInstanceID)
	if err != nil {
		switch err.(type) {
		case *p_cloud_images.PcloudCloudinstancesImagesGetNotFound:
			log.Printf("[DEBUG] image does not exist %v", err)
			d.SetId("")
			return nil
		}
		log.Printf("[DEBUG] get image failed %v", err)
		return fmt.Errorf(errors.GetImageOperationFailed, imageID, err)
	}

	imageid := *imagedata.ImageID
	d.Set("image_id", imageid)
	d.Set(helpers.PICloudInstanceId, cloudInstanceID)

	return nil
}

func resourceIBMPIImageUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceIBMPIImageDelete(d *schema.ResourceData, meta interface{}) error {
	sess, err := meta.(ClientSession).IBMPISession()
	if err != nil {
		return err
	}

	parts, err := idParts(d.Id())
	if err != nil {
		return err
	}

	cloudInstanceID := parts[0]
	imageID := parts[1]
	imageC := st.NewIBMPIImageClient(sess, cloudInstanceID)
	err = imageC.Delete(imageID, cloudInstanceID)
	if err != nil {
		return fmt.Errorf("Failed to Delete PI Image %s :%v", imageID, err)
	}

	d.SetId("")
	return nil
}

func resourceIBMPIImageExists(d *schema.ResourceData, meta interface{}) (bool, error) {

	sess, err := meta.(ClientSession).IBMPISession()
	if err != nil {
		return false, err
	}
	parts, err := idParts(d.Id())
	if err != nil {
		return false, err
	}
	name := parts[1]
	powerinstanceid := parts[0]
	client := st.NewIBMPIImageClient(sess, powerinstanceid)

	image, err := client.Get(parts[1], powerinstanceid)
	if err != nil {
		if apiErr, ok := err.(bmxerror.RequestFailure); ok {
			if apiErr.StatusCode() == 404 {
				return false, nil
			}
		}
		return false, fmt.Errorf("Error communicating with the API: %s", err)
	}
	return *image.ImageID == name, nil
}

func isWaitForIBMPIImageAvailable(client *st.IBMPIImageClient, id string, timeout time.Duration, powerinstanceid string) (interface{}, error) {
	log.Printf("Waiting for Power Image (%s) to be available.", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"retry", helpers.PIImageQueStatus},
		Target:     []string{helpers.PIImageActiveStatus},
		Refresh:    isIBMPIImageRefreshFunc(client, id, powerinstanceid),
		Timeout:    timeout,
		Delay:      20 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	return stateConf.WaitForState()
}

func isIBMPIImageRefreshFunc(client *st.IBMPIImageClient, id, powerinstanceid string) resource.StateRefreshFunc {

	log.Printf("Calling the isIBMPIImageRefreshFunc Refresh Function....")
	return func() (interface{}, string, error) {
		image, err := client.Get(id, powerinstanceid)
		if err != nil {
			return nil, "", err
		}

		if image.State == "active" {
			return image, helpers.PIImageActiveStatus, nil
		}

		return image, helpers.PIImageQueStatus, nil
	}
}

func waitForIBMPIJobCompleted(client *st.IBMPIJobClient, jobID, cloudInstanceID string, timeout time.Duration) (interface{}, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{helpers.JobStatusQueued, helpers.JobStatusReadyForProcessing, helpers.JobStatusInProgress, helpers.JobStatusRunning, helpers.JobStatusWaiting},
		Target:  []string{helpers.JobStatusCompleted, helpers.JobStatusFailed},
		Refresh: func() (interface{}, string, error) {
			job, err := client.Get(jobID, cloudInstanceID)
			if err != nil {
				log.Printf("[DEBUG] get job failed %v", err)
				return nil, "", fmt.Errorf(errors.GetJobOperationFailed, jobID, err)
			}
			if *job.Status.State == helpers.JobStatusFailed {
				log.Printf("[DEBUG] job status failed with message: %v", job.Status.Message)
				return nil, helpers.JobStatusFailed, fmt.Errorf("job status failed for job id %s with message: %v", jobID, job.Status.Message)
			}
			return job, *job.Status.State, nil
		},
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}
	return stateConf.WaitForState()
}
