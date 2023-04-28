# brightbox\_server\_group Data Source

Use this data source to get the ID of a Brightbox Server Group for use in other
resources.

## Example Usage

```hcl
data "brightbox_server_group" "default" {
	name = "^default$"
}
```

## Argument Reference

* `name` - (Optional) A regex string to apply to the Server Group list returned
by Brightbox Cloud.

* `description` - (Optional) A regex string to apply to the Server Group list
returned by Brightbox Cloud.

~> **NOTE:** arguments form a conjunction. All arguments must match to
select a server group.

~> **NOTE:** If more or less than a single match is returned by the
search, Terraform will fail. Ensure that your search is specific enough
to return a single image only, or use `most_recent` to choose the most
recent one.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Server
* `default` - Is this the default server group?
* `fqdn` - The Fully Qualified Domain Name of the server group
* `firewall_policy` - The ID of the Firewall Policy associated with this group
