package brightbox

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceBrightboxServerGroupMembership() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides non-exclusive mapping to Brightbox Server Groups",
		CreateContext: resourceBrightboxServerGroupMembershipCreate,
		ReadContext:   resourceBrightboxServerGroupMembershipRead,
		UpdateContext: resourceBrightboxServerGroupMembershipUpdate,
		DeleteContext: resourceBrightboxServerGroupMembershipDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceBrightboxServerGroupMembershipImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultTimeout),
			Delete: schema.DefaultTimeout(defaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"group": {
				Description:  "Server Group ID",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(serverGroupRegexp, "must be a valid server group ID"),
			},

			"servers": {
				Description: "List of servers to add to group",
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringMatch(serverRegexp, "must be a valid server ID"),
				},
			},
		},
	}
}

func resourceBrightboxServerGroupMembershipRead(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	objectName := "ServerGroup"
	target := d.Get("group").(string)
	reader := (*brightbox.Client).ServerGroup
	setter := setServerGroupMembershipAttributes
	client := meta.(*CompositeClient).APIClient

	log.Printf("[DEBUG] %s resource read called for %s", objectName, target)

	object, err := reader(client, ctx, target)
	if err != nil {
		var apierror *brightbox.APIError
		if !d.IsNewResource() && errors.As(err, &apierror) {
			if apierror.StatusCode == 404 {
				log.Printf("[WARN] %s not found, removing from state: %s", objectName, target)
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] setting details from returned object: %+v", *object)
	return setter(d, object)
}

func resourceBrightboxServerGroupMembershipCreate(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*CompositeClient).APIClient
	group := d.Get("group").(string)
	serverList := sliceFromStringSet(d, "servers")

	object, err := client.AddServersToServerGroup(ctx, group, mapServerGroupMemberList(serverList))
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	d.SetId(id.UniqueId())
	return append(diags, setServerGroupMembershipAttributes(d, object)...)
}

func mapServerGroupMemberList(list []string) brightbox.ServerGroupMemberList {
	var result brightbox.ServerGroupMemberList
	for _, v := range list {
		result.Servers = append(result.Servers, brightbox.ServerGroupMember{Server: v})
	}
	return result
}

func setServerGroupMembershipAttributes(
	d *schema.ResourceData,
	serverGroup *brightbox.ServerGroup,
) diag.Diagnostics {
	var diags diag.Diagnostics
	d.Set("group", serverGroup.ID)
	serverList := d.Get("servers").(*schema.Set)
	var sl []string

	for _, server := range serverGroup.Servers {
		if serverList.Contains(server.ID) {
			sl = append(sl, server.ID)
		}
	}

	if err := d.Set("servers", sl); err != nil {
		return append(diags, diag.Errorf("setting server list from group (%s), error: %s", serverGroup.ID, err)...)
	}

	return diags
}

func resourceBrightboxServerGroupMembershipUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*CompositeClient).APIClient
	if d.HasChange("servers") {
		log.Printf("[DEBUG] ServerGroupMembership change detected, updating")
		group := d.Get("group").(string)

		o, n := d.GetChange("servers")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		remove := expandStringValueList(os.Difference(ns).List())
		log.Printf("[DEBUG] removing servers: %+v", remove)
		add := expandStringValueList(ns.Difference(os).List())
		log.Printf("[DEBUG] adding servers: %+v", add)

		var object *brightbox.ServerGroup
		var err error
		if len(remove) > 0 {
			object, err = client.RemoveServersFromServerGroup(ctx, group, mapServerGroupMemberList(remove))
			if err != nil {
				diags = append(diags, diag.FromErr(err)...)
			}
		}
		if len(add) > 0 {
			object, err = client.AddServersToServerGroup(ctx, group, mapServerGroupMemberList(add))
			if err != nil {
				diags = append(diags, diag.FromErr(err)...)
			}
		}
		return append(diags, setServerGroupMembershipAttributes(d, object)...)
	}
	return resourceBrightboxServerGroupMembershipRead(ctx, d, meta)
}

func resourceBrightboxServerGroupMembershipDelete(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*CompositeClient).APIClient
	group := d.Get("group").(string)
	serverList := sliceFromStringSet(d, "servers")

	_, err := client.RemoveServersFromServerGroup(ctx, group, mapServerGroupMemberList(serverList))
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	return diags
}

func resourceBrightboxServerGroupMembershipImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) < 2 {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected <group-name>/<server-name1>/...", d.Id())
	}

	userName := idParts[0]
	groupList := idParts[1:]

	d.Set("group", userName)
	d.Set("servers", groupList)

	d.SetId(id.UniqueId())

	return []*schema.ResourceData{d}, nil
}
