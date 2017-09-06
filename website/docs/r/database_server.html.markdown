---
layout: "brightbox"
page_title: "Brightbox: brightbox_database_server"
sidebar_current: "docs-brightbox-resource-database-server"
description: |-
  Provides a Brightbox Database Server resource. This can be used to create, modify, and delete Database Servers.
---

# brightbox\_database\_server

Provides a Brightbox Database Server resource. This can be used to create,
modify, and delete Database Servers.

## Example Usage

```hcl
resource "brightbox_database_server" "default" {
	name = "Default DB"
	description = "Default DB used by servers"
	database_engine = "mysql"
	database_version = "5.6"
	maintenance_weekday = 5
	maintenance_hour = 4
	allow_access = [
		"${brightbox_server_group.barfoo.id}",
		"${brightbox_server.foobar.id}",
		"158.152.1.65/32"
	]
}

resource "brightbox_server" "foobar" {
	name = "database access"
	image = "img-testy"
}

resource "brightbox_server_group" "barfoo" {
	name = "database access group"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) A label assigned to the Database Server
* `description` - (Optional) A further description of the Database Server
* `maintenance_weekday` - (Optional) Numerical index of weekday (0 is Sunday, 1 is Monday...) to set when automatic updates may be performed. Default is 0 (Sunday). 
* `maintenance_hour` - (Optional) Number representing 24hr time start of maintenance window hour for x:00-x:59 (0-23). Default is 6
* `database_engine` - (Optional) Database engine to request. Default is mysql.
* `database_version` - (Optional) Database version to request. Default is 5.5.
* `allow_access` (Optional) - An array of server group ids, server ids or IPv4 address references the database server should be accessible from
* `snapshot` (Optional) - Database snapshot id to build from
* `zone` - (Optional) The handle of the zone required (`gb1-a`, `gb1-b`)

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Database Server
or `deleted`
* `admin_username` - The username used to log onto the database
* `admin_password` - The password used to log onto the database
* `status` - Current state of the database server, usually `active`
* `locked` - True if database server has been set to locked and cannot be deleted
* `cloud_ip_id` - The ID of the cloud ip if one is mapped. 
* `ipv4_address` - the public IPV4 address of the server. Appears if a cloud ip is mapped
* `public_hostname` - the FQDN of the public IPv4 address. Appears if a cloud ip
 is mapped

