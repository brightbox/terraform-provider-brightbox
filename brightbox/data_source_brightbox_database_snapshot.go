package brightbox

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"time"

	brightbox "github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceBrightboxDatabaseSnapshot() *schema.Resource {
	return &schema.Resource{
		Description: "Brightbox Database Snapshot",
		Read:        dataSourceBrightboxDatabaseSnapshotRead,

		Schema: map[string]*schema.Schema{

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

			"description": {
				Description: "Editable user label",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"database_version": {
				Description: "The version of the database engine used to create this snapshot",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"database_engine": {
				Description: "The engine of the database used to create this snapshot",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			//Computed Values
			"status": {
				Description: "Snapshot state",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"size": {
				Description: "Size of database partition in megabytes",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"created_at": {
				Description: "Time of resource creation (UTC)",
				Type:        schema.TypeString,
				Computed:    true,
			},

			//Locked is computed only because there is no 'null' search option
			"locked": {
				Description: "True if snapshot is locked and cannot be deleted",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

func dataSourceBrightboxDatabaseSnapshotRead(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[DEBUG] Snapshot data read called. Retrieving snapshot list")
	images, err := client.DatabaseSnapshots()
	if err != nil {
		return fmt.Errorf("Error retrieving snapshot list: %s", err)
	}

	image, err := findSnapshotByFilter(images, d)

	if err != nil {
		// Remove any existing image on error
		d.SetId("")
		return err
	}

	log.Printf("[DEBUG] Single Snapshot found: %s", image.Id)
	return dataSourceBrightboxDatabaseSnapshotsImageAttributes(d, image)
}

func dataSourceBrightboxDatabaseSnapshotsImageAttributes(
	d *schema.ResourceData,
	image *brightbox.DatabaseSnapshot,
) error {
	log.Printf("[DEBUG] Database Snapshot details: %#v", image)

	d.SetId(image.Id)
	d.Set("name", image.Name)
	d.Set("description", image.Description)
	d.Set("status", image.Status)
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
) (*brightbox.DatabaseSnapshot, error) {
	nameRe, err := regexp.Compile(d.Get("name").(string))
	if err != nil {
		return nil, err
	}

	descRe, err := regexp.Compile(d.Get("description").(string))
	if err != nil {
		return nil, err
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
			return mostRecentSnapshot(results), nil
		}
		return nil, fmt.Errorf("Your query returned more than one result (found %d entries). Please try a more "+
			"specific search criteria, or set `most_recent` attribute to true.", len(results))
	} else {
		return nil, fmt.Errorf("Your query returned no results. " +
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
	if !validImageStatus[image.Status] {
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

type snapshotSort []brightbox.DatabaseSnapshot

func (a snapshotSort) Len() int      { return len(a) }
func (a snapshotSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a snapshotSort) Less(i, j int) bool {
	itime := a[i].CreatedAt
	jtime := a[j].CreatedAt
	return itime.Unix() < jtime.Unix()
}

// Returns the most recent snapshot out of a slice of snapshots
func mostRecentSnapshot(images []brightbox.DatabaseSnapshot) *brightbox.DatabaseSnapshot {
	sortedImages := images
	sort.Sort(snapshotSort(sortedImages))
	return &sortedImages[len(sortedImages)-1]
}
