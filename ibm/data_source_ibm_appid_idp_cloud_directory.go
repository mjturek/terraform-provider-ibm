package ibm

import (
	"context"
	"fmt"

	appid "github.com/IBM/appid-management-go-sdk/appidmanagementv4"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceIBMAppIDIDPCloudDirectory() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceIBMAppIDIDPCloudDirectoryRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The service `tenantId`",
				Type:        schema.TypeString,
				Required:    true,
			},
			"is_active": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"self_service_enabled": {
				Description: "Allow users to manage their account from your app",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"signup_enabled": {
				Description: "Allow users to sign-up to your app",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"welcome_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"reset_password_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"reset_password_notification_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"identity_confirm_access_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identity_confirm_methods": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"identity_field": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceIBMAppIDIDPCloudDirectoryRead(d *schema.ResourceData, meta interface{}) error {
	appIDClient, err := meta.(ClientSession).AppIDAPI()

	if err != nil {
		return err
	}

	tenantID := d.Get("tenant_id").(string)

	config, resp, err := appIDClient.GetCloudDirectoryIDPWithContext(context.TODO(), &appid.GetCloudDirectoryIDPOptions{
		TenantID: &tenantID,
	})

	if err != nil {
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
				return fmt.Errorf("Error setting AppID Cloud Directory IDP identity_confirm_methods: %s", err)
			}
		}
	}

	d.SetId(tenantID)

	return nil
}
