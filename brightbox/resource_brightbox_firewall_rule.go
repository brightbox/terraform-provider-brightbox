package brightbox

import (
	"fmt"
	"log"
	"strings"

	"github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceBrightboxFirewallRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceBrightboxFirewallRuleCreate,
		Read:   resourceBrightboxFirewallRuleRead,
		Update: resourceBrightboxFirewallRuleUpdate,
		Delete: resourceBrightboxFirewallRuleDelete,

		Schema: map[string]*schema.Schema{
			"firewall_policy": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"protocol": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"source": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"source_port": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"destination": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"destination_port": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"icmp_type_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceBrightboxFirewallRuleCreate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	log.Printf("[INFO] Creating Firewall Rule")
	firewall_rule_opts := &brightbox.FirewallRuleOptions{
		FirewallPolicy: d.Get("firewall_policy").(string),
	}
	err := addUpdateableFirewallRuleOptions(d, firewall_rule_opts)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Firewall Rule create configuration: %#v", firewall_rule_opts)

	firewall_rule, err := client.CreateFirewallRule(firewall_rule_opts)
	if err != nil {
		return fmt.Errorf("Error creating Firewall Rule: %s", err)
	}

	d.SetId(firewall_rule.Id)

	return setFirewallRuleAttributes(d, firewall_rule)
}

func resourceBrightboxFirewallRuleRead(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	firewall_rule, err := client.FirewallRule(d.Id())
	if err != nil {
		if strings.HasPrefix(err.Error(), "missing_resource:") {
			log.Printf("[WARN] Firewall Rule not found, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving Firewall Rule details: %s", err)
	}

	return setFirewallRuleAttributes(d, firewall_rule)
}

func resourceBrightboxFirewallRuleDelete(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	log.Printf("[INFO] Deleting Firewall Rule %s", d.Id())
	err := client.DestroyFirewallRule(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting Firewall Rule (%s): %s", d.Id(), err)
	}
	return nil
}

func resourceBrightboxFirewallRuleUpdate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	firewall_rule_opts := &brightbox.FirewallRuleOptions{
		Id: d.Id(),
	}
	err := addUpdateableFirewallRuleOptions(d, firewall_rule_opts)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Firewall Rule update configuration: %#v", firewall_rule_opts)

	firewall_rule, err := client.UpdateFirewallRule(firewall_rule_opts)
	if err != nil {
		return fmt.Errorf("Error updating Firewall Rule (%s): %s", firewall_rule_opts.Id, err)
	}

	return setFirewallRuleAttributes(d, firewall_rule)
}

func addUpdateableFirewallRuleOptions(
	d *schema.ResourceData,
	opts *brightbox.FirewallRuleOptions,
) error {
	assign_string(d, &opts.Protocol, "protocol")
	assign_string(d, &opts.Source, "source")
	assign_string(d, &opts.SourcePort, "source_port")
	assign_string(d, &opts.Destination, "destination")
	assign_string(d, &opts.DestinationPort, "destination_port")
	assign_string(d, &opts.IcmpTypeName, "icmp_type_name")
	assign_string(d, &opts.Description, "description")
	return nil
}

func setFirewallRuleAttributes(
	d *schema.ResourceData,
	firewall_rule *brightbox.FirewallRule,
) error {
	d.Set("firewall_policy", firewall_rule.FirewallPolicy)
	d.Set("protocol", firewall_rule.Protocol)
	d.Set("source", firewall_rule.Source)
	d.Set("source_port", firewall_rule.SourcePort)
	d.Set("destination", firewall_rule.Destination)
	d.Set("destination_port", firewall_rule.DestinationPort)
	d.Set("icmp_type_name", firewall_rule.IcmpTypeName)
	d.Set("description", firewall_rule.Description)
	return nil
}
