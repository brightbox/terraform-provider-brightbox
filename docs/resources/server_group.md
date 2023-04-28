# brightbox\_server\_group Resource

Provides a Brightbox Server Group resource. This can be used to create,
modify, and delete Server Groups.

## Example Usage

```hcl
# Default Server Group
resource "brightbox_server_group" "default" {
  name = "Terraform controlled servers"
}

# Create a new 512Mb SSD Web Server in the gb1-a zone
resource "brightbox_server" "web" {
  image  = "img-testy"
  name   = "web-1"
  zone = "gb1-a"
  type = "512mb.ssd"
  server_groups = [ brightbox_server_group.default.id ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) A label assigned to the Server Group
* `description` - (Optional) A further description of the Server Group


## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Server
* `default` - Is this the default server group?
* `fqdn` - The Fully Qualified Domain Name of the server group
* `firewall_policy` - The ID of the Firewall Policy associated with this group

## Import

Server Groups can be imported using the server group `id`, e.g.

```
terraform import brightbox_server_group.default grp-ok8vw
```
