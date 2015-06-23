package aws

import (
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsEcsCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEcsClusterCreate,
		Read:   resourceAwsEcsClusterRead,
		Delete: resourceAwsEcsClusterDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsEcsClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	clusterName := d.Get("name").(string)
	log.Printf("[DEBUG] Creating ECS cluster %s", clusterName)

	out, err := conn.CreateCluster(&ecs.CreateClusterInput{
		ClusterName: aws.String(clusterName),
	})
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] ECS cluster %s created", *out.Cluster.ClusterARN)

	d.SetId(*out.Cluster.ClusterARN)
	d.Set("name", *out.Cluster.ClusterName)
	return nil
}

func resourceAwsEcsClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	clusterName := d.Get("name").(string)
	log.Printf("[DEBUG] Reading ECS cluster %s", clusterName)
	out, err := conn.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: []*string{aws.String(clusterName)},
	})
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Received ECS clusters: %#v", out.Clusters)

	d.SetId(*out.Clusters[0].ClusterARN)
	d.Set("name", *out.Clusters[0].ClusterName)

	return nil
}

func resourceAwsEcsClusterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	log.Printf("[DEBUG] Deleting ECS cluster %s", d.Id())

	// Retry if there are active services in cluster
	wait := resource.StateChangeConf{
		Pending:    []string{"ClusterContainsServicesException"},
		Target:     "ResourceNotFoundException",
		Timeout:    5 * time.Minute,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DeleteCluster(&ecs.DeleteClusterInput{
				Cluster: aws.String(d.Id()),
			})

			if awsErr, ok := err.(awserr.Error); ok {
				log.Printf("[DEBUG] Current error code: %q", awsErr.Code())
				return resp, awsErr.Code(), nil
			}

			return nil, "Failed", err
		},
	}

	out, err := wait.WaitForState()
	if err != nil {
		return err
	}

	// TODO: Handle ClientException: The Cluster cannot be deleted while Container Instances are active.
	// TODO: Handle ClientException: The Cluster cannot be deleted while Services are active.

	log.Printf("[DEBUG] ECS cluster %s deleted: %#v", d.Id(), out)
	d.SetId("")

	return err
}
