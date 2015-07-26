package google

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"google.golang.org/api/deploymentmanager/v2beta2"
	"google.golang.org/api/googleapi"
)

func resourceGoogleDeployment() *schema.Resource {
	return &schema.Resource{
		Create: resourceGoogleDeploymentCreate,
		Read:   resourceGoogleDeploymentRead,
		Delete: resourceGoogleDeploymentDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"target_configuration": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceGoogleDeploymentCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	deployment := deploymentmanager.Deployment{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Target: &deploymentmanager.TargetConfiguration{
			Config: d.Get("target_configuration").(string),
			// TODO: Imports:
		},
	}

	o, err := config.clientDeployment.Deployments.Insert(project, deployment).Do()
	if err != nil {
		return err
	}

	// TODO: Wait until it's ready

	return resourceGoogleDeploymentRead(d, meta)
}

func resourceGoogleDeploymentRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	deployment, err := config.clientDeployment.Deployments.Get(project, d.Get("name").(string)).Do()
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%d", deployment.Id))
	d.Set("name", deployment.Name)
	d.Set("description", deployment.Description)
	d.Set("target_configuration", deployment.Target.Config)

	return nil
}

func resourceGoogleDeploymentDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	o, err := config.clientDeployment.Deployments.Delete(project, d.Get("name").(string)).Do()
	if err != nil {
		return err
	}

	return nil
}
