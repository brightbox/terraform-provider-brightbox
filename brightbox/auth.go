package brightbox

import (
	"context"
	"fmt"
	"log"
	"net/http"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/brightbox/gobrightbox/v2/clientcredentials"
	"github.com/brightbox/gobrightbox/v2/endpoint"
	"github.com/brightbox/gobrightbox/v2/enums/accountstatus"
	"github.com/brightbox/gobrightbox/v2/passwordcredentials"
	"github.com/gophercloud/gophercloud"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"golang.org/x/oauth2"
)

func authenticatedClients(authCtx context.Context, authd authdetails) (*brightbox.Client, *gophercloud.ServiceClient, diag.Diagnostics) {
	apiContext, apiCancel := context.WithCancel(context.Background())
	defer apiCancel()
	apiContext = contextWithLoggedHTTPClient(apiContext)

	log.Printf("[DEBUG] Fetching Infrastructure Client")
	client, err := brightbox.Connect(apiContext, confFromAuthd(authd))
	if err != nil {
		return nil, nil, diag.FromErr(err)
	}

	var diags diag.Diagnostics

	if authd.Account == "" {
		log.Printf("[INFO] Obtaining default account")

		accounts, err := client.Accounts(authCtx)
		if err != nil {
			return nil, nil, diag.FromErr(err)
		}
		authd.Account = accounts[0].ID
		log.Printf("[DEBUG] default account is %v", authd.Account)
		diags = checkIsActive(diags, &accounts[0])
	} else {
		log.Printf("[INFO] Checking credentials have access to %v", authd.Account)
		account, err := client.Account(authCtx, authd.Account)
		if err != nil {
			return nil, nil, diag.Errorf("Unable to access account %v with supplied credentials", authd.Account)
		}
		log.Printf("[DEBUG] account check passsed")
		diags = checkIsActive(diags, account)
	}

	log.Printf("[DEBUG] Building Orbit Client")
	oe, err := orbitEndpointFromAuthd(authd)
	if err != nil {
		return nil, nil, append(diags, diag.FromErr(err)...)
	}

	storageContext, storageCancel := context.WithCancel(context.Background())
	defer storageCancel()
	storageContext = contextWithLoggedHTTPClient(storageContext)
	orbit, err := orbitServiceClient(storageContext, client, oe)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	return client, orbit, diags
}

func checkIsActive(diags diag.Diagnostics, account *brightbox.Account) diag.Diagnostics {
	if account.Status == accountstatus.Active {
		return diags
	}
	return append(diags,
		diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("Account %v is not active", account.ID),
			Detail:   fmt.Sprintf("The account %v is showing state %v\nIf this is unexpected, please use the GUI to contact Brightbox Support", account.ID, account.Status),
		},
	)
}

func orbitServiceClient(serviceContext context.Context, client *brightbox.Client, endpoint string) (*gophercloud.ServiceClient, error) {
	pc := &gophercloud.ProviderClient{}
	if httpClient, ok := serviceContext.Value(oauth2.HTTPClient).(*http.Client); ok {
		pc.HTTPClient = *httpClient
	}
	err := pc.SetTokenAndAuthResult(client)
	if err != nil {
		return nil, err
	}
	pc.ReauthFunc = func() error {
		return pc.SetTokenAndAuthResult(pc.GetAuthResult())
	}

	return &gophercloud.ServiceClient{
		ProviderClient: pc,
		Endpoint:       endpoint,
	}, nil
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
	if logging.IsDebugOrHigher() {
		log.Printf("[DEBUG] Enabling HTTP requests/responses tracing")
		client.Transport = logging.NewTransport("Brightbox", client.Transport)
	}
	return context.WithValue(ctx, oauth2.HTTPClient, client)
}
