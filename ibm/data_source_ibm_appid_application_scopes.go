package ibm

import (
	"context"
	"fmt"
	appid "github.com/IBM/appid-management-go-sdk/appidmanagementv4"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceIBMAppIDApplicationScopes() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceIBMAppIDApplicationScopesRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The service `tenantId`",
				Type:        schema.TypeString,
				Required:    true,
			},
			"client_id": {
				Description: "The `client_id` is a public identifier for applications",
				Type:        schema.TypeString,
				Required:    true,
			},
			"scopes": {
				Description: "A `scope` is a runtime action in your application that you register with IBM Cloud App ID to create an access permission",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
		},
	}
}

func dataSourceIBMAppIDApplicationScopesRead(d *schema.ResourceData, meta interface{}) error {
	appIDClient, err := meta.(ClientSession).AppIDAPI()

	if err != nil {
		return err
	}

	tenantID := d.Get("tenant_id").(string)
	clientID := d.Get("client_id").(string)

	scopes, resp, err := appIDClient.GetApplicationScopesWithContext(context.TODO(), &appid.GetApplicationScopesOptions{
		TenantID: &tenantID,
		ClientID: &clientID,
	})

	if err != nil {
		return fmt.Errorf("Error getting AppID application scopes: %s\n%s", err, resp)
	}

	if err := d.Set("scopes", scopes.Scopes); err != nil {
		return fmt.Errorf("Error setting AppID application scopes: %s", err)
	}

	d.SetId(fmt.Sprintf("%s/%s", tenantID, clientID))
	return nil
}
