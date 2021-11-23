// Copyright IBM Corp. 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package ibm

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	en "github.com/IBM/event-notifications-go-admin-sdk/eventnotificationsv1"
)

func dataSourceIBMEnTopic() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceIBMEnTopicRead,

		Schema: map[string]*schema.Schema{
			"instance_guid": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique identifier for IBM Cloud Event Notifications instance.",
			},
			"topic_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique identifier for Topic.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the topic.",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the topic.",
			},
			"source_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of sources.",
			},
			"sources": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of sources.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the source.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the source.",
						},
						"rules": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of rules.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Whether the rule is enabled or not.",
									},
									"event_type_filter": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Event type filter.",
									},
									"notification_filter": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Notification filter.",
									},
									"updated_at": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Last time the topic was updated.",
									},
									"id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Autogenerated rule ID.",
									},
								},
							},
						},
					},
				},
			},
			"subscription_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of subscriptions.",
			},
			"subscriptions": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of subscriptions.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Subscription ID.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Subscription name.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Subscription description.",
						},
						"destination_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of destination.",
						},
						"destination_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The destination ID.",
						},
						"topic_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Topic ID.",
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Last updated time.",
						},
					},
				},
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last time the topic was updated.",
			},
		},
	}
}

func dataSourceIBMEnTopicRead(d *schema.ResourceData, meta interface{}) error {
	enClient, err := meta.(ClientSession).EventNotificationsApiV1()
	if err != nil {
		return err
	}

	options := &en.GetTopicOptions{}

	options.SetInstanceID(d.Get("instance_guid").(string))
	options.SetID(d.Get("topic_id").(string))

	result, response, err := enClient.GetTopicWithContext(context.TODO(), options)

	if err != nil {
		return fmt.Errorf("GetTopicWithContext failed %s\n%s", err, response)
	}

	d.SetId(fmt.Sprintf("%s/%s", *options.InstanceID, *options.ID))

	d.Set("topic_id", options.ID)

	if err = d.Set("name", result.Name); err != nil {
		return fmt.Errorf("error setting name: %s", err)
	}

	if result.Description != nil {
		if err = d.Set("description", result.Description); err != nil {
			return fmt.Errorf("error setting description: %s", err)
		}
	}

	if err = d.Set("updated_at", result.UpdatedAt); err != nil {
		return fmt.Errorf("error setting updated_at: %s", err)
	}

	if err = d.Set("source_count", intValue(result.SourceCount)); err != nil {
		return fmt.Errorf("error setting source_count: %s", err)
	}

	if err = d.Set("subscription_count", intValue(result.SubscriptionCount)); err != nil {
		return fmt.Errorf("error setting subscription_count: %s", err)
	}

	if result.Sources != nil {
		err = d.Set("sources", dataSourceTopicFlattenSources(result.Sources))
		if err != nil {
			return fmt.Errorf("error setting sources %s", err)
		}
	}

	if result.Subscriptions != nil {
		err = d.Set("subscriptions", enFlattenSubscriptions(result.Subscriptions))
		if err != nil {
			return fmt.Errorf("error setting subscriptions %s", err)
		}
	}

	return nil
}

func dataSourceTopicFlattenSources(result []en.TopicSourcesItem) (sources []map[string]interface{}) {
	sources = []map[string]interface{}{}

	for _, sourcesItem := range result {
		sources = append(sources, dataSourceTopicSourcesToMap(sourcesItem))
	}

	return sources
}

func dataSourceTopicSourcesToMap(sourcesItem en.TopicSourcesItem) (sourcesMap map[string]interface{}) {
	sourcesMap = map[string]interface{}{}

	if sourcesItem.ID != nil {
		sourcesMap["id"] = sourcesItem.ID
	}
	if sourcesItem.Name != nil {
		sourcesMap["name"] = sourcesItem.Name
	}

	if sourcesItem.Rules != nil {
		rulesList := []map[string]interface{}{}
		for _, rulesItem := range sourcesItem.Rules {
			rulesList = append(rulesList, enRulesToMap(rulesItem))
		}
		sourcesMap["rules"] = rulesList
	}

	return sourcesMap
}

func enRulesToMap(rulesItem en.RulesGet) (rulesMap map[string]interface{}) {
	rulesMap = map[string]interface{}{}

	if rulesItem.ID != nil {
		rulesMap["id"] = rulesItem.ID
	}

	if rulesItem.Enabled != nil {
		rulesMap["enabled"] = rulesItem.Enabled
	}

	if rulesItem.EventTypeFilter != nil {
		rulesMap["event_type_filter"] = rulesItem.EventTypeFilter
	}

	if rulesItem.NotificationFilter != nil {
		rulesMap["notification_filter"] = rulesItem.NotificationFilter
	}

	if rulesItem.UpdatedAt != nil {
		rulesMap["updated_at"] = rulesItem.UpdatedAt
	}

	return rulesMap
}

func enFlattenSubscriptions(subscriptionList []en.SubscriptionListItem) (subscriptions []map[string]interface{}) {
	subscriptions = []map[string]interface{}{}

	for _, subscription := range subscriptionList {
		subscriptions = append(subscriptions, enSubscriptionsToMap(subscription))
	}

	return subscriptions
}

func enSubscriptionsToMap(subscription en.SubscriptionListItem) (subscriptionsMap map[string]interface{}) {
	subscriptionsMap = map[string]interface{}{}

	if subscription.ID != nil {
		subscriptionsMap["id"] = subscription.ID
	}

	if subscription.Name != nil {
		subscriptionsMap["name"] = subscription.Name
	}

	if subscription.Description != nil {
		subscriptionsMap["description"] = subscription.Description
	}

	if subscription.UpdatedAt != nil {
		subscriptionsMap["updated_at"] = subscription.UpdatedAt
	}

	if subscription.DestinationType != nil {
		subscriptionsMap["destination_type"] = subscription.DestinationType
	}

	if subscription.DestinationID != nil {
		subscriptionsMap["destination_id"] = subscription.DestinationID
	}

	if subscription.TopicID != nil {
		subscriptionsMap["topic_id"] = subscription.TopicID
	}

	return subscriptionsMap
}
