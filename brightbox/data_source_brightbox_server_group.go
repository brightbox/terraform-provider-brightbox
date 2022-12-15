package brightbox

import (
	"regexp"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceBrightboxServerGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Brightbox Server Group",
		ReadContext: datasourceBrightboxRead(
			(*brightbox.Client).ServerGroups,
			"Server Group",
			setServerGroupAttributes,
			findServerGroupFunc,
		),

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
			"firewall_policy": {
				Description: "The firewall policy associated with this server group",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func findServerGroupFunc(
	d *schema.ResourceData,
) (func(brightbox.ServerGroup) bool, diag.Diagnostics) {
	var nameRe, descRe *regexp.Regexp
	var err error
	var diags diag.Diagnostics
	if temp, ok := d.GetOk("name"); ok {
		if nameRe, err = regexp.Compile(temp.(string)); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	if temp, ok := d.GetOk("description"); ok {
		if descRe, err = regexp.Compile(temp.(string)); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	return func(object brightbox.ServerGroup) bool {
		if nameRe != nil && !nameRe.MatchString(object.Name) {
			return false
		}
		if descRe != nil && !descRe.MatchString(object.Description) {
			return false
		}
		return true
	}, diags
}
