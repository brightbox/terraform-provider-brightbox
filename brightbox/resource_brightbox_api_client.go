package brightbox

import (
	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/brightbox/gobrightbox/v2/enums/permissionsgroup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceBrightboxAPIClient() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Brightbox API Client resource",
		CreateContext: resourceBrightboxAPIClientCreate,
		ReadContext:   resourceBrightboxAPIClientRead,
		UpdateContext: resourceBrightboxAPIClientUpdate,
		DeleteContext: resourceBrightboxAPIClientDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultTimeout),
			Delete: schema.DefaultTimeout(defaultTimeout),
		},

		Schema: map[string]*schema.Schema{

			"account": {
				Description: "The account the API client relates to",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"description": {
				Description: "Verbose Description of this client",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"name": {
				Description: "Human Readable Name",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"permissions_group": {
				Description: "Summary of the permissions granted to the client (full, storage)",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     permissionsgroup.Full.String(),
				ValidateFunc: validation.StringInSlice(
					permissionsgroup.ValidStrings,
					false),
			},

			"secret": {
				Description: "A shared secret the client must present when authenticating",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

var (
	resourceBrightboxAPIClientCreate = resourceBrightboxCreate(
		(*brightbox.Client).CreateAPIClient,
		"API Client",
		addUpdateableAPIClientOptions,
		setAPIClientAttributes,
	)

	resourceBrightboxAPIClientRead = resourceBrightboxReadStatus(
		(*brightbox.Client).APIClient,
		"API Client",
		setAPIClientAttributes,
		apiClientRevoked,
	)

	resourceBrightboxAPIClientUpdate = resourceBrightboxUpdate(
		(*brightbox.Client).UpdateAPIClient,
		"API Client",
		apiClientFromID,
		addUpdateableAPIClientOptions,
		setAPIClientAttributes,
	)

	resourceBrightboxAPIClientDelete = resourceBrightboxDelete(
		(*brightbox.Client).DestroyAPIClient,
		"API Client",
	)
)

func apiClientFromID(id string) *brightbox.APIClientOptions {
	return &brightbox.APIClientOptions{
		ID: id,
	}
}

func addUpdateableAPIClientOptions(
	d *schema.ResourceData,
	opts *brightbox.APIClientOptions,
) diag.Diagnostics {
	assignString(d, &opts.Name, "name")
	assignString(d, &opts.Description, "description")
	assignEnum(d, &opts.PermissionsGroup, "permissions_group")
	return nil
}

func setAPIClientAttributes(
	d *schema.ResourceData,
	apiClient *brightbox.APIClient,
) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	d.SetId(apiClient.ID)
	err = d.Set("name", apiClient.Name)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("description", apiClient.Description)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("permissions_group", apiClient.PermissionsGroup.String())
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("account", apiClient.Account.ID)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}

	// Only update the secret if it is set
	if apiClient.Secret != "" {
		err := d.Set("secret", apiClient.Secret)
		if err != nil {
			diags = append(diags, diag.Errorf("unexpected: %s", err)...)
		}
	}
	return diags
}

func apiClientRevoked(obj *brightbox.APIClient) bool {
	return obj.RevokedAt != nil
}
