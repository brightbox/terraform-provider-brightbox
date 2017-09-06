---
layout: "brightbox"
page_title: "Brightbox: brightbox_cloudip"
sidebar_current: "docs-brightbox-resource-cloudip"
description: |-
  Provides a Brightbox CloudIP resource.
---

# brightbox\_cloudip

Provides a Brightbox CloudIP resource.

## Example Usage

```hcl
resource "brightbox_cloudip" "web-public" {
  target = "${brightbox_server.web.interface}"
  name = "web-1 public address"
}

resource "brightbox_server" "web" {
  image  = "img-testy"
  name   = "web-1"
  zone = "gb1a"
  type = "512mb.ssd"
}
```

## Argument Reference

The following arguments are supported:

* `target` - (Required) The CloudIP mapping target. This is the interface from a server, or the id of a load balancer or cloud sql resource
* `name` - (Optional) a label to assign to the CloudIP
* `reverse_dns` - (Optional) The reverse DNS entry for the CloudIP

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the CloudIP
* `fqdn` - Fully Qualified Domain Name of the CloudIP
* `public_ip` - the public IPV4 address of the CloudIP
* `status` - Current state of the CloudIP: `mapped` or `unmapped`
* `username` - The username used to log onto the server

## Import

CloudIPs can be imported using the  `id`, e.g.

```
terraform import brightbox_cloudip.mycloudip cip-vsalc
```
