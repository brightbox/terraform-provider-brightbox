package brightbox

import (
	"context"
	"fmt"
	"log"
	"regexp"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceBrightboxServerGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Brightbox Server Group",
		ReadContext: dataSourceBrightboxServerGroupRead,

		Schema: map[string]*schema.Schema{
			"default": {
				Description: "Is this the default group for the account?",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"description": {
				Description: "User Description",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"fqdn": {
				Description: "Fully Qualified Domain Name",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"name": {
				Description: "User Label",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func dataSourceBrightboxServerGroupRead(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[DEBUG] Server Group data read called. Retrieving server group list")

	groups, err := client.ServerGroups(ctx)
	if err != nil {
		return diag.Errorf("Error retrieving server group list: %s", err)
	}

	group, err := findGroupByFilter(groups, d)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Single Server Group found: %s", group.ID)
	d.SetId(group.ID)
	return setServerGroupAttributes(d, group)
}

func findGroupByFilter(
	serverGroups []brightbox.ServerGroup,
	d *schema.ResourceData,
) (*brightbox.ServerGroup, error) {
	nameRe, err := regexp.Compile(d.Get("name").(string))
	if err != nil {
		return nil, err
	}

	descRe, err := regexp.Compile(d.Get("description").(string))
	if err != nil {
		return nil, err
	}

	var results []brightbox.ServerGroup
	for _, serverGroup := range serverGroups {
		if serverGroupMatch(&serverGroup, d, nameRe, descRe) {
			results = append(results, serverGroup)
		}
	}
	if len(results) == 1 {
		return &results[0], nil
	} else if len(results) > 1 {
		return nil, fmt.Errorf("Your query returned more than one result (found %d entries). Please try a more "+
			"specific search criteria.", len(results))
	} else {
		return nil, fmt.Errorf("Your query returned no results. " +
			"Please change your search criteria and try again.")
	}
}

//Match on the search filter - if the elements exist
func serverGroupMatch(
	serverGroup *brightbox.ServerGroup,
	d *schema.ResourceData,
	nameRe *regexp.Regexp,
	descRe *regexp.Regexp,
) bool {
	_, ok := d.GetOk("name")
	if ok && !nameRe.MatchString(serverGroup.Name) {
		return false
	}
	_, ok = d.GetOk("description")
	if ok && !descRe.MatchString(serverGroup.Description) {
		return false
	}
	return true
}
