package brightbox

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/resource"
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

			"compatibility": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				Default:  nil,
				ForceNew: true,
			},

			"user_data": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				StateFunc: func(v interface{}) string {
					switch v.(type) {
					case string:
						hash := sha1.Sum([]byte(v.(string)))
						return hex.EncodeToString(hash[:])
					default:
						return ""
					}
				},
			},

			"server_groups": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
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
	client := meta.(*brightbox.Client)

	server_opts := &brightbox.ServerOptions{
		Image: d.Get("image").(string),
	}

	err := addUpdateableOptions(d, server_opts)
	if err != nil {
		return err
	}

	if attr, ok := d.GetOk("type"); ok {
		server_opts.ServerType = attr.(string)
	}
	if attr, ok := d.GetOk("zone"); ok {
		server_opts.Zone = attr.(string)
	}

	if sgs := d.Get("server_groups").(*schema.Set); sgs.Len() > 0 {
		var groups []string
		for _, v := range sgs.List() {
			groups = append(groups, v.(string))
		}
		server_opts.ServerGroups = &groups
	}

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
	client := meta.(*brightbox.Client)

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
	client := meta.(*brightbox.Client)

	err := client.DestroyServer(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting server: %s", err)
	}
	return nil
}

func resourceBrightboxServerUpdate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*brightbox.Client)

	server_opts := &brightbox.ServerOptions{
		Id: d.Id(),
	}

	err := addUpdateableOptions(d, server_opts)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Server update configuration: %#v", server_opts)

	server, err := client.UpdateServer(server_opts)
	if err != nil {
		return fmt.Errorf("Error updating server: %s", err)
	}

	setServerAttributes(d, server)

	return nil
}

func addUpdateableOptions(
	d *schema.ResourceData,
	opts *brightbox.ServerOptions,
) error {

	if attr, ok := d.GetOk("name"); ok {
		temp_name := attr.(string)
		opts.Name = &temp_name
	}

	if attr, ok := d.GetOk("userdata"); ok {
		encoded_userdata := base64.StdEncoding.EncodeToString([]byte(attr.(string)))

		if len(encoded_userdata) > userdata_size_limit {
			return fmt.Errorf(
				"The supplied user_data contains %d bytes after encoding, "+
					"this exeeds the limit of %d bytes", len(encoded_userdata), userdata_size_limit)
		}
		opts.UserData = &encoded_userdata
	}

	if attr, ok := d.GetOk("compatibility"); ok {
		temp_compat := attr.(bool)
		opts.CompatibilityMode = &temp_compat
	}

	return nil

}

func setServerAttributes(
	d *schema.ResourceData,
	server *brightbox.Server,
) {
	d.Set("image", server.Image.Id)
	d.Set("name", server.Name)
	d.Set("type", server.ServerType.Handle)
	d.Set("zone", server.Zone.Handle)
	d.Set("compatbility", server.CompatibilityMode)
	d.Set("status", server.Status)
	d.Set("locked", server.Locked)
	d.Set("hostname", server.Hostname)
	if server.Image.Username != "" {
		d.Set("username", server.Image.Username)
	}

	if len(server.Interfaces) > 0 {
		server_interface := server.Interfaces[0]
		d.Set("ipv4_address_private", server_interface.IPv4Address)
		d.Set("fqdn", server.Fqdn)
		d.Set("ipv6_address", server_interface.IPv6Address)
		d.Set("ipv6_hostname", "ipv6."+server.Fqdn)
	}

	if len(server.CloudIPs) > 0 {
		cloud_ip := server.CloudIPs[0]
		d.Set("ipv4_address", cloud_ip.PublicIP)
		d.Set("public_hostname", cloud_ip.Fqdn)
	}

	srvGrpIds := []string{}
	for _, sg := range server.ServerGroups {
		srvGrpIds = append(srvGrpIds, sg.Id)
	}
	d.Set("server_groups", srvGrpIds)

	SetConnectionDetails(d)

}

func SetConnectionDetails(d *schema.ResourceData) {
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
