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

func dataSourceBrightboxServerType() *schema.Resource {
	return &schema.Resource{
		Description: "Brightbox Cloud SQL server type",
		ReadContext: dataSourceBrightboxServerTypeRead,

		Schema: map[string]*schema.Schema{

			"cores": {
				Description: "Number of CPU Cores",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"disk_size": {
				Description: "Disk size in megabytes",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"handle": {
				Description: "Unique handle for this server type",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"name": {
				Description: "Name of this server type",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"ram": {
				Description: "RAM size in megabytes",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"status": {
				Description: "The state of this server type",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"storage_type": {
				Description: "If the server type uses local or network storage",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceBrightboxServerTypeRead(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[DEBUG] ServerType data read called. Retrieving server type list")

	serverTypes, err := client.ServerTypes(ctx)
	if err != nil {
		return diag.Errorf("Error retrieving server type list: %s", err)
	}

	serverType, err := findServerTypeByFilter(serverTypes, d)

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Single ServerType found: %s", serverType.ID)
	return dataSourceBrightboxServerTypesAttributes(d, serverType)
}

func dataSourceBrightboxServerTypesAttributes(
	d *schema.ResourceData,
	serverType *brightbox.ServerType,
) diag.Diagnostics {
	log.Printf("[DEBUG] serverType details: %#v", serverType)

	d.SetId(serverType.ID)
	d.Set("name", serverType.Name)
	d.Set("status", serverType.Status.String())
	d.Set("handle", serverType.Handle)
	d.Set("cores", serverType.Cores)
	d.Set("ram", serverType.RAM)
	d.Set("disk_size", serverType.DiskSize)
	d.Set("storage_type", serverType.StorageType.String())

	return nil
}

func findServerTypeByFilter(
	serverTypes []brightbox.ServerType,
	d *schema.ResourceData,
) (*brightbox.ServerType, error) {
	nameRe, err := regexp.Compile(d.Get("name").(string))
	if err != nil {
		return nil, err
	}

	descRe, err := regexp.Compile(d.Get("handle").(string))
	if err != nil {
		return nil, err
	}

	var results []brightbox.ServerType
	for _, serverType := range serverTypes {
		if serverTypeMatch(&serverType, d, nameRe, descRe) {
			results = append(results, serverType)
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
func serverTypeMatch(
	serverType *brightbox.ServerType,
	d *schema.ResourceData,
	nameRe *regexp.Regexp,
	descRe *regexp.Regexp,
) bool {
	_, ok := d.GetOk("name")
	if ok && !nameRe.MatchString(serverType.Name) {
		return false
	}
	_, ok = d.GetOk("handle")
	if ok && !descRe.MatchString(serverType.Handle) {
		return false
	}
	return true
}
