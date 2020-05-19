package brightbox

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"brightbox": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	p := Provider()
	if err := p.(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_badConfigs(t *testing.T) {
	if os.Getenv("TF_ACC") != "" {
		t.Skip("Skipping test that clears ENV as TF_ACC is set")
	}
	p := Provider()
	var configTests = []struct {
		name string
		raw  map[string]interface{}
		err  string
	}{
		{
			name: "Username without account",
			raw: map[string]interface{}{
				"username": "fred",
				"password": "fred",
			},
			err: "must specify Account with User Credentials",
		},
		{
			name: "Apiclient with User Credentials",
			raw: map[string]interface{}{
				"apiclient": "cli-12345",
				"apisecret": "mysecret",
				"username":  "fred",
				"password":  "fred",
				"account":   "acc-12345",
			},
			err: "user Credentials should be blank with an API Client",
		},
		{
			name: "Apiclient with User",
			raw: map[string]interface{}{
				"apiclient": "cli-12345",
				"apisecret": "mysecret",
				"username":  "fred",
			},
			err: "user Credentials should be blank with an API Client",
		},
		{
			name: "Apiclient with password",
			raw: map[string]interface{}{
				"apiclient": "cli-12345",
				"apisecret": "mysecret",
				"password":  "fred",
			},
			err: "user Credentials should be blank with an API Client",
		},
		{
			name: "Specific app id with missing user",
			raw: map[string]interface{}{
				"apiclient": "app-12345",
				"apisecret": "mysecret",
				"password":  "fred",
			},
			err: "user Credentials are missing. Please supply a Username and One Time Authentication code",
		},
		{
			name: "Default app id with missing user",
			raw: map[string]interface{}{
				"apiclient": "app-12345",
				"apisecret": "mysecret",
				"password":  "fred",
			},
			err: "user Credentials are missing. Please supply a Username and One Time Authentication code",
		},
		{
			name: "Specific app id with missing password",
			raw: map[string]interface{}{
				"apiclient": "app-12345",
				"apisecret": "mysecret",
				"user":      "fred",
			},
			err: "user Credentials are missing. Please supply a Username and One Time Authentication code",
		},
		{
			name: "Default app id with missing password",
			raw: map[string]interface{}{
				"user": "fred",
			},
			err: "user Credentials are missing. Please supply a Username and One Time Authentication code",
		},
	}

	os.Clearenv()
	for _, example := range configTests {
		t.Run(
			example.name,
			func(t *testing.T) {
				err := p.Configure(terraform.NewResourceConfigRaw(example.raw))
				if err == nil {
					t.Errorf("Expected %q, but no error was returned", example.err)
				} else {
					if err.Error() != example.err {
						t.Errorf("Got error %q, expected %q", err.Error(), example.err)
					}
				}
			},
		)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("BRIGHTBOX_CLIENT"); v == "" {
		t.Fatal("BRIGHTBOX_CLIENT must be set for acceptance tests")
	}
	if v := os.Getenv("BRIGHTBOX_CLIENT_SECRET"); v == "" {
		t.Fatal("BRIGHTBOX_CLIENT_SECRET must be set for acceptance tests")
	}

	err := testAccProvider.Configure(terraform.NewResourceConfigRaw(nil))
	if err != nil {
		t.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	acctest.UseBinaryDriver("brightbox", Provider)
	resource.TestMain(m)
}
