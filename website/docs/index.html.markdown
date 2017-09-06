---
layout: "brightbox"
page_title: "Provider: Brightbox"
sidebar_current: "docs-brightbox-index"
description: |-
  The Brightbox provider is used to interact with the resources supported by Brightbox Cloud. The provider needs to be configured with the proper credentials before it can be used.
---

# Brightbox Provider

The Brightbox provider is used to interact with the resources supported
by Brightbox Cloud. The provider needs to be configured with the proper
credentials before it can be used.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
# Configure the Brightbox Provider
provider "brightbox" {
  username = "${var.user_email_address}"
  password = "${var.user_secret_password}"
  account = "${var.account_to_work_on}"
}

# Create a web server
resource "brightbox_server" "web" {
  # ...
}
```

## Authentication

The Brightbox provider offers a flexible means of providing credentials for
authentication. The following methods are supported, in this order, and
explained below:

- Username credentials
- Static credentials
- Username Environment variables
- Static Environment variables

### Username credentials ###

Username credentials can be provided by adding a `username` and
`password` in-line in the Brightbox provider block:

Usage:

```hcl
provider "brightbox" {
  username = "someone@example.com"
  password = "secretpassword"
}
```

This will operate on the default account for the user. If you are the
collaborator on more than one account, you can select a different account
by adding an `account` argument.

```hcl
provider "brightbox" {
  username = "someone@example.com"
  password = "secretpassword"
  account  = "acc-diffr"
}
```

### Static credentials ###

Static credentials can be provided by adding an `apiclient` and
`apisecret` in-line in the Brightbox provider block:

Usage:

```hcl
provider "brightbox" {
  apiclient = "cli-testy"
  apisecret = "secretcode"
}
```

API clients will only work on the account they are generated for. 

### Username Environment variables

You can provide your username and password via the `BRIGHTBOX_USER_NAME` and
`BRIGHTBOX_PASSWORD` environment variables. If required you can provide a non-default account with the `BRIGHTBOX_ACCOUNT` variable.

```hcl
provider "brightbox" {}
```

Usage:

```hcl
$ export BRIGHTBOX_USER_NAME="someone@example.com"
$ export BRIGHTBOX_PASSWORD="secretpassword"
$ export BRIGHTBOX_ACCOUNT="acc-diffr"
$ terraform plan
```

### Static Environment variables

You can provide your api client id and secret via the `BRIGHTBOX_CLIENT` and
`BRIGHTBOX_CLIENT_SECRET` environment variables. This will operate on
the account that issued the client id.

```hcl
provider "brightbox" {}
```

Usage:

```hcl
$ export BRIGHTBOX_CLIENT="cli-testy"
$ export BRIGHTBOX_PASSWORD="secretcode"
$ terraform plan
```

## Argument Reference

The following arguments are supported:

* `apiclient` - (optional) This is the Brightbox client id for an
account. This can also be specified with the `BRIGHTBOX_CLIENT` shell
environment variable.

* `apisecret` - (optional) This is the Brightbox client secret. This can
also be specified with the `BRIGHTBOX_CLIENT_SECRET` shell environment
variable.

* `username` - (optional) This is the Brightbox user logon. This can
also be specified with the `BRIGHTBOX_USER_NAME` shell environment
variable.

* `password` - (optional) This is the Brightbox user logon password. This
can also be specified with the `BRIGHTBOX_PASSWORD` shell environment
variable.

* `account` - (optional) This is the Brightbox account you wish to
operate upon. This can also be specified with the `BRIGHTBOX_ACCOUNT`
shell environment variable.

* `apiurl` - (Optional) Use this to override the default endpoint URL
constructed for the region. It's typically used to connect to custom
Brightbox endpoints.

~> **NOTE:** At least one of `username` or `apiclient` must be specified.
