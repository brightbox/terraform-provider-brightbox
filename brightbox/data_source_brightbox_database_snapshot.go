package brightbox

import (
	"regexp"
	"time"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceBrightboxDatabaseSnapshot() *schema.Resource {
	return &schema.Resource{
		Description: "Brightbox Database Snapshot",
		ReadContext: datasourceBrightboxRecentRead(
			(*brightbox.Client).DatabaseSnapshots,
			"Database Snapshot",
			dataSourceBrightboxDatabaseSnapshotsImageAttributes,
			findDatabaseSnapshotFunc,
		),

		Schema: map[string]*schema.Schema{

			"created_at": {
				Description: "Time of resource creation (UTC)",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"database_engine": {
				Description: "The engine of the database used to create this snapshot",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ValidateFunc: validation.StringInSlice(
					validDatabaseEngines,
					false,
				),
			},

			"database_version": {
				Description:  "The version of the database engine used to create this snapshot",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},

			"description": {
				Description: "Editable user label",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			//Locked is computed only because there is no 'null' search option
			"locked": {
				Description: "True if snapshot is locked and cannot be deleted",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"most_recent": {
				Description: "Snapshot with the latest 'created_at' time",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},

			"name": {
				Description: "Editable user label",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"size": {
				Description: "Size of database partition in megabytes",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"status": {
				Description: "Snapshot state",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceBrightboxDatabaseSnapshotsImageAttributes(
	d *schema.ResourceData,
	image *brightbox.DatabaseSnapshot,
) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	d.SetId(image.ID)
	err = d.Set("name", image.Name)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("description", image.Description)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("status", image.Status.String())
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("database_engine", image.DatabaseEngine)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("database_version", image.DatabaseVersion)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("size", image.Size)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("created_at", image.CreatedAt.Format(time.RFC3339))
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("locked", image.Locked)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}

	return diags
}

func findDatabaseSnapshotFunc(
	d *schema.ResourceData,
) (func(brightbox.DatabaseSnapshot) bool, diag.Diagnostics) {
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

	return func(object brightbox.DatabaseSnapshot) bool {
		if nameRe != nil && !nameRe.MatchString(object.Name) {
			return false
		}
		if descRe != nil && !descRe.MatchString(object.Description) {
			return false
		}
		return true
	}, diags
}
