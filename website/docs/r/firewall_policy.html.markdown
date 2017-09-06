---
layout: "brightbox"
page_title: "Brightbox: brightbox_firewall_policy"
sidebar_current: "docs-brightbox-resource-firewall-policy"
description: |-
  Provides a Brightbox Firewall Policy resource.
---

# brightbox\_firewall\_policy

Provides a Brightbox Firewall Policy resource.

## Example Usage

```hcl
resource "brightbox_server_group" "default" {
  name = "Terraform"
}

resource "brightbox_firewall_policy" "default" {
  name         = "Terraform"
  server_group = "${brightbox_server_group.default.id}"
}
```

## Argument Reference

The following arguments are supported:

* `server_group` - (Optional) The ID of the Server Group the policy will be applied to
* `name` - (Optional) A label to assign to the Firewall Policy
* `description` - (Optional) A further description of the Firewall Policy

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Firewall Policy
