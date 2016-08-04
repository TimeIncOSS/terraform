package aws

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsAutoscalingGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsAutoscalingGroupRead,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"launch_configuration": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"desired_capacity": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},

			"min_elb_capacity": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},

			"min_size": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},

			"max_size": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},

			"default_cooldown": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},

			"force_delete": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},

			"health_check_grace_period": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},

			"health_check_type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"availability_zones": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"placement_group": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"load_balancers": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"vpc_zone_identifier": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"termination_policies": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"wait_for_capacity_timeout": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"wait_for_elb_capacity": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},

			"enabled_metrics": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"metrics_granularity": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"protect_from_scale_in": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},

			"tag": autoscalingTagsSchema(),
		},
	}
}

func dataSourceAwsAutoscalingGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).autoscalingconn

	name := d.Get("name").(string)
	g, err := getAwsAutoscalingGroup(name, conn)
	if err != nil {
		return err
	}
	if g == nil {
		return fmt.Errorf("Autoscaling Group %q not found", name)
	}

	d.Set("availability_zones", flattenStringList(g.AvailabilityZones))
	d.Set("default_cooldown", g.DefaultCooldown)
	d.Set("desired_capacity", g.DesiredCapacity)
	d.Set("health_check_grace_period", g.HealthCheckGracePeriod)
	d.Set("health_check_type", g.HealthCheckType)
	d.Set("launch_configuration", g.LaunchConfigurationName)
	d.Set("load_balancers", flattenStringList(g.LoadBalancerNames))
	d.Set("max_size", g.MaxSize)
	d.Set("min_size", g.MinSize)
	d.Set("placement_group", g.PlacementGroup)
	d.Set("protect_from_scale_in", g.NewInstancesProtectedFromScaleIn)
	d.Set("tag", autoscalingTagDescriptionsToSlice(g.Tags))
	d.Set("termination_policies", flattenStringList(g.TerminationPolicies))
	d.Set("vpc_zone_identifier", strings.Split(*g.VPCZoneIdentifier, ","))

	if g.EnabledMetrics != nil {
		if err := d.Set("enabled_metrics", flattenAsgEnabledMetrics(g.EnabledMetrics)); err != nil {
			log.Printf("[WARN] Error setting metrics for (%s): %s", d.Id(), err)
		}
		d.Set("metrics_granularity", g.EnabledMetrics[0].Granularity)
	}

	return nil
}
