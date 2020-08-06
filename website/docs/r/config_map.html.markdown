---
layout: "brightbox"
page_title: "Brightbox: brightbox_config_map"
sidebar_current: "docs-brightbox-resource-config-map"
description: |-
  Provides a Brightbox Config Map resource. This can be used to create, modify, and delete Config Maps.
---

# brightbox\_config\_map

Provides a Brightbox Config Map resource. This can be used to create,
modify, and delete Config Maps.

## Example Usage

```hcl
# Default Config Map
# the instances over SSH and HTTP
resource "brightbox_config_map" "default" {
  name = "Terraform config map"
  data = {"hostname":"tester", "ram": "1024"}
}

## Argument Reference

The following arguments are supported:

* `name` - (Optional) A label assigned to the Config Map
* `data` - (Required) A key value map of strings

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Server

## Import

Config Maps can be imported using the config map `id`, e.g.

```
terraform import brightbox_config_map.default cfg-ok8vw
```
