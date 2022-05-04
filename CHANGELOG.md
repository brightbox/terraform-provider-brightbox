## 3.0.0 (May 4, 2022)

IMPROVEMENTS:
- resource/volume: Add volume as a resource
- resource/server: Add ability to create servers from volumes and add data volumes to servers
- resource/server: Add 'snapshot_retention' and 'snapshot_schedule' attributes
- data/image - Add 'min_ram', 'source' and 'source_trigger' attributes
- data/database_snapshot - Add database snapshot type as a data source
- resource/cloudip: Add 'mode' attribute
- resource/database_server: Add 'snapshot_retention' attribute

NOTES:

- Update to gobrightbox v2
- Update Terraform SDK to 2.12.0
- Added additional Terraform validation to several resources
- Switch to context aware functions which respect cancellation timeouts
- Switch to Terraform diagnostics for errors, allowing multiple errors to be reported in one go

## 2.1.1 (February 16, 2022)

BUG FIXES:
- resource/server: Ensure disk resizing only allows an increase in the disk size

## 2.1.0 (February 14, 2022)

IMPROVEMENTS:

- data/server_type: Add server type as a data source
- resource/server: Add disk size attribute, for use with network block server types
- resource/server: Servers can be created by reference to a server type data resource
- resource/server: Remove force new restriction for server type
- resource/server: Allow dynamic resizing of disk size for network block server types

NOTES:

- Update to gobrightbox 0.8.2
- Update Terraform SDK to 2.10.1
- Fix errors in database type documentation

## 2.0.7 (February 1, 2022)

NOTES:

- Go version switched to 1.17 to generate darwin/windows arm64 images

## 2.0.6 (July 23, 2021)

BUG FIXES:

- resource/orbit_container: support 'created_at' attribute

NOTES:

- container 'created_at' attribute now supported by Swift drive

## 2.0.5 (May 25, 2021)

BUG FIXES:

- resource/orbit_container: stop 'created_at' attribute changing

NOTES:

- container 'created_at' attribute will be an empty string until the
  Timestamp is added to the underlying Swift driver

## 2.0.4 (April 27, 2021)

- Update Terraform SDK to version 2.6.1

## 2.0.3 (Mar 9, 2021)

BUG FIXES:

- resource/firewall_rule: allow load balancer ids in source/destination

NOTES:

- Fix repeated name in documentation

## 2.0.2 (Mar 2, 2021)

BUG FIXES:

- resource/load_balancer: add missing proxy_protocol support to listeners

NOTES:

- Update documentation to v13 HCL format

## 2.0.0 (Mar 1, 2021)

IMPROVEMENTS:

- resource/load_balancer: add https_redirect, ssl_minimum_version and domains support

NOTES:

- Added greater levels of attribute validation throughout
- Upgraded to verions 2 of the Terraform SDK
- resource/load_balancer: deprecate sslv3 attributedd
- Build with Go 1.15
- Ensure provider is lint clean
- Updated dependencies

## 1.5.0 (Nov 26, 2020)

IMPROVEMENTS:

- resource/server: add 'encryption at rest' support

NOTES:

- document new attribute

## 1.4.3 (Aug 19, 2020)

BUG FIXES:

- resource/server: ensure base64 encoded userdata is passed to API

## 1.4.2 (Aug 7, 2020)

BUG FIXES:

- resource/server_group: add missing default and FQDN attributes
- data/server_group: add missing default and FQDN attributes

NOTES:

- document missing attributes

## 1.4.1 (Aug 6, 2020)

IMPROVEMENTS:

- resource/config_map: added

BUG FIXES:

- resource/cloud_ip: add missing IPV4 and IPv6 entries
- resource/database_server: handle disabled snapshot schedule

NOTES:

- document missing cloudip entries
- document new config map resource

## 1.3.2 (Aug 3, 2020)

BUG FIXES:

- resource/firewall-rules: Fix firewall policy update

## 1.3.0 (May 21, 2020)

IMPROVEMENTS:

- Allow locking of database server, load balancer and server resources

BUG FIXES:

- resource/firewall-polices: Fix server group import

NOTES:

- Make provider lint clean
- Update dependency releases
- Update to Terraform SDK 1.12.0
- Switch to binary test runner
- Test with Go 1.14.2
- resource/cloud-ip: Deprecate erroneous lock attribute


## 1.2.0 (June 26, 2019)
- Update database versions in documentation
- Support Terraform 0.12

## 1.1.2 (April 17, 2019)
- Fix metadata delete for Orbit Containers
- Fix Port Translators for Cloud IPs

## 1.1.1 (April 09, 2019)
- Relax overly tight auth validation

## 1.1.0 (March 28, 2019)
- Add timeout support for loadbalancers
- Add timeout support for database servers
- Add timeout support for servers
- Add timeout support for Cloud IPs
- Add API client resource
- Add orbit container resource 
- Remove old container resource

## 1.0.6 (March 04, 2019)
- Switch to Go Modules #3
- Add request/response logging in debug mode
- Add import support for loadbalancers
- Add import support for database servers
- Add import support for firewall policy
- Add import support for firewall rule

## 1.0.5 (August 01, 2018)
- Terraform provider official release

## 1.0.4 (July 2, 2018)

- Rewrite structures to standard format
- Refactor Load Balancer
- Refactor Cloud Ip
- Refactor Firewall Policy
- Refactor Firewall Rule
- Refactor Server Group
- Add sensitive marker to container client key
- Remove redundant returns in brightbox server
- Refactor Container code
- Refactor Validate Functions
- Add sensitive marker the provider password entry
- Enforce healthcheck max items in schema
- Remove redundant test in cloudip
- Refactor image test to latest ubuntu image
- Use map_from_string_set function
- Use make rather than append to create string slice
- Improve name of curried function
- Tidy up code
- Allow tests to run with 1.10
- Update vendor

## 1.0.3 (March 13, 2018)

- Make database type check region independent

## 1.0.2 (March 9, 2018)

- On correct UUID
- Access un/pw github credentials properly in Jenkinsfile

## 1.0.1 (March 9, 2018)

- Switch github credentials
- Remove incorrect database fields

## 1.0.0 (March 8, 2018)

- Update README
- Update weblayer example
- Update website docs
- Add Database Type Support
- Add tagrelease shell script
- Add Database Type support

## 0.1.4 (March 8, 2018)

- Remove Jenkinsfile debugging
- Update github-token
- Set debug mode on goreleaser
- Don't do update of get-tools
- Refactor release process to makefile
- Add network capacity to docker container
- Add release process to Jenkinsfile
- Fetch tags
- Add goreleaser snapshot build for non-master branches
- Improve Orbit URL manipulation
- Fix Credentials
- Copy back junit report
- Run test from same directory
- Only create directories to dirname of git remote
- Quote shell script
- Move workspace to Go layout
- Initial Jenkinsfile
- Switch to dep vendoring
- Update index.html.markdown
- Update gitignore
- Bump version
- Ensure static binaries

## 0.1.3 (January 15, 2018)

- Ensure static binaries

## 0.1.2 (September 6, 2017)

- Add plugin installer script
- Allow userdata update
- Add encoded userdata option - to support templates
- Document resources

## 0.1.1 (August 22, 2017)

- Add import and deletion support for more resources

## 0.1.0 (August 21, 2017)

- Add brightbox image data source support
- Update provider to new breakout format
- Make sure terraform refresh removes deleted servers
- Add import support for servers
