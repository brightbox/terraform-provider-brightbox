package brightbox

import (
	"fmt"
	"log"
	"strings"
	"time"

	brightbox "github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceBrightboxServerGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceBrightboxServerGroupCreate,
		Read:   resourceBrightboxServerGroupRead,
		Update: resourceBrightboxServerGroupUpdate,
		Delete: resourceBrightboxServerGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultTimeout),
			Delete: schema.DefaultTimeout(defaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "USer editable label",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"description": {
				Description: "USer editable label",
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

	d.SetId(serverGroup.Id)

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
		Id: d.Id(),
	}
	err := addUpdateableServerGroupOptions(d, serverGroupOpts)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Server Group update configuration: %#v", serverGroupOpts)

	serverGroup, err := client.UpdateServerGroup(serverGroupOpts)
	if err != nil {
		return fmt.Errorf("Error updating Server Group (%s): %s", serverGroupOpts.Id, err)
	}

	return setServerGroupAttributes(d, serverGroup)
}

func addUpdateableServerGroupOptions(
	d *schema.ResourceData,
	opts *brightbox.ServerGroupOptions,
) error {
	assign_string(d, &opts.Name, "name")
	assign_string(d, &opts.Description, "description")
	return nil
}

func setServerGroupAttributes(
	d *schema.ResourceData,
	serverGroup *brightbox.ServerGroup,
) error {
	d.Set("name", serverGroup.Name)
	d.Set("description", serverGroup.Description)
	return nil
}

func serverIDList(servers []brightbox.Server) []string {
	var result []string
	for _, srv := range servers {
		result = append(result, srv.Id)
	}
	return result
}

func clearServerList(client *brightbox.Client, iniitialServerGroup *brightbox.ServerGroup, timeout time.Duration) error {
	serverID := iniitialServerGroup.Id
	serverList := iniitialServerGroup.Servers
	serverIds := serverIDList(serverList)
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
					fmt.Errorf("Error: servers %#v still in server group %s", serverIDList(serverGroup.Servers), serverGroup.Id),
				)
			}
			return nil
		},
	)
}
