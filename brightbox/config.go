package brightbox

import (
	"context"
	"log"
	"os"
	"strings"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/gophercloud/gophercloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

const (
	defaultClientID     = "app-dkmch"
	defaultClientSecret = "uogoelzgt0nwawb"
	clientEnvVar        = "BRIGHTBOX_CLIENT"
	clientSecretEnvVar  = "BRIGHTBOX_CLIENT_SECRET"
	usernameEnvVar      = "BRIGHTBOX_USER_NAME"
	passwordEnvVar      = "BRIGHTBOX_PASSWORD"
	accountEnvVar       = "BRIGHTBOX_ACCOUNT"
	apiURLEnvVar        = "BRIGHTBOX_API_URL"
	orbitURLEnvVar      = "BRIGHTBOX_ORBIT_URL"

	defaultTimeoutSeconds = 10
	appPrefix             = "app-"
)

// CompositeClient allows access to Honcho and Orbit
type CompositeClient struct {
	APIClient   *brightbox.Client
	OrbitClient *gophercloud.ServiceClient
}

type authdetails struct {
	APIClient string
	APISecret string
	UserName  string
	password  string
	Account   string
	APIURL    string
	OrbitURL  string
}

// obtainCloudClient creates a new Composite client using details from
// the environment
func obtainCloudClient() (*CompositeClient, diag.Diagnostics) {
	log.Printf("[DEBUG] obtainCloudClient")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return configureClient(
		ctx,
		authdetails{
			APIClient: getenvWithDefault(clientEnvVar,
				defaultClientID),
			APISecret: getenvWithDefault(clientSecretEnvVar,
				defaultClientSecret),
			UserName: os.Getenv(usernameEnvVar),
			password: os.Getenv(passwordEnvVar),
			Account:  os.Getenv(accountEnvVar),
			APIURL:   getenvWithDefault(apiURLEnvVar, ""),
			OrbitURL: getenvWithDefault(orbitURLEnvVar, ""),
		},
	)
}

// Validate account config entries
func validateConfig(authd authdetails) diag.Diagnostics {
	var result []diag.Diagnostic
	log.Printf("[DEBUG] Validating Config")
	if strings.HasPrefix(authd.APIClient, appPrefix) {
		log.Printf("[DEBUG] Detected OAuth Application. Validating User details.")
		if authd.UserName == "" || authd.password == "" {
			result = append(result, diag.Errorf("User Credentials are missing. Please supply a Username and One Time Authentication code")...)
		}
		if authd.Account == "" {
			result = append(result, diag.Errorf("Must specify Account with User Credentials")...)
		}
	} else {
		log.Printf("[DEBUG] Detected API Client.")
		if authd.UserName != "" || authd.password != "" {
			result = append(result, diag.Errorf("User Credentials should be blank with an API Client")...)
		}
	}
	return result
}

func configureClient(ctx context.Context, authd authdetails) (*CompositeClient, diag.Diagnostics) {
	log.Printf("[DEBUG] Configuring Brightbox Clients")
	if err := validateConfig(authd); err.HasError() {
		return nil, err
	}

	apiclient, orbitclient, err := authenticatedClients(ctx, authd)
	if err.HasError() {
		return nil, err
	}

	if apiclient != nil {
		log.Printf("[INFO] Brightbox Client configured for URL: %s", apiclient.ResourceBaseURL())
	}
	if orbitclient != nil {
		log.Printf("[INFO] Orbit Client configured for URL: %s", orbitclient.ResourceBaseURL())
	}

	composite := &CompositeClient{
		APIClient:   apiclient,
		OrbitClient: orbitclient,
	}

	return composite, nil
}
