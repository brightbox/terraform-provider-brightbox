resource "brightbox_cloudip" "default" {
  target = brightbox_server.web.interface
}

# Our default server group to access
# the instances over SSH and HTTP
resource "brightbox_server_group" "default" {
  name = "Used by the terraform"
}

resource "brightbox_firewall_policy" "default" {
  name         = "Used by terraform"
  server_group = brightbox_server_group.default.id
}

resource "brightbox_firewall_rule" "default_ssh" {
  destination_port = 22
  protocol         = "tcp"
  source           = "any"
  description      = "SSH access from anywhere"
  firewall_policy  = brightbox_firewall_policy.default.id
}

resource "brightbox_firewall_rule" "default_http" {
  destination_port = 80
  protocol         = "tcp"
  source           = "any"
  description      = "HTTP access from anywhere"
  firewall_policy  = brightbox_firewall_policy.default.id
}

resource "brightbox_firewall_rule" "default_https" {
  destination_port = 443
  protocol         = "tcp"
  source           = "any"
  description      = "HTTPs access from anywhere"
  firewall_policy  = brightbox_firewall_policy.default.id
}

resource "brightbox_firewall_rule" "default_outbound" {
  destination     = "any"
  description     = "Outbound internet access"
  firewall_policy = brightbox_firewall_policy.default.id
}

resource "brightbox_firewall_rule" "default_icmp" {
  protocol        = "icmp"
  source          = "any"
  icmp_type_name  = "any"
  firewall_policy = brightbox_firewall_policy.default.id
}

resource "brightbox_volume" "boot_disk" {
  name = "Terraform web server example boot disk"
  image = data.brightbox_image.ubuntu_lts.id
  size = 61440
}

resource "brightbox_volume" "data_disk" {
  name = "Terraform web server example data disk"
  size = 40960
  serial = "TWS_DATA_DISK"
  filesystem_type = "xfs"
  filesystem_label = "data_area"
}

resource "brightbox_server" "web" {

  name  = "Terraform web server example"
  volume = brightbox_volume.boot_disk.id
  data_volumes = [brightbox_volume.data_disk.id]
  type  = data.brightbox_server_type.nbs_type.id

  # Our Security group to allow HTTP and SSH access
  server_groups = [brightbox_server_group.default.id]

  # Need this so that destroy orders objects properly
  depends_on = [brightbox_firewall_policy.default]

  # We run a remote provisioner on the instance after creating it.
  # In this case, we just install nginx and start it. By default,
  # this should be on port 80
  user_data = file("userdata.sh")
}

data "brightbox_image" "ubuntu_lts" {
  name        = "^ubuntu-xenial.*server$"
  arch        = "x86_64"
  official    = true
  most_recent = true
}

data "brightbox_server_type" "nbs_type" {
  handle = "^8gb.nbs$"
}
