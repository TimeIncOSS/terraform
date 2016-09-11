package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsApiGatewayMethodSettings() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayMethodSettingsCreate,
		Read:   resourceAwsApiGatewayMethodSettingsRead,
		Update: resourceAwsApiGatewayMethodSettingsUpdate,
		Delete: resourceAwsApiGatewayMethodSettingsDelete,

		Schema: map[string]*schema.Schema{
			"rest_api_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"stage_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"path": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"metrics_enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"logging_level": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"data_trace_enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"throttling_burst_limit": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"throttling_rate_limit": &schema.Schema{
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"caching_enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"cache_ttl_in_seconds": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"cache_data_encrypted": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"require_authorization_for_cache_control": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"unauthorized_cache_control_header_strategy": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAwsApiGatewayMethodSettingsCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	operations := make([]*apigateway.PatchOperation, 0)
	operations = append(operations, &apigateway.PatchOperation{
		Op:    aws.String("replace"),
		Path:  aws.String("/methodSettings/" + d.Get("path").(string) + "/metrics/enabled"),
		Value: aws.String(fmt.Sprintf("%t", d.Get("metrics_enabled").(bool))),
	})

	if v, ok := d.GetOk("logging_level"); ok {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/methodSettings/" + d.Get("path").(string) + "/logging/loglevel"),
			Value: aws.String(v.(string)),
		})
	}

	operations = append(operations, &apigateway.PatchOperation{
		Op:    aws.String("replace"),
		Path:  aws.String("/methodSettings/" + d.Get("path").(string) + "/logging/dataTrace"),
		Value: aws.String(fmt.Sprintf("%t", d.Get("data_trace_enabled").(bool))),
	})

	operations = append(operations, &apigateway.PatchOperation{
		Op:    aws.String("replace"),
		Path:  aws.String("/methodSettings/" + d.Get("path").(string) + "/throttling/burstLimit"),
		Value: aws.String(fmt.Sprintf("%d", d.Get("throttling_burst_limit").(int))),
	})

	operations = append(operations, &apigateway.PatchOperation{
		Op:    aws.String("replace"),
		Path:  aws.String("/methodSettings/" + d.Get("path").(string) + "/throttling/rateLimit"),
		Value: aws.String(fmt.Sprintf("%f", d.Get("throttling_rate_limit").(float64))),
	})

	if v, ok := d.GetOk("caching_enabled"); ok {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/methodSettings/" + d.Get("path").(string) + "/caching/enabled"),
			Value: aws.String(fmt.Sprintf("%t", v.(bool))),
		})
	}

	operations = append(operations, &apigateway.PatchOperation{
		Op:    aws.String("replace"),
		Path:  aws.String("/methodSettings/" + d.Get("path").(string) + "/caching/ttlInSeconds"),
		Value: aws.String(fmt.Sprintf("%d", d.Get("cache_ttl_in_seconds").(int))),
	})

	if v, ok := d.GetOk("cache_data_encrypted"); ok {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/methodSettings/" + d.Get("path").(string) + "/caching/dataEncrypted"),
			Value: aws.String(fmt.Sprintf("%t", v.(bool))),
		})
	}

	if v, ok := d.GetOk("require_authorization_for_cache_control"); ok {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/methodSettings/" + d.Get("path").(string) + "/caching/requireAuthorizationForCacheControl"),
			Value: aws.String(fmt.Sprintf("%t", v.(bool))),
		})
	}

	operations = append(operations, &apigateway.PatchOperation{
		Op:    aws.String("replace"),
		Path:  aws.String("/methodSettings/" + d.Get("path").(string) + "/caching/requireAuthorizationForCacheControl"),
		Value: aws.String(d.Get("unauthorized_cache_control_header_strategy").(string)),
	})

	input := apigateway.UpdateStageInput{
		RestApiId:       aws.String(d.Get("rest_api_id").(string)),
		StageName:       aws.String(d.Get("stage_name").(string)),
		PatchOperations: operations,
	}
	conn.UpdateStage(input)

	return nil
}

func resourceAwsApiGatewayMethodSettingsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	input := apigateway.GetStageInput{
		RestApiId: aws.String(d.Get("rest_api_id").(string)),
		StageName: aws.String(d.Get("stage_name").(string)),
	}
	out, err := conn.GetStage(&input)
	if err != nil {
		return err
	}
	settings := out.MethodSettings[d.Get("path").(string)]

	d.Set("cache_data_encrypted", settings.CacheDataEncrypted)
	d.Set("cache_ttl_in_seconds", settings.CacheTtlInSeconds)
	d.Set("caching_enabled", settings.CachingEnabled)
	d.Set("data_trace_enabled", settings.DataTraceEnabled)
	d.Set("logging_level", settings.LoggingLevel)
	d.Set("metrics_enabled", settings.MetricsEnabled)
	d.Set("require_authorization_for_cache_control", settings.RequireAuthorizationForCacheControl)
	d.Set("throttling_burst_limit", settings.ThrottlingBurstLimit)
	d.Set("throttling_rate_limit", settings.ThrottlingRateLimit)
	d.Set("unauthorized_cache_control_header_strategy", settings.UnauthorizedCacheControlHeaderStrategy)

	return nil
}
