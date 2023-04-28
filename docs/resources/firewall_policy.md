# brightbox\_firewall\_policy Resource

Provides a Brightbox Firewall Policy resource.

## Example Usage

```hcl
resource "brightbox_server_group" "default" {
  name = "Terraform"
}

resource "brightbox_firewall_policy" "default" {
  name         = "Terraform"
  server_group = brightbox_server_group.default.id
}
```

## Argument Reference

The following arguments are supported:

* `server_group` - (Optional) The ID of the Server Group the policy will be applied to
* `name` - (Optional) A label to assign to the Firewall Policy
* `description` - (Optional) A further description of the Firewall Policy

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Firewall Policy

## Import

Firewall Policies can be imported using the `id`, e.g.

```
terraform import brightbox_firewall_policy.mypolicy fwp-zxcvb
```
