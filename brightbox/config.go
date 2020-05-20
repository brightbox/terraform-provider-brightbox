package brightbox

import (
	"fmt"
	"log"
	"os"
	"strings"

	brightbox "github.com/brightbox/gobrightbox"
	"github.com/gophercloud/gophercloud"
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

// obtainCloudClient creates a new Composite client using details from
// the environment
func obtainCloudClient() (*CompositeClient, error) {
	log.Printf("[DEBUG] obtainCloudClient")
	return (&authdetails{
		APIClient: getenvWithDefault(clientEnvVar,
			defaultClientID),
		APISecret: getenvWithDefault(clientSecretEnvVar,
			defaultClientSecret),
		UserName: os.Getenv(usernameEnvVar),
		password: os.Getenv(passwordEnvVar),
		Account:  os.Getenv(accountEnvVar),
		APIURL: getenvWithDefault(apiURLEnvVar,
			brightbox.DefaultRegionApiURL),
		OrbitURL: getenvWithDefault(orbitURLEnvVar,
			brightbox.DefaultOrbitAuthURL),
	}).Client()
}

// Validate account config entries
func (authd *authdetails) validateConfig() error {
	log.Printf("[DEBUG] validateConfig")
	if strings.HasPrefix(authd.APIClient, appPrefix) {
		log.Printf("[DEBUG] Detected OAuth Application. Validating User details.")
		if authd.UserName == "" || authd.password == "" {
			return fmt.Errorf("user Credentials are missing. Please supply a Username and One Time Authentication code")
		}
		if authd.Account == "" {
			return fmt.Errorf("must specify Account with User Credentials")
		}
	} else {
		log.Printf("[DEBUG] Detected API Client.")
		if authd.UserName != "" || authd.password != "" {
			return fmt.Errorf("user Credentials should be blank with an API Client")
		}
	}
	return nil
}

func (authd *authdetails) Client() (*CompositeClient, error) {
	if err := authd.validateConfig(); err != nil {
		return nil, err
	}

	apiclient, orbitclient, err := authd.authenticatedClient()
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] Brightbox Client configured for URL: %s", apiclient.BaseURL.String())
	log.Printf("[INFO] Provisioning to account %s", apiclient.AccountId)
	log.Printf("[INFO] Orbit Client configured for URL: %s", orbitclient.ResourceBaseURL())

	composite := &CompositeClient{
		APIClient:   apiclient,
		OrbitClient: orbitclient,
	}

	return composite, nil

}
