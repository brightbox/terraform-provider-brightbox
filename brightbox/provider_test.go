package brightbox

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"brightbox": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
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
				err := p.Configure(context.Background(), terraform.NewResourceConfigRaw(example.raw))
				if !err.HasError() {
					t.Errorf("Expected %q, but no error was returned", example.err)
				} else {
					for i := range err {
						if err[i].Summary != example.err {
							t.Errorf("Got error %q, expected %q", err[i].Summary, example.err)
						}
					}
				}
			},
		)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("BRIGHTBOX_CLIENT"); v == "" {
		t.Fatal("BRIGHTBOX_CLIENT must be set for acceptance tests")
	}
	if v := os.Getenv("BRIGHTBOX_CLIENT_SECRET"); v == "" {
		t.Fatal("BRIGHTBOX_CLIENT_SECRET must be set for acceptance tests")
	}

	err := testAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
	if err != nil {
		t.Fatal(err)
	}
}

// This delegation activates the sweepers
func TestMain(m *testing.M) {
	resource.TestMain(m)
}
