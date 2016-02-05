package brightbox

import (
	"log"

	"github.com/brightbox/gobrightbox"
)

type Config struct {
	authdetails
}

func (c *Config) Client() (*brightbox.Client, error) {
	client, err := c.authenticatedClient()
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] Brightbox Client configured for URL: %s", client.BaseURL.String())

	return client, nil

}
