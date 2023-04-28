# brightbox\_server Resource

Provides a Brightbox Server resource. This can be used to create,
modify, and delete Servers. Servers also support
[provisioning](/docs/provisioners/index.html).

## Example Usage

```hcl
# Create a new 512Mb SSD Web Server in the gb1-a zone
resource "brightbox_server" "web" {
  image  = "img-testy"
  name   = "web-1"
  zone = "gb1-a"
  type = "512mb.ssd"
  server_groups = [ "grp-testy" ]
}
```

## Argument Reference

The following arguments are supported:

* `image` - (Optional) The Server image ID. One of image or volume must be specified.
* `volume` - (Optional) The volume to be used to boot the server. One of image or volume must be specified.
* `server_groups` (Optional) - List of server group ids the server should be added to.
* `name` - (Optional) The Server name
* `type` - (Optional) The handle the server type required (`1gb.ssd`, etc), or a Server Type ID. 
* `zone` - (Optional) The handle of the zone required (`gb1-a`, `gb1-b`)
* `locked` - (Optional) Set to true to stop the server from being deleted
* `disk_encrypted` - (Optional) Create a server where the data on disk is
'encrypted as rest' by the cloud.
* `disk_size` - (Optional) The desired size of the disk storage for the
Server. Only usable with types using network block storage.
* `snapshots_retention` - (Optional) Keep this number of scheduled
snapshots. Keep all if unset.
* `snapshots_schedule` - (Optional) Crontab pattern for scheduled
snapshots. Must be no more frequent than hourly.
* `user_data` (Optional) - A string of the desired User Data for the Server.
* `user_data_base64` (Optional) - Already encrypted User Data - for use
with the template provider.

~> **NOTE:** Only one of `user_data` or `user_data_base64` can be specified

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Server
* `data_volumes` - List of data volumes attached to server.
* `fqdn` - Fully Qualified Domain Name of server
* `hostname` - short name of server, usually the same as the `id`
* `interface` - the id reference of the network interface. Used to target cloudips.
* `ipv4_address_private` - The RFC 1912 address of the server
* `ipv6_address` - the IPv6 address of the server
* `ipv6_hostname` - the FQDN of the IPv6 address
* `public_hostname` - the FQDN of the public IPv4 address. Appears if a cloud ip is mapped
* `ipv4_address` - the public IPV4 address of the server. Appears if a cloud ip is mapped
* `status` - Current state of the server, usually `active`, `inactive`
or `deleted`
* `snapshot_schedule_next_at` - Time in UTC of approximately when the next scheduled snapshot will run.
* `username` - The username used to log onto the server

## Import

Servers can be imported using the server `id`, e.g.

```
terraform import brightbox_server.myserver srv-ojy3o
```

<a id="timeouts"></a>
## Timeouts

`brightbox_server` provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - (Default `5 minutes`) Used for Creating Servers
- `delete` - (Default `5 minutes`) Used for Deleting Servers
