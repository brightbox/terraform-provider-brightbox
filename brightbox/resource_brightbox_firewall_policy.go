package brightbox

import (
	"fmt"
	"log"
	"strings"

	"github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceBrightboxFirewallPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceBrightboxFirewallPolicyCreate,
		Read:   resourceBrightboxFirewallPolicyRead,
		Update: resourceBrightboxFirewallPolicyUpdate,
		Delete: resourceBrightboxFirewallPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"server_group": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceBrightboxFirewallPolicyCreate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	log.Printf("[INFO] Creating Firewall Policy")
	firewall_policy_opts := &brightbox.FirewallPolicyOptions{}
	err := addUpdateableFirewallPolicyOptions(d, firewall_policy_opts)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Firewall Policy create configuration: %#v", firewall_policy_opts)

	firewall_policy, err := client.CreateFirewallPolicy(firewall_policy_opts)
	if err != nil {
		return fmt.Errorf("Error creating Firewall Policy: %s", err)
	}

	d.SetId(firewall_policy.Id)

	return setFirewallPolicyAttributes(d, firewall_policy)
}

func resourceBrightboxFirewallPolicyRead(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	firewall_policy, err := client.FirewallPolicy(d.Id())
	if err != nil {
		if strings.HasPrefix(err.Error(), "missing_resource:") {
			log.Printf("[WARN] Firewall Policy not found, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving Firewall Policy details: %s", err)
	}

	return setFirewallPolicyAttributes(d, firewall_policy)
}

func resourceBrightboxFirewallPolicyDelete(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	log.Printf("[INFO] Deleting Firewall Policy %s", d.Id())
	err := client.DestroyFirewallPolicy(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting Firewall Policy (%s): %s", d.Id(), err)
	}
	return nil
}

func resourceBrightboxFirewallPolicyUpdate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	firewall_policy_opts := &brightbox.FirewallPolicyOptions{
		Id: d.Id(),
	}
	err := addUpdateableFirewallPolicyOptions(d, firewall_policy_opts)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Firewall Policy update configuration: %#v", firewall_policy_opts)

	firewall_policy, err := client.UpdateFirewallPolicy(firewall_policy_opts)
	if err != nil {
		return fmt.Errorf("Error updating Firewall Policy (%s): %s", firewall_policy_opts.Id, err)
	}

	return setFirewallPolicyAttributes(d, firewall_policy)
}

func addUpdateableFirewallPolicyOptions(
	d *schema.ResourceData,
	opts *brightbox.FirewallPolicyOptions,
) error {
	assign_string(d, &opts.Name, "name")
	assign_string(d, &opts.Description, "description")
	assign_string(d, &opts.ServerGroup, "server_group")
	return nil
}

func setFirewallPolicyAttributes(
	d *schema.ResourceData,
	firewall_policy *brightbox.FirewallPolicy,
) error {
	d.Set("name", firewall_policy.Name)
	d.Set("description", firewall_policy.Description)
	d.Set("server_group", firewall_policy.ServerGroup)
	return nil
}
