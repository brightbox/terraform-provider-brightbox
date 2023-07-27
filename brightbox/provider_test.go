package brightbox

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccProviders() map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"brightbox": func() (*schema.Provider, error) { return Provider(), nil },
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
			err: "Must specify Account with User Credentials",
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
			err: "User Credentials should be blank with an API Client",
		},
		{
			name: "Apiclient with User",
			raw: map[string]interface{}{
				"apiclient": "cli-12345",
				"apisecret": "mysecret",
				"username":  "fred",
			},
			err: "User Credentials should be blank with an API Client",
		},
		{
			name: "Apiclient with password",
			raw: map[string]interface{}{
				"apiclient": "cli-12345",
				"apisecret": "mysecret",
				"password":  "fred",
			},
			err: "User Credentials should be blank with an API Client",
		},
		{
			name: "Specific app id with missing user",
			raw: map[string]interface{}{
				"apiclient": "app-12345",
				"apisecret": "mysecret",
				"password":  "fred",
				"account":   "acc-12345",
			},
			err: "User Credentials are missing. Please supply a Username and One Time Authentication code",
		},
		{
			name: "Default app id with missing user",
			raw: map[string]interface{}{
				"password": "fred",
				"account":  "acc-12345",
			},
			err: "User Credentials are missing. Please supply a Username and One Time Authentication code",
		},
		{
			name: "Specific app id with missing password",
			raw: map[string]interface{}{
				"apiclient": "app-12345",
				"apisecret": "mysecret",
				"user":      "fred",
				"account":   "acc-12345",
			},
			err: "User Credentials are missing. Please supply a Username and One Time Authentication code",
		},
		{
			name: "Default app id with missing password",
			raw: map[string]interface{}{
				"user":    "fred",
				"account": "acc-12345",
			},
			err: "User Credentials are missing. Please supply a Username and One Time Authentication code",
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
					for _, v := range err {
						if v.Summary != example.err {
							t.Errorf("Got error %q, expected %q", v.Summary, example.err)
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

// testAccProvider is the "main" provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// testAccPreCheck(t) must be called before using this provider instance.
var testAccProvider *schema.Provider

// testAccProviderConfigure ensures testAccProvider is only configured once
//
// The testAccPreCheck(t) function is invoked for every test and this prevents
// extraneous reconfiguration to the same values each time. However, this does
// not prevent reconfiguration that may happen should the address of
// testAccProvider be errantly reused in ProviderFactories.
var testAccProviderConfigure sync.Once

func init() {
	testAccProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	testAccProviderConfigure.Do(
		func() {
			if v := os.Getenv(clientEnvVar); v != "" {
				if v := os.Getenv(clientSecretEnvVar); v == "" {
					t.Fatalf("%s must be set for acceptance tests", clientSecretEnvVar)
				}
			} else if v := os.Getenv(usernameEnvVar); v == "" {
				t.Fatalf("%s or %s must be set for acceptance tests", clientEnvVar, usernameEnvVar)
			}

			diags := testAccProvider.Configure(context.TODO(), terraform.NewResourceConfigRaw(nil))
			if diags.HasError() {
				t.Fatal(diags[0].Summary)
			}
		},
	)
}

// This delegation activates the sweepers
func TestMain(m *testing.M) {
	resource.TestMain(m)
}
