package brightbox

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBrightboxServerGroup() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Brightbox Server Group resource",
		CreateContext: resourceBrightboxServerGroupCreate,
		ReadContext:   resourceBrightboxServerGroupRead,
		UpdateContext: resourceBrightboxServerGroupUpdate,
		DeleteContext: resourceBrightboxServerGroupDelete,
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
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[INFO] Creating Server Group")
	serverGroupOpts := brightbox.ServerGroupOptions{}
	err := addUpdateableServerGroupOptions(d, &serverGroupOpts)
	if err != nil {
		return diag.FromErr(err)
	}

	serverGroup, err := client.CreateServerGroup(ctx, serverGroupOpts)
	if err != nil {
		return diag.Errorf("Error creating Server Group: %s", err)
	}

	d.SetId(serverGroup.ID)

	return setServerGroupAttributes(d, serverGroup)
}

func resourceBrightboxServerGroupRead(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient

	serverGroup, err := client.ServerGroup(ctx, d.Id())
	if err != nil {
		if strings.HasPrefix(err.Error(), "missing_resource:") {
			log.Printf("[WARN] Server Group not found, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error retrieving Server Group details: %s", err)
	}

	return setServerGroupAttributes(d, serverGroup)
}

func resourceBrightboxServerGroupDelete(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient

	serverGroup, err := client.ServerGroup(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Error retrieving Server Group details: %s", err)
	}
	if len(serverGroup.Servers) > 0 {
		err := clearServerList(ctx, client, serverGroup, d.Timeout(schema.TimeoutDelete))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	log.Printf("[INFO] Deleting Server Group %s", d.Id())
	_, err = client.DestroyServerGroup(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Error deleting Server Group (%s): %s", d.Id(), err)
	}
	return nil
}

func resourceBrightboxServerGroupUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient

	serverGroupOpts := brightbox.ServerGroupOptions{
		ID: d.Id(),
	}
	err := addUpdateableServerGroupOptions(d, &serverGroupOpts)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] Server Group update configuration: %v", serverGroupOpts)

	serverGroup, err := client.UpdateServerGroup(ctx, serverGroupOpts)
	if err != nil {
		return diag.Errorf("Error updating Server Group (%s): %s", serverGroupOpts.ID, err)
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
) diag.Diagnostics {
	d.Set("name", serverGroup.Name)
	d.Set("description", serverGroup.Description)
	d.Set("default", serverGroup.Default)
	d.Set("fqdn", serverGroup.Fqdn)
	return nil
}

func clearServerList(ctx context.Context, client *brightbox.Client, iniitialServerGroup *brightbox.ServerGroup, timeout time.Duration) error {
	serverID := iniitialServerGroup.ID
	serverList := iniitialServerGroup.Servers
	serverIds := serverGroupMemberListFromNodes(serverList)
	log.Printf("[INFO] Removing servers %v from server group %s", serverIds, serverID)
	_, err := client.RemoveServersFromServerGroup(ctx, serverID, serverIds)
	if err != nil {
		return fmt.Errorf("Error removing servers from server group %s", serverID)
	}
	// Wait for group to empty
	return resource.Retry(
		timeout,
		func() *resource.RetryError {
			serverGroup, err := client.ServerGroup(ctx, serverID)
			if err != nil {
				return resource.NonRetryableError(
					fmt.Errorf("Error retrieving Server Group details: %s", err),
				)
			}
			if len(serverGroup.Servers) > 0 {
				return resource.RetryableError(
					fmt.Errorf("Error: servers %v still in server group %s", serverIDListFromNodes(serverGroup.Servers), serverGroup.ID),
				)
			}
			return nil
		},
	)
}
