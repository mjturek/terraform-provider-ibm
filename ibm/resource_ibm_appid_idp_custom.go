package ibm

import (
	"context"
	"fmt"
	"log"

	"github.com/IBM-Cloud/bluemix-go/helpers"
	appid "github.com/IBM/appid-management-go-sdk/appidmanagementv4"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceIBMAppIDIDPCustom() *schema.Resource {
	return &schema.Resource{
		Create: resourceIBMAppIDIDPCustomCreate,
		Read:   resourceIBMAppIDIDPCustomRead,
		Delete: resourceIBMAppIDIDPCustomDelete,
		Update: resourceIBMAppIDIDPCustomUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The service `tenantId`",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"is_active": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"public_key": {
				Description: "This is the public key used to validate your signed JWT. It is required to be a PEM in the RS256 or greater format.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func resourceIBMAppIDIDPCustomRead(d *schema.ResourceData, meta interface{}) error {
	appIDClient, err := meta.(ClientSession).AppIDAPI()

	if err != nil {
		return err
	}

	tenantID := d.Id()

	config, resp, err := appIDClient.GetCustomIDPWithContext(context.TODO(), &appid.GetCustomIDPOptions{
		TenantID: &tenantID,
	})

	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			log.Printf("[WARN] AppID instance '%s' is not found, removing AppID custom IDP configuration from state", tenantID)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error loading AppID custom IDP: %s\n%s", err, resp)
	}

	d.Set("is_active", *config.IsActive)

	if config.Config != nil && config.Config.PublicKey != nil {
		if err := d.Set("public_key", *config.Config.PublicKey); err != nil {
			return fmt.Errorf("Failed setting AppID custom IDP public_key: %s", err)
		}
	}

	d.Set("tenant_id", tenantID)

	return nil
}

func resourceIBMAppIDIDPCustomCreate(d *schema.ResourceData, meta interface{}) error {
	appIDClient, err := meta.(ClientSession).AppIDAPI()

	if err != nil {
		return err
	}

	tenantID := d.Get("tenant_id").(string)
	isActive := d.Get("is_active").(bool)

	config := &appid.SetCustomIDPOptions{
		TenantID: &tenantID,
		IsActive: &isActive,
	}

	if isActive {
		config.Config = &appid.CustomIDPConfigParamsConfig{}

		if pKey, ok := d.GetOk("public_key"); ok {
			config.Config.PublicKey = helpers.String(pKey.(string))
		}
	}

	_, resp, err := appIDClient.SetCustomIDPWithContext(context.TODO(), config)

	if err != nil {
		return fmt.Errorf("Error applying AppID custom IDP configuration: %s\n%s", err, resp)
	}

	d.SetId(tenantID)

	return resourceIBMAppIDIDPCustomRead(d, meta)
}

func appIDCustomIDPDefaults(tenantID string) *appid.SetCustomIDPOptions {
	return &appid.SetCustomIDPOptions{
		TenantID: &tenantID,
		IsActive: helpers.Bool(false),
	}
}

func resourceIBMAppIDIDPCustomDelete(d *schema.ResourceData, meta interface{}) error {
	appIDClient, err := meta.(ClientSession).AppIDAPI()

	if err != nil {
		return err
	}

	tenantID := d.Get("tenant_id").(string)
	config := appIDCustomIDPDefaults(tenantID)

	_, resp, err := appIDClient.SetCustomIDPWithContext(context.TODO(), config)

	if err != nil {
		return fmt.Errorf("Error resetting AppID custom IDP configuration: %s\n%s", err, resp)
	}

	d.SetId("")

	return nil
}

func resourceIBMAppIDIDPCustomUpdate(d *schema.ResourceData, m interface{}) error {
	// since this is configuration we can reuse create method
	return resourceIBMAppIDIDPCustomCreate(d, m)
}
