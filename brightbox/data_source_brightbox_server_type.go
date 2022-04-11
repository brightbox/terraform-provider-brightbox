package brightbox

import (
	"regexp"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceBrightboxServerType() *schema.Resource {
	return &schema.Resource{
		Description: "Brightbox Cloud SQL server type",
		ReadContext: datasourceBrightboxRead(
			(*brightbox.Client).ServerTypes,
			"Server Type",
			dataSourceBrightboxServerTypesAttributes,
			findServerTypeFunc,
		),

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

func dataSourceBrightboxServerTypesAttributes(
	d *schema.ResourceData,
	serverType *brightbox.ServerType,
) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	d.SetId(serverType.ID)
	err = d.Set("name", serverType.Name)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("status", serverType.Status.String())
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("handle", serverType.Handle)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("cores", serverType.Cores)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("ram", serverType.RAM)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("disk_size", serverType.DiskSize)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("storage_type", serverType.StorageType.String())
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	return diags
}

func findServerTypeFunc(
	d *schema.ResourceData,
) (func(brightbox.ServerType) bool, diag.Diagnostics) {
	var nameRe, descRe *regexp.Regexp
	var err error
	var diags diag.Diagnostics
	if temp, ok := d.GetOk("name"); ok {
		if nameRe, err = regexp.Compile(temp.(string)); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	if temp, ok := d.GetOk("handle"); ok {
		if descRe, err = regexp.Compile(temp.(string)); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	return func(object brightbox.ServerType) bool {
		if nameRe != nil && !nameRe.MatchString(object.Name) {
			return false
		}
		if descRe != nil && !descRe.MatchString(object.Handle) {
			return false
		}
		return true
	}, diags
}
