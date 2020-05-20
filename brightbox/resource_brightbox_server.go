package brightbox

import (
	"fmt"
	"log"

	brightbox "github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	userdataSizeLimit = 16384
)

func resourceBrightboxServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceBrightboxServerCreate,
		Read:   resourceBrightboxServerRead,
		Update: resourceBrightboxServerUpdate,
		Delete: resourceBrightboxServerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultTimeout),
			Delete: schema.DefaultTimeout(defaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"image": {
				Description: "Image used to create the server",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "Editable user label",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"type": {
				Description: "Server type of the server",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			"zone": {
				Description: "Zone where server is located",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			"user_data": {
				Description:   "Data made available to Cloud Init",
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"user_data_base64"},
				StateFunc:     hash_string,
			},
			"user_data_base64": {
				Description:   "Base64 encoded data made availalbe to Cloud Init",
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"user_data"},
				ValidateFunc:  mustBeBase64Encoded,
			},
			"server_groups": {
				Description: "Array of server groups to add server to",
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
			"locked": {
				Description: "Is true if resource has been set as locked and cannot be deleted",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},

			"status": {
				Description: "Current state of server",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"interface": {
				Description: "Network Interface connected to this server",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"ipv6_address": {
				Description: "Public IPv6 address of the interface",
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

			"hostname": {
				Description: "Short hostname",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"fqdn": {
				Description: "Fully qualified domain name",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"public_hostname": {
				Description: "Public IPv4 FQDN",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"ipv6_hostname": {
				Description: "Public IPv6 FQDN",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"username": {
				Description: "Username to use when logging into a server",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceBrightboxServerCreate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[DEBUG] Server create called")
	serverOpts := &brightbox.ServerOptions{
		Image: d.Get("image").(string),
	}

	err := addUpdateableServerOptions(d, serverOpts)
	if err != nil {
		return err
	}

	serverType := &serverOpts.ServerType
	assign_string(d, &serverType, "type")
	zone := &serverOpts.Zone
	assign_string(d, &zone, "zone")

	log.Printf("[DEBUG] Server create configuration: %#v", serverOpts)

	server, err := client.CreateServer(serverOpts)
	if err != nil {
		return fmt.Errorf("Error creating server: %s", err)
	}

	d.SetId(server.Id)

	locked := d.Get("locked").(bool)
	log.Printf("[INFO] Setting lock state to %v", locked)
	if err := setLockState(client, locked, server); err != nil {
		return err
	}

	log.Printf("[INFO] Waiting for Server (%s) to become available", d.Id())

	stateConf := resource.StateChangeConf{
		Pending:    []string{"creating"},
		Target:     []string{"active", "inactive"},
		Refresh:    serverStateRefresh(client, server.Id),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      checkDelay,
		MinTimeout: minimumRefreshWait,
	}
	activeServer, err := stateConf.WaitForState()
	if err != nil {
		return err
	}

	return setServerAttributes(d, activeServer.(*brightbox.Server))
}

func resourceBrightboxServerRead(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[DEBUG] Server read called for %s", d.Id())
	server, err := client.Server(d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving server details: %s", err)
	}
	if server.Status == "deleted" {
		log.Printf("[WARN] Server not found, removing from state: %s", d.Id())
		d.SetId("")
		return nil
	}

	return setServerAttributes(d, server)
}

func resourceBrightboxServerDelete(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[DEBUG] Server delete called for %s", d.Id())
	err := client.DestroyServer(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting server: %s", err)
	}
	stateConf := resource.StateChangeConf{
		Pending:    []string{"deleting", "active", "inactive"},
		Target:     []string{"deleted"},
		Refresh:    serverStateRefresh(client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      checkDelay,
		MinTimeout: minimumRefreshWait,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func resourceBrightboxServerUpdate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[DEBUG] Server update called for %s", d.Id())
	serverOpts := &brightbox.ServerOptions{
		Id: d.Id(),
	}

	err := addUpdateableServerOptions(d, serverOpts)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Server update configuration: %#v", serverOpts)

	server, err := client.UpdateServer(serverOpts)
	if err != nil {
		return fmt.Errorf("Error updating server: %s", err)
	}
	if d.HasChange("locked") {
		locked := d.Get("locked").(bool)
		log.Printf("[INFO] Setting lock state to %v", locked)
		if err := setLockState(client, locked, server); err != nil {
			return err
		}
		return resourceBrightboxServerRead(d, meta)
	}
	return setServerAttributes(d, server)
}

func addUpdateableServerOptions(
	d *schema.ResourceData,
	opts *brightbox.ServerOptions,
) error {
	assign_string(d, &opts.Name, "name")
	assign_string_set(d, &opts.ServerGroups, "server_groups")
	if d.HasChange("user_data") {
		encodedUserData := ""
		if userData, ok := d.GetOk("user_data"); ok {
			log.Printf("[DEBUG] UserData to encode: %s", userData.(string))
			encodedUserData = base64Encode(userData.(string))
		} else if userData, ok := d.GetOk("user_data_base64"); ok {
			log.Printf("[DEBUG] Encoded Userdata found, passing through")
			encodedUserData = userData.(string)
		}
		if encodedUserData == "" {
			// Nothing found, nothing to do
		} else if len(encodedUserData) > userdataSizeLimit {
			return fmt.Errorf(
				"The supplied user_data contains %d bytes after encoding, this exeeds the limit of %d bytes",
				len(encodedUserData),
				userdataSizeLimit,
			)
		} else {
			opts.UserData = &encodedUserData
		}
	}
	return nil
}

func setServerAttributes(
	d *schema.ResourceData,
	server *brightbox.Server,
) error {
	d.Set("image", server.Image.Id)
	d.Set("name", server.Name)
	d.Set("type", server.ServerType.Handle)
	d.Set("zone", server.Zone.Handle)
	d.Set("status", server.Status)
	d.Set("locked", server.Locked)
	d.Set("hostname", server.Hostname)
	d.Set("username", server.Image.Username)

	if len(server.Interfaces) > 0 {
		serverInterface := server.Interfaces[0]
		d.Set("interface", serverInterface.Id)
		d.Set("ipv4_address_private", serverInterface.IPv4Address)
		d.Set("fqdn", server.Fqdn)
		d.Set("ipv6_address", serverInterface.IPv6Address)
		d.Set("ipv6_hostname", "ipv6."+server.Fqdn)
	}

	if len(server.CloudIPs) > 0 {
		setPrimaryCloudIP(d, &server.CloudIPs[0])
	}

	if err := d.Set("server_groups", schema.NewSet(schema.HashString, flattenServerGroups(server.ServerGroups))); err != nil {
		return fmt.Errorf("error setting server_groups: %s", err)
	}

	setUserDataDetails(d, server.UserData)
	setConnectionDetails(d)
	return nil

}

func flattenServerGroups(list []brightbox.ServerGroup) []interface{} {
	srvGrpIds := make([]interface{}, len(list))
	for i, sg := range list {
		srvGrpIds[i] = sg.Id
	}
	return srvGrpIds
}

func setUserDataDetails(d *schema.ResourceData, base64Userdata string) {
	if len(base64Userdata) <= 0 {
		log.Printf("[DEBUG] No user data found, skipping set")
		return
	}
	_, b64 := d.GetOk("user_data_base64")
	if b64 {
		log.Printf("[DEBUG] encoded user_data requested, setting user_data_base64")
		d.Set("user_data_base64", base64Userdata)
	} else {
		log.Printf("[DEBUG] decrypted user_data requested, setting user_data")
		d.Set("user_data", userDataHashSum(base64Userdata))
	}
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

func serverStateRefresh(client *brightbox.Client, serverID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		server, err := client.Server(serverID)
		if err != nil {
			log.Printf("Error on Server State Refresh: %s", err)
			return nil, "", err
		}
		return server, server.Status, nil
	}
}
