# brightbox\_database\_server Resource

Provides a Brightbox Database Server resource. This can be used to create,
modify, and delete Database Servers.

## Example Usage

```hcl
resource "brightbox_database_server" "default" {
	name = "Default DB"
	description = "Default DB used by servers"
	database_engine = "mysql"
	database_version = "8.0"
	database_type = data.brightbox_database_type.4gb.id
	maintenance_weekday = 5
	maintenance_hour = 4
	snapshots_schedule = "0 5 * * *"
	allow_access = [
		brightbox_server_group.barfoo.id,
		brightbox_server.foobar.id,
		"158.152.1.65/32"
	]
}

data "brightbox_database_type" "4gb" {
	name = "^SSD 4GB$"
}

resource "brightbox_server" "foobar" {
	name = "database access"
	image = "img-testy"
	server_groups = [ brightbox_server_group.barfoo.id ]
}

resource "brightbox_server_group" "barfoo" {
	name = "database access group"
}
```

## Argument Reference

The following arguments are supported:

* `allow_access` (Required) - A list of server group ids, server ids or IPv4 address references the database server should be accessible from. There must be at least one entry in the list
* `name` - (Optional) A label assigned to the Database Server
* `description` - (Optional) A further description of the Database Server
* `maintenance_weekday` - (Optional) Numerical index of weekday (0 is Sunday, 1 is Monday...) to set when automatic updates may be performed. Default is 0 (Sunday). 
* `maintenance_hour` - (Optional) Number representing 24hr time start of maintenance window hour for x:00-x:59 (0-23). Default is 6
* `snapshots_retention` - (Optional) Keep this number of scheduled snapshots. Keep all if unset.
* `snapshots_schedule` - (Optional) A crontab pattern to determine approximately when scheduled snapshots will run (must be at least hourly)
* `database_engine` - (Optional) Database engine to request. Default is mysql
* `database_version` - (Optional) Database version to request. Default is 8.0
* `database_type` - (Optional) ID of the Database Type required
* `snapshot` (Optional) - Database snapshot id to build from
* `zone` - (Optional) The handle of the zone required (`gb1-a`, `gb1-b`)
* `locked` - (Optional) Set to true to stop the database server from being deleted

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Database Server
* `admin_username` - The username used to log onto the database
* `admin_password` - The password used to log onto the database
* `status` - Current state of the database server, usually `active` or `deleted`
* `snapshots_schedule_next_at` - The approximate UTC time when the next snapshot is scheduled

## Import

Database Servers can be imported using the `id`, e.g.

```
terraform import brightbox_database_server.mydatabase dbs-qwert
```

<a id="timeouts"></a>
## Timeouts

`brightbox_database_server` provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - (Default `5 minutes`) Used for Creating Databases
- `delete` - (Default `5 minutes`) Used for Deleting Databases
