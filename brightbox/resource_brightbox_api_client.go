package brightbox

import (
	"context"
	"log"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/brightbox/gobrightbox/v2/status/permissionsgroup"
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

func resourceBrightboxAPIClientCreate(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[INFO] Creating Api Client")
	apiClientOpts := brightbox.APIClientOptions{}
	errs := addUpdateableAPIClientOptions(d, &apiClientOpts)
	if errs.HasError() {
		return errs
	}
	log.Printf("[INFO] Api Client create configuration: %v", apiClientOpts)
	apiClient, err := client.CreateAPIClient(ctx, apiClientOpts)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(apiClient.ID)

	return setAPIClientAttributes(d, apiClient)
}

func resourceBrightboxAPIClientRead(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient

	apiClient, err := client.APIClient(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if apiClient.RevokedAt != nil {
		log.Printf("[WARN] Api Client revoked, removing from state: %s", d.Id())
		d.SetId("")
		return nil
	}

	log.Printf("[DEBUG] Api Client read: %#v", apiClient)
	return setAPIClientAttributes(d, apiClient)
}

func resourceBrightboxAPIClientDelete(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[INFO] Deleting Api Client %s", d.Id())
	_, err := client.DestroyAPIClient(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceBrightboxAPIClientUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient

	apiClientOpts := brightbox.APIClientOptions{
		ID: d.Id(),
	}
	errs := addUpdateableAPIClientOptions(d, &apiClientOpts)
	if errs.HasError() {
		return errs
	}
	log.Printf("[DEBUG] Api Client update configuration: %v", apiClientOpts)

	apiClient, err := client.UpdateAPIClient(ctx, apiClientOpts)
	if err != nil {
		return diag.FromErr(err)
	}

	return setAPIClientAttributes(d, apiClient)
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
	d.Set("name", apiClient.Name)
	d.Set("description", apiClient.Description)
	d.Set("permissions_group", apiClient.PermissionsGroup.String())
	d.Set("account", apiClient.Account.ID)

	// Only update the secret if it is set
	if apiClient.Secret != "" {
		d.Set("secret", apiClient.Secret)
	}
	return nil
}
