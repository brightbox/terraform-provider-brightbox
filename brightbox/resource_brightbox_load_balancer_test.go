package brightbox

import (
	"fmt"
	"log"
	"testing"

	brightbox "github.com/brightbox/gobrightbox"
	"github.com/brightbox/gobrightbox/status"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccBrightboxLoadBalancer_BasicUpdates(t *testing.T) {
	resourceName := "brightbox_load_balancer.default"
	var loadBalancer brightbox.LoadBalancer

	resource.Test(t, resource.TestCase{
		DisableBinaryDriver: true,
		PreCheck:            func() { testAccPreCheck(t) },
		Providers:           testAccProviders,
		CheckDestroy:        testAccCheckBrightboxLoadBalancerAndServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxLoadBalancerConfig_locked,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxLoadBalancerExists(resourceName, &loadBalancer),
					resource.TestCheckResourceAttr(
						resourceName, "locked", "true"),
					resource.TestCheckResourceAttr(
						resourceName, "healthcheck.0.request", "/"),
					resource.TestCheckResourceAttr(
						resourceName, "healthcheck.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "listener.#", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "listener.1595392858.timeout", "50000"),
					resource.TestCheckResourceAttr(
						resourceName, "listener.1462547963.timeout", "10000"),
					resource.TestCheckResourceAttr(
						resourceName, "nodes.#", "1"),
				),
			},
			{
				Config: testAccCheckBrightboxLoadBalancerConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxLoadBalancerExists(resourceName, &loadBalancer),
					testAccCheckBrightboxEmptyLoadBalancerAttributes(&loadBalancer),
					resource.TestCheckResourceAttr(
						resourceName, "locked", "false"),
				),
			},
			{
				Config: testAccCheckBrightboxLoadBalancerConfig_locked,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "locked", "true"),
				),
			},
			{
				Config: testAccCheckBrightboxLoadBalancerConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxLoadBalancerExists(resourceName, &loadBalancer),
					testAccCheckBrightboxEmptyLoadBalancerAttributes(&loadBalancer),
					resource.TestCheckResourceAttr(
						resourceName, "locked", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCheckBrightboxLoadBalancerConfig_new_timeout,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxLoadBalancerExists(resourceName, &loadBalancer),
					resource.TestCheckResourceAttr(
						resourceName, "locked", "false"),
					resource.TestCheckResourceAttr(
						resourceName, "listener.3297149260.timeout", "10000"),
					resource.TestCheckResourceAttr(
						resourceName, "listener.3297149260.out", "8080"),
					resource.TestCheckResourceAttr(
						resourceName, "listener.1462547963.timeout", "10000"),
				),
			},
			{
				Config: testAccCheckBrightboxLoadBalancerConfig_new_healthcheck,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxLoadBalancerExists(resourceName, &loadBalancer),
					resource.TestCheckResourceAttr(
						resourceName, "healthcheck.0.type", "tcp"),
					resource.TestCheckResourceAttr(
						resourceName, "healthcheck.0.port", "23"),
					resource.TestCheckResourceAttr(
						resourceName, "healthcheck.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "listener.#", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "listener.1595392858.timeout", "50000"),
					resource.TestCheckResourceAttr(
						resourceName, "listener.1462547963.timeout", "10000"),
				),
			},
			{
				Config: testAccCheckBrightboxLoadBalancerConfig_addListener,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxLoadBalancerExists(resourceName, &loadBalancer),
					resource.TestCheckResourceAttr(
						resourceName, "listener.#", "3"),
					resource.TestCheckResourceAttr(
						resourceName, "listener.2238538594.protocol", "https"),
					resource.TestCheckResourceAttr(
						resourceName, "listener.2238538594.in", "443"),
					resource.TestCheckResourceAttr(
						resourceName, "listener.2238538594.timeout", "50000"),
					resource.TestCheckResourceAttr(
						resourceName, "certificate_private_key", "63158de92c07f5a53ee8bd56c5750deaa654aabf"),
					resource.TestCheckResourceAttr(
						resourceName, "certificate_pem", "a5f8997fb16293ae7827f974b9cc120c8c776d02"),
					resource.TestCheckResourceAttr(
						resourceName, "sslv3", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"certificate_pem", "certificate_private_key"},
			},
			{
				Config: testAccCheckBrightboxLoadBalancerConfig_remove_listener,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxLoadBalancerExists(resourceName, &loadBalancer),
					resource.TestCheckResourceAttr(
						resourceName, "listener.#", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "certificate_private_key", "da39a3ee5e6b4b0d3255bfef95601890afd80709"),
					resource.TestCheckResourceAttr(
						resourceName, "certificate_pem", "da39a3ee5e6b4b0d3255bfef95601890afd80709"),
				),
			},
		},
	})
}

func testAccCheckBrightboxLoadBalancerAndServerDestroy(s *terraform.State) error {
	err := testAccCheckBrightboxLoadBalancerDestroy(s)
	if err != nil {
		return err
	}
	return testAccCheckBrightboxServerDestroy(s)
}

func testAccCheckBrightboxLoadBalancerDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*CompositeClient).APIClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "brightbox_load_balancer" {
			continue
		}

		// Try to find the LoadBalancer
		_, err := client.LoadBalancer(rs.Primary.ID)

		// Wait

		if err != nil {
			apierror := err.(brightbox.ApiError)
			if apierror.StatusCode != 404 {
				return fmt.Errorf(
					"Error waiting for load_balancer %s to be destroyed: %s",
					rs.Primary.ID, err)
			}
		}
	}

	return nil
}

func testAccCheckBrightboxLoadBalancerExists(n string, loadBalancer *brightbox.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No LoadBalancer ID is set")
		}

		client := testAccProvider.Meta().(*CompositeClient).APIClient

		// Try to find the LoadBalancer
		retrieveLoadBalancer, err := client.LoadBalancer(rs.Primary.ID)

		if err != nil {
			return err
		}

		if retrieveLoadBalancer.Id != rs.Primary.ID {
			return fmt.Errorf("LoadBalancer not found")
		}

		*loadBalancer = *retrieveLoadBalancer

		return nil
	}
}

func testAccCheckBrightboxEmptyLoadBalancerAttributes(loadBalancer *brightbox.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if loadBalancer.Name != "foo-20200513" {
			return fmt.Errorf("Bad name: %s", loadBalancer.Name)
		}
		if loadBalancer.Locked != false {
			return fmt.Errorf("Bad locked: %v", loadBalancer.Locked)
		}
		if loadBalancer.Status != "active" {
			return fmt.Errorf("Bad status: %s", loadBalancer.Status)
		}
		if loadBalancer.Policy != "least-connections" {
			return fmt.Errorf("Bad policy: %s", loadBalancer.Policy)
		}
		if loadBalancer.BufferSize != 4096 {
			return fmt.Errorf("Bad buffer size: %d", loadBalancer.BufferSize)
		}
		return nil
	}
}

var testAccCheckBrightboxLoadBalancerConfig_basic = fmt.Sprintf(`

resource "brightbox_load_balancer" "default" {
	name = "foo-20200513"
	listener {
		protocol = "http"
		in = 80
		out = 8080
	}
	listener {
		protocol = "http+ws"
		in = 81
		out = 81
		timeout = 10000
	}
	
	healthcheck {
		type = "http"
		port = 8080
	}
	nodes = ["${brightbox_server.foobar.id}"]
}

resource "brightbox_server" "foobar" {
	image = "${data.brightbox_image.foobar.id}"
	name = "foo-20200426"
	type = "1gb.ssd"
	server_groups = ["${data.brightbox_server_group.default.id}"]

}

%s%s`, TestAccBrightboxImageDataSourceConfig_blank_disk,
	TestAccBrightboxDataServerGroupConfig_default)

var testAccCheckBrightboxLoadBalancerConfig_locked = fmt.Sprintf(`

resource "brightbox_load_balancer" "default" {
	name = "foo-20200513"
	locked = true
	listener {
		protocol = "http"
		in = 80
		out = 8080
	}
	listener {
		protocol = "http+ws"
		in = 81
		out = 81
		timeout = 10000
	}
	
	healthcheck {
		type = "http"
		port = 8080
	}
	nodes = ["${brightbox_server.foobar.id}"]
}

resource "brightbox_server" "foobar" {
	image = "${data.brightbox_image.foobar.id}"
	name = "foo-20200426"
	type = "1gb.ssd"
	server_groups = ["${data.brightbox_server_group.default.id}"]

}

%s%s`, TestAccBrightboxImageDataSourceConfig_blank_disk,
	TestAccBrightboxDataServerGroupConfig_default)

var testAccCheckBrightboxLoadBalancerConfig_new_timeout = fmt.Sprintf(`

resource "brightbox_load_balancer" "default" {
	name = "foo-20200513"
	listener {
		protocol = "http"
		in = 80
		out = 8080
		timeout = 10000
	}
	listener {
		protocol = "http+ws"
		in = 81
		out = 81
		timeout = 10000
	}
	
	healthcheck {
		type = "http"
		port = 8080
	}
	nodes = ["${brightbox_server.foobar.id}"]
}

resource "brightbox_server" "foobar" {
	image = "${data.brightbox_image.foobar.id}"
	name = "foo-20200426"
	type = "1gb.ssd"
	server_groups = ["${data.brightbox_server_group.default.id}"]
}

%s%s`, TestAccBrightboxImageDataSourceConfig_blank_disk,
	TestAccBrightboxDataServerGroupConfig_default)

var testAccCheckBrightboxLoadBalancerConfig_new_healthcheck = fmt.Sprintf(`

resource "brightbox_load_balancer" "default" {
	name = "foo-20200513"
	listener {
		protocol = "http"
		in = 80
		out = 8080
	}
	listener {
		protocol = "http+ws"
		in = 81
		out = 81
		timeout = 10000
	}
	
	healthcheck {
		type = "tcp"
		port = 23
	}
	nodes = ["${brightbox_server.foobar.id}"]
}

resource "brightbox_server" "foobar" {
	image = "${data.brightbox_image.foobar.id}"
	name = "foo-20200426"
	type = "1gb.ssd"
	server_groups = ["${data.brightbox_server_group.default.id}"]
}

%s%s`, TestAccBrightboxImageDataSourceConfig_blank_disk,
	TestAccBrightboxDataServerGroupConfig_default)

var testAccCheckBrightboxLoadBalancerConfig_addListener = fmt.Sprintf(`

resource "brightbox_load_balancer" "default" {
	name = "foo-20200513"
	listener {
		protocol = "https"
		in = 443
		out = 8080
	}
	listener {
		protocol = "http"
		in = 80
		out = 8080
	}
	listener {
		protocol = "http+ws"
		in = 81
		out = 81
		timeout = 10000
	}
	
	healthcheck {
		type = "tcp"
		port = 23
	}
	nodes = ["${brightbox_server.foobar.id}"]
	sslv3 = true

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

	certificate_private_key= <<EOF
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

resource "brightbox_server" "foobar" {
	image = "${data.brightbox_image.foobar.id}"
	name = "foo-20200513"
	type = "1gb.ssd"
	server_groups = ["${data.brightbox_server_group.default.id}"]
}

%s%s`, TestAccBrightboxImageDataSourceConfig_blank_disk,
	TestAccBrightboxDataServerGroupConfig_default)

var testAccCheckBrightboxLoadBalancerConfig_remove_listener = fmt.Sprintf(`

resource "brightbox_load_balancer" "default" {
	name = "foo-20201513"
	listener {
		protocol = "http"
		in = 80
		out = 8080
	}
	listener {
		protocol = "http+ws"
		in = 81
		out = 81
		timeout = 10000
	}
	
	healthcheck {
		type = "tcp"
		port = 23
	}
	nodes = ["${brightbox_server.foobar.id}"]

	certificate_pem = ""
	certificate_private_key= ""

}

resource "brightbox_server" "foobar" {
	image = "${data.brightbox_image.foobar.id}"
	name = "foo-20200513"
	type = "1gb.ssd"
	server_groups = ["${data.brightbox_server_group.default.id}"]
}

%s%s`, TestAccBrightboxImageDataSourceConfig_blank_disk,
	TestAccBrightboxDataServerGroupConfig_default)

// Sweeper

func init() {
	resource.AddTestSweepers("load_balancer", &resource.Sweeper{
		Name: "load_balancer",
		F: func(_ string) error {
			client, err := obtainCloudClient()
			if err != nil {
				return err
			}
			objects, err := client.APIClient.LoadBalancers()
			if err != nil {
				return err
			}
			for _, object := range objects {
				if object.Status != status.Active {
					continue
				}
				if isTestName(object.Name) {
					log.Printf("[INFO] removing %s named %s", object.Id, object.Name)
					if err := setLockState(client.APIClient, false, brightbox.LoadBalancer{Id: object.Id}); err != nil {
						log.Printf("error unlocking %s during sweep: %s", object.Id, err)
					}
					if err := client.APIClient.DestroyLoadBalancer(object.Id); err != nil {
						log.Printf("error destroying %s during sweep: %s", object.Id, err)
					}
				}
			}
			return nil
		},
	})
}
