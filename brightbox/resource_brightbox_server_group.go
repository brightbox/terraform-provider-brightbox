package brightbox

import (
	"context"
	"fmt"
	"log"
	"time"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBrightboxServerGroup() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Brightbox Server Group resource",
		CreateContext: resourceBrightboxServerGroupCreate,
		ReadContext:   resourceBrightboxServerGroupRead,
		UpdateContext: resourceBrightboxServerGroupUpdate,
		DeleteContext: resourceBrightboxServerGroupClearAndDelete,
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
			"firewall_policy": {
				Description: "The firewall policy associated with this server group",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

var (
	resourceBrightboxServerGroupCreate = resourceBrightboxCreate(
		(*brightbox.Client).CreateServerGroup,
		"Server Group",
		addUpdateableServerGroupOptions,
		setServerGroupAttributes,
	)
	resourceBrightboxServerGroupRead = resourceBrightboxRead(
		(*brightbox.Client).ServerGroup,
		"Server Group",
		setServerGroupAttributes,
	)
	resourceBrightboxServerGroupUpdate = resourceBrightboxUpdate(
		(*brightbox.Client).UpdateServerGroup,
		"Server Group",
		serverGroupFromID,
		addUpdateableServerGroupOptions,
		setServerGroupAttributes,
	)

	resourceBrightboxServerGroupDelete = resourceBrightboxDelete(
		(*brightbox.Client).DestroyServerGroup,
		"Server Group",
	)
)

func serverGroupFromID(id string) *brightbox.ServerGroupOptions {
	return &brightbox.ServerGroupOptions{
		ID: id,
	}
}

func addUpdateableServerGroupOptions(
	d *schema.ResourceData,
	opts *brightbox.ServerGroupOptions,
) diag.Diagnostics {
	assignString(d, &opts.Name, "name")
	assignString(d, &opts.Description, "description")
	return nil
}

func setServerGroupAttributes(
	d *schema.ResourceData,
	serverGroup *brightbox.ServerGroup,
) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	d.SetId(serverGroup.ID)
	err = d.Set("name", serverGroup.Name)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("description", serverGroup.Description)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("default", serverGroup.Default)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("fqdn", serverGroup.Fqdn)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	firewallTarget := ""
	if serverGroup.FirewallPolicy != nil {
		firewallTarget = serverGroup.FirewallPolicy.ID
	}
	err = d.Set("firewall_policy", firewallTarget)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	return diags
}

func resourceBrightboxServerGroupClearAndDelete(
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
	return resourceBrightboxServerGroupDelete(ctx, d, meta)
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
		func() *retry.RetryError {
			serverGroup, err := client.ServerGroup(ctx, serverID)
			if err != nil {
				return retry.NonRetryableError(
					fmt.Errorf("Error retrieving Server Group details: %s", err),
				)
			}
			if len(serverGroup.Servers) > 0 {
				return retry.RetryableError(
					fmt.Errorf("Error: servers %v still in server group %s", serverIDListFromNodes(serverGroup.Servers), serverGroup.ID),
				)
			}
			return nil
		},
	)
}
