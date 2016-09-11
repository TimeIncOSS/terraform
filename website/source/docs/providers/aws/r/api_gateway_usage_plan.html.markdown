---
layout: "aws"
page_title: "AWS: aws_api_gateway_usage_plan"
sidebar_current: "docs-aws-resource-api-gateway-usage-plan"
description: |-
  Provides an API Gateway Usage Plan.
---

# aws\_api\_gateway\_usage\_plan

Provides an API Gateway Usage Plan.

## Example Usage

```
resource "aws_api_gateway_usage_plan" "demo" {
  
}

```

## Argument Reference

The following arguments are supported:

* `description` - (Optional) The description of a usage plan.
* `name` - (Optional) The name of a usage plan.
* `quota` - (Optional) The maximum number of permitted requests per a given interval.
* `throttle` - (Optional) The request throttle limits of a usage plan. See supported fields below.

## Import

API Gateway Usage Plans can be imported using the id, e.g.

```
$ terraform import aws_api_gateway_usage_plan.demo ab1cqe
```
