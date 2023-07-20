package provider

import (
	"context"
	"fmt"
	"net/http"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/brightbox/gobrightbox/v2/clientcredentials"
	"github.com/brightbox/gobrightbox/v2/endpoint"
	"github.com/brightbox/gobrightbox/v2/enums/accountstatus"
	"github.com/brightbox/gobrightbox/v2/passwordcredentials"
	"github.com/brightbox/terraform-provider-brightbox/internal/diagerrors"
	"github.com/gophercloud/gophercloud"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/oauth2"
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

func authenticatedClients(authCtx context.Context, authd authdetails) (*brightbox.Client, *gophercloud.ServiceClient, diag.Diagnostics) {
	apiContext, apiCancel := context.WithCancel(context.Background())
	defer apiCancel()
	apiContext = contextWithLoggedHTTPClient(apiContext)

	tflog.Debug(authCtx, "Fetching Infrastructure Client")
	var diags diag.Diagnostics
	client, err := brightbox.Connect(apiContext, confFromAuthd(authd))
	if err != nil {
		diags.Append(diagerrors.FromErr(err))
		return nil, nil, diags
	}

	if authd.Account == "" {
		tflog.Info(authCtx, "Obtaining default account")

		accounts, err := client.Accounts(authCtx)
		if err != nil {
			diags.Append(diagerrors.FromErr(err))
			return nil, nil, diags
		}
		authd.Account = accounts[0].ID
		tflog.Debug(authCtx, fmt.Sprintf("default account is %v", authd.Account))
		diags = checkIsActive(diags, &accounts[0])
	} else {
		tflog.Info(authCtx, fmt.Sprintf("Checking credentials have access to %v", authd.Account))
		account, err := client.Account(authCtx, authd.Account)
		if err != nil {
			diags.Append(diagerrors.FromErr(err))
			return nil, nil, diags
		}
		tflog.Debug(authCtx, "account check passsed")
		diags = checkIsActive(diags, account)
	}

	tflog.Debug(authCtx, "Building Orbit Client")
	oe, err := orbitEndpointFromAuthd(authd)
	if err != nil {
		diags.Append(diagerrors.FromErr(err))
		return nil, nil, diags
	}

	storageContext, storageCancel := context.WithCancel(context.Background())
	defer storageCancel()
	storageContext = contextWithLoggedHTTPClient(storageContext)
	orbit, err := orbitServiceClient(storageContext, client, oe)
	if err != nil {
		diags.Append(diagerrors.FromErr(err))
		return nil, nil, diags
	}
	return client, orbit, diags
}

func checkIsActive(diags diag.Diagnostics, account *brightbox.Account) diag.Diagnostics {
	if account.Status == accountstatus.Active {
		return diags
	}
	return append(diags,
		diag.NewWarningDiagnostic(
			fmt.Sprintf("Account %v is not active", account.ID),
			fmt.Sprintf("The account %v is showing state %v\nIf this is unexpected, please use the GUI to contact Brightbox Support", account.ID, account.Status),
		),
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
	// if logging.IsDebugOrHigher() {
	// 	log.Printf("[DEBUG] Enabling HTTP requests/responses tracing")
	// 	client.Transport = logging.NewTransport("Brightbox", client.Transport)
	// }
	return context.WithValue(ctx, oauth2.HTTPClient, client)
}
