---
layout: "brightbox"
page_title: "Brightbox: brightbox_server_group"
sidebar_current: "docs-brightbox-resource-server-group"
description: |-
  Provides a Brightbox Server Group resource. This can be used to create, modify, and delete Server Groups.
---

# brightbox\_server\_group

Provides a Brightbox Server Group resource. This can be used to create,
modify, and delete Server Groups.

## Example Usage

```hcl
# Default Server Group
# the instances over SSH and HTTP
resource "brightbox_server_group" "default" {
  name = "Terraform controlled servers"
}

# Create a new 512Mb SSD Web Server in the gb1-a zone
resource "brightbox_server" "web" {
  image  = "img-testy"
  name   = "web-1"
  zone = "gb1-a"
  type = "512mb.ssd"
  server_groups = ["${brightbox_server_group.default.id}"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) A label assigned to the Server Group
* `description` - (Optional) A further description of the Server Group

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Server

## Import

Server Groups can be imported using the server group `id`, e.g.

```
terraform import brightbox_server_group.default grp-ok8vw
```
