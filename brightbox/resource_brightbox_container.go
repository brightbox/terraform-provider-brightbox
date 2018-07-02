package brightbox

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	default_container_permission = "storage"
	defaultOrbitAuthUrl          = "https://orbit.brightbox.com/v1/"
)

func resourceBrightboxContainer() *schema.Resource {
	return &schema.Resource{
		Create: resourceBrightboxContainerCreate,
		Read:   resourceBrightboxContainerRead,
		Update: resourceBrightboxContainerUpdate,
		Delete: resourceBrightboxContainerDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: mustNotBeEmptyString,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"auth_user": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"auth_key": &schema.Schema{
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"account_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"orbit_url": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("BRIGHTBOX_ORBIT_URL", defaultOrbitAuthUrl),
				Description: "Brightbox Cloud Orbit URL for selected Region",
			},
		},
	}
}

func resourceBrightboxContainerCreate(
	d *schema.ResourceData,
	meta interface{},
) error {
	composite := meta.(*CompositeClient)
	client := composite.ApiClient

	log.Printf("[INFO] Creating API client")
	api_client, err := createApiClient(d, client)
	if err != nil {
		return fmt.Errorf("Error creating Api Client: %s", err)
	}
	setInitialAccountAttributes(d, api_client)
	container_url, err := createContainerUrl(d, api_client.Name)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Creating container at %s in Orbit", container_url)
	log.Printf("[DEBUG] Using Auth Token: %s", *composite.OrbitAuthToken)
	err = createContainer(container_url, composite.OrbitAuthToken)
	if err != nil {
		return err
	}
	return setContainerAttributes(d, api_client)
}

func resourceBrightboxContainerDelete(
	d *schema.ResourceData,
	meta interface{},
) error {
	composite := meta.(*CompositeClient)
	client := composite.ApiClient

	container_url, err := createContainerUrl(d, d.Get("name").(string))
	if err != nil {
		return err
	}
	log.Printf("[INFO] Removing container %s in Orbit", container_url)
	log.Printf("[DEBUG] Using Auth Token: %s", *composite.OrbitAuthToken)
	err = destroyContainer(container_url, composite.OrbitAuthToken)
	if err != nil {
		return err
	}
	user := d.Get("auth_user").(string)
	log.Printf("[INFO] Deleting ApiClient %s", user)
	err = client.DestroyApiClient(user)
	if err != nil {
		return fmt.Errorf("Error deleting ApiClient (%s): %s", user, err)
	}
	return nil
}

func resourceBrightboxContainerUpdate(
	d *schema.ResourceData,
	meta interface{},
) error {
	composite := meta.(*CompositeClient)
	client := composite.ApiClient

	if d.HasChange("name") {
		oraw, nraw := d.GetChange("name")
		old_name := oraw.(string)
		new_name := nraw.(string)
		old_url, err := createContainerUrl(d, old_name)
		if err != nil {
			return err
		}
		log.Printf("[INFO] Removing old container %s in Orbit", old_url)
		log.Printf("[DEBUG] Using Auth Token: %s", *composite.OrbitAuthToken)
		err = destroyContainer(old_url, composite.OrbitAuthToken)
		if err != nil {
			return err
		}
		new_url, err := createContainerUrl(d, new_name)
		if err != nil {
			return err
		}
		log.Printf("[INFO] Creating new container %s in Orbit", new_url)
		log.Printf("[DEBUG] Using Auth Token: %s", *composite.OrbitAuthToken)
		err = createContainer(new_url, composite.OrbitAuthToken)
		if err != nil {
			return err
		}
	}
	if d.HasChange("name") || d.HasChange("description") {
		log.Printf("[INFO] Updating API Client")
		api_client_opts := &brightbox.ApiClientOptions{
			Id: d.Get("auth_user").(string),
		}
		addUpdateableApiClientOptions(d, api_client_opts)
		log.Printf("[DEBUG] ApiClient update configuration: %#v", api_client_opts)

		api_client, err := client.UpdateApiClient(api_client_opts)
		if err != nil {
			return fmt.Errorf("Error updating ApiClient (%s): %s", api_client_opts.Id, err)
		}
		return setContainerAttributes(d, api_client)
	}
	return resourceBrightboxContainerRead(d, meta)
}

func resourceBrightboxContainerRead(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	api_client, err := client.ApiClient(d.Get("auth_user").(string))
	if err != nil {
		return fmt.Errorf("Error retrieving ApiClient details: %s", err)
	}

	return setContainerAttributes(d, api_client)
}

func setInitialAccountAttributes(
	d *schema.ResourceData,
	api_client *brightbox.ApiClient,
) {
	log.Printf("[DEBUG] Setting Partial")
	d.Partial(true)
	setAccountAttributes(d, api_client)
	log.Printf("[DEBUG] Setting Key details")
	d.Set("auth_key", api_client.Secret)
	d.SetPartial("auth_key")
}

func setAccountAttributes(
	d *schema.ResourceData,
	api_client *brightbox.ApiClient,
) {
	log.Printf("[DEBUG] Setting Account details")
	d.SetId(api_client.Id)
	d.Set("auth_user", api_client.Id)
	d.SetPartial("auth_user")
	d.Set("account_id", api_client.Account.Id)
	d.SetPartial("account_id")
}

func setContainerAttributes(
	d *schema.ResourceData,
	api_client *brightbox.ApiClient,
) error {
	setAccountAttributes(d, api_client)
	log.Printf("[DEBUG] Setting Container details")
	d.Set("name", api_client.Name)
	d.SetPartial("name")
	d.Set("description", api_client.Description)
	d.SetPartial("description")
	log.Printf("[DEBUG] Clearing Partial")
	d.Partial(false)
	return nil
}

func addUpdateableApiClientOptions(
	d *schema.ResourceData,
	opts *brightbox.ApiClientOptions,
) {
	assign_string(d, &opts.Name, "name")
	assign_string(d, &opts.Description, "description")
}

func createApiClient(
	d *schema.ResourceData,
	client *brightbox.Client,
) (*brightbox.ApiClient, error) {
	permission_group := default_container_permission
	api_client_opts := &brightbox.ApiClientOptions{
		PermissionsGroup: &permission_group,
	}
	addUpdateableApiClientOptions(d, api_client_opts)
	return client.CreateApiClient(api_client_opts)
}

func createContainerUrl(d *schema.ResourceData, name string) (string, error) {

	base_url, err := url.Parse(d.Get("orbit_url").(string))
	if err != nil {
		return "", err
	}
	rel_url, err := url.Parse(filepath.Join(base_url.EscapedPath(), d.Get("account_id").(string), name))
	if err != nil {
		return "", err
	}
	return base_url.ResolveReference(rel_url).String(), nil
}

func manipulateContainer(url string, token *string, action string) error {
	req, err := http.NewRequest(action, url, nil)
	if err != nil {
		return fmt.Errorf("Error creating Orbit request %s", err)
	}
	req.Header.Set("X-Auth-Token", *token)
	resp, err := makeHttpRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func createContainer(url string, token *string) error {
	return manipulateContainer(url, token, "PUT")
}

func destroyContainer(url string, token *string) error {
	return manipulateContainer(url, token, "DELETE")
}

func makeHttpRequest(req *http.Request) (resp *http.Response, err error) {
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		if resp != nil {
			defer resp.Body.Close()
		}
		return resp, fmt.Errorf("Error accessing Orbit: %s", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusPartialContent {
		defer resp.Body.Close()
		return resp, fmt.Errorf("HTTP error response %v", resp.Status)
	}
	return resp, nil
}
