package brightbox

import (
	"log"

	"github.com/brightbox/gobrightbox"
	"github.com/gophercloud/gophercloud"
)

type CompositeClient struct {
	ApiClient   *brightbox.Client
	OrbitClient *gophercloud.ServiceClient
}

func (c *authdetails) Client() (*CompositeClient, error) {
	apiclient, orbitclient, err := c.authenticatedClient()
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] Brightbox Client configured for URL: %s", apiclient.BaseURL.String())
	log.Printf("[INFO] Provisioning to account %s", apiclient.AccountId)
	log.Printf("[INFO] Orbit Client configured for URL: %s", orbitclient.ResourceBaseURL())

	composite := &CompositeClient{
		ApiClient:   apiclient,
		OrbitClient: orbitclient,
	}

	return composite, nil

}
