package brightbox

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/brightbox/gobrightbox"
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

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceBrightboxServerGroupCreate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	log.Printf("[INFO] Creating Server Group")
	server_group_opts := &brightbox.ServerGroupOptions{}
	err := addUpdateableServerGroupOptions(d, server_group_opts)
	if err != nil {
		return err
	}

	server_group, err := client.CreateServerGroup(server_group_opts)
	if err != nil {
		return fmt.Errorf("Error creating Server Group: %s", err)
	}

	d.SetId(server_group.Id)

	return setServerGroupAttributes(d, server_group)
}

func resourceBrightboxServerGroupRead(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	server_group, err := client.ServerGroup(d.Id())
	if err != nil {
		if strings.HasPrefix(err.Error(), "missing_resource:") {
			log.Printf("[WARN] Server Group not found, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving Server Group details: %s", err)
	}

	return setServerGroupAttributes(d, server_group)
}

func resourceBrightboxServerGroupDelete(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	server_group, err := client.ServerGroup(d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving Server Group details: %s", err)
	}
	if len(server_group.Servers) > 0 {
		err := clearServerList(client, server_group)
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
	client := meta.(*CompositeClient).ApiClient

	server_group_opts := &brightbox.ServerGroupOptions{
		Id: d.Id(),
	}
	err := addUpdateableServerGroupOptions(d, server_group_opts)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Server Group update configuration: %#v", server_group_opts)

	server_group, err := client.UpdateServerGroup(server_group_opts)
	if err != nil {
		return fmt.Errorf("Error updating Server Group (%s): %s", server_group_opts.Id, err)
	}

	return setServerGroupAttributes(d, server_group)
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
	server_group *brightbox.ServerGroup,
) error {
	d.Set("name", server_group.Name)
	d.Set("description", server_group.Description)
	return nil
}

func serverIdList(servers []brightbox.Server) []string {
	var result []string
	for _, srv := range servers {
		result = append(result, srv.Id)
	}
	return result
}

func clearServerList(client *brightbox.Client, initial_server_group *brightbox.ServerGroup) error {
	serverID := initial_server_group.Id
	server_list := initial_server_group.Servers
	serverIds := serverIdList(server_list)
	log.Printf("[INFO] Removing servers %#v from server group %s", serverIds, serverID)
	_, err := client.RemoveServersFromServerGroup(serverID, serverIds)
	if err != nil {
		return fmt.Errorf("Error removing servers from server group %s", serverID)
	}
	// Wait for group to empty
	return resource.Retry(
		1*time.Minute,
		func() *resource.RetryError {
			server_group, err := client.ServerGroup(serverID)
			if err != nil {
				return resource.NonRetryableError(
					fmt.Errorf("Error retrieving Server Group details: %s", err),
				)
			}
			if len(server_group.Servers) > 0 {
				return resource.RetryableError(
					fmt.Errorf("Error: servers %#v still in server group %s", serverIdList(server_group.Servers), server_group.Id),
				)
			}
			return nil
		},
	)
}
