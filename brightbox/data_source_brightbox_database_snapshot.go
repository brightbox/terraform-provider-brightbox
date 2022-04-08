package brightbox

import (
	"context"
	"log"
	"regexp"
	"time"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/brightbox/gobrightbox/v2/status/databasesnapshot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceBrightboxDatabaseSnapshot() *schema.Resource {
	return &schema.Resource{
		Description: "Brightbox Database Snapshot",
		ReadContext: dataSourceBrightboxDatabaseSnapshotRead,

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

func dataSourceBrightboxDatabaseSnapshotRead(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[DEBUG] Snapshot data read called. Retrieving snapshot list")
	images, err := client.DatabaseSnapshots(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	image, errs := findSnapshotByFilter(images, d)

	if errs.HasError() {
		// Remove any existing image on error
		d.SetId("")
		return errs
	}

	log.Printf("[DEBUG] Single Snapshot found: %s", image.ID)
	return dataSourceBrightboxDatabaseSnapshotsImageAttributes(d, image)
}

func dataSourceBrightboxDatabaseSnapshotsImageAttributes(
	d *schema.ResourceData,
	image *brightbox.DatabaseSnapshot,
) diag.Diagnostics {
	log.Printf("[DEBUG] Database Snapshot details: %#v", image)

	d.SetId(image.ID)
	d.Set("name", image.Name)
	d.Set("description", image.Description)
	d.Set("status", image.Status.String())
	d.Set("database_engine", image.DatabaseEngine)
	d.Set("database_version", image.DatabaseVersion)
	d.Set("size", image.Size)
	d.Set("created_at", image.CreatedAt.Format(time.RFC3339))
	d.Set("locked", image.Locked)

	return nil
}

func findSnapshotByFilter(
	images []brightbox.DatabaseSnapshot,
	d *schema.ResourceData,
) (*brightbox.DatabaseSnapshot, diag.Diagnostics) {
	var diags diag.Diagnostics
	nameRe, err := regexp.Compile(d.Get("name").(string))
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	descRe, err := regexp.Compile(d.Get("description").(string))
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	if diags.HasError() {
		return nil, diags
	}

	var results []brightbox.DatabaseSnapshot
	for _, image := range images {
		if snapshotMatch(&image, d, nameRe, descRe) {
			results = append(results, image)
		}
	}
	if len(results) == 1 {
		return &results[0], nil
	} else if len(results) > 1 {
		recent := d.Get("most_recent").(bool)
		log.Printf("[DEBUG] Multiple results found and `most_recent` is set to: %t", recent)
		if recent {
			return mostRecent(results), nil
		}
		return nil, diag.Errorf("Your query returned more than one result (found %d entries). Please try a more "+
			"specific search criteria, or set `most_recent` attribute to true.", len(results))
	} else {
		return nil, diag.Errorf("Your query returned no results. " +
			"Please change your search criteria and try again.")
	}
}

//Match on the search filter - if the elements exist
func snapshotMatch(
	image *brightbox.DatabaseSnapshot,
	d *schema.ResourceData,
	nameRe *regexp.Regexp,
	descRe *regexp.Regexp,
) bool {
	// Only check available snapshots
	if image.Status != databasesnapshot.Available {
		return false
	}
	_, ok := d.GetOk("name")
	if ok && !nameRe.MatchString(image.Name) {
		return false
	}
	_, ok = d.GetOk("description")
	if ok && !descRe.MatchString(image.Description) {
		return false
	}
	engine, ok := d.GetOk("database_engine")
	if ok && engine.(string) != image.DatabaseEngine {
		return false
	}
	version, ok := d.GetOk("database_version")
	if ok && version.(string) != image.DatabaseVersion {
		return false
	}
	return true
}
