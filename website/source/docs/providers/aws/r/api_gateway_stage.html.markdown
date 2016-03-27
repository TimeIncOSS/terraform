---
layout: "aws"
page_title: "AWS: aws_api_gateway_stage"
sidebar_current: "docs-aws-resource-api-gateway-stage"
description: |-
  Provides an API Gateway Stage.
---

# aws\_api\_gateway\_stage

Provides an API Gateway Stage.

## Example Usage

```
resource "aws_api_gateway_rest_api" "demo" {
  name = "MyDemoAPI"
  description = "This is my API for demonstration purposes"
}

resource "aws_api_gateway_deployment" "demo" {
  depends_on = ["aws_api_gateway_integration.test"]
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  stage_name = "dev"
}

resource "aws_api_gateway_stage" "demo" {
  stage_name = "prod"
  rest_api_id = "${aws_api_gateway_rest_api.demo.id}"
  deployment_id = "${aws_api_gateway_deployment.demo.id}"
}
```

## Argument Reference

The following arguments are supported:

* `rest_api_id` - (Required) The ID of the associated REST API
* `stage_name` - (Required) The name of the stage
* `deployment_id` - (Required) The ID of the deployment that the stage points to
* `cache_cluster_enabled` - (Optional) Specifies whether a cache cluster is enabled for the stage
* `cache_cluster_size` - (Optional) The size of the cache cluster for the stage, if enabled.
	Allowed values include `0.5`, `1.6`, `6.1`, `13.5`, `28.4`, `58.2`, `118` and `237`.
* `client_certificate_id` - (Optional) The identifier of a client certificate for the stage.
* `description` - (Optional) The description of the stage
* `variables` - (Optional) A map that defines the stage variables
