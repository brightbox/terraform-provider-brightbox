---
layout: "brightbox"
page_title: "Brightbox: brightbox_image"
sidebar_current: "docs-brightbox-datasource-image"
description: |-
  Get information about a Brightbox Image.
---

# brightbox\_image

Use this data source to get the ID of a Brightbox Image for use in other
resources.

## Example Usage

```hcl
data "brightbox_image" "ubuntu_lts" {
	name = "^ubuntu-xenial.*server$"
	arch = "x86_64"
	official = true
	most_recent = true
}
```

## Argument Reference

* `most_recent` - (Optional) If more than one result is returned, use
the most recent image based upon the `created_at` time.

* `name` - (Optional) A regex string to apply to the Image list returned
by Brightbox Cloud.

* `description` - (Optional) A regex string to apply to the Image list
returned by Brightbox Cloud.

* `source_type` - (Optional) Either `upload` or `snapshot`.

* `owner` - (Optional) The account id that owns the image. Matches
exactly.

* `arch` - (Optional) The architecture of the image: either `x86_64` or
`i686`.

* `public` - (Optional) Boolean to select a public image.

* `official` - (Optional) Boolean to select an official image.

* `compatibility_mode` - (Optional) Boolean to match the compatibility
mode flag.

* `username` - (Optional) The username used to logon to the image. Matches
exactly.

* `ancestor_id` - (Optional) The image id of the parent of the image
you are looking for.

* `licence_name` - (Optional) The name of the licence for the
image. Matches exactly.

~> **NOTE:** arguments form a conjunction. All arguments must match to
select an image.

~> **NOTE:** If more or less than a single match is returned by the
search, Terraform will fail. Ensure that your search is specific enough
to return a single image only, or use `most_recent` to choose the most
recent one.

## Attributes Reference

`id` is set to the ID of the found Image. In addition, the following attributes
are exported:

* `status` - The state the image is in. Usually `available`, `deprecated`
or `deleted`.
* `created_at` - The time and date the image was created/registered (UTC)
* `locked` - true if image has been set as locked and can not be deleted
* `virtual_size` - The virtual size of the disk image "container" in MB
* `disk_size` - The actual size of the data within the Image in MB
