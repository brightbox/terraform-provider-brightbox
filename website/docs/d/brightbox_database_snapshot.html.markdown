---
layout: "brightbox"
page_title: "Brightbox: brightbox_database_snapshot"
sidebar_current: "docs-brightbox-datasource-image"
description: |-
  Get information about a Brightbox Database Snapshot.
---

# brightbox\_database\_snapshot

Use this data source to get the ID of a Brightbox Database Snapshot for
use in other resources.

## Example Usage

```hcl
data "brightbox_database_snapshot" "today" {
	name = "Main db"
	most_recent = true
}
```

## Argument Reference

* `most_recent` - (Optional) If more than one result is returned, use
the most recent image based upon the `created_at` time.

* `name` - (Optional) A regex string to apply to the Database Snapshot
list returned by Brightbox Cloud.

* `description` - (Optional) A regex string to apply to the Database
Snapshot list returned by Brightbox Cloud.

* `database_engine` = (Optional) The engine of the database used to create the snapshot, e.g. mysql

* `database_version` = (Optional) The version of the database used to create the snapshot, e.g. 8.0

~> **NOTE:** arguments form a conjunction. All arguments must match to
select an image.

~> **NOTE:** If more or less than a single match is returned by the
search, Terraform will fail. Ensure that your search is specific enough
to return a single image only, or use `most_recent` to choose the most
recent one.

## Attributes Reference

`id` is set to the ID of the found Database Snapshot. In addition, the following attributes are exported:

* `size` - The size of database partition in megabytes
* `status` - The state the image is in. Usually `available`, or `deleted`.
* `created_at` - The time and date the image was created/registered (UTC)
* `locked` - true if image has been set as locked and can not be deleted
