package brightbox

import (
	"context"
	"log"
	"net/http"
	"strings"

	brightbox "github.com/brightbox/gobrightbox"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/logging"
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
	OrbitURL     string
	currentToken oauth2.TokenSource
}

// Authenticate the details and return a client
func (authd *authdetails) authenticatedClient() (*brightbox.Client, *gophercloud.ServiceClient, error) {
	authContext := contextWithLoggedHttpClient()
	if authd.currentToken == nil {
		switch {
		case authd.UserName != "" || authd.password != "":
			if err := authd.getUserTokenSource(authContext); err != nil {
				return nil, nil, err
			}
		default:
			authd.getApiClientTokenSource(authContext)
		}
	}
	log.Printf("[DEBUG] Fetching API Client")
	httpClient := oauth2.NewClient(authContext, authd.currentToken)
	apiclient, err := brightbox.NewClient(authd.APIURL, authd.Account, httpClient)
	if err != nil {
		return nil, nil, err
	}
	if apiclient.AccountId == "" {
		log.Printf("[INFO] Obtaining default account")
		authedAPIClient, err := apiclient.ApiClient(authd.APIClient)
		if err != nil {
			return nil, nil, err
		}
		apiclient.AccountId = authedAPIClient.Account.Id
		authd.Account = apiclient.AccountId
	}

	log.Printf("[DEBUG] Fetching Orbit Service Client")
	serviceclient, err := authd.getServiceClient(authContext)
	if err != nil {
		return nil, nil, err
	}
	return apiclient, serviceclient, nil
}

func (authd *authdetails) tokenURL() string {
	return strings.TrimSuffix(authd.APIURL, "/") + "/token"
}

func (authd *authdetails) storageURL() string {
	return strings.TrimSuffix(authd.OrbitURL, "/") + "/" + authd.Account + "/"
}

func (authd *authdetails) getUserTokenSource(ctx context.Context) error {
	conf := oauth2.Config{
		ClientID:     authd.APIClient,
		ClientSecret: authd.APISecret,
		Scopes:       infrastructureScope,
		Endpoint: oauth2.Endpoint{
			AuthURL:   authd.APIURL,
			TokenURL:  authd.tokenURL(),
			AuthStyle: oauth2.AuthStyleInHeader,
		},
	}
	log.Printf("[DEBUG] Obtaining Tokensource for user %s", authd.UserName)
	token, err := conf.PasswordCredentialsToken(ctx, authd.UserName, authd.password)
	if err != nil {
		return err
	}
	authd.currentToken = conf.TokenSource(ctx, token)
	return nil
}

func (authd *authdetails) getApiClientTokenSource(ctx context.Context) {
	conf := clientcredentials.Config{
		ClientID:     authd.APIClient,
		ClientSecret: authd.APISecret,
		Scopes:       infrastructureScope,
		TokenURL:     authd.tokenURL(),
		AuthStyle:    oauth2.AuthStyleInHeader,
	}
	log.Printf("[DEBUG] Obtaining Tokensource for client %s", authd.APIClient)
	authd.currentToken = conf.TokenSource(ctx)
}

func contextWithLoggedHttpClient() context.Context {
	client := cleanhttp.DefaultClient()
	client.Transport = logging.NewTransport("Brightbox", client.Transport)
	return context.WithValue(context.Background(), oauth2.HTTPClient, client)
}

func (authd *authdetails) getServiceClient(ctx context.Context) (*gophercloud.ServiceClient, error) {
	pc, err := authd.getProviderClient(ctx)
	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] Obtaining Orbit Service Client")
	return openstack.NewObjectStorageV1(pc, gophercloud.EndpointOpts{})
}

func (authd *authdetails) getProviderClient(ctx context.Context) (*gophercloud.ProviderClient, error) {
	log.Printf("[DEBUG] Obtaining Provider Client")
	client, err := openstack.NewClient(authd.OrbitURL)
	if err != nil {
		return nil, err
	}
	client.EndpointLocator = func(opts gophercloud.EndpointOpts) (string, error) {
		return authd.storageURL(), nil
	}
	if httpClient, ok := ctx.Value(oauth2.HTTPClient).(*http.Client); ok {
		client.HTTPClient = *httpClient
	}
	client.Context = ctx
	client.ReauthFunc = func() error {
		return client.SetTokenAndAuthResult(authd)
	}
	err = client.ReauthFunc()
	return client, err
}

func (authd *authdetails) ExtractTokenID() (string, error) {
	token, err := authd.currentToken.Token()
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}
