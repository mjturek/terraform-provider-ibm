package ibm

import (
	"context"
	"fmt"
	"log"

	"github.com/IBM-Cloud/bluemix-go/helpers"
	appid "github.com/IBM/appid-management-go-sdk/appidmanagementv4"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceIBMAppIDIDPCloudDirectory() *schema.Resource {
	return &schema.Resource{
		Create: resourceIBMAppIDIDPCloudDirectoryCreate,
		Read:   resourceIBMAppIDIDPCloudDirectoryRead,
		Delete: resourceIBMAppIDIDPCloudDirectoryDelete,
		Update: resourceIBMAppIDIDPCloudDirectoryUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"is_active": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"self_service_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"signup_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"welcome_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"reset_password_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"reset_password_notification_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"identity_confirm_access_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "FULL",
				ValidateFunc: validation.StringInSlice([]string{"FULL", "RESTRICTIVE", "OFF"}, false),
			},
			"identity_confirm_methods": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"identity_field": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceIBMAppIDIDPCloudDirectoryRead(d *schema.ResourceData, meta interface{}) error {
	appIDClient, err := meta.(ClientSession).AppIDAPI()

	if err != nil {
		return err
	}

	tenantID := d.Id()

	config, resp, err := appIDClient.GetCloudDirectoryIDPWithContext(context.TODO(), &appid.GetCloudDirectoryIDPOptions{
		TenantID: &tenantID,
	})

	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			log.Printf("[WARN] AppID instance '%s' is not found, removing IDP Cloud Directory from state", tenantID)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error loading AppID Cloud Directory IDP: %s\n%s", err, resp)
	}

	d.Set("is_active", *config.IsActive)

	if config.Config != nil {
		d.Set("self_service_enabled", *config.Config.SelfServiceEnabled)

		if config.Config.SignupEnabled != nil {
			d.Set("signup_enabled", *config.Config.SignupEnabled)
		}

		if config.Config.IdentityField != nil {
			d.Set("identity_field", *config.Config.IdentityField)
		}

		if config.Config.Interactions != nil {
			d.Set("welcome_enabled", *config.Config.Interactions.WelcomeEnabled)
			d.Set("reset_password_enabled", *config.Config.Interactions.ResetPasswordEnabled)
			d.Set("reset_password_notification_enabled", *config.Config.Interactions.ResetPasswordNotificationEnable)
			d.Set("identity_confirm_access_mode", *config.Config.Interactions.IdentityConfirmation.AccessMode)

			if err := d.Set("identity_confirm_methods", config.Config.Interactions.IdentityConfirmation.Methods); err != nil {
				return fmt.Errorf("Error setting AppID IDP Cloud Directory identity confirm methods: %s", err)
			}
		}
	}

	d.Set("tenant_id", tenantID)

	return nil
}

func resourceIBMAppIDIDPCloudDirectoryCreate(d *schema.ResourceData, meta interface{}) error {
	appIDClient, err := meta.(ClientSession).AppIDAPI()

	if err != nil {
		return err
	}

	tenantID := d.Get("tenant_id").(string)
	isActive := d.Get("is_active").(bool)

	config := &appid.SetCloudDirectoryIDPOptions{
		TenantID: &tenantID,
		IsActive: &isActive,
		Config: &appid.CloudDirectoryConfigParams{
			SelfServiceEnabled: helpers.Bool(d.Get("self_service_enabled").(bool)),
			SignupEnabled:      helpers.Bool(d.Get("signup_enabled").(bool)),
			Interactions: &appid.CloudDirectoryConfigParamsInteractions{
				WelcomeEnabled:                  helpers.Bool(d.Get("welcome_enabled").(bool)),
				ResetPasswordEnabled:            helpers.Bool(d.Get("reset_password_enabled").(bool)),
				ResetPasswordNotificationEnable: helpers.Bool(d.Get("reset_password_notification_enabled").(bool)),
				IdentityConfirmation: &appid.CloudDirectoryConfigParamsInteractionsIdentityConfirmation{
					AccessMode: helpers.String(d.Get("identity_confirm_access_mode").(string)),
				},
			},
		},
	}

	if idField, ok := d.GetOk("identity_field"); ok {
		config.Config.IdentityField = helpers.String(idField.(string))
	}

	if methods, ok := d.GetOk("identity_confirm_methods"); ok {
		config.Config.Interactions.IdentityConfirmation.Methods = expandStringList(methods.([]interface{}))
	}

	_, resp, err := appIDClient.SetCloudDirectoryIDPWithContext(context.TODO(), config)

	if err != nil {
		return fmt.Errorf("Error applying AppID Cloud Directory IDP configuration: %s\n%s", err, resp)
	}

	d.SetId(tenantID)

	return resourceIBMAppIDIDPCloudDirectoryRead(d, meta)
}

func resourceIBMAppIDIDPCloudDirectoryUpdate(d *schema.ResourceData, m interface{}) error {
	// since this is configuration we can reuse create method
	return resourceIBMAppIDIDPCloudDirectoryCreate(d, m)
}

func resourceIBMAppIDIDPCloudDirectoryDelete(d *schema.ResourceData, meta interface{}) error {
	appIDClient, err := meta.(ClientSession).AppIDAPI()

	if err != nil {
		return err
	}

	tenantID := d.Get("tenant_id").(string)
	config := ibmAppIDIDPCloudDirectoryDefaults(tenantID)

	_, resp, err := appIDClient.SetCloudDirectoryIDPWithContext(context.TODO(), config)

	if err != nil {
		return fmt.Errorf("Error resetting AppID Cloud Directory IDP configuration: %s\n%s", err, resp)
	}

	d.SetId("")

	return nil
}

func ibmAppIDIDPCloudDirectoryDefaults(tenantID string) *appid.SetCloudDirectoryIDPOptions {
	return &appid.SetCloudDirectoryIDPOptions{
		TenantID: &tenantID,
		IsActive: helpers.Bool(false),
		Config: &appid.CloudDirectoryConfigParams{
			SignupEnabled:      helpers.Bool(true),
			SelfServiceEnabled: helpers.Bool(true),
			Interactions: &appid.CloudDirectoryConfigParamsInteractions{
				IdentityConfirmation: &appid.CloudDirectoryConfigParamsInteractionsIdentityConfirmation{
					AccessMode: helpers.String("FULL"),
					Methods:    []string{"email"},
				},
				WelcomeEnabled:                  helpers.Bool(true),
				ResetPasswordEnabled:            helpers.Bool(true),
				ResetPasswordNotificationEnable: helpers.Bool(true),
			},
		},
	}
}
