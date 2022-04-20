package brightbox

import (
	"context"
	"log"

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
		ReadContext: resourceBrightboxRead(
			(*brightbox.Client).Server,
			"Server",
			setServerAttributes,
		),
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

			"disk_encrypted": {
				Description: "Is true if the server has been built with an encrypted disk",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},

			"disk_size": {
				Description:  "Disk size in megabytes for server types with variable block storage",
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
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
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(imageRegexp, "must be a valid image ID"),
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
				Description: "Array of server groups to add server to",
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringMatch(serverGroupRegexp, "must be a valid server group ID"),
				},
				Set: schema.HashString,
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
		Refresh:    serverStateRefresh(ctx, client, server.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      checkDelay,
		MinTimeout: minimumRefreshWait,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceBrightboxSetServerLockState(ctx, d, meta)
}

var resourceBrightboxServerDelete = resourceBrightboxDelete(
	(*brightbox.Client).DestroyServer,
	"Server",
)

func resourceBrightboxServerDeleteAndWait(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	diags := resourceBrightboxServerDelete(ctx, d, meta)
	if diags.HasError() {
		return diags
	}

	client := meta.(*CompositeClient).APIClient
	stateConf := resource.StateChangeConf{
		Pending: []string{
			serverConst.Deleting.String(),
			serverConst.Active.String(),
			serverConst.Inactive.String(),
		},
		Target:     []string{serverConst.Deleted.String()},
		Refresh:    serverStateRefresh(ctx, client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      checkDelay,
		MinTimeout: minimumRefreshWait,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
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

	if d.HasChanges("name", "server_groups", "user_data", "user_data_base64") {
		errs := addUpdateableServerOptions(d, &serverOpts)
		if errs.HasError() {
			return errs
		}

		log.Printf("[DEBUG] Server update configuration: %+v", serverOpts)

		server, err = client.UpdateServer(ctx, serverOpts)
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		server, err = client.Server(ctx, d.Id())
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChange("disk_size") {
		volumeID := server.Volumes[0].ID
		oldSize, newSize := d.GetChange("disk_size")
		oldSizeInt, ok := oldSize.(int)
		if !ok {
			return diag.Errorf("expected type of old disk size to be Integer")
		}
		newSizeInt, ok := newSize.(int)
		if !ok {
			return diag.Errorf("expected type of new disk size to be Integer")
		}
		if oldSizeInt > newSizeInt {
			return diag.Errorf("expected new disk size (%v) to be bigger than old disk size (%v)", newSizeInt, oldSizeInt)

		}
		log.Printf("[INFO] Resizing volume %v from %v to %v", volumeID, oldSizeInt, newSizeInt)
		_, err = client.ResizeVolume(
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
		server, err = client.Server(ctx, d.Id())
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChange("type") {
		newServerType := d.Get("type").(string)
		log.Printf("[INFO] Changing server type to %v", newServerType)
		server, err = client.ResizeServer(
			ctx,
			d.Id(),
			brightbox.ServerNewSize{newServerType},
		)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChange("locked") {
		return resourceBrightboxSetServerLockState(ctx, d, meta)
	}
	return setServerAttributes(d, server)
}

var resourceBrightboxSetServerLockState = resourceBrightboxSetLockState(
	(*brightbox.Client).LockServer,
	(*brightbox.Client).UnlockServer,
	setServerAttributes,
)

func addBlockStorageOptions(
	d *schema.ResourceData,
	opts *brightbox.ServerOptions,
) {
	var diskSize *uint
	image := d.Get("image").(string)
	assignInt(d, &diskSize, "disk_size")
	if diskSize == nil {
		opts.Image = &image
	} else {
		opts.Volumes = []brightbox.VolumeOptions{
			brightbox.VolumeOptions{
				Image: image,
				Size:  *diskSize,
			},
		}
	}
}

func addUpdateableServerOptions(
	d *schema.ResourceData,
	opts *brightbox.ServerOptions,
) diag.Diagnostics {
	assignString(d, &opts.Name, "name")
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
	if server.ServerType.DiskSize != 0 {
		err = d.Set("disk_size", server.ServerType.DiskSize)
		if err != nil {
			diags = append(diags, diag.Errorf("unexpected: %s", err)...)
		}
	} else {
		err = d.Set("disk_size", server.Volumes[0].Size)
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

	err = d.Set("server_groups", serverGroupIDListFromGroups(server.ServerGroups))
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}

	diags = append(diags, setUserDataDetails(d, server.UserData)...)
	setConnectionDetails(d)
	diags = append(diags, setServerTypeDetails(d, server.ServerType)...)
	return diags

}

func serverGroupIDListFromGroups(
	list []brightbox.ServerGroup,
) []string {
	srvGrpIds := make([]string, len(list))
	for i, sg := range list {
		srvGrpIds[i] = sg.ID
	}
	return srvGrpIds
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

func serverStateRefresh(ctx context.Context, client *brightbox.Client, serverID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		serverInstance, err := client.Server(ctx, serverID)
		if err != nil {
			log.Printf("Error on Server State Refresh: %s", err)
			return nil, "", err
		}
		return serverInstance, serverInstance.Status.String(), nil
	}
}
