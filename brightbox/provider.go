package brightbox

import (
	brightbox "github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// Provider is the Brightbox Terraform driver root
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
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
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(usernameEnvVar, nil),
				Description: "Brightbox Cloud User Name",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc(passwordEnvVar, nil),
				Description: "Brightbox Cloud Password for User Name",
			},
			"account": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(accountEnvVar, nil),
				Description: "Brightbox Cloud Account to operate on",
			},
			"apiurl": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(apiURLEnvVar, brightbox.DefaultRegionApiURL),
				Description: "Brightbox Cloud Api URL for selected Region",
			},
			"orbit_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(orbitURLEnvVar, brightbox.DefaultOrbitAuthURL),
				Description: "Brightbox Cloud Orbit URL for selected Region",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"brightbox_image":         dataSourceBrightboxImage(),
			"brightbox_database_type": dataSourceBrightboxDatabaseType(),
			"brightbox_server_group":  dataSourceBrightboxServerGroup(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"brightbox_server":          resourceBrightboxServer(),
			"brightbox_cloudip":         resourceBrightboxCloudip(),
			"brightbox_server_group":    resourceBrightboxServerGroup(),
			"brightbox_firewall_policy": resourceBrightboxFirewallPolicy(),
			"brightbox_firewall_rule":   resourceBrightboxFirewallRule(),
			"brightbox_load_balancer":   resourceBrightboxLoadBalancer(),
			"brightbox_database_server": resourceBrightboxDatabaseServer(),
			"brightbox_orbit_container": resourceBrightboxContainer(),
			"brightbox_api_client":      resourceBrightboxAPIClient(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	return (&authdetails{
		APIClient: d.Get("apiclient").(string),
		APISecret: d.Get("apisecret").(string),
		UserName:  d.Get("username").(string),
		password:  d.Get("password").(string),
		Account:   d.Get("account").(string),
		APIURL:    d.Get("apiurl").(string),
		OrbitURL:  d.Get("orbit_url").(string),
	}).Client()
}
