package brightbox

import (
	"log"

	"github.com/brightbox/gobrightbox"
)

type CompositeClient struct {
	ApiClient      *brightbox.Client
	OrbitAuthToken *string
}

func (c *authdetails) Client() (*CompositeClient, error) {
	client, err := c.authenticatedClient()
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] Brightbox Client configured for URL: %s", client.BaseURL.String())
	if client.AccountId == "" {
		log.Printf("[INFO] Provisioning on default account")
	} else {
		log.Printf("[INFO] Provisioning to account %s", client.AccountId)
	}

	composite := &CompositeClient{
		ApiClient:      client,
		OrbitAuthToken: &(c.currentToken.AccessToken),
	}

	log.Printf("[DEBUG] Current Access Token is %s", *composite.OrbitAuthToken)

	return composite, nil

}
