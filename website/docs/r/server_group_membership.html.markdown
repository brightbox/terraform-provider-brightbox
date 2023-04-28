---
layout: "brightbox"
page_title: "Brightbox: brightbox_server_group_membership"
sidebar_current: "docs-brightbox-resource-server-group-membership"
description: |-
  Provides a resource for adding servers to server groups without conflicting with itself.
---

# brightbox\_server\_group\_membership

Provides a resource for adding an [Servers][2] to [Server Groups][1]. This
resource can be used multiple times with the same group for non-overlapping servers.

To exclusively manage the groups, use the `server_groups` attribute within [Servers.][2]

## Example Usage

```terraform
resource "brightbox_server_group_membership" "foobar" {
	group = brightbox_server_group.group1.id
	servers = [
		brightbox_server.server1.id,
	]
}

resource "brightbox_server_group_membership" "foobar2" {
	group = brightbox_server_group.group1.id
	servers = [
		brightbox_server.server3.id,
	]
}

resource "brightbox_server_group_membership" "barfoo" {
	group = brightbox_server_group.group2.id
	servers = [
		brightbox_server.server1.id,
	]
}

resource "brightbox_server_group" "group1" {
    name = "group1"
}

resource "brightbox_server_group" "group2" {
    name = "group2"
}

resource "brightbox_server" "server1" {
    name = "server1"
    image = data.brightbox_image.foobar.id
	type = "1gb.ssd"
}

resource "brightbox_server" "server2" {
    name = "server2"
    image = data.brightbox_image.foobar.id
	type = "1gb.ssd"
}

resource "brightbox_server" "server3" {
    name = "server3"
    image = data.brightbox_image.foobar.id
	type = "1gb.ssd"
}
```

## Argument Reference

The following arguments are supported:

* `group` - (Required) The name of the [Server Group.][1]
* `servers` - (Required) A list of [Servers][2] to add to the group.

## Attributes Reference

No additional attributes are exported.

[1]: /docs/providers/brightbox/r/server_group.html
[2]: /docs/providers/brightbox/r/server.html

## Import

Server group membership can be imported using the group id and server ids separated by `/`.

```
$ terraform import brightbox_server_group_membership.example1 grp-12345/srv-abcde/srv-fghij
```
