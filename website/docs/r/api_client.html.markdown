---
layout: "brightbox"
page_title: "Brightbox: brightbox_api_client"
sidebar_current: "docs-brightbox-resource-api-client"
description: |-
  Provides a Brightbox API Client resource.
---

# brightbox\_firewall\_policy

Provides a Brightbox API Client resource.

## Example Usage

```hcl
resource "brightbox_api_client" "default" {
  name              = "Terraform"
  description       = "Terraform API Client"
  permissions_group = "storage"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) A label to assign to the API Client
* `description` - (Optional) A further description of the API Client
* `permissions_group` - (Optional) The type of API Client required, either `full` or `storage`. The default is `full`.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the API Client
* `secret` - The initial secret key of the API Client
* `account` - The ID of the account the API Client is linked to

