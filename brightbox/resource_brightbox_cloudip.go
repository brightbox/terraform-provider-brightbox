package brightbox

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	brightbox "github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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
				Description: "Name assigned to the Cloud IP",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"target": {
				Description: "The object this Cloud IP maps to",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"status": {
				Description: "Current state of the Cloud IP",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"locked": {
				Description: "No lock on Cloud IPs",
				Type:        schema.TypeString,
				Computed:    true,
				Deprecated:  "No lock on Cloud IPs",
			},

			"public_ip": {
				Description: "Old alias of the IPv4 address",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"fqdn": {
				Description: "Full Domain name entry for the Cloud IP",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"reverse_dns": {
				Description: "Reverse DNS entry for the Cloud IP",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"port_translator": {
				Description: "Array of Port Translators",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"incoming": {
							Description:  "Incoming Port",
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(minPort, maxPort),
						},

						"outgoing": {
							Description:  "Outgoing Port",
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(minPort, maxPort),
						},
						"protocol": {
							Description:  "Transport protocol to port translate (tcp/udp)",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"tcp", "udp"}, false),
						},
					},
				},
				Set: resourceBrightboxPortTranslationHash,
			},
		},
	}
}

func resourceBrightboxCloudipCreate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[INFO] Creating CloudIP")
	cloudipOpts := &brightbox.CloudIPOptions{}
	err := addUpdateableCloudipOptions(d, cloudipOpts)
	if err != nil {
		return err
	}

	cloudip, err := client.CreateCloudIP(cloudipOpts)
	if err != nil {
		return fmt.Errorf("Error creating Cloud IP: %s", err)
	}

	d.SetId(cloudip.Id)

	if targetID, ok := d.GetOk("target"); ok {
		cloudip, err = assignCloudIP(client, cloudip.Id, targetID.(string), d.Timeout(schema.TimeoutCreate))
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
	client := meta.(*CompositeClient).APIClient

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
	client := meta.(*CompositeClient).APIClient
	return removeCloudIP(client, d.Id(), d.Timeout(schema.TimeoutDelete))
}

func resourceBrightboxCloudipUpdate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	if d.HasChange("target") {
		err := unmapCloudIP(client, d.Id(), d.Timeout(schema.TimeoutDelete))
		if err != nil {
			return err
		}
		if targetID, ok := d.GetOk("target"); ok {
			_, err := assignCloudIP(client, d.Id(), targetID.(string), d.Timeout(schema.TimeoutCreate))
			if err != nil {
				return err
			}
		}
	}

	cloudipOpts := &brightbox.CloudIPOptions{
		Id: d.Id(),
	}
	err := addUpdateableCloudipOptions(d, cloudipOpts)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Cloud IP update configuration: %#v", cloudipOpts)

	cloudip, err := client.UpdateCloudIP(cloudipOpts)
	if err != nil {
		return fmt.Errorf("Error updating Cloud IP (%s): %s", cloudipOpts.Id, err)
	}

	return setCloudipAttributes(d, cloudip)
}

func assignCloudIP(
	client *brightbox.Client,
	cloudipID string,
	targetID string,
	timeout time.Duration,
) (*brightbox.CloudIP, error) {
	log.Printf("[INFO] Assigning Cloud IP %s to target %s", cloudipID, targetID)
	err := client.MapCloudIP(cloudipID, targetID)
	if err != nil {
		return nil, fmt.Errorf("Error assigning Cloud IP %s to target %s: %s", cloudipID, targetID, err)
	}
	cloudip, err := waitForMappedCloudIP(client, cloudipID, timeout)
	if err != nil {
		return nil, err
	}
	return cloudip, err
}

func unmapCloudIP(
	client *brightbox.Client,
	cloudipID string,
	timeout time.Duration,
) error {
	log.Printf("[INFO] Checking mapping of Cloud IP %s", cloudipID)
	cloudip, err := client.CloudIP(cloudipID)
	if err != nil {
		return fmt.Errorf("Error retrieving details of Cloud IP %s: %s", cloudipID, err)
	}
	if cloudip.Status == mapped {
		log.Printf("[INFO] Unmapping Cloud IP %s", cloudipID)
		err := client.UnMapCloudIP(cloudipID)
		if err != nil {
			return fmt.Errorf("Error unmapping Cloud IP %s: %s", cloudipID, err)
		}
		_, err = waitForUnmappedCloudIP(client, cloudipID, timeout)
		if err != nil {
			return err
		}
	} else {
		log.Printf("[DEBUG] Cloud IP %s is already unmapped", cloudipID)
	}
	return nil
}

func waitForCloudip(
	client *brightbox.Client,
	cloudipID string,
	timeout time.Duration,
	pending string,
	target string,
) (*brightbox.CloudIP, error) {
	stateConf := resource.StateChangeConf{
		Pending:    []string{pending},
		Target:     []string{target},
		Refresh:    cloudipStateRefresh(client, cloudipID),
		Timeout:    timeout,
		MinTimeout: minimumRefreshWait,
	}

	activeCloudIP, err := stateConf.WaitForState()
	if err != nil {
		return nil, err
	}

	return activeCloudIP.(*brightbox.CloudIP), err
}

func waitForMappedCloudIP(
	client *brightbox.Client,
	cloudipID string,
	timeout time.Duration,
) (*brightbox.CloudIP, error) {
	return waitForCloudip(client, cloudipID, timeout, unmapped, mapped)
}

func waitForUnmappedCloudIP(
	client *brightbox.Client,
	cloudipID string,
	timeout time.Duration,
) (*brightbox.CloudIP, error) {
	return waitForCloudip(client, cloudipID, timeout, mapped, unmapped)
}

func cloudipStateRefresh(client *brightbox.Client, cloudipID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cloudip, err := client.CloudIP(cloudipID)
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
	log.Printf("[DEBUG] PortTranslator details are %#v", cloudip.PortTranslators)
	portTranslators := make([]map[string]interface{}, len(cloudip.PortTranslators))
	for i, portTranslator := range cloudip.PortTranslators {
		portTranslators[i] = map[string]interface{}{
			"incoming": portTranslator.Incoming,
			"outgoing": portTranslator.Outgoing,
			"protocol": portTranslator.Protocol,
		}
	}
	if err := d.Set("port_translator", portTranslators); err != nil {
		return fmt.Errorf("error setting port_translator: %s", err)
	}

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
	assignPortTranslators(d, &opts.PortTranslators)
	return nil
}

func resourceBrightboxPortTranslationHash(
	v interface{},
) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%d-", m["incoming"].(int)))
	buf.WriteString(fmt.Sprintf("%s-",
		strings.ToLower(m["protocol"].(string))))
	buf.WriteString(fmt.Sprintf("%d-", m["outgoing"].(int)))

	return hashcode.String(buf.String())
}

func assignPortTranslators(d *schema.ResourceData, target *[]brightbox.PortTranslator) {
	if d.HasChange("port_translator") {
		*target = expandPortTranslators(d.Get("port_translator").(*schema.Set).List())
	}
}

func expandPortTranslators(configured []interface{}) []brightbox.PortTranslator {
	portTranslators := make([]brightbox.PortTranslator, len(configured))

	for i, portTranslationSource := range configured {
		data := portTranslationSource.(map[string]interface{})
		portTranslators[i].Protocol = data["protocol"].(string)
		portTranslators[i].Incoming = data["incoming"].(int)
		portTranslators[i].Outgoing = data["outgoing"].(int)
	}
	return portTranslators
}
