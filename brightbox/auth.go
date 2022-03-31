package brightbox

import (
	"context"
	"log"
	"net/http"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/brightbox/gobrightbox/v2/clientcredentials"
	"github.com/brightbox/gobrightbox/v2/endpoint"
	"github.com/brightbox/gobrightbox/v2/passwordcredentials"
	"github.com/gophercloud/gophercloud"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"golang.org/x/oauth2"
)

func authenticatedClients(ctx context.Context, authd authdetails) (*brightbox.Client, *gophercloud.ServiceClient, diag.Diagnostics) {
	authContext := context.Background()
	if logging.IsDebugOrHigher() {
		log.Printf("[DEBUG] Enabling HTTP requests/responses tracing")
		authContext = contextWithLoggedHTTPClient(authContext)
	}

	log.Printf("[DEBUG] Fetching Infrastructure Client")
	client, err := brightbox.Connect(authContext, confFromAuthd(authd))
	if err != nil {
		return nil, nil, diag.FromErr(err)
	}

	log.Printf("[DEBUG] Building Orbit Client")
	oe, err := orbitEndpointFromAuthd(authd)
	if err != nil {
		return nil, nil, diag.FromErr(err)
	}

	token, err := client.AuthToken()
	if err != nil {
		return nil, nil, diag.FromErr(err)
	}

	return client, orbitServiceClient(token, oe, client.HTTPClient()), nil
}

func orbitServiceClient(token, endpoint string, httpClient *http.Client) *gophercloud.ServiceClient {
	result := &gophercloud.ServiceClient{
		ProviderClient: &gophercloud.ProviderClient{
			TokenID: token,
		},
		Endpoint: endpoint,
	}
	if httpClient != nil {
		result.ProviderClient.HTTPClient = *httpClient
	}
	return result
}

func orbitEndpointFromAuthd(authd authdetails) (string, error) {
	conf := &endpoint.Config{
		BaseURL: authd.OrbitURL,
		Account: authd.Account,
	}
	return conf.StorageURL()
}

func confFromAuthd(authd authdetails) brightbox.Oauth2 {
	if authd.UserName != "" || authd.password != "" {
		return &passwordcredentials.Config{
			UserName: authd.UserName,
			Password: authd.password,
			ID:       authd.APIClient,
			Secret:   authd.APISecret,
			Config: endpoint.Config{
				BaseURL: authd.APIURL,
				Account: authd.Account,
				Scopes:  endpoint.FullScope,
			},
		}
	}
	return &clientcredentials.Config{
		ID:     authd.APIClient,
		Secret: authd.APISecret,
		Config: endpoint.Config{
			BaseURL: authd.APIURL,
			Scopes:  endpoint.FullScope,
		},
	}
}

func contextWithLoggedHTTPClient(ctx context.Context) context.Context {
	client := cleanhttp.DefaultClient()
	client.Transport = logging.NewTransport("Brightbox", client.Transport)
	return context.WithValue(ctx, oauth2.HTTPClient, client)
}
