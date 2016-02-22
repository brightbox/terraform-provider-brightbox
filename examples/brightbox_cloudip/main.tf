# Specify the provider and access details
provider "brightbox" {
}

resource "brightbox_cloudip" "default" {
  target = "${brightbox_server.web.interface}"
}

# Our default server group to access
# the instances over SSH and HTTP
resource "brightbox_server_group" "default" {
  name = "Used by the terraform"
}

resource "brightbox_firewall_policy" "default" {
  name = "Used by terraform"
  server_group = "${brightbox_server_group.default.id}"
}

resource "brightbox_firewall_rule" "default_ssh" {
    destination_port = 22
    protocol = "tcp"
    source = "any"
    description = "SSH access from anywhere"
    firewall_policy = "${brightbox_firewall_policy.default.id}"
}

resource "brightbox_firewall_rule" "default_http" {
    destination_port = 80
    protocol = "tcp"
    source = "any"
    description = "HTTP access from anywhere"
    firewall_policy = "${brightbox_firewall_policy.default.id}"
}

resource "brightbox_firewall_rule" "default_outbound" {
    destination = "any"
    description = "Outbound internet access"
    firewall_policy = "${brightbox_firewall_policy.default.id}"
}

resource "brightbox_server" "web" {

  depends_on = ["brightbox_firewall_policy.default"]

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
