---
layout: "brightbox"
page_title: "Brightbox: brightbox_database_type"
sidebar_current: "docs-brightbox-datasource-database-type"
description: |-
  Get information about a Brightbox Database Type.
---

# brightbox\_database\_type

Use this data source to get the ID of a Brightbox Database Type for use in other
resources.

## Example Usage

```hcl
data "brightbox_database_type" "4gb" {
	name = "^SSD 4GB$"
}
```

## Argument Reference

* `name` - (Optional) A regex string to apply to the Database Type list returned
by Brightbox Cloud.

* `description` - (Optional) A regex string to apply to the Database Type list
returned by Brightbox Cloud.

~> **NOTE:** arguments form a conjunction. All arguments must match to
select a type.

~> **NOTE:** If more or less than a single match is returned by the
search, Terraform will fail. Ensure that your search is specific enough
to return a single type.

## Attributes Reference

`id` is set to the ID of the found Database Type. In addition, the
following attributes are exported:

* `disk_size` - The disk size of the database server for this type
* `ram` - The memory size of the database server for this type
