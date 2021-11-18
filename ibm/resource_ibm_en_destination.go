// Copyright IBM Corp. 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package ibm

import (
	"context"
	"fmt"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	en "github.com/IBM/event-notifications-go-admin-sdk/eventnotificationsv1"
)

func resourceIBMEnDestination() *schema.Resource {
	return &schema.Resource{
		Create:   resourceIBMEnDestinationCreate,
		Read:     resourceIBMEnDestinationRead,
		Update:   resourceIBMEnDestinationUpdate,
		Delete:   resourceIBMEnDestinationDelete,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"instance_guid": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Unique identifier for IBM Cloud Event Notifications instance.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Destintion name.",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: InvokeValidator("ibm_en_destination", "type"),
				Description:  "The type of Destination Webhook.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Destination description.",
			},
			"config": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Payload describing a destination configuration.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"params": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"url": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "URL of webhook.",
									},
									"verb": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "HTTP method of webhook.",
									},
									"custom_headers": {
										Type:        schema.TypeMap,
										Optional:    true,
										Description: "Custom headers (Key-Value pair) for webhook call.",
										Elem:        &schema.Schema{Type: schema.TypeString},
									},
									"sensitive_headers": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: "List of sensitive headers from custom headers.",
										Elem:        &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			"destination_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Destination ID",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last updated time.",
			},
			"subscription_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of subscriptions.",
			},
			"subscription_names": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of subscriptions.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceIBMEnDestinationValidator() *ResourceValidator {
	validateSchema := make([]ValidateSchema, 1)
	validateSchema = append(validateSchema,
		ValidateSchema{
			Identifier:                 "type",
			ValidateFunctionIdentifier: ValidateAllowedStringValue,
			Type:                       TypeString,
			Required:                   true,
			AllowedValues:              "webhook",
			MinValueLength:             1,
		},
	)

	resourceValidator := ResourceValidator{ResourceName: "ibm_en_destination", Schema: validateSchema}
	return &resourceValidator
}

func resourceIBMEnDestinationCreate(d *schema.ResourceData, meta interface{}) error {
	enClient, err := meta.(ClientSession).EventNotificationsApiV1()
	if err != nil {
		return err
	}

	options := &en.CreateDestinationOptions{}

	options.SetInstanceID(d.Get("instance_guid").(string))
	options.SetName(d.Get("name").(string))
	options.SetType(d.Get("type").(string))

	if _, ok := d.GetOk("description"); ok {
		options.SetDescription(d.Get("description").(string))
	}
	if _, ok := d.GetOk("config"); ok {
		config := destinationConfigMapToDestinationConfig(d.Get("config.0.params.0").(map[string]interface{}))
		options.SetConfig(&config)
	}

	result, response, err := enClient.CreateDestinationWithContext(context.TODO(), options)
	if err != nil {
		return fmt.Errorf("CreateDestinationWithContext failed %s\n%s", err, response)
	}

	d.SetId(fmt.Sprintf("%s/%s", *options.InstanceID, *result.ID))

	return resourceIBMEnDestinationRead(d, meta)
}

func resourceIBMEnDestinationRead(d *schema.ResourceData, meta interface{}) error {
	enClient, err := meta.(ClientSession).EventNotificationsApiV1()
	if err != nil {
		return err
	}

	options := &en.GetDestinationOptions{}

	parts, err := sepIdParts(d.Id(), "/")
	if err != nil {
		return err
	}

	options.SetInstanceID(parts[0])
	options.SetID(parts[1])

	result, response, err := enClient.GetDestinationWithContext(context.TODO(), options)
	if err != nil {
		if response != nil && response.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("GetDestinationWithContext failed %s\n%s", err, response)
	}

	if err = d.Set("instance_guid", options.InstanceID); err != nil {
		return fmt.Errorf("error setting instance_guid: %s", err)
	}

	if err = d.Set("destination_id", options.ID); err != nil {
		return fmt.Errorf("error setting destination_id: %s", err)
	}

	if err = d.Set("name", result.Name); err != nil {
		return fmt.Errorf("error setting name: %s", err)
	}

	if err = d.Set("type", result.Type); err != nil {
		return fmt.Errorf("error setting type: %s", err)
	}

	if err = d.Set("description", result.Description); err != nil {
		return fmt.Errorf("error setting description: %s", err)
	}

	if result.Config != nil {
		err = d.Set("config", enDestinationFlattenConfig(*result.Config))
		if err != nil {
			return fmt.Errorf("error setting config %s", err)
		}
	}

	if err = d.Set("updated_at", dateTimeToString(result.UpdatedAt)); err != nil {
		return fmt.Errorf("error setting updated_at: %s", err)
	}

	if err = d.Set("subscription_count", intValue(result.SubscriptionCount)); err != nil {
		return fmt.Errorf("error setting subscription_count: %s", err)
	}

	if result.Config != nil {
		if err = d.Set("subscription_names", result.SubscriptionNames); err != nil {
			return fmt.Errorf("error setting subscription_names: %s", err)
		}
	}

	return nil
}

func resourceIBMEnDestinationUpdate(d *schema.ResourceData, meta interface{}) error {
	enClient, err := meta.(ClientSession).EventNotificationsApiV1()
	if err != nil {
		return err
	}

	options := &en.UpdateDestinationOptions{}

	parts, err := sepIdParts(d.Id(), "/")
	if err != nil {
		return err
	}

	options.SetInstanceID(parts[0])
	options.SetID(parts[1])

	if ok := d.HasChanges("name", "description", "config"); ok {
		options.SetName(d.Get("name").(string))

		if _, ok := d.GetOk("description"); ok {
			options.SetDescription(d.Get("description").(string))
		}
		if _, ok := d.GetOk("config"); ok {
			config := destinationConfigMapToDestinationConfig(d.Get("config.0.params.0").(map[string]interface{}))
			options.SetConfig(&config)
		}
		_, response, err := enClient.UpdateDestinationWithContext(context.TODO(), options)
		if err != nil {
			return fmt.Errorf("UpdateDestinationWithContext failed %s\n%s", err, response)
		}

		return resourceIBMEnDestinationRead(d, meta)
	}

	return nil
}

func resourceIBMEnDestinationDelete(d *schema.ResourceData, meta interface{}) error {
	enClient, err := meta.(ClientSession).EventNotificationsApiV1()
	if err != nil {
		return err
	}

	options := &en.DeleteDestinationOptions{}

	parts, err := sepIdParts(d.Id(), "/")
	if err != nil {
		return err
	}

	options.SetInstanceID(parts[0])
	options.SetID(parts[1])

	response, err := enClient.DeleteDestinationWithContext(context.TODO(), options)
	if err != nil {
		if response != nil && response.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("DeleteDestinationWithContext failed %s\n%s", err, response)
	}

	d.SetId("")

	return nil
}

func destinationConfigMapToDestinationConfig(configParams map[string]interface{}) en.DestinationConfig {
	params := new(en.DestinationConfigParams)
	if configParams["url"] != nil {
		params.URL = core.StringPtr(configParams["url"].(string))
	}

	if configParams["verb"] != nil {
		params.Verb = core.StringPtr(configParams["verb"].(string))
	}

	if configParams["custom_headers"] != nil {
		var customHeaders = make(map[string]string)
		for k, v := range configParams["custom_headers"].(map[string]interface{}) {
			customHeaders[k] = v.(string)
		}
		params.CustomHeaders = customHeaders
	}

	if configParams["sensitive_headers"] != nil {
		sensitiveHeaders := []string{}
		for _, sensitiveHeadersItem := range configParams["sensitive_headers"].([]interface{}) {
			sensitiveHeaders = append(sensitiveHeaders, sensitiveHeadersItem.(string))
		}
		params.SensitiveHeaders = sensitiveHeaders
	}

	destinationConfig := new(en.DestinationConfig)
	destinationConfig.Params = params
	return *destinationConfig
}
