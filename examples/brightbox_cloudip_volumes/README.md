# Cloud IP Example

This Cloud IP example launches a web server, installs nginx. It also creates security group.

This version uses a boot volume from an image and creates an xfs formatted volume which is attached to the server.  

## Running the example

run `terraform apply` 

Give couple of mins for userdata to install nginx, and then type the DNS name from outputs in your browser and see the nginx welcome page
