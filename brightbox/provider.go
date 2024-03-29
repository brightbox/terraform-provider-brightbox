package brightbox

import (
	"context"
	"time"

	"github.com/brightbox/gobrightbox/v2/endpoint"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	defaultTimeout     = 5 * time.Minute
	minimumRefreshWait = 3 * time.Second
	checkDelay         = 10 * time.Second
)

// Provider is the Brightbox Terraform driver root
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"account": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(accountEnvVar, nil),
				Description: "Brightbox Cloud Account to operate upon",
			},
			"apiclient": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(clientEnvVar, defaultClientID),
				Description: "Brightbox Cloud API Client/OAuth Application ID",
			},
			"apisecret": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(clientSecretEnvVar, defaultClientSecret),
				Description: "Brightbox Cloud API Client/OAuth Application Secret",
			},
			"apiurl": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(apiURLEnvVar, endpoint.DefaultBaseURL),
				Description: "Brightbox Cloud Api URL for selected Region",
			},
			"orbit_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(orbitURLEnvVar, endpoint.DefaultOrbitBaseURL),
				Description: "Brightbox Cloud Orbit URL for selected Region",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc(passwordEnvVar, nil),
				Description: "Brightbox Cloud Password for User Name",
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(usernameEnvVar, nil),
				Description: "Brightbox Cloud User Name",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"brightbox_image":             dataSourceBrightboxImage(),
			"brightbox_database_type":     dataSourceBrightboxDatabaseType(),
			"brightbox_server_group":      dataSourceBrightboxServerGroup(),
			"brightbox_server_type":       dataSourceBrightboxServerType(),
			"brightbox_database_snapshot": dataSourceBrightboxDatabaseSnapshot(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"brightbox_server":                  resourceBrightboxServer(),
			"brightbox_cloudip":                 resourceBrightboxCloudIP(),
			"brightbox_server_group":            resourceBrightboxServerGroup(),
			"brightbox_server_group_membership": resourceBrightboxServerGroupMembership(),
			"brightbox_firewall_policy":         resourceBrightboxFirewallPolicy(),
			"brightbox_firewall_rule":           resourceBrightboxFirewallRule(),
			"brightbox_load_balancer":           resourceBrightboxLoadBalancer(),
			"brightbox_database_server":         resourceBrightboxDatabaseServer(),
			"brightbox_orbit_container":         resourceBrightboxContainer(),
			"brightbox_api_client":              resourceBrightboxAPIClient(),
			"brightbox_config_map":              resourceBrightboxConfigMap(),
			"brightbox_volume":                  resourceBrightboxVolume(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return configureClient(
		ctx,
		authdetails{
			APIClient: d.Get("apiclient").(string),
			APISecret: d.Get("apisecret").(string),
			UserName:  d.Get("username").(string),
			password:  d.Get("password").(string),
			Account:   d.Get("account").(string),
			APIURL:    d.Get("apiurl").(string),
			OrbitURL:  d.Get("orbit_url").(string),
		},
	)
}
