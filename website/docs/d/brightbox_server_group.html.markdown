---
layout: "brightbox"
page_title: "Brightbox: brightbox_server_group"
sidebar_current: "docs-brightbox-datasource-server-group"
description: |-
  Get information about a Brightbox Server Group
---

# brightbox\_server\_group

Use this data source to get the ID of a Brightbox Server Group for use in other
resources.

## Example Usage

```hcl
data "brightbox_server_group" "defaul" {
	name = "^default$"
}
```

## Argument Reference

* `name` - (Optional) A regex string to apply to the Server Group list returned
by Brightbox Cloud.

* `description` - (Optional) A regex string to apply to the Server Group list
returned by Brightbox Cloud.

~> **NOTE:** arguments form a conjunction. All arguments must match to
select an image.

~> **NOTE:** If more or less than a single match is returned by the
search, Terraform will fail. Ensure that your search is specific enough
to return a single image only, or use `most_recent` to choose the most
recent one.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Server
