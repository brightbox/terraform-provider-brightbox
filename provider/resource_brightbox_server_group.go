package brightbox

import (
	"fmt"
	"log"
	"time"

	"github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceBrightboxServerGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceBrightboxServerGroupCreate,
		Read:   resourceBrightboxServerGroupRead,
		Update: resourceBrightboxServerGroupUpdate,
		Delete: resourceBrightboxServerGroupDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},

			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},
		},
	}
}

func resourceBrightboxServerGroupCreate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*brightbox.Client)

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

	setServerGroupAttributes(d, server_group)

	return nil
}

func setServerGroupAttributes(
	d *schema.ResourceData,
	server_group *brightbox.ServerGroup,
) {
	d.Set("name", server_group.Name)
	d.Set("description", server_group.Description)

}

func resourceBrightboxServerGroupRead(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*brightbox.Client)

	server_group, err := client.ServerGroup(d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving Server Group details: %s", err)
	}

	setServerGroupAttributes(d, server_group)

	return nil
}

func resourceBrightboxServerGroupDelete(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*brightbox.Client)

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
	client := meta.(*brightbox.Client)

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

	setServerGroupAttributes(d, server_group)
	return nil
}

func addUpdateableServerGroupOptions(
	d *schema.ResourceData,
	opts *brightbox.ServerGroupOptions,
) error {
	if d.HasChange("name") {
		var temp string
		if attr, ok := d.GetOk("name"); ok {
			temp = attr.(string)
		}
		opts.Name = &temp
	}
	if d.HasChange("description") {
		var temp string
		if attr, ok := d.GetOk("description"); ok {
			temp = attr.(string)
		}
		opts.Description = &temp
	}
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
		func() error {
			server_group, err := client.ServerGroup(serverID)
			if err != nil {
				return resource.RetryError{
					Err: fmt.Errorf("Error retrieving Server Group details: %s", err),
				}
			}
			if len(server_group.Servers) > 0 {
				return fmt.Errorf("Error: servers %#v still in server group %s", serverIdList(server_group.Servers), server_group.Id)
			}
			return nil
		},
	)
}
