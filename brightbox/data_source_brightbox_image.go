package brightbox

import (
	"log"
	"regexp"
	"time"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/brightbox/gobrightbox/v2/status/arch"
	imageConst "github.com/brightbox/gobrightbox/v2/status/image"
	"github.com/brightbox/gobrightbox/v2/status/sourcetrigger"
	"github.com/brightbox/gobrightbox/v2/status/sourcetype"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	validImageStatusMap = map[imageConst.Status]bool{
		imageConst.Available:  true,
		imageConst.Deprecated: true,
	}
)

func dataSourceBrightboxImage() *schema.Resource {
	return &schema.Resource{
		Description: "Brightbox Image",
		ReadContext: datasourceBrightboxRecentRead(
			(*brightbox.Client).Images,
			"Image",
			dataSourceBrightboxImagesImageAttributes,
			findImageFunc,
		),

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
					arch.ValidStrings,
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
				Description:  "The actual size of the data within this image in Megabytes",
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
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
					sourcetrigger.ValidStrings,
					false,
				),
			},

			"source_type": {
				Description: "Source type for this image (upload or snapshot)",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ValidateFunc: validation.StringInSlice(
					sourcetype.ValidStrings,
					false,
				),
			},

			"status": {
				Description: "State of the image",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ValidateFunc: validation.StringInSlice(
					imageConst.ValidStrings,
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

func dataSourceBrightboxImagesImageAttributes(
	d *schema.ResourceData,
	image *brightbox.Image,
) diag.Diagnostics {
	log.Printf("[DEBUG] Image details: %#v", image)

	var diags diag.Diagnostics
	var err error

	d.SetId(image.ID)
	err = d.Set("name", image.Name)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("username", image.Username)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("status", image.Status.String())
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("locked", image.Locked)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("description", image.Description)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("arch", image.Arch.String())
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("created_at", image.CreatedAt.Format(time.RFC3339))
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("official", image.Official)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("public", image.Public)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("owner", image.Owner)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("source", image.Source)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("source_trigger", image.SourceTrigger.String())
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("source_type", image.SourceType.String())
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("virtual_size", image.VirtualSize)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("disk_size", image.DiskSize)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("compatibility_mode", image.CompatibilityMode)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("licence_name", image.LicenceName)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	if image.MinRAM != nil {
		err = d.Set("min_ram", image.MinRAM)
		if err != nil {
			diags = append(diags, diag.Errorf("unexpected: %s", err)...)
		}
	}
	if image.Ancestor != nil {
		err = d.Set("ancestor_id", image.Ancestor.ID)
		if err != nil {
			diags = append(diags, diag.Errorf("unexpected: %s", err)...)
		}
	}
	return diags
}

func findImageFunc(
	d *schema.ResourceData,
) (func(brightbox.Image) bool, diag.Diagnostics) {
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
	source, sourceok := d.GetOk("source")
	sourceTrigger, sourceTriggerok := d.GetOk("source_trigger")
	sourceType, sourceTypeok := d.GetOk("source_type")
	status, statusok := d.GetOk("status")
	owner, ownerok := d.GetOk("owner")
	username, usernameok := d.GetOk("username")
	licenceName, licenceNameok := d.GetOk("licence_name")
	arch, archok := d.GetOk("arch")
	ancestorID, ancestorIDok := d.GetOk("ancestor_id")
	minRAM, minRAMok := d.GetOk("min_ram")
	public, publicok := d.GetOk("public")
	official, officialok := d.GetOk("official")
	compat := d.Get("compatibility_mode")
	return func(image brightbox.Image) bool {
		// Only check available images
		if !validImageStatusMap[image.Status] {
			return false
		}
		if nameRe != nil && !nameRe.MatchString(image.Name) {
			return false
		}
		if descRe != nil && !descRe.MatchString(image.Description) {
			return false
		}
		if sourceok && source.(string) != image.Source {
			return false
		}
		if sourceTriggerok && sourceTrigger.(string) != image.SourceTrigger.String() {
			return false
		}
		if sourceTypeok && sourceType.(string) != image.SourceType.String() {
			return false
		}
		if statusok && status.(string) != image.Status.String() {
			return false
		}
		if ownerok && owner.(string) != image.Owner {
			return false
		}
		if usernameok && username.(string) != image.Username {
			return false
		}
		if licenceNameok && licenceName.(string) != image.LicenceName {
			return false
		}
		if archok && arch.(string) != image.Arch.String() {
			return false
		}
		if image.Ancestor != nil {
			if ancestorIDok && ancestorID.(string) != image.Ancestor.ID {
				return false
			}
		}
		if image.MinRAM != nil {
			if minRAMok && minRAM.(uint) != *image.MinRAM {
				return false
			}
		}
		// Binary choices are treated as Yes/Not bothered
		// due to false being treated by Terraform as null
		if publicok && public.(bool) != image.Public {
			return false
		}
		if officialok && official.(bool) != image.Official {
			return false
		}
		//Make Compatibility mode Yes/No, rather than Yes/Not bothered
		if compat.(bool) != image.CompatibilityMode {
			return false
		}
		return true
	}, diags
}
