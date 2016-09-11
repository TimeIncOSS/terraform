package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsApiGatewayUsagePlan() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayUsagePlanCreate,
		Read:   resourceAwsApiGatewayUsagePlanRead,
		Update: resourceAwsApiGatewayUsagePlanUpdate,
		Delete: resourceAwsApiGatewayUsagePlanDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"quota": {
				Type:     schema.TypeList,
				MaxItems: 1,
			},
			"throttle": {
				Type:     schema.TypeList,
				MaxItems: 1,
			},
		},
	}
}

func resourceAwsApiGatewayUsagePlanCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	input := apigateway.CreateUsagePlanInput{}
	log.Printf("[DEBUG] Creating API Gateway Usage Plan: %s", input)
	out, err := conn.CreateUsagePlan(&input)
	if err != nil {
		return fmt.Errorf("Failed to create usage plan: %s", err)
	}

	d.SetId(*out.Id)
	d.Set("description", out.Description)
	d.Set("name", out.Name)
	d.Set("quota", flattenApiGatewayQuota(out.Quota))
	d.Set("throttle", flattenApiGatewayThrottle(out.Throttle))

	return nil
}

func resourceAwsApiGatewayUsagePlanRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	input := apigateway.GetUsagePlanInput{
		UsagePlanId: aws.String(d.Id()),
	}
	out, err := conn.GetUsagePlan(&input)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NotFoundException" {
			log.Printf("[WARN] API Gateway Usage Plan %s not found, removing", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}
	log.Printf("[DEBUG] Received API Gateway Usage Plan: %s", out)

	d.Set("description", out.Description)
	d.Set("name", out.Name)
	d.Set("quota", flattenApiGatewayQuota(out.Quota))
	d.Set("throttle", flattenApiGatewayThrottle(out.Throttle))

	return nil
}

func resourceAwsApiGatewayUsagePlanUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	operations := make([]*apigateway.PatchOperation, 0)
	if d.HasChange("description") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/description"),
			Value: aws.String(d.Get("description").(string)),
		})
	}

	input := apigateway.UpdateUsagePlanInput{
		UsagePlanId:     aws.String(d.Id()),
		PatchOperations: operations,
	}

	log.Printf("[DEBUG] Updating API Gateway Usage Plan: %s", input)
	_, err := conn.UpdateClientCertificate(&input)
	if err != nil {
		return fmt.Errorf("Updating API Gateway Usage Plan failed: %s", err)
	}

	return resourceAwsApiGatewayUsagePlanRead(d, meta)
}

func resourceAwsApiGatewayUsagePlanDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	log.Printf("[DEBUG] Deleting API Gateway Usage Plan: %s", d.Id())
	input := apigateway.DeleteUsagePlanInput{
		UsagePlanId: aws.String(d.Id()),
	}
	_, err := conn.DeleteUsagePlan(&input)
	if err != nil {
		return fmt.Errorf("Deleting API Gateway Usage Plan failed: %s", err)
	}

	return nil
}
