package brightbox

import (
	"context"
	"fmt"
	"log"
	"time"

	brightbox "github.com/brightbox/gobrightbox/v2"
	serverConst "github.com/brightbox/gobrightbox/v2/status/server"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	userdataSizeLimit = 16384
)

func resourceBrightboxServer() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Brightbox Server resource",
		CreateContext: resourceBrightboxServerCreateAndWait,
		ReadContext:   resourceBrightboxServerRead,
		UpdateContext: resourceBrightboxServerUpdate,
		DeleteContext: resourceBrightboxServerDeleteAndWait,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultTimeout),
			Delete: schema.DefaultTimeout(defaultTimeout),
		},

		Schema: map[string]*schema.Schema{

			"data_volumes": {
				Description: "List of volumes to attach to server",
				Type:        schema.TypeSet,
				Optional:    true,
				MinItems:    0,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringMatch(volumeRegexp, "must be a valid volume ID"),
				},
				Set: schema.HashString,
			},

			"disk_encrypted": {
				Description:  "Is true if the server has been built with an encrypted disk",
				Type:         schema.TypeBool,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				RequiredWith: []string{"image"},
			},

			"disk_size": {
				Description:  "Disk size in megabytes for server types with variable block storage",
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				RequiredWith: []string{"image"},
			},

			"fqdn": {
				Description: "Fully qualified domain name",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"hostname": {
				Description: "Short hostname",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"image": {
				Description:  "Image used to create the server",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(imageRegexp, "must be a valid image ID"),
				ExactlyOneOf: []string{"image", "volume"},
			},

			"interface": {
				Description: "Network Interface connected to this server",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"ipv4_address": {
				Description: "Public IPv4 address of the interface",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"ipv4_address_private": {
				Description: "Private IPv4 address of the interface",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"ipv6_address": {
				Description: "Public IPv6 address of the interface",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"ipv6_hostname": {
				Description: "Public IPv6 FQDN",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"locked": {
				Description: "Is true if resource has been set as locked and cannot be deleted",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},

			"name": {
				Description: "Editable user label",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"public_hostname": {
				Description: "Public IPv4 FQDN",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"server_groups": {
				Description: "List of server groups to add server to",
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringMatch(serverGroupRegexp, "must be a valid server group ID"),
				},
				Set: schema.HashString,
			},

			"snapshots_retention": {
				Description:  "Keep this number of scheduled snapshots. Keep all if unset",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},

			"snapshots_schedule": {
				Description:  "Crontab pattern for scheduled snapshots. Must be at least hourly",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "0 7 * * *",
				ValidateFunc: ValidateCronString,
			},

			"snapshots_schedule_next_at": {
				Description: "time in UTC when next approximate scheduled snapshot will be run",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"status": {
				Description: "Current state of server",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"type": {
				Description: "Server type of the server",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ValidateFunc: validation.Any(
					validation.StringMatch(serverTypeRegexp, "must be a valid server type"),
					validation.StringIsNotWhiteSpace,
				),
			},

			"user_data": {
				Description:   "Data made available to Cloud Init",
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"user_data_base64"},
				StateFunc:     hashString,
				ValidateFunc:  validation.StringIsNotWhiteSpace,
			},

			"user_data_base64": {
				Description:   "Base64 encoded data made available to Cloud Init",
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"user_data"},
				ValidateFunc:  validation.StringIsBase64,
			},

			"username": {
				Description: "Username to use when logging into a server",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"volume": {
				Description:  "Volume used to boot the server",
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.StringMatch(volumeRegexp, "must be a valid volume ID"),
				ExactlyOneOf: []string{"image", "volume"},
			},

			"zone": {
				Description:  "Zone where server is located",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(zoneRegexp, "must be a valid zone ID or handle"),
			},
		},
	}
}

var (
	resourceBrightboxServerRead = resourceBrightboxReadStatus(
		(*brightbox.Client).Server,
		"Server",
		setServerAttributes,
		serverUnavailable,
	)

	resourceBrightboxServerDeleteAndWait = resourceBrightboxDeleteAndWait(
		(*brightbox.Client).DestroyServer,
		"Server",
		[]string{
			serverConst.Deleting.String(),
			serverConst.Active.String(),
			serverConst.Inactive.String(),
		},
		[]string{
			serverConst.Deleted.String(),
		},
		serverStateRefresh,
	)

	resourceBrightboxSetServerLockState = resourceBrightboxSetLockState(
		(*brightbox.Client).LockServer,
		(*brightbox.Client).UnlockServer,
		setServerAttributes,
	)
)

func addUpdateableServerOptions(
	d *schema.ResourceData,
	opts *brightbox.ServerOptions,
) diag.Diagnostics {
	assignString(d, &opts.Name, "name")
	assignString(d, &opts.SnapshotsSchedule, "snapshots_schedule")
	assignString(d, &opts.SnapshotsRetention, "snapshots_retention")
	assignStringSet(d, &opts.ServerGroups, "server_groups")
	encodedUserData := ""
	if d.HasChange("user_data") {
		if userData, ok := d.GetOk("user_data"); ok {
			log.Printf("[DEBUG] UserData to encode: %s", userData.(string))
			encodedUserData = base64Encode(userData.(string))
		}
	}
	if d.HasChange("user_data_base64") {
		if userData, ok := d.GetOk("user_data_base64"); ok {
			log.Printf("[DEBUG] Encoded Userdata found, passing through")
			encodedUserData = userData.(string)
		}
	}
	if len(encodedUserData) > userdataSizeLimit {
		return diag.Errorf(
			"The supplied user_data contains %d bytes after encoding, this exeeds the limit of %d bytes",
			len(encodedUserData),
			userdataSizeLimit,
		)
	}
	if encodedUserData != "" {
		opts.UserData = &encodedUserData
	}
	return nil
}

func setServerAttributes(
	d *schema.ResourceData,
	server *brightbox.Server,
) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	d.SetId(server.ID)
	err = d.Set("image", server.Image.ID)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("name", server.Name)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("zone", server.Zone.Handle)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("status", server.Status.String())
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("locked", server.Locked)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("disk_encrypted", server.DiskEncrypted)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("hostname", server.Hostname)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("username", server.Image.Username)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("snapshots_retention", server.SnapshotsRetention)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("snapshots_schedule", server.SnapshotsSchedule)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	var snapshotTime string
	if server.SnapshotsScheduleNextAt != nil {
		snapshotTime = server.SnapshotsScheduleNextAt.Format(time.RFC3339)
	}
	err = d.Set("snapshots_schedule_next_at", snapshotTime)
	if server.ServerType.DiskSize != 0 {
		err = d.Set("disk_size", server.ServerType.DiskSize)
		if err != nil {
			diags = append(diags, diag.Errorf("unexpected: %s", err)...)
		}
	}

	if len(server.Interfaces) > 0 {
		serverInterface := server.Interfaces[0]
		err = d.Set("interface", serverInterface.ID)
		if err != nil {
			diags = append(diags, diag.Errorf("unexpected: %s", err)...)
		}
		err = d.Set("ipv4_address_private", serverInterface.IPv4Address)
		if err != nil {
			diags = append(diags, diag.Errorf("unexpected: %s", err)...)
		}
		err = d.Set("fqdn", server.Fqdn)
		if err != nil {
			diags = append(diags, diag.Errorf("unexpected: %s", err)...)
		}
		err = d.Set("ipv6_address", serverInterface.IPv6Address)
		if err != nil {
			diags = append(diags, diag.Errorf("unexpected: %s", err)...)
		}
		err = d.Set("ipv6_hostname", "ipv6."+server.Fqdn)
		if err != nil {
			diags = append(diags, diag.Errorf("unexpected: %s", err)...)
		}
	}

	if len(server.CloudIPs) > 0 {
		setPrimaryCloudIP(d, &server.CloudIPs[0])
	}

	bootVolumes := filter(server.Volumes, func(v brightbox.Volume) bool { return v.Boot })

	if len(bootVolumes) > 0 {
		bootVolume := &bootVolumes[0]
		err = d.Set("volume", bootVolume.ID)
		if err != nil {
			diags = append(diags, diag.Errorf("unexpected: %s", err)...)
		}
		err = d.Set("disk_size", bootVolume.Size)
		if err != nil {
			diags = append(diags, diag.Errorf("unexpected: %s", err)...)
		}
		err = d.Set("disk_encrypted", bootVolume.Encrypted)
		if err != nil {
			diags = append(diags, diag.Errorf("unexpected: %s", err)...)
		}
	} else {
		diags = append(diags,
			diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("No boot volume detected in volume list"),
			})
	}

	err = d.Set(
		"server_groups",
		idList(
			server.ServerGroups,
			func(v brightbox.ServerGroup) string { return v.ID },
		),
	)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}

	err = d.Set(
		"data_volumes",
		idList(
			filter(server.Volumes, func(v brightbox.Volume) bool { return !v.Boot }),
			func(v brightbox.Volume) string { return v.ID },
		),
	)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}

	diags = append(diags, setUserDataDetails(d, server.UserData)...)
	setConnectionDetails(d)
	diags = append(diags, setServerTypeDetails(d, server.ServerType)...)
	return diags

}

func serverUnavailable(obj *brightbox.Server) bool {
	return obj.Status == serverConst.Deleted ||
		obj.Status == serverConst.Failed
}

func serverStateRefresh(client *brightbox.Client, ctx context.Context, serverID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		serverInstance, err := client.Server(ctx, serverID)
		if err != nil {
			log.Printf("Error on Server State Refresh: %s", err)
			return nil, "", err
		}
		return serverInstance, serverInstance.Status.String(), nil
	}
}

func resourceBrightboxServerCreateAndWait(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[INFO] Creating Server")
	var serverOpts brightbox.ServerOptions

	errs := addUpdateableServerOptions(d, &serverOpts)
	if errs.HasError() {
		return errs
	}

	assignString(d, &serverOpts.ServerType, "type")
	assignString(d, &serverOpts.Zone, "zone")
	assignBool(d, &serverOpts.DiskEncrypted, "disk_encrypted")
	addBlockStorageOptions(d, &serverOpts)

	log.Printf("[DEBUG] Server create configuration: %+v", serverOpts)

	server, err := client.CreateServer(ctx, serverOpts)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(server.ID)

	log.Printf("[INFO] Waiting for Server (%s) to become available", d.Id())

	stateConf := resource.StateChangeConf{
		Pending: []string{
			serverConst.Creating.String(),
		},
		Target: []string{
			serverConst.Active.String(),
			serverConst.Inactive.String(),
		},
		Refresh:    serverStateRefresh(client, ctx, server.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      checkDelay,
		MinTimeout: minimumRefreshWait,
	}
	result, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	server = result.(*brightbox.Server)

	if errs := adjustDiskAttachment(ctx, d, meta, server.Volumes); errs.HasError() {
		return errs
	}

	return resourceBrightboxSetServerLockState(ctx, d, meta)
}

func resourceBrightboxServerUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[DEBUG] Server update called for %s", d.Id())
	serverOpts := brightbox.ServerOptions{
		ID: d.Id(),
	}
	var server *brightbox.Server
	var err error
	var diags diag.Diagnostics

	if d.HasChanges("name", "server_groups", "user_data", "user_data_base64") {
		diags = append(diags, addUpdateableServerOptions(d, &serverOpts)...)
		if diags.HasError() {
			return diags
		}
		log.Printf("[DEBUG] Server update configuration: %+v", serverOpts)

		server, err = client.UpdateServer(ctx, serverOpts)
		if err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	} else {
		server, err = client.Server(ctx, d.Id())
		if err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}
	if d.HasChange("disk_size") {
		diags = append(diags, resizeBrightboxVolume(ctx, d, meta, server.Volumes[0].ID, "disk_size")...)
		server, err = client.Server(ctx, d.Id())
		if err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}
	if d.HasChange("type") {
		newServerType := d.Get("type").(string)
		log.Printf("[INFO] Changing server type to %v", newServerType)
		server, err = client.ResizeServer(
			ctx,
			d.Id(),
			brightbox.ServerNewSize{NewType: newServerType},
		)
		if err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}
	if diags.HasError() {
		return diags
	}

	if errs := adjustDiskAttachment(ctx, d, meta, server.Volumes); errs.HasError() {
		return errs
	}
	if d.HasChange("locked") {
		return resourceBrightboxSetServerLockState(ctx, d, meta)
	}
	return resourceBrightboxServerRead(ctx, d, meta)
}

func adjustDiskAttachment(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
	currentVolumes []brightbox.Volume,
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient
	log.Printf("[DEBUG] adjust Disks called for %s", d.Id())
	var requiredVolumeList []string
	assignStringSet(d, &requiredVolumeList, "data_volumes")

	currentVolumeList := idList(
		filter(currentVolumes, func(v brightbox.Volume) bool { return !v.Boot }),
		func(v brightbox.Volume) string { return v.ID },
	)
	diags := attachDisks(
		ctx,
		client,
		Difference(requiredVolumeList, currentVolumeList),
		brightbox.VolumeAttachment{Server: d.Id()},
	)
	return append(diags, detachDisks(
		ctx,
		client,
		Difference(currentVolumeList, requiredVolumeList),
	)...)
}

func attachDisks(
	ctx context.Context,
	client *brightbox.Client,
	volumeIDs []string,
	attachment brightbox.VolumeAttachment,
) diag.Diagnostics {
	log.Printf("[DEBUG] attaching %v", volumeIDs)
	var diags diag.Diagnostics
	for _, id := range volumeIDs {
		_, err := client.AttachVolume(
			ctx,
			id,
			attachment,
		)
		if err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}
	return diags
}

func detachDisks(
	ctx context.Context,
	client *brightbox.Client,
	volumeIDs []string,
) diag.Diagnostics {
	log.Printf("[DEBUG] detaching %v", volumeIDs)
	var diags diag.Diagnostics
	for _, id := range volumeIDs {
		_, err := client.DetachVolume(
			ctx,
			id,
		)
		if err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}
	return diags
}

func addBlockStorageOptions(
	d *schema.ResourceData,
	opts *brightbox.ServerOptions,
) {
	if volume, ok := d.GetOk("volume"); ok {
		opts.Volumes = []brightbox.VolumeEntry{
			brightbox.VolumeEntry{
				Volume: volume.(string),
			},
		}
		return
	}
	image := d.Get("image").(string)
	if diskSize, ok := d.GetOk("disk_size"); ok {
		opts.Volumes = []brightbox.VolumeEntry{
			brightbox.VolumeEntry{
				Image: image,
				Size:  uint(diskSize.(int)),
			},
		}
		return
	}
	opts.Image = &image
}

func setUserDataDetails(d *schema.ResourceData, base64Userdata string) diag.Diagnostics {
	if len(base64Userdata) <= 0 {
		log.Printf("[DEBUG] No user data found, skipping set")
		return nil
	}
	_, b64 := d.GetOk("user_data_base64")
	if b64 {
		log.Printf("[DEBUG] encoded user_data requested, setting user_data_base64")
		if err := d.Set("user_data_base64", base64Userdata); err != nil {
			return diag.FromErr(err)
		}
	} else {
		log.Printf("[DEBUG] decrypted user_data requested, setting user_data")
		if err := d.Set("user_data", userDataHashSum(base64Userdata)); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func setConnectionDetails(d *schema.ResourceData) {
	var preferredSSHAddress string
	if attr, ok := d.GetOk("public_hostname"); ok {
		preferredSSHAddress = attr.(string)
	} else if attr, ok := d.GetOk("ipv6_hostname"); ok {
		preferredSSHAddress = attr.(string)
	} else if attr, ok := d.GetOk("fqdn"); ok {
		preferredSSHAddress = attr.(string)
	}

	if preferredSSHAddress != "" {
		connectionDetails := map[string]string{
			"type": "ssh",
			"host": preferredSSHAddress,
		}
		if attr, ok := d.GetOk("username"); ok {
			connectionDetails["user"] = attr.(string)
		}
		d.SetConnInfo(connectionDetails)
	}
}

func setServerTypeDetails(d *schema.ResourceData, serverType *brightbox.ServerType) diag.Diagnostics {
	if currentType, ok := d.GetOk("type"); ok {
		if !serverTypeRegexp.MatchString(currentType.(string)) {
			if err := d.Set("type", serverType.Handle); err != nil {
				return diag.FromErr(err)
			}
			return nil
		}
	}
	if err := d.Set("type", serverType.ID); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
