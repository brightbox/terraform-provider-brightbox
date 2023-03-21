package brightbox

import (
	"context"
	"log"
	"regexp"
	"time"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/brightbox/gobrightbox/v2/enums/filesystemtype"
	"github.com/brightbox/gobrightbox/v2/enums/volumestatus"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceBrightboxVolume() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Brightbox Volume resource",
		CreateContext: resourceBrightboxVolumeCreateAndWait,
		ReadContext:   resourceBrightboxVolumeRead,
		UpdateContext: resourceBrightboxVolumeUpdateAndResize,
		DeleteContext: resourceBrightboxVolumeDetachAndDelete,
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
				RequiredWith: []string{"filesystem_type"},
			},

			"filesystem_type": {
				Description: "Format of the filesystem on the volume",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				ValidateFunc: validation.StringInSlice(
					filesystemtype.ValidStrings,
					false),
				ConflictsWith: []string{"image", "source"},
			},

			"image": {
				Description:   "Image used to create the volume",
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  validation.StringMatch(imageRegexp, "must be a valid image ID"),
				ConflictsWith: []string{"filesystem_type", "source"},
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

			"server": {
				Description: "ID of the server this volume is to be mapped to",
				Type:        schema.TypeString,
				Optional:    true,
				ValidateFunc: validation.Any(
					validation.StringMatch(serverRegexp, "must be a valid server ID"),
					validation.StringIsEmpty,
				),
			},

			"size": {
				Description:  "Disk size in megabytes",
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},

			"source": {
				Description:   "ID of the source volume for this image",
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  validation.StringMatch(volumeRegexp, "must be a valid volume ID"),
				ConflictsWith: []string{"filesystem_type", "image"},
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

func resourceBrightboxVolumeDetachAndDelete(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	err := detachVolume(ctx, d, meta, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		log.Printf("[INFO] Detach on delete issue: %s", err)
	}
	return resourceBrightboxVolumeDelete(ctx, d, meta)
}

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
	if volume.Image == nil {
		d.Set("image", "")
	} else {
		err = d.Set("image", volume.Image.ID)
		if err != nil {
			diags = append(diags, diag.Errorf("unexpected: %s", err)...)
		}
	}
	if volume.Server == nil {
		d.Set("server", "")
	} else {
		err = d.Set("server", volume.Server.ID)
		if err != nil {
			diags = append(diags, diag.Errorf("unexpected: %s", err)...)
		}
	}
	return diags
}

func volumeUnavailable(obj *brightbox.Volume) bool {
	return obj.Status == volumestatus.Deleted ||
		obj.Status == volumestatus.Deleting ||
		obj.Status == volumestatus.Failed
}

func volumeStateRefresh(client *brightbox.Client, ctx context.Context, volumeID string) retry.StateRefreshFunc {
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

	stateConf := retry.StateChangeConf{
		Pending: []string{
			volumestatus.Creating.String(),
		},
		Target: []string{
			volumestatus.Detached.String(),
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

	if serverID, ok := d.GetOk("server"); ok {
		if target := serverID.(string); target != "" {
			if err := attachVolume(ctx, d, meta, target, d.Timeout(schema.TimeoutUpdate)); err != nil {
				diags = append(diags, diag.FromErr(err)...)
			}
		}
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
	log.Printf("[DEBUG] Checking if server attachment has changed")
	if d.HasChange("server") {
		log.Printf("[INFO] Volume server attachment has changed, updating...")
		err := detachVolume(ctx, d, meta, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			diags = append(diags, diag.FromErr(err)...)
		} else if serverID, ok := d.GetOk("server"); ok {
			if target := serverID.(string); target != "" {
				err := attachVolume(ctx, d, meta, target, d.Timeout(schema.TimeoutUpdate))
				if err != nil {
					diags = append(diags, diag.FromErr(err)...)
				}
			}
		}
	}
	log.Printf("[DEBUG] Updating volume")
	return append(diags, resourceBrightboxVolumeUpdate(ctx, d, meta)...)
}

func attachVolume(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
	server string,
	timeout time.Duration,
) error {
	log.Printf("[DEBUG] attaching %v to %v", d.Id(), server)
	client := meta.(*CompositeClient).APIClient
	_, err := client.AttachVolume(
		ctx,
		d.Id(),
		brightbox.VolumeAttachment{Server: server},
	)
	if err != nil {
		return err
	}
	stateConf := retry.StateChangeConf{
		Pending: []string{
			volumestatus.Detached.String(),
		},
		Target: []string{
			volumestatus.Attached.String(),
		},
		Refresh:    volumeStateRefresh(client, ctx, d.Id()),
		Timeout:    timeout,
		Delay:      checkDelay,
		MinTimeout: minimumRefreshWait,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	return err
}

func detachVolume(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
	timeout time.Duration,
) error {
	log.Printf("[DEBUG] detaching %v", d.Id())
	client := meta.(*CompositeClient).APIClient
	_, err := client.DetachVolume(
		ctx,
		d.Id(),
	)
	if err != nil {
		return err
	}
	stateConf := retry.StateChangeConf{
		Pending: []string{
			volumestatus.Attached.String(),
		},
		Target: []string{
			volumestatus.Detached.String(),
		},
		Refresh:    volumeStateRefresh(client, ctx, d.Id()),
		Timeout:    timeout,
		Delay:      checkDelay,
		MinTimeout: minimumRefreshWait,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	return err
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
