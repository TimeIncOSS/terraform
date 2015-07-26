---
layout: "google"
page_title: "Google: google_deployment"
sidebar_current: "docs-google-resource-deployment"
description: |-
  Creates a Google Cloud deployment for Deployment Manager.
---

# google\_deployment

...


## Example Usage

```
resource "google_deployment" "default" {
	name = "test-cluster"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) 
* `description` - (Optional) 
* `target_configuration` - (Required) 
