# brightbox\_server\_type Data Source

Use this data source to get the ID of a Brightbox Server Type for use in other
resources.

## Example Usage

```hcl
data "brightbox_server_type" "4gb" {
	handle = "^4gb.nbs$"
}
```

## Argument Reference

* `name` - (Optional) A regex string to apply to the Server Type list returned
by Brightbox Cloud.

* `handle` - (Optional) A regex string to apply to the Server Type list
returned by Brightbox Cloud.

~> **NOTE:** arguments form a conjunction. All arguments must match to
select a type.

~> **NOTE:** If more or less than a single match is returned by the
search, Terraform will fail. Ensure that your search is specific enough
to return a single type.

## Attributes Reference

`id` is set to the ID of the found Server Type. In addition, the
following attributes are exported:

* `disk_size` - The disk size of the server for this type. Will be zero for network storage types.
* `ram` - The memory size of the server for this type.
* `cores` - The memory size of the server for this type.
* `status` - 'experimental', 'available' or 'deprecated'.
* `storage_type` - the type of block storage available with this type: 'network' or 'local'.
