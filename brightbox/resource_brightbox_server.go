package brightbox

import (
	"fmt"
	"log"
	"time"

	"github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	userdata_size_limit = 16384
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

		Schema: map[string]*schema.Schema{
			"image": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},

			"type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Default:  nil,
				ForceNew: true,
			},

			"zone": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Default:  nil,
				ForceNew: true,
			},

			"user_data": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"user_data_base64"},
				StateFunc:     hash_string,
			},

			"user_data_base64": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"user_data"},
				ValidateFunc: func(v interface{}, name string) (warns []string, errs []error) {
					s := v.(string)
					if !isBase64Encoded(s) {
						errs = append(errs, fmt.Errorf(
							"%s: must be base64-encoded", name,
						))
					}
					return
				},
			},

			"server_groups": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"status": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"locked": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},

			"interface": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"ipv6_address": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"ipv4_address": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"ipv4_address_private": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"hostname": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"fqdn": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"public_hostname": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"ipv6_hostname": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"username": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceBrightboxServerCreate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	log.Printf("[DEBUG] Server create called")
	server_opts := &brightbox.ServerOptions{
		Image: d.Get("image").(string),
	}

	err := addUpdateableServerOptions(d, server_opts)
	if err != nil {
		return err
	}

	server_type := &server_opts.ServerType
	assign_string(d, &server_type, "type")
	zone := &server_opts.Zone
	assign_string(d, &zone, "zone")

	log.Printf("[DEBUG] Server create configuration: %#v", server_opts)

	server, err := client.CreateServer(server_opts)
	if err != nil {
		return fmt.Errorf("Error creating server: %s", err)
	}

	d.SetId(server.Id)

	log.Printf("[INFO] Waiting for Server (%s) to become available", d.Id())

	stateConf := resource.StateChangeConf{
		Pending:    []string{"creating"},
		Target:     []string{"active", "inactive"},
		Refresh:    serverStateRefresh(client, server.Id),
		Timeout:    5 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	active_server, err := stateConf.WaitForState()
	if err != nil {
		return err
	}

	setServerAttributes(d, active_server.(*brightbox.Server))

	return nil
}

func resourceBrightboxServerRead(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	log.Printf("[DEBUG] Server read called for %s", d.Id())
	server, err := client.Server(d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving server details: %s", err)
	}

	setServerAttributes(d, server)

	return nil

}

func resourceBrightboxServerDelete(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	log.Printf("[DEBUG] Server delete called for %s", d.Id())
	err := client.DestroyServer(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting server: %s", err)
	}
	stateConf := resource.StateChangeConf{
		Pending:    []string{"deleting", "active", "inactive"},
		Target:     []string{"deleted"},
		Refresh:    serverStateRefresh(client, d.Id()),
		Timeout:    5 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
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
	client := meta.(*CompositeClient).ApiClient

	log.Printf("[DEBUG] Server update called for %s", d.Id())
	server_opts := &brightbox.ServerOptions{
		Id: d.Id(),
	}

	err := addUpdateableServerOptions(d, server_opts)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Server update configuration: %#v", server_opts)

	server, err := client.UpdateServer(server_opts)
	if err != nil {
		return fmt.Errorf("Error updating server: %s", err)
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
		encoded_userdata := ""
		if user_data, ok := d.GetOk("user_data"); ok {
			log.Printf("[DEBUG] UserData to encode: %s", user_data.(string))
			encoded_userdata = base64Encode(user_data.(string))
		} else if user_data, ok := d.GetOk("user_data_base64"); ok {
			log.Printf("[DEBUG] Encoded Userdata found, passing through")
			encoded_userdata = user_data.(string)
		}
		if encoded_userdata == "" {
			// Nothing found, nothing to do
		} else if len(encoded_userdata) > userdata_size_limit {
			return fmt.Errorf(
				"The supplied user_data contains %d bytes after encoding, this exeeds the limit of %d bytes",
				len(encoded_userdata),
				userdata_size_limit,
			)
		} else {
			opts.UserData = &encoded_userdata
		}
	}
	return nil
}

func setServerAttributes(
	d *schema.ResourceData,
	server *brightbox.Server,
) error {
	// If server is deleted, clear the Id to tell Terraform to remove the record
	if server.Status == "deleted" {
		d.SetId("")
		return nil
	}
	d.Set("image", server.Image.Id)
	d.Set("name", server.Name)
	d.Set("type", server.ServerType.Handle)
	d.Set("zone", server.Zone.Handle)
	d.Set("status", server.Status)
	d.Set("locked", server.Locked)
	d.Set("hostname", server.Hostname)
	d.Set("username", server.Image.Username)

	if len(server.Interfaces) > 0 {
		server_interface := server.Interfaces[0]
		d.Set("interface", server_interface.Id)
		d.Set("ipv4_address_private", server_interface.IPv4Address)
		d.Set("fqdn", server.Fqdn)
		d.Set("ipv6_address", server_interface.IPv6Address)
		d.Set("ipv6_hostname", "ipv6."+server.Fqdn)
	}

	if len(server.CloudIPs) > 0 {
		setPrimaryCloudIp(d, &server.CloudIPs[0])
	}

	srvGrpIds := []string{}
	for _, sg := range server.ServerGroups {
		srvGrpIds = append(srvGrpIds, sg.Id)
	}
	d.Set("server_groups", srvGrpIds)

	setUserDataDetails(d, server.UserData)
	setConnectionDetails(d)
	return nil

}

func setUserDataDetails(d *schema.ResourceData, base64_userdata string) {
	if len(base64_userdata) <= 0 {
		log.Printf("[DEBUG] No user data found, skipping set")
		return
	}
	_, b64 := d.GetOk("user_data_base64")
	if b64 {
		log.Printf("[DEBUG] encoded user_data requested, setting user_data_base64")
		d.Set("user_data_base64", base64_userdata)
	} else {
		log.Printf("[DEBUG] decrypted user_data requested, setting user_data")
		d.Set("user_data", userDataHashSum(base64_userdata))
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
		connection_details := map[string]string{
			"type": "ssh",
			"host": preferredSSHAddress,
		}
		if attr, ok := d.GetOk("username"); ok {
			connection_details["user"] = attr.(string)
		}
		d.SetConnInfo(connection_details)
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
