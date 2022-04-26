package brightbox

import (
	"context"
	"log"
	"regexp"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/brightbox/gobrightbox/v2/status/filesystemtype"
	volumeConst "github.com/brightbox/gobrightbox/v2/status/volume"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceBrightboxVolume() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Brightbox Volume resource",
		CreateContext: resourceBrightboxVolumeCreateAndWait,
		ReadContext:   resourceBrightboxVolumeRead,
		UpdateContext: resourceBrightboxVolumeUpdateAndResize,
		DeleteContext: resourceBrightboxVolumeDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultTimeout),
			Delete: schema.DefaultTimeout(defaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Description:  "Verbose Description of this volume",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},

			"encrypted": {
				Description: "Is true if the volume is encrypted",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},

			"filesystem_label": {
				Description:  "Label given to the filesystem on the volume",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[\w-]{0,12}$`), "must be a valid filesystem label"),
			},

			"filesystem_type": {
				Description: "Format of the filesystem on the volume",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				ValidateFunc: validation.StringInSlice(
					filesystemtype.ValidStrings,
					false),
			},

			"image": {
				Description:  "Image used to create the volume",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(imageRegexp, "must be a valid image ID"),
			},

			"locked": {
				Description: "Is true if the image is set as locked and cannot be deleted",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},

			"name": {
				Description:  "Human Readable Name",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},

			"serial": {
				Description:  "Volume Serial Number",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[\w.+-]{0,20}$`), "must be a valid serial number"),
			},

			"size": {
				Description:  "Disk size in megabytes",
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},

			"source": {
				Description:  "ID of the source volume for this image",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(volumeRegexp, "must be a valid volume ID"),
			},

			"source_type": {
				Description: "Source type for this image (image, volume or raw)",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"status": {
				Description: "Current state of volume",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"storage_type": {
				Description: "Storage type for this volume (local or network)",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

var (
	resourceBrightboxVolumeRead = resourceBrightboxReadStatus(
		(*brightbox.Client).Volume,
		"Volume",
		setVolumeAttributes,
		volumeUnavailable,
	)

	resourceBrightboxVolumeUpdate = resourceBrightboxUpdateWithLock(
		(*brightbox.Client).UpdateVolume,
		"Volume",
		volumeFromID,
		addUpdateableVolumeOptions,
		setVolumeAttributes,
		resourceBrightboxSetVolumeLockState,
	)

	resourceBrightboxVolumeDelete = resourceBrightboxDelete(
		(*brightbox.Client).DestroyVolume,
		"Volume",
	)

	resourceBrightboxSetVolumeLockState = resourceBrightboxSetLockState(
		(*brightbox.Client).LockVolume,
		(*brightbox.Client).UnlockVolume,
		setVolumeAttributes,
	)
)

func volumeFromID(id string) *brightbox.VolumeOptions {
	return &brightbox.VolumeOptions{
		ID: id,
	}
}

func addUpdateableVolumeOptions(
	d *schema.ResourceData,
	opts *brightbox.VolumeOptions,
) diag.Diagnostics {
	assignString(d, &opts.Name, "name")
	assignString(d, &opts.Description, "description")
	assignString(d, &opts.Serial, "serial")
	return nil
}

func addVolumeCreateOptions(
	d *schema.ResourceData,
	opts *brightbox.VolumeOptions,
) diag.Diagnostics {
	assignString(d, &opts.Image, "image")
	assignString(d, &opts.FilesystemLabel, "filesystem_label")
	assignEnum(d, &opts.FilesystemType, "filesystem_type")
	assignInt(d, &opts.Size, "size")
	assignBool(d, &opts.Encrypted, "encrypted")
	return nil
}

func setVolumeAttributes(
	d *schema.ResourceData,
	volume *brightbox.Volume,
) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	d.SetId(volume.ID)
	err = d.Set("name", volume.Name)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("description", volume.Description)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("status", volume.Status.String())
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("locked", volume.Locked)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("encrypted", volume.Encrypted)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("filesystem_label", volume.FilesystemLabel)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("filesystem_type", volume.FilesystemType.String())
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("serial", volume.Serial)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("size", volume.Size)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("source", volume.Source)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("storage_type", volume.StorageType.String())
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("source_type", volume.SourceType.String())
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	if volume.Image != nil {
		err = d.Set("image", volume.Image.ID)
		if err != nil {
			diags = append(diags, diag.Errorf("unexpected: %s", err)...)
		}
	}
	return diags
}

func volumeUnavailable(obj *brightbox.Volume) bool {
	return obj.Status == volumeConst.Deleted ||
		obj.Status == volumeConst.Deleting ||
		obj.Status == volumeConst.Failed
}

func volumeStateRefresh(client *brightbox.Client, ctx context.Context, volumeID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		volume, err := client.Volume(ctx, volumeID)
		if err != nil {
			log.Printf("Error on Volume State Refresh: %s", err)
			return nil, "", err
		}
		return volume, volume.Status.String(), nil
	}
}

func resourceBrightboxVolumeCreateAndWait(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[INFO] Creating volume")
	var volumeOpts brightbox.VolumeOptions

	var diags diag.Diagnostics
	diags = append(diags, addUpdateableVolumeOptions(d, &volumeOpts)...)
	diags = append(diags, addVolumeCreateOptions(d, &volumeOpts)...)
	if diags.HasError() {
		return diags
	}

	log.Printf("[DEBUG] volume create configuration: %+v", volumeOpts)

	object, err := client.CreateVolume(ctx, volumeOpts)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(object.ID)

	log.Printf("[INFO] Waiting for Volume (%s) to become available", d.Id())

	stateConf := resource.StateChangeConf{
		Pending: []string{
			volumeConst.Creating.String(),
		},
		Target: []string{
			volumeConst.Detached.String(),
		},
		Refresh:    volumeStateRefresh(client, ctx, object.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      checkDelay,
		MinTimeout: minimumRefreshWait,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceBrightboxSetVolumeLockState(ctx, d, meta)
}

func resourceBrightboxVolumeUpdateAndResize(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("[DEBUG] Checking if volume size has changed")
	if d.HasChange("size") {
		diags = append(diags, resizeBrightboxVolume(ctx, d, meta, d.Id(), "size")...)
	}
	log.Printf("[DEBUG] Updating volume")
	return append(diags, resourceBrightboxVolumeUpdate(ctx, d, meta)...)
}

func resizeBrightboxVolume(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
	volumeID string,
	sizeAttr string,
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient
	var diags diag.Diagnostics

	log.Printf("[DEBUG] Checking size change parameters")
	oldSize, newSize := d.GetChange(sizeAttr)
	oldSizeInt, ok := oldSize.(int)
	if !ok {
		diags = append(diags, diag.Errorf("expected type of old disk size to be Integer")...)
	}
	newSizeInt, ok := newSize.(int)
	if !ok {
		diags = append(diags, diag.Errorf("expected type of new disk size to be Integer")...)
	}
	if oldSizeInt > newSizeInt {
		diags = append(diags, diag.Errorf("expected new disk size (%v) to be bigger than old disk size (%v)", newSizeInt, oldSizeInt)...)

	}
	if diags.HasError() {
		return diags
	}
	log.Printf("[INFO] Resizing volume %v from %v to %v", volumeID, oldSizeInt, newSizeInt)
	_, err := client.ResizeVolume(
		ctx,
		volumeID,
		brightbox.VolumeNewSize{
			From: uint(oldSizeInt),
			To:   uint(newSizeInt),
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}
