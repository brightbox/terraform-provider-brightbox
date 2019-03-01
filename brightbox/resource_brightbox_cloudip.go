package brightbox

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	mapped             = "mapped"
	unmapped           = "unmapped"
	defaultTimeout     = 5 * time.Minute
	minimumRefreshWait = 3 * time.Second
	checkDelay         = 10 * time.Second
)

func resourceBrightboxCloudip() *schema.Resource {
	return &schema.Resource{
		Create: resourceBrightboxCloudipCreate,
		Read:   resourceBrightboxCloudipRead,
		Update: resourceBrightboxCloudipUpdate,
		Delete: resourceBrightboxCloudipDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultTimeout),
			Delete: schema.DefaultTimeout(defaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"target": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"locked": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"public_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"fqdn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"reverse_dns": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceBrightboxCloudipCreate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	log.Printf("[INFO] Creating CloudIP")
	cloudip_opts := &brightbox.CloudIPOptions{}
	err := addUpdateableCloudipOptions(d, cloudip_opts)
	if err != nil {
		return err
	}

	cloudip, err := client.CreateCloudIP(cloudip_opts)
	if err != nil {
		return fmt.Errorf("Error creating Cloud IP: %s", err)
	}

	d.SetId(cloudip.Id)

	if target_id, ok := d.GetOk("target"); ok {
		cloudip, err = assignCloudIP(client, cloudip.Id, target_id.(string), d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return err
		}
	}

	return setCloudipAttributes(d, cloudip)
}

func resourceBrightboxCloudipRead(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	cloudip, err := client.CloudIP(d.Id())
	if err != nil {
		if strings.HasPrefix(err.Error(), "missing_resource:") {
			log.Printf("[WARN] CloudIP not found, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving Cloud IP details: %s", err)
	}

	return setCloudipAttributes(d, cloudip)
}

func resourceBrightboxCloudipDelete(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient
	return removeCloudIP(client, d.Id(), d.Timeout(schema.TimeoutDelete))
}

func resourceBrightboxCloudipUpdate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	d.Partial(true)

	if d.HasChange("target") {
		err := unmapCloudIP(client, d.Id(), d.Timeout(schema.TimeoutDelete))
		if err != nil {
			return err
		}
		if target_id, ok := d.GetOk("target"); ok {
			_, err := assignCloudIP(client, d.Id(), target_id.(string), d.Timeout(schema.TimeoutCreate))
			if err != nil {
				return err
			}
		}
		d.SetPartial("target")
	}

	cloudip_opts := &brightbox.CloudIPOptions{
		Id: d.Id(),
	}
	err := addUpdateableCloudipOptions(d, cloudip_opts)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Cloud IP update configuration: %#v", cloudip_opts)

	cloudip, err := client.UpdateCloudIP(cloudip_opts)
	if err != nil {
		return fmt.Errorf("Error updating Cloud IP (%s): %s", cloudip_opts.Id, err)
	}

	return setCloudipAttributes(d, cloudip)
}

func assignCloudIP(
	client *brightbox.Client,
	cloudip_id string,
	target_id string,
	timeout time.Duration,
) (*brightbox.CloudIP, error) {
	log.Printf("[INFO] Assigning Cloud IP %s to target %s", cloudip_id, target_id)
	err := client.MapCloudIP(cloudip_id, target_id)
	if err != nil {
		return nil, fmt.Errorf("Error assigning Cloud IP %s to target %s: %s", cloudip_id, target_id, err)
	}
	cloudip, err := waitForMappedCloudIp(client, cloudip_id, timeout)
	if err != nil {
		return nil, err
	}
	return cloudip, err
}

func unmapCloudIP(
	client *brightbox.Client,
	cloudip_id string,
	timeout time.Duration,
) error {
	log.Printf("[INFO] Checking mapping of Cloud IP %s", cloudip_id)
	cloudip, err := client.CloudIP(cloudip_id)
	if err != nil {
		return fmt.Errorf("Error retrieving details of Cloud IP %s: %s", cloudip_id, err)
	}
	if cloudip.Status == mapped {
		log.Printf("[INFO] Unmapping Cloud IP %s", cloudip_id)
		err := client.UnMapCloudIP(cloudip_id)
		if err != nil {
			return fmt.Errorf("Error unmapping Cloud IP %s: %s", cloudip_id, err)
		}
		_, err = waitForUnmappedCloudIp(client, cloudip_id, timeout)
		if err != nil {
			return err
		}
	} else {
		log.Printf("[DEBUG] Cloud IP %s is already unmapped", cloudip_id)
	}
	return nil
}

func waitForCloudip(
	client *brightbox.Client,
	cloudip_id string,
	timeout time.Duration,
	pending string,
	target string,
) (*brightbox.CloudIP, error) {
	stateConf := resource.StateChangeConf{
		Pending:    []string{pending},
		Target:     []string{target},
		Refresh:    cloudipStateRefresh(client, cloudip_id),
		Timeout:    timeout,
		MinTimeout: minimumRefreshWait,
	}

	active_cloudip, err := stateConf.WaitForState()
	if err != nil {
		return nil, err
	}

	return active_cloudip.(*brightbox.CloudIP), err
}

func waitForMappedCloudIp(
	client *brightbox.Client,
	cloudip_id string,
	timeout time.Duration,
) (*brightbox.CloudIP, error) {
	return waitForCloudip(client, cloudip_id, timeout, unmapped, mapped)
}

func waitForUnmappedCloudIp(
	client *brightbox.Client,
	cloudip_id string,
	timeout time.Duration,
) (*brightbox.CloudIP, error) {
	return waitForCloudip(client, cloudip_id, timeout, mapped, unmapped)
}

func cloudipStateRefresh(client *brightbox.Client, cloudip_id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cloudip, err := client.CloudIP(cloudip_id)
		if err != nil {
			log.Printf("Error on Cloud IP State Refresh: %s", err)
			return nil, "", err
		}
		return cloudip, cloudip.Status, nil
	}
}

func setCloudipAttributes(
	d *schema.ResourceData,
	cloudip *brightbox.CloudIP,
) error {
	d.Set("name", cloudip.Name)
	d.Set("public_ip", cloudip.PublicIP)
	d.Set("status", cloudip.Status)
	d.Set("locked", cloudip.Locked)
	d.Set("reverse_dns", cloudip.ReverseDns)
	d.Set("fqdn", cloudip.Fqdn)
	// Set the server id first and let interface override it
	// Server and interface should appear together, but catch at least one
	if cloudip.Server != nil {
		d.Set("target", cloudip.Server.Id)
	}
	if cloudip.Interface != nil {
		d.Set("target", cloudip.Interface.Id)
	}
	if cloudip.LoadBalancer != nil {
		d.Set("target", cloudip.LoadBalancer.Id)
	}
	if cloudip.DatabaseServer != nil {
		d.Set("target", cloudip.DatabaseServer.Id)
	}
	if cloudip.ServerGroup != nil {
		d.Set("target", cloudip.ServerGroup.Id)
	}
	d.Partial(false)
	return nil
}

func removeCloudIP(client *brightbox.Client, id string, timeout time.Duration) error {
	log.Printf("[DEBUG] Unmapping Cloud IP %s", id)
	err := unmapCloudIP(client, id, timeout)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Deleting Cloud IP %s", id)
	err = client.DestroyCloudIP(id)
	if err != nil {
		return fmt.Errorf("Error deleting Cloud IP (%s): %s", id, err)
	}
	return nil
}

func addUpdateableCloudipOptions(
	d *schema.ResourceData,
	opts *brightbox.CloudIPOptions,
) error {
	assign_string(d, &opts.Name, "name")
	assign_string(d, &opts.ReverseDns, "reverse_dns")
	return nil
}
