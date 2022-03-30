package brightbox

import (
	"context"
	"log"

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
	var conf brightbox.Oauth2
	if authd.UserName != "" || authd.password != "" {
		conf = &passwordcredentials.Config{
			UserName: authd.UserName,
			Password: authd.password,
			ID:       authd.APIClient,
			Secret:   authd.APISecret,
			Config: endpoint.Config{
				BaseURL: authd.APIURL,
				Account: authd.Account,
				Scopes:  endpoint.InfrastructureScope,
			},
		}
	} else {
		conf = &clientcredentials.Config{
			ID:     authd.APIClient,
			Secret: authd.APISecret,
			Config: endpoint.Config{
				BaseURL: authd.APIURL,
				Scopes:  endpoint.InfrastructureScope,
			},
		}
	}

	authContext := context.Background()
	if logging.IsDebugOrHigher() {
		log.Printf("[DEBUG] Enabling HTTP requests/responses tracing")
		authContext = contextWithLoggedHTTPClient(authContext)
	}
	log.Printf("[DEBUG] Fetching Infrastructure Client")
	client, err := brightbox.Connect(authContext, conf)
	if err != nil {
		return nil, nil, diag.FromErr(err)
	}
	return client, nil, nil
}

func contextWithLoggedHTTPClient(ctx context.Context) context.Context {
	client := cleanhttp.DefaultClient()
	client.Transport = logging.NewTransport("Brightbox", client.Transport)
	return context.WithValue(ctx, oauth2.HTTPClient, client)
}
