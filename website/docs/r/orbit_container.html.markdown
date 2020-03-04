---
layout: "brightbox"
page_title: "Brightbox: brightbox_orbit_container"
sidebar_current: "docs-brightbox-resource-orbit-container"
description: |-
  Provides a Brightbox Orbit Container resource. This can be used to create, modify, and delete Containers in Orbit.
---

# brightbox\_orbit\_container

Provides a Brightbox Orbit Container resource. This can be used to create,
modify, and delete Containers in Orbit.

## Example Usage

```hcl
# Example Container
resource "brightbox_orbit_container" "initial" {
  name = "initial"
  metadata = {
    "description" = "Initial database snapshots"
  }
  container_read = ["acc-testy", "acc-12345"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A label assigned to the Orbit container
* `metadata` - (Optional) A dictionary of metadata key/value items. The key must be in lower case with no underscores or spaces
* `container_read` (Optional) A set of accounts and referrals that are allowed to read the Orbit container
* `container_write` (Optional) A set of accounts and referrals that are allowed to write to the Orbit container
* `container_sync_key` (Optional) Sets the secret key for Orbit container synchronization. If this is cleared synchronisation stops
* `container_sync_to` (Optional) Sets the destination for Orbit container synchronization. Used with `container_sync_key`
* `versions_location` (Optional) The Orbit container to hold previous versions of this Orbit container's contents, which are automatically restored if an item is deleted. Cannot be used at the same time as `history_location`
* `history_location` (Optional) The Orbit container to hold previous versions of this Orbit container's contents, where delete copies the item to history from this container. Cannot be used at the same time as `versions_location`

## Attributes Reference

The following attributes are exported:

* `object_count` - The number of items in the Orbit Container
* `bytes_used` - The total size of the items in the Orbit Container
* `storage_policy` - The storage policy in place for this container. Always 'Policy-0' at present
* `created_at` - The time the container was created

## Import

Orbit Containers can be imported using the `name`, e.g.

```
terraform import brightbox_orbit_container.myorbitcontainer initial
```

