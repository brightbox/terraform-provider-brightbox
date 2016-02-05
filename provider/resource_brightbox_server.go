package brightbox

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform/helper/schema"
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

	addUpdateableOptions(d, server_opts)

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

	log.Printf("[INFO] Server ID: %s", d.Id())

	setServerAttributes(d, server)

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

	addUpdateableOptions(d, server_opts)

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
) {

	if attr, ok := d.GetOk("name"); ok {
		temp_name := attr.(string)
		opts.Name = &temp_name
	}

	if attr, ok := d.GetOk("userdata"); ok {
		encoded_userdata := base64.StdEncoding.EncodeToString([]byte(attr.(string)))
		opts.UserData = &encoded_userdata
	}

	if attr, ok := d.GetOk("compatibility"); ok {
		temp_compat := attr.(bool)
		opts.CompatibilityMode = &temp_compat
	}

}

func setServerAttributes(
	d *schema.ResourceData,
	server *brightbox.Server,
) {
	d.Set("image", server.Image.Id)
	d.Set("name", server.Name)
	d.Set("type", server.ServerType.Id)
	d.Set("zone", server.Zone.Id)
	d.Set("compatbility", server.CompatibilityMode)
	d.Set("status", server.Status)
	d.Set("locked", server.Locked)
	d.Set("hostname", server.Hostname)
	d.Set("fqdn", server.Fqdn)
	d.Set("ipv6_hostname", "ipv6."+server.Fqdn)

	server_interface := server.Interfaces[0]
	d.Set("ipv6_address", server_interface.IPv6Address)
	d.Set("ipv4_address_private", server_interface.IPv4Address)

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
}
