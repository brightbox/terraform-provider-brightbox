package brightbox

import (
	"context"
	"log"
	"regexp"
	"sort"
	"time"

	brightbox "github.com/brightbox/gobrightbox/v2"
	imageConst "github.com/brightbox/gobrightbox/v2/status/image"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	validImageStatusMap = map[imageConst.Status]bool{
		imageConst.Available:  true,
		imageConst.Deprecated: true,
	}
	validImageStatus         = []string{"available", "deprecated"}
	validImageArchitectures  = []string{"x86_64", "i686"}
	validImageSourceTypes    = []string{"upload", "snapshot"}
	validImageSourceTriggers = []string{"manual", "schedule"}
)

func dataSourceBrightboxImage() *schema.Resource {
	return &schema.Resource{
		Description: "Brightbox Image",
		ReadContext: dataSourceBrightboxImageRead,

		Schema: map[string]*schema.Schema{

			"ancestor_id": {
				Description: "Image this image was derived from",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"arch": {
				Description: "OS Architecture",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ValidateFunc: validation.StringInSlice(
					validImageArchitectures,
					false,
				),
			},

			"compatibility_mode": {
				Description: "Does this image require a non-virtio VM shell",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},

			"created_at": {
				Description: "The time this image was created/registered (UTC)",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"description": {
				Description: "A Description of the image",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"disk_size": {
				Description: "The actual size of the data within this image in Megabytes",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"licence_name": {
				Description: "The licence name for this image",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			//Locked is computed only because there is no 'null' search option
			"locked": {
				Description: "Is true if the image is set as locked and cannot be deleted",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"min_ram": {
				Description: "The actual size of the data within this image in Megabytes",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},

			"most_recent": {
				Description: "Select the most recent image",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},

			"name": {
				Description: "User Label for this image",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"official": {
				Description: "Is this image an official Brightbox provided one?",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},

			"owner": {
				Description: "Account ID this image belongs to",
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

			"source": {
				Description: "Name of theSource for this image",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"source_trigger": {
				Description: "Source trigger for this image (manual or schedule)",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ValidateFunc: validation.StringInSlice(
					validImageSourceTriggers,
					false,
				),
			},

			"source_type": {
				Description: "Source type for this image (upload or snapshot)",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ValidateFunc: validation.StringInSlice(
					validImageSourceTypes,
					false,
				),
			},

			"status": {
				Description: "State of the image",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ValidateFunc: validation.StringInSlice(
					validImageStatus,
					false,
				),
			},

			"username": {
				Description:  "Username to use when logging into a server booted with this image",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},

			"virtual_size": {
				Description: "The virtual size of the disk image container in Megabytes",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

func dataSourceBrightboxImageRead(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[DEBUG] Image data read called. Retrieving image list")

	images, err := client.Images(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	image, errs := findImageByFilter(images, d)

	if errs.HasError() {
		// Remove any existing image id on error
		d.SetId("")
		return errs
	}

	log.Printf("[DEBUG] Single Image found: %s", image.ID)

	return dataSourceBrightboxImagesImageAttributes(d, image)
}

func dataSourceBrightboxImagesImageAttributes(
	d *schema.ResourceData,
	image *brightbox.Image,
) diag.Diagnostics {
	log.Printf("[DEBUG] Image details: %#v", image)

	d.SetId(image.ID)
	d.Set("name", image.Name)
	d.Set("username", image.Username)
	d.Set("status", image.Status.String())
	d.Set("locked", image.Locked)
	d.Set("description", image.Description)
	d.Set("arch", image.Arch.String())
	d.Set("created_at", image.CreatedAt.Format(time.RFC3339))
	d.Set("official", image.Official)
	d.Set("public", image.Public)
	d.Set("owner", image.Owner)
	d.Set("source", image.Source)
	d.Set("source_trigger", image.SourceTrigger.String())
	d.Set("source_type", image.SourceType.String())
	d.Set("virtual_size", image.VirtualSize)
	d.Set("disk_size", image.DiskSize)
	d.Set("compatibility_mode", image.CompatibilityMode)
	d.Set("licence_name", image.LicenceName)
	if image.MinRAM != nil {
		d.Set("min_ram", image.MinRAM)
	}
	if image.Ancestor != nil {
		d.Set("ancestor_id", image.Ancestor.ID)
	}

	return nil
}

func findImageByFilter(
	images []brightbox.Image,
	d *schema.ResourceData,
) (*brightbox.Image, diag.Diagnostics) {
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
		}
		return nil, diag.Errorf("Your query returned more than one result (found %d entries). Please try a more "+
			"specific search criteria, or set `most_recent` attribute to true.", len(results))
	}
	return nil, diag.Errorf("Your query returned no results. " +
		"Please change your search criteria and try again.")
}

//Match on the search filter - if the elements exist
func imageMatch(
	image *brightbox.Image,
	d *schema.ResourceData,
	nameRe *regexp.Regexp,
	descRe *regexp.Regexp,
) bool {
	// Only check available images
	if !validImageStatusMap[image.Status] {
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
	source, ok := d.GetOk("source")
	if ok && source.(string) != image.Source {
		return false
	}
	sourceTrigger, ok := d.GetOk("source_trigger")
	if ok && sourceTrigger.(string) != image.SourceTrigger.String() {
		return false
	}
	sourceType, ok := d.GetOk("source_type")
	if ok && sourceType.(string) != image.SourceType.String() {
		return false
	}
	status, ok := d.GetOk("status")
	if ok && status.(string) != image.Status.String() {
		return false
	}
	owner, ok := d.GetOk("owner")
	if ok && owner.(string) != image.Owner {
		return false
	}
	username, ok := d.GetOk("username")
	if ok && username.(string) != image.Username {
		return false
	}
	licenceName, ok := d.GetOk("licence_name")
	if ok && licenceName.(string) != image.LicenceName {
		return false
	}
	arch, ok := d.GetOk("arch")
	if ok && arch.(string) != image.Arch.String() {
		return false
	}
	if image.Ancestor != nil {
		ancestorID, ok := d.GetOk("ancestor_id")
		if ok && ancestorID.(string) != image.Ancestor.ID {
			return false
		}
	}
	if image.MinRAM != nil {
		minRAM, ok := d.GetOk("min_ram")
		if ok && minRAM.(int) != *image.MinRAM {
			return false
		}
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

// Returns the most recent Image out of a slice of images
func mostRecentImage(images []brightbox.Image) *brightbox.Image {
	sortedImages := images
	sort.Slice(sortedImages, func(i, j int) bool {
		itime := sortedImages[i].CreatedAt
		jtime := sortedImages[j].CreatedAt
		return itime.Unix() > jtime.Unix()
	})
	return &sortedImages[0]
}
