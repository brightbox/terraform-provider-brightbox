package brightbox

import (
	"fmt"
	"log"
	"strings"
	"time"

	brightbox "github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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

		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(defaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"server_group": {
				Description: "The server group using this policy",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description: "Optional name for this policy",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"description": {
				Description: "Optional description of the policy",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func resourceBrightboxFirewallPolicyCreate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[INFO] Creating Firewall Policy")
	firewallPolicyOpts := &brightbox.FirewallPolicyOptions{}
	err := addUpdateableFirewallPolicyOptions(d, firewallPolicyOpts)
	if err != nil {
		return err
	}
	assign_string(d, &firewallPolicyOpts.ServerGroup, "server_group")

	log.Printf("[INFO] Firewall Policy create configuration: %#v", firewallPolicyOpts)

	firewallPolicy, err := client.CreateFirewallPolicy(firewallPolicyOpts)
	if err != nil {
		return fmt.Errorf("Error creating Firewall Policy: %s", err)
	}

	d.SetId(firewallPolicy.Id)

	return setFirewallPolicyAttributes(d, firewallPolicy)
}

func resourceBrightboxFirewallPolicyRead(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	firewallPolicy, err := client.FirewallPolicy(d.Id())
	if err != nil {
		if strings.HasPrefix(err.Error(), "missing_resource:") {
			log.Printf("[WARN] Firewall Policy not found, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving Firewall Policy details: %s", err)
	}

	return setFirewallPolicyAttributes(d, firewallPolicy)
}

func resourceBrightboxFirewallPolicyDelete(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

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
	client := meta.(*CompositeClient).APIClient

	firewallPolicyOpts := &brightbox.FirewallPolicyOptions{
		Id: d.Id(),
	}
	err := addUpdateableFirewallPolicyOptions(d, firewallPolicyOpts)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Firewall Policy update configuration: %#v", firewallPolicyOpts)
	firewallPolicy, err := client.UpdateFirewallPolicy(firewallPolicyOpts)
	if err != nil {
		return fmt.Errorf("Error updating Firewall Policy (%s): %s", firewallPolicyOpts.Id, err)
	}

	if d.HasChange("server_group") {
		var err error
		log.Printf("[INFO] Server Group changed, updating...")
		o, n := d.GetChange("server_group")
		newServerGroupID := n.(string)
		oldServerGroupID := o.(string)
		if firewallPolicy.ServerGroup != nil {
			log.Printf("[INFO] Detaching %s from Server Group %s", firewallPolicyOpts.Id, oldServerGroupID)
			err = retryServerGroupChange(
				func() error {
					_, err := client.RemoveFirewallPolicy(firewallPolicyOpts.Id, oldServerGroupID)
					return err
				},
				d.Timeout(schema.TimeoutUpdate),
			)
			if err != nil {
				return fmt.Errorf("Error removing group from Firewall Policy (%s): %s", firewallPolicyOpts.Id, err)
			}
		}
		if newServerGroupID != "" {
			log.Printf("[INFO] Attaching %s to Server Group %s", firewallPolicyOpts.Id, newServerGroupID)
			err = retryServerGroupChange(
				func() error {
					_, err := client.ApplyFirewallPolicy(firewallPolicyOpts.Id, newServerGroupID)
					return err
				},
				d.Timeout(schema.TimeoutUpdate),
			)
			if err != nil {
				return fmt.Errorf("Error adding group to Firewall Policy (%s): %s", firewallPolicyOpts.Id, err)
			}
		}

	}

	return resourceBrightboxFirewallPolicyRead(d, meta)
}

func addUpdateableFirewallPolicyOptions(
	d *schema.ResourceData,
	opts *brightbox.FirewallPolicyOptions,
) error {
	assign_string(d, &opts.Name, "name")
	assign_string(d, &opts.Description, "description")
	return nil
}

func setFirewallPolicyAttributes(
	d *schema.ResourceData,
	firewallPolicy *brightbox.FirewallPolicy,
) error {
	d.Set("name", firewallPolicy.Name)
	d.Set("description", firewallPolicy.Description)
	if firewallPolicy.ServerGroup == nil {
		d.Set("server_group", "")
	} else {
		d.Set("server_group", firewallPolicy.ServerGroup.Id)
	}
	return nil
}

func retryServerGroupChange(changeFunc func() error, timeout time.Duration) error {
	// Wait for group to change
	return resource.Retry(
		timeout,
		func() *resource.RetryError {
			if err := changeFunc(); err != nil {
				apierror := err.(brightbox.ApiError)
				if apierror.StatusCode == 409 {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		},
	)
}
