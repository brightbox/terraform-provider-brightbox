package brightbox

import (
	"fmt"
	"log"
	"strings"
	"time"

	brightbox "github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBrightboxServerGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Brightbox Server Group resource",
		Create:      resourceBrightboxServerGroupCreate,
		Read:        resourceBrightboxServerGroupRead,
		Update:      resourceBrightboxServerGroupUpdate,
		Delete:      resourceBrightboxServerGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultTimeout),
			Delete: schema.DefaultTimeout(defaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"default": {
				Description: "Is this the default group for the account?",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"description": {
				Description: "User editable label",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"fqdn": {
				Description: "Fully Qualified Domain Name",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"name": {
				Description: "User editable label",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func resourceBrightboxServerGroupCreate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[INFO] Creating Server Group")
	serverGroupOpts := &brightbox.ServerGroupOptions{}
	err := addUpdateableServerGroupOptions(d, serverGroupOpts)
	if err != nil {
		return err
	}

	serverGroup, err := client.CreateServerGroup(serverGroupOpts)
	if err != nil {
		return fmt.Errorf("Error creating Server Group: %s", err)
	}

	d.SetId(serverGroup.ID)

	return setServerGroupAttributes(d, serverGroup)
}

func resourceBrightboxServerGroupRead(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	serverGroup, err := client.ServerGroup(d.Id())
	if err != nil {
		if strings.HasPrefix(err.Error(), "missing_resource:") {
			log.Printf("[WARN] Server Group not found, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving Server Group details: %s", err)
	}

	return setServerGroupAttributes(d, serverGroup)
}

func resourceBrightboxServerGroupDelete(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	serverGroup, err := client.ServerGroup(d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving Server Group details: %s", err)
	}
	if len(serverGroup.Servers) > 0 {
		err := clearServerList(client, serverGroup, d.Timeout(schema.TimeoutDelete))
		if err != nil {
			return err
		}
	}

	log.Printf("[INFO] Deleting Server Group %s", d.Id())
	err = client.DestroyServerGroup(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting Server Group (%s): %s", d.Id(), err)
	}
	return nil
}

func resourceBrightboxServerGroupUpdate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	serverGroupOpts := &brightbox.ServerGroupOptions{
		ID: d.Id(),
	}
	err := addUpdateableServerGroupOptions(d, serverGroupOpts)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Server Group update configuration: %#v", serverGroupOpts)

	serverGroup, err := client.UpdateServerGroup(serverGroupOpts)
	if err != nil {
		return fmt.Errorf("Error updating Server Group (%s): %s", serverGroupOpts.ID, err)
	}

	return setServerGroupAttributes(d, serverGroup)
}

func addUpdateableServerGroupOptions(
	d *schema.ResourceData,
	opts *brightbox.ServerGroupOptions,
) error {
	assignString(d, &opts.Name, "name")
	assignString(d, &opts.Description, "description")
	return nil
}

func setServerGroupAttributes(
	d *schema.ResourceData,
	serverGroup *brightbox.ServerGroup,
) error {
	d.Set("name", serverGroup.Name)
	d.Set("description", serverGroup.Description)
	d.Set("default", serverGroup.Default)
	d.Set("fqdn", serverGroup.Fqdn)
	return nil
}

func clearServerList(client *brightbox.Client, iniitialServerGroup *brightbox.ServerGroup, timeout time.Duration) error {
	serverID := iniitialServerGroup.ID
	serverList := iniitialServerGroup.Servers
	serverIds := serverIDListFromNodes(serverList)
	log.Printf("[INFO] Removing servers %#v from server group %s", serverIds, serverID)
	_, err := client.RemoveServersFromServerGroup(serverID, serverIds)
	if err != nil {
		return fmt.Errorf("Error removing servers from server group %s", serverID)
	}
	// Wait for group to empty
	return resource.Retry(
		timeout,
		func() *resource.RetryError {
			serverGroup, err := client.ServerGroup(serverID)
			if err != nil {
				return resource.NonRetryableError(
					fmt.Errorf("Error retrieving Server Group details: %s", err),
				)
			}
			if len(serverGroup.Servers) > 0 {
				return resource.RetryableError(
					fmt.Errorf("Error: servers %#v still in server group %s", serverIDListFromNodes(serverGroup.Servers), serverGroup.ID),
				)
			}
			return nil
		},
	)
}
