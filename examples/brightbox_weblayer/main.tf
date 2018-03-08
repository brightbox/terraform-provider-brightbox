# Specify the provider and access details
provider "brightbox" {}

resource "brightbox_cloudip" "lb" {
  target = "${brightbox_load_balancer.lb.id}"
}

resource "brightbox_cloudip" "server1" {
  target = "${brightbox_server.server1.interface}"
}

resource "brightbox_cloudip" "server2" {
  target = "${brightbox_server.server2.interface}"
}

resource "brightbox_cloudip" "database" {
  target = "${brightbox_database_server.database.id}"
}

resource "brightbox_load_balancer" "lb" {
  name = "Terraform weblayer example"

  listener {
    protocol = "https"
    in       = 443
    out      = 8080
  }

  listener {
    protocol = "http"
    in       = 80
    out      = 8080
    timeout  = 10000
  }

  listener {
    protocol = "http+ws"
    in       = 81
    out      = 81
    timeout  = 10000
  }

  healthcheck {
    type = "http"
    port = 8080
  }

  nodes = [
    "${brightbox_server.server2.id}",
    "${brightbox_server.server1.id}",
  ]

  certificate_pem = <<EOF
-----BEGIN CERTIFICATE-----
MIIDBzCCAe+gAwIBAgIJAPD+BTBqIVp6MA0GCSqGSIb3DQEBBQUAMBoxGDAWBgNV
BAMMD3d3dy5leGFtcGxlLmNvbTAeFw0xNjAzMDIxMTU0MDFaFw0yNjAyMjgxMTU0
MDFaMBoxGDAWBgNVBAMMD3d3dy5leGFtcGxlLmNvbTCCASIwDQYJKoZIhvcNAQEB
BQADggEPADCCAQoCggEBANuA/TLmuCbZdHcMKUwFadRpNnjg3S3PuP9AECDu+mIC
rOBmNqeZ66dEkzJqNMq4pEo30L9ZlZXl7fAvsIZTPYLEb0ieYGyTTdqAKrHi8GPP
ZeC+iAySKXnTKjpnciTWFv2T8R9tLsgPrsv54okM59bYC5mSnD7pL6RR/aQ0oi4f
X2eJex5fpfFlcxm9HvvVEdWq9/CQNoCOpGhLT911MRVMUl3S10BmzTG8Q87P76ji
Axt3t5piPg8JGiSBHTUJmKw/jxcwhybWHaf/217RmSmeoTo40wMCB2b05RqdSOm5
39qLotrjt2w3nFKzm423cVok3y2w55hLkDCbDlxUK1kCAwEAAaNQME4wHQYDVR0O
BBYEFCX20aoQddqjbga66nppwRlJdvB8MB8GA1UdIwQYMBaAFCX20aoQddqjbga6
6nppwRlJdvB8MAwGA1UdEwQFMAMBAf8wDQYJKoZIhvcNAQEFBQADggEBAJkFZvAL
joeAiWaEItIPr8+98OJam7Pnta29HoKu4jAHkiunzXxNTQutUMMx1WhBF8OJJX1P
pHhKEfK47W8z4PbsM/hudZfm2xXlFMfvYNAusptJxOhMKNJgJz+gjY5FaTCGD9Ao
JkcshhUgXQ9zvu2Ol390qo0zlxMvnlVacRgKGY/I6hJaktrbdXm7qcReZp06Pw3a
adoKmzXeUlPvlbb+8KLXSD7hgUaojLDEgOLpAE++muiAAuwOP2UX3XJOPUQZdicB
sbrBMXO6F253YTqZiwAg9hgEHTHdXgqrd3TQT9P9mazrHxskqk9uWmIgN8oolHjp
OsWSdvMP2tRS8Oo=
-----END CERTIFICATE-----
EOF

  certificate_private_key = <<EOF
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA24D9Mua4Jtl0dwwpTAVp1Gk2eODdLc+4/0AQIO76YgKs4GY2
p5nrp0STMmo0yrikSjfQv1mVleXt8C+whlM9gsRvSJ5gbJNN2oAqseLwY89l4L6I
DJIpedMqOmdyJNYW/ZPxH20uyA+uy/niiQzn1tgLmZKcPukvpFH9pDSiLh9fZ4l7
Hl+l8WVzGb0e+9UR1ar38JA2gI6kaEtP3XUxFUxSXdLXQGbNMbxDzs/vqOIDG3e3
mmI+DwkaJIEdNQmYrD+PFzCHJtYdp//bXtGZKZ6hOjjTAwIHZvTlGp1I6bnf2oui
2uO3bDecUrObjbdxWiTfLbDnmEuQMJsOXFQrWQIDAQABAoIBAHzvoC42sB48q1OP
Mno4opHqCL0oj/uhPdTa69My8oSSrT9ULkubCkw8deO+G6o/ChPMTR58qO2W36VU
H491FY+2qviUXKGv/iIdzS9O0jCdPYl8KQeusbjLfj+b3ZYl3RQb/qQ6iuQIOR+U
bWJAXD0m3wNcNV6Bb0KCAHJUGvNQjiueMMVEND1Pvb9WogFWY7yvteoxv9ASFiRv
1N2LDlm/199/Tpmb9a9vVrIuT8pZfAtmVfZ5HhwV8xU1q2qbys1j9DpZPggHnT4l
CzIw7pALbaE8/sG17h6+icl13cKLpgp63HyJFgik1v1NDnCmzckrNAiSW4lZsgzM
BV3m9hkCgYEA7qboVDv6FvwwwyILbd3aYjLjCqNjDzpvngJrOl6/cDDQR34NQPzI
3ePYO1p99xRYmQe0FJ7ZuJtOQHJOdeLEJqeo6lNMI9T+FhKnqk7Gy7ZQI0PNP2x6
tpfoa27emeDblu+AVSBIZjByS+Cpf/Mnf4/DhhofAMdT4TFyng/JbbMCgYEA63XA
tHE8BwxY/6NxR/pGlRi0AbZfjfU40/q+309NNGrGyDZfoYpbG9I6Wo09Rc+QDhEq
2+zk59ubO1jkgh9eIOBm6+5yHjcbwftBsxesQQdabAg34ppFYMfvKsLgDnejvYEW
pfLmMAvcmIFGWid9hX5/ShbjjkJnIKSbu/vN9MMCgYEAoFhyZw45NTJSjPkV1sal
0S7Bj0dB6lxn3DFh3EEGvRl/B1nxC8YMK9HHWfGuCtGXyZH8c5JbVIa8p95lSx2G
jzY87tvyn2yfHzN/hZUSSpL++wK2J7P6Ky6bkXtXguoqgBoBDrD3E/nfAY48NGSq
GDH+u95XEE3c1MRFb1/KBbMCgYEAo2VgqBdYR6/a5vPd/cwBRSASconDf7inifsc
j8zxT6m1bmTFMk3X8dOOqR4QYiyq1Ag3zMx1AS0VaTbDxETORlRTN/CNgshNW+zn
Z8fKwom+xu9hEMBr2sCECRGY+JEvsKcvN1P7R2ZD3BUB5Dg5U/U3kguWODd+Z1mz
tN0FzI8CgYBx9giIe7aAItxl43p6tPsMW6ROlXEjWit2XBlaDdY5t48k8KJ2clk/
IHu8B12R2mN+lMn9mkOa4mSb9MrVQZ2FGg4lUAQro519NVBcVqoRsEDn1kHd+hhl
L6c41r4AZ3Iyvr3MYoSohogBbAnd6TW14NjvBHceREhAqvmIWlWmAQ==
-----END RSA PRIVATE KEY-----
EOF
}

resource "brightbox_container" "backups" {
  name        = "terraform_backups"
  description = "Terraform example backup area"
}

resource "brightbox_database_server" "database" {
  name                = "Terraform weblayer example database"
  database_engine     = "mysql"
  database_version    = "5.6"
  database_type       = "${data.brightbox_database_type.4gb.id}"
  maintenance_weekday = 6
  maintenance_hour    = 6
  allow_access        = ["${brightbox_server_group.weblayer.id}"]
}

resource "brightbox_server_group" "weblayer" {
  name = "Used by the terraform"
}

resource "brightbox_firewall_policy" "weblayer" {
  name         = "Used by terraform"
  server_group = "${brightbox_server_group.weblayer.id}"
}

resource "brightbox_firewall_rule" "weblayer_ssh" {
  destination_port = 22
  protocol         = "tcp"
  source           = "any"
  description      = "SSH access from anywhere"
  firewall_policy  = "${brightbox_firewall_policy.weblayer.id}"
}

resource "brightbox_firewall_rule" "weblayer_http" {
  destination_port = 80
  protocol         = "tcp"
  source           = "any"
  description      = "HTTP access from anywhere"
  firewall_policy  = "${brightbox_firewall_policy.weblayer.id}"
}

resource "brightbox_firewall_rule" "weblayer_https" {
  destination_port = 443
  protocol         = "tcp"
  source           = "any"
  description      = "HTTPs access from anywhere"
  firewall_policy  = "${brightbox_firewall_policy.weblayer.id}"
}

resource "brightbox_firewall_rule" "weblayer_outbound" {
  destination     = "any"
  description     = "Outbound internet access"
  firewall_policy = "${brightbox_firewall_policy.weblayer.id}"
}

resource "brightbox_firewall_rule" "weblayer_icmp" {
  protocol        = "icmp"
  source          = "any"
  icmp_type_name  = "any"
  firewall_policy = "${brightbox_firewall_policy.weblayer.id}"
}

resource "brightbox_server" "server1" {
  depends_on = ["brightbox_firewall_policy.weblayer"]

  name  = "Terraform web server example 1"
  image = "${data.brightbox_image.ubuntu_lts.id}"
  type  = "${var.web_type}"

  # Our Security group to allow HTTP and SSH access
  server_groups = ["${brightbox_server_group.weblayer.id}"]
}

resource "brightbox_server" "server2" {
  depends_on = ["brightbox_firewall_policy.weblayer"]

  name  = "Terraform web server example 2"
  image = "${data.brightbox_image.ubuntu_lts.id}"
  type  = "${var.web_type}"

  # Our Security group to allow HTTP and SSH access
  server_groups = ["${brightbox_server_group.weblayer.id}"]
}

data "brightbox_image" "ubuntu_lts" {
  name        = "${var.web_image}"
  arch        = "x86_64"
  official    = true
  most_recent = true
}

data "brightbox_database_type" "4gb" {
  name = "${var.database_type}"
}
