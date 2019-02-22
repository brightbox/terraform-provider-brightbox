package brightbox

import (
	"log"
	"net/http"

	"github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform/helper/logging"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

var infrastructureScope = []string{"infrastructure, orbit"}

type authdetails struct {
	APIClient    string
	APISecret    string
	UserName     string
	password     string
	Account      string
	APIURL       string
	currentToken *oauth2.Token
}

// Authenticate the details and return a client
func (authd *authdetails) authenticatedClient() (*brightbox.Client, error) {
	switch {
	case authd.currentToken != nil:
		return authd.tokenisedAuth()
	case authd.UserName != "" || authd.password != "":
		return authd.tokenisedAuth()
	default:
		return authd.apiClientAuth()
	}
}

func (authd *authdetails) tokenURL() string {
	return authd.APIURL + "/token"
}

func (authd *authdetails) tokenisedAuth() (*brightbox.Client, error) {
	conf := oauth2.Config{
		ClientID:     authd.APIClient,
		ClientSecret: authd.APISecret,
		Scopes:       infrastructureScope,
		Endpoint: oauth2.Endpoint{
			TokenURL: authd.tokenURL(),
		},
	}
	if authd.currentToken == nil {
		log.Printf("[DEBUG] Obtaining authentication for user %s", authd.UserName)
		token, err := conf.PasswordCredentialsToken(oauth2.NoContext, authd.UserName, authd.password)
		if err != nil {
			return nil, err
		}
		authd.currentToken = token
	}
	log.Printf("[DEBUG] Refreshing current token if required")
	oauthClient := conf.Client(oauth2.NoContext, authd.currentToken)

	return newClient(authd.APIURL, authd.Account, oauthClient)
}

func (authd *authdetails) apiClientAuth() (*brightbox.Client, error) {
	conf := clientcredentials.Config{
		ClientID:     authd.APIClient,
		ClientSecret: authd.APISecret,
		Scopes:       infrastructureScope,
		TokenURL:     authd.tokenURL(),
	}
	log.Printf("[DEBUG] Obtaining API client authorisation for client %s", authd.APIClient)
	oauthClient := conf.Client(oauth2.NoContext)
	if authd.currentToken == nil {
		log.Printf("[DEBUG] Retrieving auth token for %s", conf.ClientID)
		token, err := conf.Token(oauth2.NoContext)
		if err != nil {
			return nil, err
		}
		authd.currentToken = token
	}

	return newClient(authd.APIURL, authd.Account, oauthClient)
}

func newClient(apiURL, account string, client *http.Client) (*brightbox.Client, error) {
	client.Transport = logging.NewTransport("Brightbox", client.Transport)

	return brightbox.NewClient(apiURL, account, client)
}
