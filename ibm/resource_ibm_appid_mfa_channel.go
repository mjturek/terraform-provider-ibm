package ibm

import (
	"context"
	"fmt"

	"github.com/IBM-Cloud/bluemix-go/helpers"
	appid "github.com/IBM/appid-management-go-sdk/appidmanagementv4"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceIBMAppIDMFAChannel() *schema.Resource {
	return &schema.Resource{
		Description: "Update MFA channel configuration",
		Read:        resourceIBMAppIDMFAChannelRead,
		Create:      resourceIBMAppIDMFAChannelCreate,
		Update:      resourceIBMAppIDMFAChannelCreate,
		Delete:      resourceIBMAppIDMFAChannelDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The AppID instance GUID",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"active": {
				Description:  "Allowed values: `email`, `sms`",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"email", "sms"}, false),
			},
			"sms_config": {
				Description: "Configuration for `sms` channel. Create Vonage account (https://dashboard.nexmo.com/sign-up) to get an API key",
				Type:        schema.TypeList,
				Optional:    true,
				Sensitive:   true, // terraform does not yet support nested sensitive attributes, this is temporary workaround
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Description: "API key",
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
						},
						"secret": {
							Description: "API secret",
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
						},
						"from": {
							Description: "Sender's phone number",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func resourceIBMAppIDMFAChannelRead(d *schema.ResourceData, meta interface{}) error {
	appIDClient, err := meta.(ClientSession).AppIDAPI()

	if err != nil {
		return err
	}

	tenantID := d.Id()

	ch, resp, err := appIDClient.ListChannelsWithContext(context.TODO(), &appid.ListChannelsOptions{
		TenantID: &tenantID,
	})

	if err != nil {
		return fmt.Errorf("Error getting AppID MFA channels: %s\n%s", err, resp)
	}

	for _, channel := range ch.Channels {
		if *channel.IsActive {
			d.Set("active", *channel.Type)
		}

		if *channel.Type == "sms" && channel.Config != nil {
			config := map[string]interface{}{
				"key":    *channel.Config.Key,
				"secret": *channel.Config.Secret,
				"from":   *channel.Config.From,
			}

			if err := d.Set("sms_config", []interface{}{config}); err != nil {
				return fmt.Errorf("Error setting AppID MFA channel config: %s", err)
			}
		}
	}

	d.Set("tenant_id", tenantID)

	return nil
}

func resourceIBMAppIDMFAChannelCreate(d *schema.ResourceData, meta interface{}) error {
	appIDClient, err := meta.(ClientSession).AppIDAPI()

	if err != nil {
		return err
	}

	tenantID := d.Get("tenant_id").(string)
	isActive := d.Get("active").(string) == "sms"

	input := &appid.UpdateChannelOptions{
		TenantID: &tenantID,
		IsActive: &isActive,
		// email does not have any options, it is sufficient to just update nexmo configuration, if it is set to disabled, email will be automatically enabled
		Channel: helpers.String("nexmo"),
	}

	if cfg, ok := d.GetOk("sms_config"); ok {
		config := cfg.([]interface{})

		if len(config) > 0 {
			input.Config = cfg.([]interface{})[0]
		}
	}

	_, resp, err := appIDClient.UpdateChannelWithContext(context.TODO(), input)

	if err != nil {
		return fmt.Errorf("Error updating AppID MFA configuration: %s\n%s", err, resp)
	}

	d.SetId(tenantID)

	return resourceIBMAppIDMFAChannelRead(d, meta)
}

func resourceIBMAppIDMFAChannelDelete(d *schema.ResourceData, meta interface{}) error {
	appIDClient, err := meta.(ClientSession).AppIDAPI()

	if err != nil {
		return err
	}

	tenantID := d.Get("tenant_id").(string)

	// defaults
	input := &appid.UpdateChannelOptions{
		TenantID: &tenantID,
		IsActive: helpers.Bool(false),
		Channel:  helpers.String("nexmo"),
		Config: map[string]interface{}{
			"provider": "nexmo",
			"from":     "+12223334444", // AppID default
			"key":      "<key>",
			"secret":   "<secret>",
		},
	}

	_, resp, err := appIDClient.UpdateChannelWithContext(context.TODO(), input)

	if err != nil {
		return fmt.Errorf("Error resetting AppID MFA configuration: %s\n%s", err, resp)
	}

	d.SetId("")
	return nil
}
