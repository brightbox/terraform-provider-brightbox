package brightbox

import (
	"fmt"

	"github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

const (
	// Terraform application client credentials
	defaultClientID     = "app-dkmch"
	defaultClientSecret = "uogoelzgt0nwawb"
	passwordEnvVar      = "BRIGHTBOX_PASSWORD"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"apiclient": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("BRIGHTBOX_CLIENT", defaultClientID),
				Description: "Brightbox Cloud API Client",
			},
			"apisecret": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("BRIGHTBOX_CLIENT_SECRET", defaultClientSecret),
				Description: "Brightbox Cloud API Client Secret",
			},
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("BRIGHTBOX_USER_NAME", nil),
				Description: "Brightbox Cloud User Name",
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(passwordEnvVar, nil),
				Description: "Brightbox Cloud Password for User Name",
			},
			"account": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("BRIGHTBOX_ACCOUNT", nil),
				Description: "Brightbox Cloud Account to operate on",
			},
			"apiurl": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("BRIGHTBOX_API_URL", brightbox.DefaultRegionApiURL),
				Description: "Brightbox Cloud Api URL for selected Region",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"brightbox_image": dataSourceBrightboxImage(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"brightbox_server":          resourceBrightboxServer(),
			"brightbox_cloudip":         resourceBrightboxCloudip(),
			"brightbox_server_group":    resourceBrightboxServerGroup(),
			"brightbox_firewall_policy": resourceBrightboxFirewallPolicy(),
			"brightbox_firewall_rule":   resourceBrightboxFirewallRule(),
			"brightbox_load_balancer":   resourceBrightboxLoadBalancer(),
			"brightbox_database_server": resourceBrightboxDatabaseServer(),
			"brightbox_container":       resourceBrightboxContainer(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := &authdetails{
		APIClient: d.Get("apiclient").(string),
		APISecret: d.Get("apisecret").(string),
		UserName:  d.Get("username").(string),
		password:  d.Get("password").(string),
		Account:   d.Get("account").(string),
		APIURL:    d.Get("apiurl").(string),
	}

	if config.APIClient == defaultClientID && config.APISecret == defaultClientSecret {
		if config.Account == "" {
			return nil,
				fmt.Errorf("Must specify Account with User Credentials")
		}
	} else {
		if config.UserName != "" || config.password != "" {
			return nil,
				fmt.Errorf("User Credentials not used with API Client.")
		}
	}

	return config.Client()
}
