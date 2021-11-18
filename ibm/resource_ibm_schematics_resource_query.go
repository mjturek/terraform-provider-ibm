// Copyright IBM Corp. 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package ibm

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/schematics-go-sdk/schematicsv1"
)

func resourceIBMSchematicsResourceQuery() *schema.Resource {
	return &schema.Resource{
		Create:   resourceIBMSchematicsResourceQueryCreate,
		Read:     resourceIBMSchematicsResourceQueryRead,
		Update:   resourceIBMSchematicsResourceQueryUpdate,
		Delete:   resourceIBMSchematicsResourceQueryDelete,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: InvokeValidator("ibm_schematics_resource_query", "type"),
				Description:  "Resource type (cluster, vsi, icd, vpc).",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Resource query name.",
			},
			"queries": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query_type": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Type of the query(workspaces).",
						},
						"query_condition": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": &schema.Schema{
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Name of the resource query param.",
									},
									"value": &schema.Schema{
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Value of the resource query param.",
									},
									"description": &schema.Schema{
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Description of resource query param variable.",
									},
								},
							},
						},
						"query_select": &schema.Schema{
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of query selection parameters.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"created_at": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Resource query creation time.",
			},
			"created_by": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Email address of user who created the Resource query.",
			},
			"updated_at": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Resource query updation time.",
			},
			"updated_by": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Email address of user who updated the Resource query.",
			},
		},
	}
}

func resourceIBMSchematicsResourceQueryValidator() *ResourceValidator {
	validateSchema := make([]ValidateSchema, 1)
	validateSchema = append(validateSchema,
		ValidateSchema{
			Identifier:                 "type",
			ValidateFunctionIdentifier: ValidateAllowedStringValue,
			Type:                       TypeString,
			Optional:                   true,
			AllowedValues:              "vsi",
		},
	)

	resourceValidator := ResourceValidator{ResourceName: "ibm_schematics_resource_query", Schema: validateSchema}
	return &resourceValidator
}

func resourceIBMSchematicsResourceQueryCreate(d *schema.ResourceData, meta interface{}) error {
	schematicsClient, err := meta.(ClientSession).SchematicsV1()
	if err != nil {
		return err
	}

	createResourceQueryOptions := &schematicsv1.CreateResourceQueryOptions{}

	if _, ok := d.GetOk("type"); ok {
		createResourceQueryOptions.SetType(d.Get("type").(string))
	}
	if _, ok := d.GetOk("name"); ok {
		createResourceQueryOptions.SetName(d.Get("name").(string))
	}
	if _, ok := d.GetOk("queries"); ok {
		var queries []schematicsv1.ResourceQuery
		for _, e := range d.Get("queries").([]interface{}) {
			value := e.(map[string]interface{})
			queriesItem := resourceIBMSchematicsResourceQueryMapToResourceQuery(value)
			queries = append(queries, queriesItem)
		}
		createResourceQueryOptions.SetQueries(queries)
	}

	resourceQueryRecord, response, err := schematicsClient.CreateResourceQueryWithContext(context.TODO(), createResourceQueryOptions)
	if err != nil {
		log.Printf("[DEBUG] CreateResourceQueryWithContext failed %s\n%s", err, response)
		return fmt.Errorf("CreateResourceQueryWithContext failed %s\n%s", err, response)
	}

	d.SetId(*resourceQueryRecord.ID)

	return resourceIBMSchematicsResourceQueryRead(d, meta)
}

func resourceIBMSchematicsResourceQueryMapToResourceQuery(resourceQueryMap map[string]interface{}) schematicsv1.ResourceQuery {
	resourceQuery := schematicsv1.ResourceQuery{}

	if resourceQueryMap["query_type"] != nil {
		resourceQuery.QueryType = core.StringPtr(resourceQueryMap["query_type"].(string))
	}
	if resourceQueryMap["query_condition"] != nil {
		queryCondition := []schematicsv1.ResourceQueryParam{}
		for _, queryConditionItem := range resourceQueryMap["query_condition"].([]interface{}) {
			queryConditionItemModel := resourceIBMSchematicsResourceQueryMapToResourceQueryParam(queryConditionItem.(map[string]interface{}))
			queryCondition = append(queryCondition, queryConditionItemModel)
		}
		resourceQuery.QueryCondition = queryCondition
	}
	if resourceQueryMap["query_select"] != nil {
		querySelect := []string{}
		for _, querySelectItem := range resourceQueryMap["query_select"].([]interface{}) {
			querySelect = append(querySelect, querySelectItem.(string))
		}
		resourceQuery.QuerySelect = querySelect
	}

	return resourceQuery
}

func resourceIBMSchematicsResourceQueryMapToResourceQueryParam(resourceQueryParamMap map[string]interface{}) schematicsv1.ResourceQueryParam {
	resourceQueryParam := schematicsv1.ResourceQueryParam{}

	if resourceQueryParamMap["name"] != nil {
		resourceQueryParam.Name = core.StringPtr(resourceQueryParamMap["name"].(string))
	}
	if resourceQueryParamMap["value"] != nil {
		resourceQueryParam.Value = core.StringPtr(resourceQueryParamMap["value"].(string))
	}
	if resourceQueryParamMap["description"] != nil {
		resourceQueryParam.Description = core.StringPtr(resourceQueryParamMap["description"].(string))
	}

	return resourceQueryParam
}

func resourceIBMSchematicsResourceQueryRead(d *schema.ResourceData, meta interface{}) error {
	schematicsClient, err := meta.(ClientSession).SchematicsV1()
	if err != nil {
		return err
	}

	getResourcesQueryOptions := &schematicsv1.GetResourcesQueryOptions{}

	getResourcesQueryOptions.SetQueryID(d.Id())

	resourceQueryRecord, response, err := schematicsClient.GetResourcesQueryWithContext(context.TODO(), getResourcesQueryOptions)
	if err != nil {
		if response != nil && response.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		log.Printf("[DEBUG] GetResourcesQueryWithContext failed %s\n%s", err, response)
		return fmt.Errorf("GetResourcesQueryWithContext failed %s\n%s", err, response)
	}

	if err = d.Set("type", resourceQueryRecord.Type); err != nil {
		return fmt.Errorf("Error setting type: %s", err)
	}
	if err = d.Set("name", resourceQueryRecord.Name); err != nil {
		return fmt.Errorf("Error setting name: %s", err)
	}
	if resourceQueryRecord.Queries != nil {
		queries := []map[string]interface{}{}
		for _, queriesItem := range resourceQueryRecord.Queries {
			queriesItemMap := resourceIBMSchematicsResourceQueryResourceQueryToMap(queriesItem)
			queries = append(queries, queriesItemMap)
		}
		if err = d.Set("queries", queries); err != nil {
			return fmt.Errorf("Error setting queries: %s", err)
		}
	}
	if err = d.Set("created_at", dateTimeToString(resourceQueryRecord.CreatedAt)); err != nil {
		return fmt.Errorf("Error setting created_at: %s", err)
	}
	if err = d.Set("created_by", resourceQueryRecord.CreatedBy); err != nil {
		return fmt.Errorf("Error setting created_by: %s", err)
	}
	if err = d.Set("updated_at", dateTimeToString(resourceQueryRecord.UpdatedAt)); err != nil {
		return fmt.Errorf("Error setting updated_at: %s", err)
	}
	if err = d.Set("updated_by", resourceQueryRecord.UpdatedBy); err != nil {
		return fmt.Errorf("Error setting updated_by: %s", err)
	}

	return nil
}

func resourceIBMSchematicsResourceQueryResourceQueryToMap(resourceQuery schematicsv1.ResourceQuery) map[string]interface{} {
	resourceQueryMap := map[string]interface{}{}

	if resourceQuery.QueryType != nil {
		resourceQueryMap["query_type"] = resourceQuery.QueryType
	}
	if resourceQuery.QueryCondition != nil {
		queryCondition := []map[string]interface{}{}
		for _, queryConditionItem := range resourceQuery.QueryCondition {
			queryConditionItemMap := resourceIBMSchematicsResourceQueryResourceQueryParamToMap(queryConditionItem)
			queryCondition = append(queryCondition, queryConditionItemMap)
			// TODO: handle QueryCondition of type TypeList -- list of non-primitive, not model items
		}
		resourceQueryMap["query_condition"] = queryCondition
	}
	if resourceQuery.QuerySelect != nil {
		resourceQueryMap["query_select"] = resourceQuery.QuerySelect
	}

	return resourceQueryMap
}

func resourceIBMSchematicsResourceQueryResourceQueryParamToMap(resourceQueryParam schematicsv1.ResourceQueryParam) map[string]interface{} {
	resourceQueryParamMap := map[string]interface{}{}

	if resourceQueryParam.Name != nil {
		resourceQueryParamMap["name"] = resourceQueryParam.Name
	}
	if resourceQueryParam.Value != nil {
		resourceQueryParamMap["value"] = resourceQueryParam.Value
	}
	if resourceQueryParam.Description != nil {
		resourceQueryParamMap["description"] = resourceQueryParam.Description
	}

	return resourceQueryParamMap
}

func resourceIBMSchematicsResourceQueryUpdate(d *schema.ResourceData, meta interface{}) error {
	schematicsClient, err := meta.(ClientSession).SchematicsV1()
	if err != nil {
		return err
	}

	replaceResourcesQueryOptions := &schematicsv1.ReplaceResourcesQueryOptions{}

	replaceResourcesQueryOptions.SetQueryID(d.Id())
	if _, ok := d.GetOk("type"); ok {
		replaceResourcesQueryOptions.SetType(d.Get("type").(string))
	}
	if _, ok := d.GetOk("name"); ok {
		replaceResourcesQueryOptions.SetName(d.Get("name").(string))
	}
	if _, ok := d.GetOk("queries"); ok {
		var queries []schematicsv1.ResourceQuery
		for _, e := range d.Get("queries").([]interface{}) {
			value := e.(map[string]interface{})
			queriesItem := resourceIBMSchematicsResourceQueryMapToResourceQuery(value)
			queries = append(queries, queriesItem)
		}
		replaceResourcesQueryOptions.SetQueries(queries)
	}

	_, response, err := schematicsClient.ReplaceResourcesQueryWithContext(context.TODO(), replaceResourcesQueryOptions)
	if err != nil {
		log.Printf("[DEBUG] ReplaceResourcesQueryWithContext failed %s\n%s", err, response)
		return fmt.Errorf("ReplaceResourcesQueryWithContext failed %s\n%s", err, response)
	}

	return resourceIBMSchematicsResourceQueryRead(d, meta)
}

func resourceIBMSchematicsResourceQueryDelete(d *schema.ResourceData, meta interface{}) error {
	schematicsClient, err := meta.(ClientSession).SchematicsV1()
	if err != nil {
		return err
	}

	deleteResourcesQueryOptions := &schematicsv1.DeleteResourcesQueryOptions{}

	deleteResourcesQueryOptions.SetQueryID(d.Id())

	response, err := schematicsClient.DeleteResourcesQueryWithContext(context.TODO(), deleteResourcesQueryOptions)
	if err != nil {
		log.Printf("[DEBUG] DeleteResourcesQueryWithContext failed %s\n%s", err, response)
		return fmt.Errorf("DeleteResourcesQueryWithContext failed %s\n%s", err, response)
	}

	d.SetId("")

	return nil
}
