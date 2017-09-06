---
layout: "brightbox"
page_title: "Brightbox: brightbox_firewall_rule"
sidebar_current: "docs-brightbox-resource-firewall-rule"
description: |-
  Provides a Brightbox Firewall Rule resource.
---

# brightbox\_firewall\_rule

Provides a Brightbox Firewall Rule resource.

## Example Usage

```hcl
resource "brightbox_server_group" "default" {
  name = "Terraform"
}

resource "brightbox_firewall_policy" "default" {
  name         = "Terraform"
  server_group = "${brightbox_server_group.default.id}"
}

resource "brightbox_firewall_rule" "default_ssh" {
  destination_port = 22
  protocol         = "tcp"
  source           = "any"
  description      = "SSH access from anywhere"
  firewall_policy  = "${brightbox_firewall_policy.default.id}"
}

```

## Argument Reference

The following arguments are supported:

* `firewall_policy` - (Required) The ID of the firewall policy this rule belongs to
* `protocol` - (Optional) Protocol Number or one of `tcp`, `udp`, `icmp`
* `source` - (Optional) Subnet, ServerGroup or ServerID. `any`,`10.1.1.23/32` or `srv-4ktk4`
* `source_port` - (Optional) single port, multiple ports or range separated by `-` or `:`; upto 255 characters. Example - `80`, `80,443,21` or `3000-3999`
* `destination` - (Optional) Subnet, ServerGroup or ServerID. `any`,`10.1.1.23/32` or `srv-4ktk4`
* `destination_port` - (Optional) single port, multiple ports or range separated by `-` or `:`; upto 255 characters. Example - `80`, `80,443,21` or `3000-3999`
* `icmp_type_name` - (Optional) ICMP type name. `echo-request`, `echo-reply`. Only allowed if protocol is `icmp`.
* `description` - (Optional) A further description of the Firewall Rule

~> **NOTE:** Only one of `source` or `destination` can be specified

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Firewall Rule
