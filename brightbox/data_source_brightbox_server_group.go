package brightbox

import (
	"fmt"
	"log"
	"regexp"

	brightbox "github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceBrightboxServerGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Brightbox Server Group",
		Read:        dataSourceBrightboxServerGroupRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "User Label",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"description": {
				Description: "User Description",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func dataSourceBrightboxServerGroupRead(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[DEBUG] Server Group data read called. Retrieving server group list")

	groups, err := client.ServerGroups()
	if err != nil {
		return fmt.Errorf("Error retrieving server group list: %s", err)
	}

	group, err := findGroupByFilter(groups, d)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Single Server Group found: %s", group.Id)
	d.SetId(group.Id)
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
