---
layout: "pingdom"
page_title: "Provider: Pingdom"
sidebar_current: "docs-pingdom-index"
description: |-
  The Pingdom provider is used to interact with the many resources supported by pingdom. The provider needs to be configured with the proper credentials before it can be used.
---

# Pingdom Provider

The Pingdom provider is used to interact with the
many resources supported by Pingdom. The provider needs to be configured
with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources.

## Example Usage

```
# Configure the Pingdom provider
provider "pingdom" {
    username = "john@doe.tld"
    password = "..."
    api_key = "..."
}

# Create a HTTP check
resource "pingdom_http_check" "test" {
  name = "test_check"
  hostname = "google.co.uk"
  notify = ["android", "sms"]
}
```

## Argument Reference

The following arguments are supported in the `provider` block:

* `username` - (Required) Your Pingdom username (email address)

* `password` - (Required) Your Pingdom password

* `api_key` - (Required) The Pingdom API key
  which you can generate and manage at https://my.pingdom.com/account/appkeys

