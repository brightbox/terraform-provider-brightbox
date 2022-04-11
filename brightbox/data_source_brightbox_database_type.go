package brightbox

import (
	"regexp"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceBrightboxDatabaseType() *schema.Resource {
	return &schema.Resource{
		Description: "Brightbox Cloud SQL Database Type",
		ReadContext: datasourceBrightboxRead(
			(*brightbox.Client).DatabaseServerTypes,
			"Database Type",
			dataSourceBrightboxDatabaseTypesAttributes,
			findDatabaseServerTypeFunc,
		),

		Schema: map[string]*schema.Schema{

			"default": {
				Description: "Is this the default database type",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"description": {
				Description: "Description of this database type",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"disk_size": {
				Description: "Disk size in megabytes",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"name": {
				Description: "Name of this database type",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"ram": {
				Description: "RAM size in megabytes",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

func dataSourceBrightboxDatabaseTypesAttributes(
	d *schema.ResourceData,
	databaseType *brightbox.DatabaseServerType,
) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	d.SetId(databaseType.ID)
	err = d.Set("name", databaseType.Name)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("description", databaseType.Description)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("default", databaseType.Default)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("disk_size", databaseType.DiskSize)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("ram", databaseType.RAM)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}

	return nil
}

func findDatabaseServerTypeFunc(
	d *schema.ResourceData,
) (func(brightbox.DatabaseServerType) bool, diag.Diagnostics) {
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

	return func(object brightbox.DatabaseServerType) bool {
		if nameRe != nil && !nameRe.MatchString(object.Name) {
			return false
		}
		if descRe != nil && !descRe.MatchString(object.Description) {
			return false
		}
		return true
	}, diags
}
