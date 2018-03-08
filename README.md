Terraform Provider for [Brightbox Cloud](https://www.brightbox.com)
=======================================

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.11.x
-	[Go](https://golang.org/doc/install) 1.10 (to build the provider plugin)
-	[Dep](https://golang.github.io/dep/) v0.4.1 (to manage dependencies)

Usage
---------------------

```
# For example, restrict brightbox version to 1.x.x
provider "brightbox" {
  version = "~> 1.0"
}
```

Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/brightbox/terraform-provider-brightbox`

```sh
$ mkdir -p $GOPATH/src/github.com/brightbox; cd $GOPATH/src/github.com/brightbox
$ git clone git@github.com:brightbox/terraform-provider-brightbox
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/brightbox/terraform-provider-brightbox
$ make build
```

Using the provider
----------------------
This version supports managing:

* [Cloud Servers](https://www.brightbox.com/cloud/servers/)
* [Load Balancers](https://www.brightbox.com/cloud/load-balancing/)
* [Firewall Policies](https://www.brightbox.com/docs/reference/firewall/)
* [Cloud SQL Instances](https://www.brightbox.com/cloud/database/)
* [Cloud IPs](https://www.brightbox.com/blog/2014/02/27/design-decisions-cloud-ip-policy/)
* [Orbit Cloud Storage](https://www.brightbox.com/cloud/storage/) containers

Documentation
-------------------------

The announcement blog post gives a good overview:

https://www.brightbox.com/blog/2016/05/13/terraforming-brightbox-cloud/

And the getting started guide goes into more detail on how to use it

https://www.brightbox.com/docs/guides/terraform/getting-started/

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.10+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make bin
...
$ $GOPATH/bin/terraform-provider-brightbox
...
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```
