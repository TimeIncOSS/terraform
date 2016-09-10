package aws

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsEcsTaskDefinition() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEcsTaskDefinitionCreate,
		Read:   resourceAwsEcsTaskDefinitionRead,
		Delete: resourceAwsEcsTaskDefinitionDelete,

		ListVersions: func(d *schema.ResourceData, limit int, meta interface{}) ([]string, error) {
			conn := meta.(*AWSClient).ecsconn
			var versions = make([]string, 0)
			out, err := conn.ListTaskDefinitionsPages(&ecs.ListTaskDefinitionsInput{
				FamilyPrefix: aws.String(p.Get("family").(string)),
			}, func(page *ecs.ListTaskDefinitionsOutput, lastPage bool) bool {
				for _, arn := range page.TaskDefinitionArns {
					versions = append(versions, *arn)
				}
				return !lastPage
			})
			return versions, err
		},
		SetVersion: schema.SetVersionPassthrough,

		Schema: map[string]*schema.Schema{
			"arn": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"family": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"revision": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},

			"container_definitions": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(v interface{}) string {
					hash := sha1.Sum([]byte(v.(string)))
					return hex.EncodeToString(hash[:])
				},
			},

			"task_role_arn": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"network_mode": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsEcsTaskDefinitionNetworkMode,
			},

			"volume": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						"host_path": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				Set: resourceAwsEcsTaskDefinitionVolumeHash,
			},
		},
	}
}

func validateAwsEcsTaskDefinitionNetworkMode(v interface{}, k string) (ws []string, errors []error) {
	value := strings.ToLower(v.(string))
	validTypes := map[string]struct{}{
		"bridge": struct{}{},
		"host":   struct{}{},
		"none":   struct{}{},
	}

	if _, ok := validTypes[value]; !ok {
		errors = append(errors, fmt.Errorf("ECS Task Definition network_mode %q is invalid, must be `bridge`, `host` or `none`", value))
	}
	return
}

func resourceAwsEcsTaskDefinitionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	rawDefinitions := d.Get("container_definitions").(string)
	definitions, err := expandEcsContainerDefinitions(rawDefinitions)
	if err != nil {
		return err
	}

	input := ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions: definitions,
		Family:               aws.String(d.Get("family").(string)),
	}

	if v, ok := d.GetOk("task_role_arn"); ok {
		input.TaskRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("network_mode"); ok {
		input.NetworkMode = aws.String(v.(string))
	}

	if v, ok := d.GetOk("volume"); ok {
		volumes, err := expandEcsVolumes(v.(*schema.Set).List())
		if err != nil {
			return err
		}
		input.Volumes = volumes
	}

	log.Printf("[DEBUG] Registering ECS task definition: %s", input)
	out, err := conn.RegisterTaskDefinition(&input)
	if err != nil {
		return err
	}

	taskDefinition := *out.TaskDefinition

	log.Printf("[DEBUG] ECS task definition registered: %q (rev. %d)",
		*taskDefinition.TaskDefinitionArn, *taskDefinition.Revision)

	d.SetId(*taskDefinition.Family)
	d.Set("arn", taskDefinition.TaskDefinitionArn)

	return resourceAwsEcsTaskDefinitionRead(d, meta)
}

func resourceAwsEcsTaskDefinitionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	log.Printf("[DEBUG] Reading task definition %s", d.Id())
	out, err := conn.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(d.Get("arn").(string)),
	})
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Received task definition %s", out)

	taskDefinition := out.TaskDefinition

	d.SetId(*taskDefinition.Family)
	d.Set("arn", taskDefinition.TaskDefinitionArn)
	d.Set("family", taskDefinition.Family)
	d.Set("revision", taskDefinition.Revision)
	d.Set("container_definitions", taskDefinition.ContainerDefinitions)
	d.Set("task_role_arn", taskDefinition.TaskRoleArn)
	d.Set("network_mode", taskDefinition.NetworkMode)
	d.Set("volumes", flattenEcsVolumes(taskDefinition.Volumes))

	return nil
}

func resourceAwsEcsTaskDefinitionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	_, err := conn.DeregisterTaskDefinition(&ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: aws.String(d.Get("arn").(string)),
	})
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Task definition %q deregistered.", d.Get("arn").(string))

	return nil
}

func resourceAwsEcsTaskDefinitionVolumeHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["host_path"].(string)))

	return hashcode.String(buf.String())
}
