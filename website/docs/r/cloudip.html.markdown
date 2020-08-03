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
  port_translator {
	  protocol = "tcp"
	  incoming = 80
	  outgoing = 8080
  }
  port_translator {
	  protocol = "udp"
	  incoming = 53
	  outgoing = 8053
  }
}

resource "brightbox_server" "web" {
  image  = "img-testy"
  name   = "web-1"
  zone = "gb1-a"
  type = "512mb.ssd"
  server_groups = [ "grp-testy" ]
}
```

Cloud ips can just be reserved

```hcl
resource "brightbox_cloudip" "myapp-public" {
  name = "Reserved for use by application"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) a label to assign to the CloudIP
* `reverse_dns` - (Optional) The reverse DNS entry for the CloudIP
* `target` - (Optional) The CloudIP mapping target. This is the interface id from a server, or the id of a load balancer, server group or cloud sql resource.
* `port_translator` - (Optional) An array of port translator blocks. The Port Translator block is descibed below

Note that the default group for each account cannot be used as the target for a cloud ip.

Port Translator (`port_translator`) supports the following:
* `incoming` - (Required) The Port number traffic is coming in on the network
* `outgoing` - (Required) The Port number traffic is received at the mapped device
* `protocol` - (Required) The protocol of the port translator. Either `tcp` or `udp`

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the CloudIP
* `fqdn` - Fully Qualified Domain Name of the CloudIP
* `public_ip` - the public IPV4 address of the CloudIP
* `public_ipv4` - the public IPV4 address of the CloudIP
* `public_ipv6` - the public IPV6 address of the CloudIP
* `status` - Current state of the CloudIP: `mapped` or `unmapped`
* `username` - The username used to log onto the server

## Import

CloudIPs can be imported using the `id`, e.g.

```
terraform import brightbox_cloudip.mycloudip cip-vsalc
```

<a id="timeouts"></a>
## Timeouts

`brightbox_cloudip` provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - (Default `5 minutes`) Used for Mapping Cloud IPs
- `delete` - (Default `5 minutes`) Used for Unmapping Cloud IPs
