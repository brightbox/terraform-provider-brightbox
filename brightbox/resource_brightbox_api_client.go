package brightbox

import (
	"fmt"
	"log"
	"strings"

	"github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

var valid_permissions_groups = []string{"full", "storage"}

func resourceBrightboxApiClient() *schema.Resource {
	return &schema.Resource{
		Create: resourceBrightboxApiClientCreate,
		Read:   resourceBrightboxApiClientRead,
		Update: resourceBrightboxApiClientUpdate,
		Delete: resourceBrightboxApiClientDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"secret": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"permissions_group": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      valid_permissions_groups[0],
				ValidateFunc: validation.StringInSlice(valid_permissions_groups, false),
			},
			"account": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceBrightboxApiClientCreate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	log.Printf("[INFO] Creating Api Client")
	api_client_opts := &brightbox.ApiClientOptions{}
	err := addUpdateableApiClientOptions(d, api_client_opts)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Api Client create configuration: %#v", api_client_opts)
	api_client, err := client.CreateApiClient(api_client_opts)
	if err != nil {
		return fmt.Errorf("Error creating Api Client: %s", err)
	}

	d.SetId(api_client.Id)

	return setApiClientAttributes(d, api_client)
}

func resourceBrightboxApiClientRead(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	api_client, err := client.ApiClient(d.Id())
	if err != nil {
		if strings.HasPrefix(err.Error(), "missing_resource:") {
			log.Printf("[WARN] Api Client not found, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving Api Client details: %s", err)
	}

	log.Printf("[DEBUG] Api Client read: %#v", api_client)
	return setApiClientAttributes(d, api_client)
}

func resourceBrightboxApiClientDelete(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	log.Printf("[INFO] Deleting Api Client %s", d.Id())
	err := client.DestroyApiClient(d.Id())
	if err != nil {
		if strings.HasPrefix(err.Error(), "missing_resource:") {
			log.Printf("[WARN] Api Client not found, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error deleting Api Client (%s): %s", d.Id(), err)
	}
	return nil
}

func resourceBrightboxApiClientUpdate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	api_client_opts := &brightbox.ApiClientOptions{
		Id: d.Id(),
	}
	err := addUpdateableApiClientOptions(d, api_client_opts)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Api Client update configuration: %#v", api_client_opts)

	api_client, err := client.UpdateApiClient(api_client_opts)
	if err != nil {
		return fmt.Errorf("Error updating Api Client (%s): %s", api_client_opts.Id, err)
	}

	return setApiClientAttributes(d, api_client)
}

func addUpdateableApiClientOptions(
	d *schema.ResourceData,
	opts *brightbox.ApiClientOptions,
) error {
	assign_string(d, &opts.Name, "name")
	assign_string(d, &opts.Description, "description")
	assign_string(d, &opts.PermissionsGroup, "permissions_group")
	return nil
}

func setApiClientAttributes(
	d *schema.ResourceData,
	api_client *brightbox.ApiClient,
) error {
	d.Set("name", api_client.Name)
	d.Set("description", api_client.Description)
	d.Set("permissions_group", api_client.PermissionsGroup)
	d.Set("account", api_client.Account.Id)

	// Only update the secret if it is set
	if api_client.Secret != "" {
		d.Set("secret", api_client.Secret)
	}
	return nil
}
