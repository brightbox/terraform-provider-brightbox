# Specify the provider and access details
provider "brightbox" {
}

resource "brightbox_cloudip" "default" {
  target = "${brightbox_server.web.id}"
}

# Our default server group to access
# the instances over SSH and HTTP
resource "brightbox_server_group" "default" {
  description = "Used by the terraform"

  # SSH access from anywhere
  ingress {
    from_port = 22
    to_port = 22
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # HTTP access from anywhere
  ingress {
    from_port = 80
    to_port = 80
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # outbound internet access
  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "brightbox_server" "web" {

  name = "Terraform web server example"
  image = "${var.web_image}"
  type = "${var.web_type}"

  # Our Security group to allow HTTP and SSH access
  server_groups = ["${brightbox_server_group.default.id}"]

  # We run a remote provisioner on the instance after creating it.
  # In this case, we just install nginx and start it. By default,
  # this should be on port 80
  user_data = "${file("userdata.sh")}"
}
