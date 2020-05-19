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

var (
	validImageStatus = map[string]bool{
		"available":  true,
		"deprecated": true,
	}
)

func dataSourceBrightboxImage() *schema.Resource {
	return &schema.Resource{
		Description: "Brightbox Image",
		Read:        dataSourceBrightboxImageRead,

		Schema: map[string]*schema.Schema{

			"most_recent": {
				Description: "Select the most recent image",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},

			"source_type": {
				Description: "Source type for this image (upload or snapshot)",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"owner": {
				Description: "Account ID this image belongs to",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"status": {
				Description: "State of the image",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"arch": {
				Description: "OS Architecture",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"name": {
				Description: "User Label for this image",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"description": {
				Description: "A Description of the image",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"public": {
				Description: "Is this image available to other customers?",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},

			"official": {
				Description: "Is this image an official Brightbox provided one?",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},

			"compatibility_mode": {
				Description: "Does this image require a non-virtio VM shell",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},

			"username": {
				Description: "Username to use when logging into a server booted with this image",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"ancestor_id": {
				Description: "Image this image was derived from",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"licence_name": {
				Description: "The licence name for this image",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			//Computed Values
			"created_at": {
				Description: "The time this image was created/registered (UTC)",
				Type:        schema.TypeString,
				Computed:    true,
			},

			//Locked is computed only because there is no 'null' search option
			"locked": {
				Description: "Is true if the image is set as locked and cannot be deleted",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"virtual_size": {
				Description: "The virtual size of the disk image container in Megabytes",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"disk_size": {
				Description: "The actual size of the data within this image in Megabytes",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

func dataSourceBrightboxImageRead(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[DEBUG] Image data read called. Retrieving image list")

	images, err := client.Images()
	if err != nil {
		return fmt.Errorf("Error retrieving image list: %s", err)
	}

	image, err := findImageByFilter(images, d)

	if err != nil {
		// Remove any existing image id on error
		d.SetId("")
		return err
	}

	log.Printf("[DEBUG] Single Image found: %s", image.Id)
	return dataSourceBrightboxImagesImageAttributes(d, image)
}

func dataSourceBrightboxImagesImageAttributes(
	d *schema.ResourceData,
	image *brightbox.Image,
) error {
	log.Printf("[DEBUG] Image details: %#v", image)

	d.SetId(image.Id)
	d.Set("name", image.Name)
	d.Set("username", image.Username)
	d.Set("status", image.Status)
	d.Set("locked", image.Locked)
	d.Set("description", image.Description)
	d.Set("source", image.Source)
	d.Set("arch", image.Arch)
	d.Set("created_at", image.CreatedAt.Format(time.RFC3339))
	d.Set("official", image.Official)
	d.Set("public", image.Public)
	d.Set("owner", image.Owner)
	d.Set("source_type", image.SourceType)
	d.Set("virtual_size", image.VirtualSize)
	d.Set("disk_size", image.DiskSize)
	d.Set("compatibility_mode", image.CompatibilityMode)
	d.Set("ancestor_id", image.AncestorId)
	d.Set("licence_name", image.LicenceName)

	return nil
}

func findImageByFilter(
	images []brightbox.Image,
	d *schema.ResourceData,
) (*brightbox.Image, error) {
	nameRe, err := regexp.Compile(d.Get("name").(string))
	if err != nil {
		return nil, err
	}

	descRe, err := regexp.Compile(d.Get("description").(string))
	if err != nil {
		return nil, err
	}

	var results []brightbox.Image
	for _, image := range images {
		if imageMatch(&image, d, nameRe, descRe) {
			results = append(results, image)
		}
	}
	if len(results) == 1 {
		return &results[0], nil
	} else if len(results) > 1 {
		recent := d.Get("most_recent").(bool)
		log.Printf("[DEBUG] Multiple results found and `most_recent` is set to: %t", recent)
		if recent {
			return mostRecentImage(results), nil
		} else {
			return nil, fmt.Errorf("Your query returned more than one result (found %d entries). Please try a more "+
				"specific search criteria, or set `most_recent` attribute to true.", len(results))
		}
	} else {
		return nil, fmt.Errorf("Your query returned no results. " +
			"Please change your search criteria and try again.")
	}
}

//Match on the search filter - if the elements exist
func imageMatch(
	image *brightbox.Image,
	d *schema.ResourceData,
	nameRe *regexp.Regexp,
	descRe *regexp.Regexp,
) bool {
	// Only check available images
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
	source_type, ok := d.GetOk("source_type")
	if ok && source_type.(string) != image.SourceType {
		return false
	}
	status, ok := d.GetOk("status")
	if ok && status.(string) != image.Status {
		return false
	}
	owner, ok := d.GetOk("owner")
	if ok && owner.(string) != image.Owner {
		return false
	}
	arch, ok := d.GetOk("arch")
	if ok && arch.(string) != image.Arch {
		return false
	}
	// Binary choices are treated as Yes/Not bothered
	// due to false being treated by Terraform as null
	public, ok := d.GetOk("public")
	if ok && public.(bool) != image.Public {
		return false
	}
	official, ok := d.GetOk("official")
	if ok && official.(bool) != image.Official {
		return false
	}
	//Make Compatibility mode Yes/No, rather than Yes/Not bothered
	if d.Get("compatibility_mode").(bool) != image.CompatibilityMode {
		return false
	}
	return true
}

type imageSort []brightbox.Image

func (a imageSort) Len() int      { return len(a) }
func (a imageSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a imageSort) Less(i, j int) bool {
	itime := a[i].CreatedAt
	jtime := a[j].CreatedAt
	return itime.Unix() < jtime.Unix()
}

// Returns the most recent Image out of a slice of images
func mostRecentImage(images []brightbox.Image) *brightbox.Image {
	sortedImages := images
	sort.Sort(imageSort(sortedImages))
	return &sortedImages[len(sortedImages)-1]
}
