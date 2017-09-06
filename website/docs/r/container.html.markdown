---
layout: "brightbox"
page_title: "Brightbox: brightbox_container"
sidebar_current: "docs-brightbox-resource-container"
description: |-
  Provides a Brightbox Container resource. This can be used to create, modify, and delete Containers in Orbit.
---

# brightbox\_container

Provides a Brightbox Container resource. This can be used to create,
modify, and delete Containers in Orbit.

## Example Usage

```hcl
# Example Container
resource "brightbox_container" "initial" {
  name = "initial"
  description = "Initial database snapshots"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A label assigned to the Container
* `description` - (Optional) A further description of the Container
* `orbit_url` - (Optional) The Orbit URL you wish to talk to. This defaults to either `https://orbit.brightbox.com/v1/` or the contents of the `BRIGHTBOX_ORBIT_URL` environment variable if set.


## Attributes Reference

The following attributes are exported:

* `auth_user` - the api client id used to access the container
* `auth_key` - the client secret used to access the container
* `account_id` - the account under which the container is stored

