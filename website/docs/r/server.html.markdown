---
layout: "brightbox"
page_title: "Brightbox: brightbox_server"
sidebar_current: "docs-brightbox-resource-server"
description: |-
  Provides a Brightbox Server resource. This can be used to create, modify, and delete Servers. Servers also support provisioning.
---

# brightbox\_server

Provides a Brightbox Server resource. This can be used to create,
modify, and delete Servers. Servers also support
[provisioning](/docs/provisioners/index.html).

## Example Usage

```hcl
# Create a new 512Mb SSD Web Server in the gb1a zone
resource "brightbox_server" "web" {
  image  = "img-testy"
  name   = "web-1"
  zone = "gb1a"
  type = "512mb.ssd"
}
```

## Argument Reference

The following arguments are supported:

* `image` - (Required) The Server image ID
* `name` - (Optional) The Server name
* `type` - (Optional) The handle of the server type required (`1gb.ssd`, etc)
* `zone` - (Optional) The handle of the zone required (`gb1-a`, `gb1-b`)
* `user_data` (Optional) - A string of the desired User Data for the Server.
* `user_data_base64` (Optional) - Already encrypted User Data - for use
with the template provider.
* `server_groups` (Optional) - An array of server group ids the server should be added to.

~> **NOTE:** Only one of `user_data` or `user_data_base64` can be specified

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Server
* `fqdn` - Fully Qualified Domain Name of server
* `hostname` - short name of server, usually the same as the `id`
* `interface` - the id reference of the network interface. Used to target cloudips.
* `ipv4_address_private` - The RFC 1912 address of the server
* `ipv6_address` - the IPv6 address of the server
* `ipv6_hostname` - the FQDN of the IPv6 address
* `public_hostname` - the FQDN of the public IPv4 address. Appears if a cloud ip is mapped
* `ipv4_address` - the public IPV4 address of the server. Appears if a cloud ip is mapped
* `locked` - True if server has been set to locked and cannot be deleted
* `status` - Current state of the server, usually `active`, `inactive`
or `deleted`
* `username` - The username used to log onto the server

## Import

Servers can be imported using the server `id`, e.g.

```
terraform import brightbox_server.myserver srv-ojy3o
```
